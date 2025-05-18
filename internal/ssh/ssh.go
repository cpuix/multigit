package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// CommandRunner is a function type that matches the signature of exec.Command
// This is used to make the package more testable by allowing tests to mock the exec.Command function.
type CommandRunner func(name string, arg ...string) *exec.Cmd

// ExecCommand is a variable that holds the function to execute commands.
// By default, it's set to exec.Command, but can be replaced in tests.
var ExecCommand CommandRunner = exec.Command

const (
	rsaKeyBitSize = 4096
	// ED25519 keys are always 256 bits (32 bytes) when using the standard implementation
)

// marshalED25519PrivateKey converts an ED25519 private key to OpenSSH format
func marshalED25519PrivateKey(key ed25519.PrivateKey, comment string) []byte {
	// The public key is the last 32 bytes of the private key
	publicKey := key.Public().(ed25519.PublicKey)

	// The private key is the seed (first 32 bytes of the private key)
	seed := key.Seed()

	// Create the key data in OpenSSH format
	keyData := struct {
		CipherName   string
		KdfName      string
		KdfOpts      string
		NumKeys      uint32
		PubKey       []byte
		PrivKeyBlock []byte
	}{
		CipherName: "none",
		KdfName:    "none",
		KdfOpts:    "",
		NumKeys:    1,
	}

	// Public key
	pubKeyData := struct {
		KeyType string
		Key     []byte
	}{
		KeyType: "ssh-ed25519",
		Key:     publicKey,
	}
	keyData.PubKey = ssh.Marshal(pubKeyData)

	// Private key block
	privKeyData := struct {
		Check1  uint64
		Check2  uint64
		KeyType string
		PubKey  []byte
		PrivKey []byte
		Comment string
		Pad     []byte `ssh:"rest"`
	}{
		Check1:  0, // checkint1
		Check2:  0, // checkint2 (same as checkint1)
		KeyType: "ssh-ed25519",
		PubKey:  publicKey,
		PrivKey: append(seed, publicKey...), // seed + public key
		Comment: comment,
	}

	// Marshal the private key data
	keyData.PrivKeyBlock = ssh.Marshal(privKeyData)

	// Add padding to make the private key length a multiple of the cipher block size (8 bytes)
	if pad := len(keyData.PrivKeyBlock) % 8; pad != 0 {
		padding := make([]byte, 8-pad)
		for i := range padding {
			padding[i] = byte(i + 1)
		}
		keyData.PrivKeyBlock = append(keyData.PrivKeyBlock, padding...)
	}

	// Create the OpenSSH private key format
	magic := append([]byte("openssh-key-v1\x00"), 0)
	data := ssh.Marshal(keyData)

	// Create the PEM block
	pemBlock := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: append(magic, data...),
	}

	return pem.EncodeToMemory(pemBlock)
}

// validatePrivateKey attempts to parse the private key to ensure it's in a valid format
func validatePrivateKey(keyData []byte) error {
	// Try to parse the private key
	_, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("invalid private key format: %w", err)
	}
	return nil
}

