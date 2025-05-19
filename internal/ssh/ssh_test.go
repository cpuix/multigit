package ssh_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test için gerekli wrapper fonksiyonlar
var (
	sshPublicKeyED25519 = func(pubKey ed25519.PublicKey, comment string) ([]byte, error) {
		return ssh.SSHPublicKeyED25519(pubKey, comment)
	}

	marshalED25519PrivateKey = func(key ed25519.PrivateKey, comment string) []byte {
		return ssh.MarshalED25519PrivateKey(key, comment)
	}

	validatePrivateKey = func(keyData []byte) error {
		return ssh.ValidatePrivateKey(keyData)
	}
)

// mockCommand is used to mock exec.Command
func mockCommand(command string, success bool) ssh.CommandRunner {
	return func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		if success {
			cs = append(cs, "success")
		} else {
			cs = append(cs, "fail")
		}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

// TestHelperProcess is used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Mock successful command
	if strings.Contains(strings.Join(os.Args[3:], " "), "success") {
		os.Exit(0)
	}
	// Mock failed command
	if strings.Contains(strings.Join(os.Args[3:], " "), "fail") {
		os.Stderr.WriteString("ssh-add: invalid key")
		os.Exit(1)
	}
	os.Exit(0)
}

func TestCreateAndDeleteSSHKey(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

	// Test account details
	accountName := "test-account"
	accountEmail := "test@example.com"
	keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)

	// Test key creation
	t.Run("CreateSSHKey", func(t *testing.T) {
		err := ssh.CreateSSHKey(accountName, accountEmail, keyFile, ssh.KeyTypeED25519)
		require.NoError(t, err, "Failed to create SSH key")

		// Check if key files were created (ED25519 keys)
		publicKeyPath := keyFile + ".pub"

		_, err = os.Stat(keyFile)
		assert.NoError(t, err, "Private key file should exist")

		_, err = os.Stat(publicKeyPath)
		assert.NoError(t, err, "Public key file should exist")
	})

	// Test key deletion
	t.Run("DeleteSSHKey", func(t *testing.T) {
		err := ssh.DeleteSSHKey(accountName)
		require.NoError(t, err, "Failed to delete SSH key")

		// Check if key files were deleted (ED25519 keys)
		publicKeyPath := keyFile + ".pub"

		_, err = os.Stat(keyFile)
		assert.True(t, os.IsNotExist(err), "Private key file should not exist")

		_, err = os.Stat(publicKeyPath)
		assert.True(t, os.IsNotExist(err), "Public key file should not exist")
	})
}

func setupTestSSHConfig(t *testing.T) (string, string) {
	tempDir := t.TempDir()

	// Override the home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

	accountName := "test-config-account"

	// Create a test key file first
	keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)
	require.NoError(t, os.WriteFile(keyFile, []byte("test private key"), 0600), "Failed to create test key file")

	return sshDir, accountName
}

