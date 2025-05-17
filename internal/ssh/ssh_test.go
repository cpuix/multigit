package ssh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndDeleteSSHKey(t *testing.T) {
	tempDir := t.TempDir()
	
	// Override the home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Test account details
	accountName := "test-account"
	accountEmail := "test@example.com"
	passphrase := "test-passphrase"

	// Test key creation
	t.Run("CreateSSHKey", func(t *testing.T) {
		err := ssh.CreateSSHKey(accountName, accountEmail, passphrase, ssh.KeyTypeED25519)
		require.NoError(t, err, "Failed to create SSH key")

		// Check if key files were created (ED25519 keys)
		privateKeyPath := filepath.Join(tempDir, ".ssh", "id_ed25519_"+accountName)
		publicKeyPath := privateKeyPath + ".pub"

		_, err = os.Stat(privateKeyPath)
		assert.NoError(t, err, "Private key file should exist")

		_, err = os.Stat(publicKeyPath)
		assert.NoError(t, err, "Public key file should exist")
	})

	// Test key deletion
	t.Run("DeleteSSHKey", func(t *testing.T) {
		err := ssh.DeleteSSHKey(accountName)
		require.NoError(t, err, "Failed to delete SSH key")

		// Check if key files were deleted (ED25519 keys)
		privateKeyPath := filepath.Join(tempDir, ".ssh", "id_ed25519_"+accountName)
		publicKeyPath := privateKeyPath + ".pub"

		_, err = os.Stat(privateKeyPath)
		assert.True(t, os.IsNotExist(err), "Private key file should not exist")

		_, err = os.Stat(publicKeyPath)
		assert.True(t, os.IsNotExist(err), "Public key file should not exist")
	})
}

func TestSSHConfigManagement(t *testing.T) {
	tempDir := t.TempDir()
	
	// Override the home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700), "Failed to create .ssh directory")

	accountName := "test-config-account"

	t.Run("AddSSHConfigEntry", func(t *testing.T) {
		err := ssh.AddSSHConfigEntry(accountName)
		require.NoError(t, err, "Failed to add SSH config entry")

		// Check if config file was created
		configPath := filepath.Join(sshDir, "config")
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err, "Failed to read SSH config file")

		// Check if the entry was added
		assert.Contains(t, string(configData), "github.com-"+accountName, "Config should contain the new host entry")
	})

	t.Run("RemoveSSHConfigEntry", func(t *testing.T) {
		err := ssh.RemoveSSHConfigEntry(accountName)
		require.NoError(t, err, "Failed to remove SSH config entry")

		// Verify the entry was removed
		configPath := filepath.Join(sshDir, "config")
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err, "Failed to read SSH config file")

		assert.NotContains(t, string(configData), "github.com-"+accountName, "Config should not contain the removed host entry")
	})
}