// CreateSSHKey creates a new SSH key pair for the given account
func CreateSSHKey(accountName, accountEmail, keyFile string, keyType KeyType) error {
	if keyType == "" {
		keyType = KeyTypeED25519 // Default to ED25519
	}

	// If keyFile is not provided, use default location in .ssh directory
	if keyFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Create .ssh directory if it doesn't exist
		sshDir := filepath.Join(homeDir, ".ssh")
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return fmt.Errorf("failed to create .ssh directory: %w", err)
		}

		// Generate default key file path based on key type
		switch keyType {
		case KeyTypeRSA:
			keyFile = filepath.Join(sshDir, fmt.Sprintf("id_rsa_%s", accountName))
		case KeyTypeED25519:
			keyFile = filepath.Join(sshDir, fmt.Sprintf("id_ed25519_%s", accountName))
		}
	} else {
		// Ensure parent directory exists
		parentDir := filepath.Dir(keyFile)
		if err := os.MkdirAll(parentDir, 0700); err != nil {
			return fmt.Errorf("failed to create directory for key file: %w", err)
		}
	}

	// Format the comment as "accountName <email>"
	comment := fmt.Sprintf("%s <%s>", accountName, accountEmail)

	switch keyType {
	case KeyTypeRSA:
		// Generate RSA key pair
		privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBitSize)
		if err != nil {
			return fmt.Errorf("failed to generate RSA private key: %w", err)
		}

		// Create private key in PKCS#1 format
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		// Write private key to file
		if err := os.WriteFile(keyFile, pem.EncodeToMemory(privateKeyPEM), 0600); err != nil {
			return fmt.Errorf("failed to write private key: %w", err)
		}

		// Generate public key
		sshPubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to generate SSH public key: %w", err)
		}

		// Write public key to file
		publicKeyPath := keyFile + ".pub"
		publicKey := fmt.Sprintf("%s %s %s", ssh.KeyAlgoRSA, base64.StdEncoding.EncodeToString(sshPubKey.Marshal()), comment)
		if err := os.WriteFile(publicKeyPath, []byte(publicKey+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write public key: %w", err)
		}

	case KeyTypeED25519:
		// Generate ED25519 key pair
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate ED25519 key pair: %w", err)
		}

		// Create a signer from the private key
		signer, err := ssh.NewSignerFromSigner(privKey)
		if err != nil {
			return fmt.Errorf("failed to create signer from private key: %w", err)
		}

		// Marshal private key in OpenSSH format
		privateKeyPEM := &pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: ssh.MarshalAuthorizedKey(signer.PublicKey()),
		}

		// Write private key to file
		if err := os.WriteFile(keyFile, pem.EncodeToMemory(privateKeyPEM), 0600); err != nil {
			return fmt.Errorf("failed to write private key: %w", err)
		}

		// Generate public key
		sshPubKey, err := ssh.NewPublicKey(pubKey)
		if err != nil {
			return fmt.Errorf("failed to generate SSH public key: %w", err)
		}

		// Write public key to file
		publicKeyPath := keyFile + ".pub"
		publicKey := fmt.Sprintf("%s %s %s", ssh.KeyAlgoED25519, base64.StdEncoding.EncodeToString(sshPubKey.Marshal()), comment)
		if err := os.WriteFile(publicKeyPath, []byte(publicKey+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write public key: %w", err)
		}

	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}

	// Print success message with public key
	fmt.Printf("✅ SSH key pair created successfully for %s\n", accountName)
	fmt.Printf("Private key: %s\n", keyFile)
	fmt.Printf("Public key: %s\n", keyFile+".pub")

	return nil
}

// sshPublicKeyRSA generates the authorized_keys format for an RSA public key
func sshPublicKeyRSA(pubKey *rsa.PublicKey, comment string) ([]byte, error) {
	// Generate the public key in SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Format as authorized_keys entry
	keyType := ssh.KeyAlgoRSA
	keyBytes := base64.StdEncoding.EncodeToString(sshPubKey.Marshal())

	return []byte(fmt.Sprintf("%s %s %s", keyType, keyBytes, comment)), nil
}

// sshPublicKeyED25519 generates the authorized_keys format for an ED25519 public key
func sshPublicKeyED25519(pubKey ed25519.PublicKey, comment string) ([]byte, error) {
	// Generate the public key in SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Format as authorized_keys entry
	keyType := ssh.KeyAlgoED25519
	keyBytes := base64.StdEncoding.EncodeToString(sshPubKey.Marshal())

	return []byte(fmt.Sprintf("%s %s %s", keyType, keyBytes, comment)), nil
}