func TestCreateSSHKey(t *testing.T) {
	tests := []struct {
		name        string
		keyType     string
		expectError bool
		setup       func(string) string
	}{
		{
			name:        "Create RSA Key with custom path",
			keyType:     "rsa",
			expectError: false,
			setup: func(baseDir string) string {
				return filepath.Join(baseDir, "custom_rsa_key")
			},
		},
		{
			name:        "Create ED25519 Key with custom path",
			keyType:     "ed25519",
			expectError: false,
			setup: func(baseDir string) string {
				return filepath.Join(baseDir, "custom_ed25519_key")
			},
		},
		{
			name:        "Create with Invalid Type",
			keyType:     "dsa",
			expectError: true,
			setup: func(baseDir string) string {
				return filepath.Join(baseDir, "invalid_key")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for this test case
			tempDir := t.TempDir()
			keyFile := filepath.Join(tempDir, "id_"+tt.keyType)

			// Create .ssh directory
			sshDir := filepath.Join(tempDir, ".ssh")
			require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", oldHome)

			err := ssh.CreateSSHKey("test-account", "test@example.com", keyFile, ssh.KeyType(tt.keyType))

			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Failed to create SSH key")

			// Verify key files exist at the specified location
			require.FileExists(t, keyFile, "Private key file should exist at the specified location")
			require.FileExists(t, keyFile+".pub", "Public key file should exist at the specified location")

			// Verify the key files contain valid content
			privateKey, err := os.ReadFile(keyFile)
			require.NoError(t, err, "Failed to read private key file")
			require.True(t, len(privateKey) > 0, "Private key file should not be empty")

			publicKey, err := os.ReadFile(keyFile + ".pub")
			require.NoError(t, err, "Failed to read public key file")
			require.True(t, len(publicKey) > 0, "Public key file should not be empty")

			// Verify the public key starts with the correct prefix
			if tt.keyType == "rsa" {
				require.True(t, strings.HasPrefix(string(publicKey), "ssh-rsa "),
					"RSA public key should start with 'ssh-rsa'")
			} else if tt.keyType == "ed25519" {
				require.True(t, strings.HasPrefix(string(publicKey), "ssh-ed25519 "),
					"ED25519 public key should start with 'ssh-ed25519'")
			}
		})
	}
}

