package ssh_test

import (
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
				// Create a test ED25519 key
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)
				return ssh.CreateSSHKey(accountName, "test@example.com", keyFile, ssh.KeyTypeED25519)
			},
			verify: func(t *testing.T, sshDir, accountName string) {
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
				return ssh.RemoveSSHConfigEntry(accountName)
			},
		},
		{
			name: "Add duplicate SSH config entry",
			setup: func(sshDir, accountName string) error {
				// Create a test ED25519 key first
				keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)
				if err := ssh.CreateSSHKey(accountName, "test@example.com", keyFile, ssh.KeyTypeED25519); err != nil {
					return err
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
				return ssh.RemoveSSHConfigEntry(accountName)
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

			// Create a test account name
			accountName := "test-account" + t.Name()

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
				// Verify the error message indicates a permission issue
				require.Contains(t, err.Error(), "permission denied", "Expected a permission denied error")
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
				// Create a test private key file
				err := ssh.CreateSSHKey(accountName, "test@example.com", keyFile, ssh.KeyTypeED25519)
				require.NoError(t, err, "Failed to create test key")

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
				// Create a test private key file
				err := ssh.CreateSSHKey(accountName, "test@example.com", keyFile, ssh.KeyTypeED25519)
				require.NoError(t, err, "Failed to create test key")

				// Mock failing ssh-add command
				ssh.ExecCommand = mockCommand("ssh-add", false)

				// Set SSH_AUTH_SOCK to simulate running agent
				os.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-agent.sock")
			},
			expectedError: "failed to add key to SSH agent",
		},
		{
			name: "SSH add command fails",
			setup: func() {
				// Create a test private key file
				err := ssh.CreateSSHKey(accountName, "test@example.com", keyFile, ssh.KeyTypeED25519)
				require.NoError(t, err, "Failed to create test key")

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