// AddSSHKeyToAgent adds the SSH key to the SSH agent
// accountOrKeyFile can be either an account name (e.g., "github") or a direct path to the private key file
func AddSSHKeyToAgent(accountOrKeyFile string) error {
	var privateKeyFile string

	// If the input is a direct file path, use it directly
	if filepath.IsAbs(accountOrKeyFile) || strings.HasPrefix(accountOrKeyFile, ".") || strings.Contains(accountOrKeyFile, "/") {
		privateKeyFile = accountOrKeyFile
	} else {
		// Otherwise, treat it as an account name and look for the key in the default location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		privateKeyFile = filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s", accountOrKeyFile))
	}

	// Convert to absolute path for consistent error messages
	absKeyPath, err := filepath.Abs(privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for key file: %w", err)
	}
	privateKeyFile = absKeyPath

	// Check if key exists
	if _, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("private key file %s does not exist", privateKeyFile)
	}

	// Check if SSH_AUTH_SOCK is set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return fmt.Errorf("SSH agent is not running. Please start the SSH agent and try again")
	}

	// Use ssh-add to add the key to the agent
	cmd := ExecCommand("ssh-add", privateKeyFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add key to SSH agent: %s - %v", string(output), err)
	}

	fmt.Printf("✅ SSH key %s added to SSH agent\n", privateKeyFile)
	return nil
}

// AddSSHConfigEntry adds an entry to the SSH config file
func AddSSHConfigEntry(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshConfigFile := filepath.Join(homeDir, ".ssh", "config")

	// Check for both RSA and ED25519 key files
	rsaKeyFile := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName))
	ed25519KeyFile := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s", accountName))

	var privateKeyFile string
	if _, err := os.Stat(rsaKeyFile); err == nil {
		privateKeyFile = rsaKeyFile
	} else if _, err := os.Stat(ed25519KeyFile); err == nil {
		privateKeyFile = ed25519KeyFile
	} else {
		return fmt.Errorf("no SSH key found for account %s", accountName)
	}

	// Read existing config with atomic operation
	var configData []byte
	if _, err := os.Stat(sshConfigFile); err == nil {
		configData, err = os.ReadFile(sshConfigFile)
		if err != nil {
			return fmt.Errorf("failed to read SSH config file: %w", err)
		}
	}

	hostPattern := fmt.Sprintf("github.com-%s", accountName)

	// Check if host already exists in config
	if containsHost(string(configData), hostPattern) {
		return fmt.Errorf("SSH config entry for %s already exists", accountName)
	}

	hostEntry := fmt.Sprintf(
		"\n# Multigit managed config for %s\n"+
			"Host %s\n"+
			"\tHostName github.com\n"+
			"\tUser git\n"+
			"\tIdentityFile %s\n"+
			"\tIdentitiesOnly yes\n"+
			"# End of Multigit config for %s\n",
		accountName, hostPattern, privateKeyFile, accountName,
	)

	// Create a backup of the original config
	backupFile := sshConfigFile + ".multigit.bak"
	if err := os.WriteFile(backupFile, configData, 0600); err != nil {
		return fmt.Errorf("failed to create backup of SSH config: %w", err)
	}
	defer os.Remove(backupFile) // Remove backup if everything goes well

	// Write new config with atomic operation
	tempFile := sshConfigFile + ".tmp"
	if err := os.WriteFile(tempFile, append(configData, []byte(hostEntry)...), 0600); err != nil {
		return fmt.Errorf("failed to write temporary SSH config: %w", err)
	}

	// Atomic replace
	if err := os.Rename(tempFile, sshConfigFile); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	fmt.Printf("✅ Added SSH config entry for %s\n", accountName)
	return nil
}