func TestDeleteSSHKey(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, tempDir string) (string, []string) // Returns (accountName/keyFile, filesToCheck)
		cleanup     func(t *testing.T, tempDir string)
		expectError bool
		useFileFunc bool // Whether to use DeleteSSHKeyFile instead of DeleteSSHKey
		fileTypes   []ssh.KeyType
	}{
		{
			name: "Delete existing RSA key by account name",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_rsa_test-account")
				err := ssh.CreateSSHKey("test-account", "test@example.com", keyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create test key")
				return "test-account", []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			expectError: false,
			useFileFunc: false,
			fileTypes:   []ssh.KeyType{ssh.KeyTypeRSA},
		},
		{
			name: "Delete existing ED25519 key by account name",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_ed25519_test-account-ed25519")
				err := ssh.CreateSSHKey("test-account-ed25519", "test@example.com", keyFile, ssh.KeyTypeED25519)
				require.NoError(t, err, "Failed to create test key")
				return "test-account-ed25519", []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			expectError: false,
			useFileFunc: false,
			fileTypes:   []ssh.KeyType{ssh.KeyTypeED25519},
		},
		{
			name: "Delete non-existent account",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				return "nonexistent-account", nil
			},
			expectError: false, // Should not return error for non-existent accounts
			useFileFunc: false,
		},
		{
			name: "Delete existing key by file path",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_rsa_test_file")
				err := ssh.CreateSSHKey("test-account-file", "test@example.com", keyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create test key")
				return keyFile, []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			expectError: false,
			useFileFunc: true,
		},
		{
			name: "Delete non-existent key file",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				nonExistentFile := filepath.Join(tempDir, "nonexistent")
				// Make sure it doesn't exist
				os.Remove(nonExistentFile)
				os.Remove(nonExistentFile + ".pub")
				return nonExistentFile, []string{}
			},
			expectError: false, // Should not return error for non-existent files
			useFileFunc: true,
		},
		{
			name: "Delete multiple key types for same account",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				// Create both RSA and ED25519 keys for the same account
				rsaKeyFile := filepath.Join(tempDir, "id_rsa_multi-account")
				ed25519KeyFile := filepath.Join(tempDir, "id_ed25519_multi-account")

				err := ssh.CreateSSHKey("multi-account", "test@example.com", rsaKeyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create RSA test key")

				err = ssh.CreateSSHKey("multi-account", "test@example.com", ed25519KeyFile, ssh.KeyTypeED25519)
				require.NoError(t, err, "Failed to create ED25519 test key")

				return "multi-account", []string{
					rsaKeyFile,
					rsaKeyFile + ".pub",
					ed25519KeyFile,
					ed25519KeyFile + ".pub",
				}
			},
			expectError: false,
			useFileFunc: false,
			fileTypes:   []ssh.KeyType{ssh.KeyTypeRSA, ssh.KeyTypeED25519},
		},
		{
			name: "Fail when cannot delete private key",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_rsa_protected")
				err := ssh.CreateSSHKey("protected-account", "test@example.com", keyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create test key")

				// Make the directory read-only to prevent file deletion
				err = os.Chmod(tempDir, 0555) // Read and execute, no write
				require.NoError(t, err, "Failed to set directory permissions")

				t.Cleanup(func() {
					// Restore permissions for cleanup
					os.Chmod(tempDir, 0700)
				})

				// Verify the files exist and are not writable
				_, err = os.Stat(keyFile)
				require.NoError(t, err, "Key file should exist")

				return keyFile, []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			cleanup: func(t *testing.T, tempDir string) {
				// Restore permissions for cleanup
				os.Chmod(tempDir, 0700)
			},
			expectError: true,
			useFileFunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Override home directory for this test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", oldHome)

			// Create .ssh directory
			sshDir := filepath.Join(tempDir, ".ssh")
			err := os.MkdirAll(sshDir, 0700)
			require.NoError(t, err, "Failed to create .ssh directory")

			// Setup test case
			keyOrAccount, filesToCheck := tt.setup(t, sshDir)

			// Run cleanup after test if provided
			if tt.cleanup != nil {
				defer tt.cleanup(t, sshDir)
			}

			// Execute the function being tested
			var deleteErr error
			if tt.useFileFunc {
				deleteErr = ssh.DeleteSSHKeyFile(keyOrAccount)
			} else {
				deleteErr = ssh.DeleteSSHKey(keyOrAccount)
			}

			// Check for expected error
			if tt.expectError {
				require.Error(t, deleteErr, "Expected an error but got none")
				// Verify the error message indicates a permission issue
				require.Contains(t, deleteErr.Error(), "permission denied", "Expected a permission denied error")
			} else {
				require.NoError(t, deleteErr, "Unexpected error")
			}

			// Verify results
			if tt.expectError {
				// Verify files still exist when error is expected
				for _, file := range filesToCheck {
					_, err := os.Stat(file)
					if os.IsNotExist(err) {
						t.Errorf("File %s was deleted but should not have been: %v", file, err)
					} else if err != nil {
						t.Errorf("Error checking file %s: %v", file, err)
					} else {
						t.Logf("Verified that %s still exists as expected", file)
					}
				}
			} else {
				require.NoError(t, deleteErr, "Unexpected error deleting SSH key")

				// Verify files were deleted
				for _, file := range filesToCheck {
					_, err := os.Stat(file)
					if !os.IsNotExist(err) {
						if err == nil {
							t.Errorf("File %s still exists but should have been deleted", file)
						} else {
							t.Errorf("Error checking file %s: %v", file, err)
						}
					} else {
						t.Logf("Successfully verified that %s was deleted", file)
					}
				}
			}
		})
	}
}

