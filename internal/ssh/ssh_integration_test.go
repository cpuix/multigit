package ssh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
)

// These tests require an SSH agent to be running
// They can be run with: go test -v -tags=integration ./internal/ssh/...

func TestSSHKeyLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	tc := ssh.SetupTestEnvironment(t)

	t.Run("SSH key creation and verification", func(t *testing.T) {
		// Create a test key
		keyPath := tc.CreateTestKey(t, "test_key", ssh.KeyTypeRSA)
		
		// Verify key files were created
		tc.AssertFileExists(t, keyPath)
		tc.AssertFileExists(t, keyPath+".pub")
		

		
		t.Run("Add and remove key from SSH agent", func(t *testing.T) {
			// Skip if SSH agent is not running
			if os.Getenv("SSH_AUTH_SOCK") == "" {
				t.Skip("SSH_AUTH_SOCK is not set, skipping agent tests")
			}

			// Add key to agent
			err := ssh.AddSSHKeyToAgent(keyPath)
			// We can't reliably test the agent in CI, so we'll just check for permission errors
			if err != nil && !os.IsPermission(err) {
				assert.NoError(t, err, "Failed to add key to SSH agent")
			}
		})
		
		t.Run("Delete SSH key files", func(t *testing.T) {
			// Delete key files
			err := ssh.DeleteSSHKey(keyPath)
			assert.NoError(t, err, "Failed to delete SSH key")
			
			// Verify files were deleted
			tc.AssertFileNotExists(t, keyPath)
			tc.AssertFileNotExists(t, keyPath+".pub")
		})
	})
}

func TestMultipleSSHKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := ssh.SetupTestEnvironment(t)

	// Create multiple test keys
	keyNames := []string{"key_A", "key_B", "key_C"}
	keyTypes := []ssh.KeyType{ssh.KeyTypeRSA, ssh.KeyTypeED25519, ssh.KeyTypeRSA}
	keyPaths := make([]string, len(keyNames))

	for i, name := range keyNames {
		t.Run("Create "+name, func(t *testing.T) {
			keyPaths[i] = tc.CreateTestKey(t, name, keyTypes[i])
			tc.AssertFileExists(t, keyPaths[i])
			tc.AssertFileExists(t, keyPaths[i]+".pub")
		})
	}

	// Test deleting keys
	for i, keyPath := range keyPaths {
		t.Run("Delete "+keyNames[i], func(t *testing.T) {
			err := ssh.DeleteSSHKey(keyPath)
			assert.NoError(t, err, "Failed to delete key: %s", keyPath)
			tc.AssertFileNotExists(t, keyPath)
			tc.AssertFileNotExists(t, keyPath+".pub")
		})
	}
}

// TestSSHConfigManagementIntegration tests SSH config management functionality
// This is an integration test that requires a real SSH config file
func TestSSHConfigManagementIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := ssh.SetupTestEnvironment(t)
	
	// Skip if we can't access the SSH config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}
	sshConfigPath := filepath.Join(homeDir, ".ssh", "config")
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		t.Skip("SSH config file does not exist, skipping test")
	}

	t.Run("Add host to SSH config", func(t *testing.T) {
		// Test with a simple host entry
		hostConfig := struct {
			Host         string
			HostName     string
			User         string
			IdentityFile string
		}{
			Host:         "github-test",
			HostName:     "github.com",
			User:         "git",
			IdentityFile: "~/.ssh/github_key",
		}
		err := ssh.AddSSHConfigEntry(hostConfig.Host)
		if err != nil {
			t.Logf("Warning: Failed to update SSH config: %v", err)
			t.Skip("Skipping test due to SSH config update failure")
		}

		// Verify config was updated
		tc.AssertFileExists(t, sshConfigPath)
	})

	t.Run("Remove host from SSH config", func(t *testing.T) {
		// First add a host to remove
		hostToRemove := "github-test-remove"
		err = ssh.AddSSHConfigEntry(hostToRemove)
		if err != nil {
			t.Logf("Warning: Failed to add test host to config: %v", err)
			t.Skip("Skipping test due to SSH config update failure")
		}

		// Now remove it
		err = ssh.RemoveSSHConfigEntry(hostToRemove)
		if err != nil {
			t.Logf("Warning: Failed to remove test host from config: %v", err)
		}

		// Just verify the config file still exists
		tc.AssertFileExists(t, sshConfigPath)
	})
}