// DeleteSSHKey deletes the SSH key pair for the given account name or key file path
func DeleteSSHKey(accountOrKeyFile string) error {
	// First, check if the input is a file path
	if _, err := os.Stat(accountOrKeyFile); err == nil {
		// It's a file path, delete it directly
		return DeleteSSHKeyFile(accountOrKeyFile)
	}

	// If not a file path, treat it as an account name
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Define key files to check
	var keyFiles []string

	// Get the base name of the account (in case it's a path)
	accountName := filepath.Base(accountOrKeyFile)

	// First, try to find all key files in the .ssh directory that match the account name
	sshDir := filepath.Join(homeDir, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err == nil {
		// Look for files that contain the account name
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				// Check if the filename contains the account name
				if strings.Contains(name, accountName) {
					keyFiles = append(keyFiles, filepath.Join(sshDir, name))
				}
			}
		}
	}

	// Also try standard naming patterns with both the full account name and just the base name
	patterns := []string{
		// Standard patterns with key type prefix and account name
		"id_rsa_%s",
		"id_ed25519_%s",
		// Just the account name
		"%s",
	}

	// For each pattern, generate the full path with both full and base account names
	for _, pattern := range patterns {
		// Try with full account name first
		keyFile := filepath.Join(sshDir, fmt.Sprintf(pattern, accountOrKeyFile))
		keyFiles = append(keyFiles, keyFile, keyFile+".pub")

		// Then try with just the base name
		if accountName != accountOrKeyFile {
			keyFile = filepath.Join(sshDir, fmt.Sprintf(pattern, accountName))
			keyFiles = append(keyFiles, keyFile, keyFile+".pub")
		}
	}

	// Remove all key files
	var lastErr error
	dedupedFiles := make(map[string]bool)

	for _, keyFile := range keyFiles {
		if keyFile == "" {
			continue
		}

		if _, exists := dedupedFiles[keyFile]; !exists {
			dedupedFiles[keyFile] = true

			// Check if file exists before trying to delete
			if _, err := os.Stat(keyFile); os.IsNotExist(err) {
				// File doesn't exist, skip to next file
				fmt.Printf("File %s does not exist, skipping\n", keyFile)
				continue
			}

			// File exists, try to delete it
			fmt.Printf("Attempting to delete key file: %s\n", keyFile)
			err := DeleteSSHKeyFile(keyFile)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("Key file %s already deleted\n", keyFile)
				} else {
					errMsg := fmt.Errorf("failed to delete key %s: %w", keyFile, err)
					fmt.Println(errMsg)
					lastErr = errMsg
				}
			} else {
				fmt.Printf("Successfully deleted key file: %s\n", keyFile)
			}
		}
	}

	if lastErr != nil {
		return lastErr
	}

	fmt.Printf("✅ Removed SSH key pair for %s\n", accountOrKeyFile)
	return nil
}

// DeleteSSHKeyFile deletes the SSH key pair files for the given private key file path
func DeleteSSHKeyFile(keyFile string) error {
	var lastErr error

	// Remove private key if it exists
	fmt.Printf("Checking if private key exists: %s\n", keyFile)
	if _, err := os.Stat(keyFile); err == nil {
		fmt.Printf("Attempting to remove private key: %s\n", keyFile)
		if err := os.Remove(keyFile); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Private key %s already deleted\n", keyFile)
			} else {
				errMsg := fmt.Errorf("failed to remove private key %s: %w", keyFile, err)
				fmt.Println(errMsg)
				lastErr = errMsg
			}
		} else {
			fmt.Printf("Successfully removed private key: %s\n", keyFile)
		}
	} else if os.IsNotExist(err) {
		fmt.Printf("Private key %s does not exist, skipping\n", keyFile)
	} else {
		errMsg := fmt.Errorf("error checking private key %s: %w", keyFile, err)
		fmt.Println(errMsg)
		lastErr = errMsg
	}

	// Remove public key if it exists
	publicKeyFile := fmt.Sprintf("%s.pub", keyFile)
	fmt.Printf("Checking if public key exists: %s\n", publicKeyFile)
	if _, err := os.Stat(publicKeyFile); err == nil {
		fmt.Printf("Attempting to remove public key: %s\n", publicKeyFile)
		if err := os.Remove(publicKeyFile); err != nil {
			errMsg := fmt.Errorf("failed to remove public key %s: %w", publicKeyFile, err)
			fmt.Println(errMsg)
			if lastErr != nil {
				lastErr = fmt.Errorf("%v, %w", lastErr, errMsg)
			} else {
				lastErr = errMsg
			}
		} else {
			fmt.Printf("Successfully removed public key: %s\n", publicKeyFile)
		}
	} else if os.IsNotExist(err) {
		fmt.Printf("Public key %s does not exist, skipping\n", publicKeyFile)
	} else {
		errMsg := fmt.Errorf("error checking public key %s: %w", publicKeyFile, err)
		fmt.Println(errMsg)
		if lastErr != nil {
			lastErr = fmt.Errorf("%v, %w", lastErr, errMsg)
		} else {
			lastErr = errMsg
		}
	}

	return lastErr
}