func TestSSHConfigManagement(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string, string) error
		verify      func(*testing.T, string, string)
		cleanup     func(string, string) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "AddSSHConfigEntry with ED25519 key",
			setup: func(sshDir, accountName string) error {
				// Create .ssh directory if it doesn't exist
				if err := os.MkdirAll(sshDir, 0700); err != nil {
					return fmt.Errorf("failed to create .ssh directory: %w", err)
				}

				// Create a test ED25519 key directly
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)

				// Generate ED25519 key pair
				pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return fmt.Errorf("failed to generate ED25519 key pair: %w", err)
				}

				// Write private key
				privateKeyBytes := ssh.MarshalED25519PrivateKey(privKey, "test@example.com")
				if err := os.WriteFile(keyFile, privateKeyBytes, 0600); err != nil {
					return fmt.Errorf("failed to write private key: %w", err)
				}

				// Write public key
				publicKeyBytes, err := ssh.SSHPublicKeyED25519(pubKey, "test@example.com")
				if err != nil {
					return fmt.Errorf("failed to generate public key: %w", err)
				}
				if err := os.WriteFile(keyFile+".pub", publicKeyBytes, 0644); err != nil {
					return fmt.Errorf("failed to write public key: %w", err)
				}

				return nil
			},
			verify: func(t *testing.T, sshDir, accountName string) {
				// Mevcut verify fonksiyonunun içeriği aynı kalacak
				err := ssh.AddSSHConfigEntry(accountName)
				require.NoError(t, err, "Failed to add SSH config entry")

				// Verify the config file was updated
				configPath := filepath.Join(sshDir, "config")
				configData, err := os.ReadFile(configPath)
				require.NoError(t, err, "Failed to read SSH config file")

				// Check if the host entry was added with ED25519 key
				hostEntry := fmt.Sprintf("github.com-%s", accountName)
				assert.Contains(t, string(configData), hostEntry, "SSH config should contain the host entry")
				assert.Contains(t, string(configData), "id_ed25519_"+accountName, "SSH config should use ED25519 key")
			},
			cleanup: func(sshDir, accountName string) error {
				// Remove the config entry
				if err := ssh.RemoveSSHConfigEntry(accountName); err != nil && !os.IsNotExist(err) {
					return err
				}
				// Remove the key files
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)
				os.Remove(keyFile)
				os.Remove(keyFile + ".pub")
				return nil
			},
		},
		{
			name: "Add duplicate SSH config entry",
			setup: func(sshDir, accountName string) error {
				// Create .ssh directory if it doesn't exist
				if err := os.MkdirAll(sshDir, 0700); err != nil {
					return fmt.Errorf("failed to create .ssh directory: %w", err)
				}

				// Create a test ED25519 key directly
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)

				// Generate ED25519 key pair
				pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return fmt.Errorf("failed to generate ED25519 key pair: %w", err)
				}

				// Write private key
				privateKeyBytes := ssh.MarshalED25519PrivateKey(privKey, "test@example.com")
				if err := os.WriteFile(keyFile, privateKeyBytes, 0600); err != nil {
					return fmt.Errorf("failed to write private key: %w", err)
				}

				// Write public key
				publicKeyBytes, err := ssh.SSHPublicKeyED25519(pubKey, "test@example.com")
				if err != nil {
					return fmt.Errorf("failed to generate public key: %w", err)
				}
				if err := os.WriteFile(keyFile+".pub", publicKeyBytes, 0644); err != nil {
					return fmt.Errorf("failed to write public key: %w", err)
				}

				// Add the entry once
				return ssh.AddSSHConfigEntry(accountName)
			},
			verify: func(t *testing.T, sshDir, accountName string) {
				// Try to add the same entry again
				err := ssh.AddSSHConfigEntry(accountName)
				require.Error(t, err, "Expected error when adding duplicate entry")
				assert.Contains(t, err.Error(), "already exists", "Error message should indicate entry exists")
			},
			cleanup: func(sshDir, accountName string) error {
				// Remove the config entry
				if err := ssh.RemoveSSHConfigEntry(accountName); err != nil && !os.IsNotExist(err) {
					return err
				}
				// Remove the key files
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)
				os.Remove(keyFile)
				os.Remove(keyFile + ".pub")
				return nil
			},
		},
		{
			name: "Remove non-existent SSH config entry",
			verify: func(t *testing.T, sshDir, accountName string) {
				err := ssh.RemoveSSHConfigEntry("nonexistent-account")
				require.NoError(t, err, "Removing non-existent entry should not return an error")
			},
		},
		{
			name: "Add entry with non-existent key",
			verify: func(t *testing.T, sshDir, accountName string) {
				err := ssh.AddSSHConfigEntry("nonexistent-account")
				require.Error(t, err, "Expected error when adding entry with non-existent key")
				assert.Contains(t, err.Error(), "no SSH key found", "Error message should indicate key not found")
			},
			expectError: true,
			errorMsg:    "no SSH key found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()
			sshDir := filepath.Join(tempDir, ".ssh")
			require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

			// Override home directory for this test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", oldHome)

			// Create a test account name (remove slashes and other invalid characters)
			testName := strings.ReplaceAll(t.Name(), "/", "_")
			testName = strings.ReplaceAll(testName, " ", "_")
			accountName := "test-account-" + testName

			// Run setup if provided
			if tc.setup != nil {
				require.NoError(t, tc.setup(sshDir, accountName), "Setup failed")
			}

			// Run the test
			tc.verify(t, sshDir, accountName)

			// Run cleanup if provided
			if tc.cleanup != nil {
				require.NoError(t, tc.cleanup(sshDir, accountName), "Cleanup failed")
			}

			// Verify cleanup was successful
			if tc.cleanup != nil {
				configPath := filepath.Join(sshDir, "config")
				if _, err := os.Stat(configPath); err == nil {
					configData, err := os.ReadFile(configPath)
					require.NoError(t, err, "Failed to read config file during cleanup verification")
					assert.NotContains(t, string(configData), accountName, "Account should be removed from config after cleanup")
				}
			}
		})
	}
}

func TestDeleteSSHKeyFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, tempDir string) (string, []string) // returns keyFile and list of files that should exist
		expectError bool
		expectFiles []string // files that should exist after the operation
	}{
		{
			name: "Delete existing key file",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_rsa_test")
				err := ssh.CreateSSHKey("test-account", "test@example.com", keyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create test key")
				return keyFile, []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			expectError: false,
			expectFiles: []string{}, // All files should be deleted
		},
		{
			name: "Delete non-existent key file",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				nonExistentFile := filepath.Join(tempDir, "nonexistent")
				// Make sure it doesn't exist
				os.Remove(nonExistentFile)
				os.Remove(nonExistentFile + ".pub")
				return nonExistentFile, []string{}
			},
			expectError: false, // Should not return error for non-existent files
			expectFiles: []string{},
		},
		{
			name: "Delete only public key when private key doesn't exist",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				// Create only a public key file
				keyFile := filepath.Join(tempDir, "id_rsa_public_only")
				pubKeyFile := keyFile + ".pub"
				
				// Create a dummy public key file
				err := os.WriteFile(pubKeyFile, []byte("ssh-rsa AAAAB3NzaC1yc2E test@example.com"), 0644)
				require.NoError(t, err, "Failed to create test public key")
				
				return keyFile, []string{
					pubKeyFile,
				}
			},
			expectError: false,
			expectFiles: []string{}, // Public key should be deleted
		},
		{
			name: "Error checking private key file",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				// Create a directory with the same name as the key file to cause a stat error
				keyDir := filepath.Join(tempDir, "key_dir")
				err := os.MkdirAll(keyDir, 0700)
				require.NoError(t, err, "Failed to create directory")
				
				// Make it inaccessible to cause a stat error
				err = os.Chmod(keyDir, 0000) // No permissions
				require.NoError(t, err, "Failed to set directory permissions")
				
				t.Cleanup(func() {
					// Restore permissions for cleanup
					os.Chmod(keyDir, 0700)
				})
				
				keyFile := filepath.Join(keyDir, "id_rsa")
				return keyFile, []string{}
			},
			expectError: true,
			expectFiles: []string{},
		},
		{
			name: "Error checking public key file",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				// Create a key file but make the public key a directory to cause an error
				keyFile := filepath.Join(tempDir, "id_rsa_error")
				err := os.WriteFile(keyFile, []byte("dummy private key"), 0600)
				require.NoError(t, err, "Failed to create test key")
				
				// Create a directory with the same name as the public key file
				pubKeyDir := keyFile + ".pub"
				err = os.MkdirAll(pubKeyDir, 0700)
				require.NoError(t, err, "Failed to create directory")
				
				// Create a file inside the directory to ensure os.Remove fails
				// (can't remove non-empty directory)
				dummyFile := filepath.Join(pubKeyDir, "dummy")
				err = os.WriteFile(dummyFile, []byte("dummy"), 0600)
				require.NoError(t, err, "Failed to create dummy file")
				
				return keyFile, []string{
					keyFile,
					pubKeyDir,
				}
			},
			expectError: true,
			expectFiles: []string{"id_rsa_error.pub"}, // Directory should still exist
		},
		{
			name: "Fail when cannot delete private key",
			setup: func(t *testing.T, tempDir string) (string, []string) {
				keyFile := filepath.Join(tempDir, "id_rsa_protected")
				err := ssh.CreateSSHKey("protected-account", "test@example.com", keyFile, ssh.KeyTypeRSA)
				require.NoError(t, err, "Failed to create test key")

				// Make the directory read-only to prevent file deletion
				err = os.Chmod(tempDir, 0555) // Read and execute, no write
				require.NoError(t, err, "Failed to set directory permissions")

				t.Cleanup(func() {
					// Restore permissions for cleanup
					os.Chmod(tempDir, 0700)
				})

				// Verify the files exist and are not writable
				_, err = os.Stat(keyFile)
				require.NoError(t, err, "Key file should exist")

				return keyFile, []string{
					keyFile,
					keyFile + ".pub",
				}
			},
			expectError: true,
			expectFiles: []string{"id_rsa_protected", "id_rsa_protected.pub"}, // Files should still exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup test case
			keyFile, filesToCheck := tt.setup(t, tempDir)

			// Run the function being tested
			err := ssh.DeleteSSHKeyFile(keyFile)

			// Check for expected error
			if tt.expectError {
				require.Error(t, err, "Expected an error but got none")
				// Verify the error message based on the test case
				if tt.name == "Error checking public key file" {
					require.Contains(t, err.Error(), "directory not empty", "Expected a 'directory not empty' error")
				} else {
					require.Contains(t, err.Error(), "permission denied", "Expected a permission denied error")
				}
			} else {
				require.NoError(t, err, "Unexpected error")
			}

			// Verify file existence
			for _, file := range tt.expectFiles {
				fullPath := filepath.Join(tempDir, file)
				_, err := os.Stat(fullPath)
				assert.NoError(t, err, "Expected file %s to exist but it doesn't", fullPath)
			}

			// Verify files that should be deleted are gone
			for _, file := range filesToCheck {
				if !contains(tt.expectFiles, filepath.Base(file)) {
					_, err := os.Stat(file)
					assert.True(t, os.IsNotExist(err), "Expected file %s to be deleted but it still exists", file)
				}
			}
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func TestSSHPublicKeyED25519(t *testing.T) {
	// Generate a new ED25519 key pair
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "Failed to generate ED25519 key pair")

	comment := "test@example.com"

	// Test successful key generation
	t.Run("Success", func(t *testing.T) {
		keyData, err := sshPublicKeyED25519(pubKey, comment)
		require.NoError(t, err, "Failed to generate SSH public key")
		assert.NotEmpty(t, keyData, "Generated key should not be empty")
		assert.Contains(t, string(keyData), "ssh-ed25519", "Key should be ED25519 type")
	})

	// Test with empty public key
	t.Run("EmptyPublicKey", func(t *testing.T) {
		_, err := sshPublicKeyED25519(nil, comment)
		assert.Error(t, err, "Should return error for nil public key")
	})
}

