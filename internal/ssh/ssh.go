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

const (
	rsaKeyBitSize = 4096
	// ED25519 keys are always 256 bits (32 bytes) when using the standard implementation
)

// CreateSSHKey creates a new SSH key pair for the given account
func CreateSSHKey(accountName, accountEmail, passphrase string, keyType KeyType) error {
	if keyType == "" {
		keyType = KeyTypeED25519 // Default to ED25519
	}
	// Create .ssh directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	var privateKeyPEM *pem.Block
	var privateKeyFile string
	var pubKey []byte

	switch keyType {
	case KeyTypeRSA:
		privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBitSize)
		if err != nil {
			return fmt.Errorf("failed to generate RSA private key: %w", err)
		}

		privateKeyFile = filepath.Join(sshDir, fmt.Sprintf("id_rsa_%s", accountName))
		privateKeyPEM = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		// Create public key
		pubKey, err = sshPublicKeyRSA(&privateKey.PublicKey, fmt.Sprintf("%s <%s>", accountName, accountEmail))
		if err != nil {
			return fmt.Errorf("failed to generate RSA public key: %w", err)
		}

	case KeyTypeED25519:
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate ED25519 private key: %w", err)
		}

		privateKeyFile = filepath.Join(sshDir, fmt.Sprintf("id_ed25519_%s", accountName))
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			return fmt.Errorf("failed to marshal ED25519 private key: %w", err)
		}

		privateKeyPEM = &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyBytes,
		}

		// Create public key
		pubKey, err = sshPublicKeyED25519(privKey.Public().(ed25519.PublicKey), fmt.Sprintf("%s <%s>", accountName, accountEmail))
		if err != nil {
			return fmt.Errorf("failed to generate ED25519 public key: %w", err)
		}

	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}

	// If passphrase is provided, encrypt the private key
	if passphrase != "" {
		encryptedPEM, err := x509.EncryptPEMBlock(
			rand.Reader,
			privateKeyPEM.Type,
			privateKeyPEM.Bytes,
			[]byte(passphrase),
			x509.PEMCipherAES256,
		)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
		privateKeyPEM = encryptedPEM
	}



	// Write private key file
	privateKeyData := pem.EncodeToMemory(privateKeyPEM)
	if err := os.WriteFile(privateKeyFile, privateKeyData, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key file
	publicKeyFile := fmt.Sprintf("%s.pub", privateKeyFile)
	if err := os.WriteFile(publicKeyFile, pubKey, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	// Set correct permissions
	if err := os.Chmod(privateKeyFile, 0600); err != nil {
		return fmt.Errorf("failed to set private key permissions: %w", err)
	}
	if err := os.Chmod(publicKeyFile, 0644); err != nil {
		return fmt.Errorf("failed to set public key permissions: %w", err)
	}

	// Print success message with public key
	fmt.Printf("✅ SSH key pair created successfully for %s\n", accountName)
	fmt.Printf("Private key: %s\n", privateKeyFile)
	fmt.Printf("Public key: %s\n", publicKeyFile)
	fmt.Println("\nPublic key (add this to your GitHub account):")
	fmt.Println(string(pubKey))

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
func AddSSHKeyToAgent(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	privateKeyFile := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName))
	
	// Check if key exists
	if _, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("private key file %s does not exist", privateKeyFile)
	}

	// Check if SSH_AUTH_SOCK is set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return fmt.Errorf("SSH agent is not running. Please start the SSH agent and try again")
	}

	// Use ssh-add to add the key to the agent
	cmd := exec.Command("ssh-add", privateKeyFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add key to SSH agent: %s - %v", string(output), err)
	}

	fmt.Printf("✅ SSH key for %s added to SSH agent\n", accountName)
	return nil
}

// AddSSHConfigEntry adds an entry to the SSH config file
func AddSSHConfigEntry(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshConfigFile := filepath.Join(homeDir, ".ssh", "config")
	privateKeyFile := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName))

	// Create config file if it doesn't exist
	f, err := os.OpenFile(sshConfigFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open SSH config file: %w", err)
	}
	defer f.Close()

	// Check if the entry already exists
	configData, err := os.ReadFile(sshConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read SSH config file: %w", err)
	}

	hostEntry := fmt.Sprintf(
		"\nHost github.com-%s\n"+
		"\tHostName github.com\n"+
		"\tUser git\n"+
		"\tIdentityFile %s\n"+
		"\tIdentitiesOnly yes\n",
		accountName, privateKeyFile,
	)

	// Check if host already exists in config
	if containsHost(string(configData), fmt.Sprintf("github.com-%s", accountName)) {
		return fmt.Errorf("SSH config entry for %s already exists", accountName)
	}

	// Append the new host entry
	if _, err := f.WriteString(hostEntry); err != nil {
		return fmt.Errorf("failed to write to SSH config: %w", err)
	}

	fmt.Printf("✅ Added SSH config entry for %s\n", accountName)
	return nil
}

// DeleteSSHKey deletes the SSH key pair for the given account
func DeleteSSHKey(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Define key types to check
	keyTypes := []KeyType{KeyTypeRSA, KeyTypeED25519}
	var keyFiles []string

	// Generate all possible key file paths
	for _, keyType := range keyTypes {
		var keyFile string
		switch keyType {
		case KeyTypeRSA:
			keyFile = filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName))
		case KeyTypeED25519:
			keyFile = filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s", accountName))
		}
		keyFiles = append(keyFiles, keyFile)
	}

	// Remove all key files
	for _, keyFile := range keyFiles {
		// Remove private key
		if err := os.Remove(keyFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove private key %s: %w", keyFile, err)
		}

		// Remove public key
		publicKeyFile := fmt.Sprintf("%s.pub", keyFile)
		if err := os.Remove(publicKeyFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove public key %s: %w", publicKeyFile, err)
		}
	}

	fmt.Printf("✅ Removed SSH key pair for %s\n", accountName)
	return nil
}

// RemoveSSHConfigEntry removes the SSH config entry for the given account
func RemoveSSHConfigEntry(accountName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshConfigFile := filepath.Join(homeDir, ".ssh", "config")
	
	// Check if config file exists
	configData, err := os.ReadFile(sshConfigFile)
	if os.IsNotExist(err) {
		return nil // Nothing to remove
	} else if err != nil {
		return fmt.Errorf("failed to read SSH config file: %w", err)
	}

	hostPattern := fmt.Sprintf("github.com-%s", accountName)
	updatedConfig := removeHostEntry(string(configData), hostPattern)

	// Only write if config was changed
	if updatedConfig != string(configData) {
		if err := os.WriteFile(sshConfigFile, []byte(updatedConfig), 0600); err != nil {
			return fmt.Errorf("failed to update SSH config: %w", err)
		}
		fmt.Printf("✅ Removed SSH config entry for %s\n", accountName)
	} else {
		fmt.Printf("ℹ️ No SSH config entry found for %s\n", accountName)
	}

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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "Host ") {
			// Check if this is the host we want to remove
			if strings.Contains(line, hostPattern) {
				inHostBlock = true
				continue
			}
			inHostBlock = false
		}

		if !inHostBlock {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