// RemoveSSHConfigEntry removes the SSH config entry for the given account
func RemoveSSHConfigEntry(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshConfigFile := filepath.Join(homeDir, ".ssh", "config")

	// Check if config file exists
	if _, err := os.Stat(sshConfigFile); os.IsNotExist(err) {
		return nil // Config file doesn't exist, nothing to remove
	}

	// Read the config file
	configData, err := os.ReadFile(sshConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read SSH config file: %w", err)
	}

	hostPattern := fmt.Sprintf("github.com-%s", accountName)
	configStr := string(configData)

	// Check if the entry exists
	if !containsHost(configStr, hostPattern) &&
		!strings.Contains(configStr, "# Multigit managed config for "+accountName) {
		return nil // Entry doesn't exist, nothing to do
	}

	// Create a backup of the original config
	backupFile := sshConfigFile + ".multigit.bak"
	if err := os.WriteFile(backupFile, configData, 0600); err != nil {
		return fmt.Errorf("failed to create backup of SSH config: %w", err)
	}
	defer os.Remove(backupFile) // Remove backup if everything goes well

	// First try to remove by host pattern (for backward compatibility)
	updatedConfig := removeHostEntry(configStr, hostPattern)

	// If the config still contains the account name, try to remove by account name
	if strings.Contains(updatedConfig, accountName) {
		updatedConfig = removeHostEntry(updatedConfig, accountName)
	}

	// Clean up any empty lines at the end
	lines := strings.Split(updatedConfig, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	updatedConfig = strings.Join(lines, "\n")

	// Ensure there's a single newline at the end
	if updatedConfig != "" {
		updatedConfig += "\n"
	}

	// Write the updated config with atomic operation
	tempFile := sshConfigFile + ".tmp"
	if err := os.WriteFile(tempFile, []byte(updatedConfig), 0600); err != nil {
		return fmt.Errorf("failed to write temporary SSH config: %w", err)
	}

	// Atomic replace
	if err := os.Rename(tempFile, sshConfigFile); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	fmt.Printf("✅ Removed SSH config entry for %s\n", accountName)
	return nil
}

// Helper function to check if a host entry already exists
func containsHost(configData, host string) bool {
	hostPattern := fmt.Sprintf("\nHost %s", host)
	return strings.Contains(configData, hostPattern) ||
		strings.HasPrefix(configData, fmt.Sprintf("Host %s", host))
}

// Helper function to remove a host entry from SSH config
func removeHostEntry(configData, hostPattern string) string {
	lines := strings.Split(configData, "\n")
	var result []string
	inHostBlock := false
	inMultigitBlock := false
	var multigitBlockStart int = -1

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for Multigit config block start
		if strings.HasPrefix(line, "# Multigit managed config for ") && (hostPattern == "" || strings.Contains(line, hostPattern)) {
			inMultigitBlock = true
			multigitBlockStart = len(result) // Store the start of the block
			continue
		}

		// Check for Multigit config block end
		if inMultigitBlock && strings.HasPrefix(line, "# End of Multigit config for ") && (hostPattern == "" || strings.Contains(line, hostPattern)) {
			inMultigitBlock = false
			continue // Skip the end marker
		}

		// Skip all lines within the Multigit block
		if inMultigitBlock {
			continue
		}

		// Handle regular SSH host blocks
		if strings.HasPrefix(line, "Host ") {
			// Check if this is the host we want to remove
			if hostPattern != "" && strings.Contains(line, hostPattern) {
				inHostBlock = true
				continue
			}
			inHostBlock = false
		}

		if !inHostBlock && !inMultigitBlock {
			// Only add the line if it's not part of any block we want to remove
			result = append(result, lines[i]) // Use original line to preserve formatting
		}
	}

	// If we were in a Multigit block but didn't find the end, remove the partial block
	if inMultigitBlock && multigitBlockStart >= 0 && multigitBlockStart < len(result) {
		result = result[:multigitBlockStart]
	}

	// Clean up any empty lines at the end
	for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
		result = result[:len(result)-1]
	}

	return strings.Join(result, "\n")
}