func TestMarshalED25519PrivateKey(t *testing.T) {
	// Generate a new ED25519 key pair
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "Failed to generate ED25519 key pair")

	comment := "test@example.com"

	t.Run("ValidKey", func(t *testing.T) {
		keyData := marshalED25519PrivateKey(privKey, comment)
		assert.NotEmpty(t, keyData, "Marshaled key should not be empty")
		assert.Contains(t, string(keyData), "PRIVATE KEY", "Should contain PRIVATE KEY header")
	})

	t.Run("NilKey", func(t *testing.T) {
		// This should panic with nil key, so we'll recover from the panic
		assert.Panics(t, func() {
			_ = marshalED25519PrivateKey(nil, comment)
		}, "Should panic with nil key")
	})
}

func TestValidatePrivateKey(t *testing.T) {
	// Generate a valid ED25519 private key
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "Failed to generate ED25519 key pair")

	t.Run("ValidKey", func(t *testing.T) {
		keyData := marshalED25519PrivateKey(privKey, "test@example.com")
		err := validatePrivateKey(keyData)
		assert.NoError(t, err, "Valid key should pass validation")
	})

	t.Run("InvalidKey", func(t *testing.T) {
		err := validatePrivateKey([]byte("invalid-key-data"))
		assert.Error(t, err, "Invalid key should fail validation")
	})

	t.Run("EmptyKey", func(t *testing.T) {
		err := validatePrivateKey([]byte("")) // Empty string as JSON
		assert.Error(t, err, "Empty key should fail validation")
	})
}

func TestAddSSHKeyToAgent(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

	// Test account details
	accountName := "test-agent-account"
	keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)

	// Save original ExecCommand and restore it after the test
	oldExecCommand := ssh.ExecCommand
	defer func() { ssh.ExecCommand = oldExecCommand }()

	// Set SSH_AUTH_SOCK for agent tests
	oldSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", oldSSHAuthSock)
	os.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-agent.sock")

	tests := []struct {
		name          string
		setup         func()
		expectedError string
	}{
		{
			name: "Successfully add key to agent",
			setup: func() {
				// Create a test private key file directly
				_, privKey, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err, "Failed to generate test key")
				keyData := ssh.MarshalED25519PrivateKey(privKey, "test@example.com")
				require.NoError(t, os.WriteFile(keyFile, keyData, 0600), "Failed to write test key")

				// Mock successful ssh-add command
				ssh.ExecCommand = mockCommand("ssh-add", true)
			},
			expectedError: "",
		},
		{
			name: "Key file does not exist",
			setup: func() {
				// Remove the key file if it exists
				os.Remove(keyFile)
				// Set a mock that shouldn't be called
				ssh.ExecCommand = func(name string, arg ...string) *exec.Cmd {
					t.Error("ssh-add should not be called when key file doesn't exist")
					return exec.Command("true")
				}
			},
			expectedError: "private key file",
		},
		{
			name: "SSH agent not running",
			setup: func() {
				// Create a test private key file directly
				_, privKey, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err, "Failed to generate test key")
				keyData := ssh.MarshalED25519PrivateKey(privKey, "test@example.com")
				require.NoError(t, os.WriteFile(keyFile, keyData, 0600), "Failed to write test key")

				// Unset SSH_AUTH_SOCK to simulate agent not running
				os.Unsetenv("SSH_AUTH_SOCK")
			},
			expectedError: "SSH agent is not running",
		},
		{
			name: "SSH add command fails",
			setup: func() {
				// Create a test private key file directly
				_, privKey, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err, "Failed to generate test key")
				keyData := ssh.MarshalED25519PrivateKey(privKey, "test@example.com")
				require.NoError(t, os.WriteFile(keyFile, keyData, 0600), "Failed to write test key")

				// Mock failing ssh-add command
				ssh.ExecCommand = mockCommand("ssh-add", false)

				// Set SSH_AUTH_SOCK to simulate running agent
				os.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-agent.sock")
			},
			expectedError: "failed to add key to SSH agent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset SSH_AUTH_SOCK to default for each test case
			os.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-agent.sock")
			
			// Setup test case
			tc.setup()

			// Run the function
			err := ssh.AddSSHKeyToAgent(accountName)

			// Verify results
			if tc.expectedError != "" {
				require.Error(t, err, "Expected error but got none")
				assert.Contains(t, err.Error(), tc.expectedError, "Error message should contain expected text")
			} else {
				require.NoError(t, err, "Unexpected error")
			}
		})
	}
}
