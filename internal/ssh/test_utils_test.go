package ssh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigAssertions tests the AssertConfigContains and AssertConfigNotContains utility functions
func TestConfigAssertions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a test config file
	configFile := filepath.Join(tempDir, "config")
	configContent := `# SSH Config File for Testing
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_github

Host gitlab.com
  HostName gitlab.com
  User git
  IdentityFile ~/.ssh/id_rsa_gitlab
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Create a TestConfig instance
	testConfig := &ssh.TestConfig{
		TempDir:    tempDir,
		ConfigFile: configFile,
	}

	// Test cases for AssertConfigContains with existing content
	t.Run("AssertConfigContains with existing content", func(t *testing.T) {
		// These should not panic
		assert.NotPanics(t, func() {
			testConfig.AssertConfigContains(t, "Host github.com")
		})
		assert.NotPanics(t, func() {
			testConfig.AssertConfigContains(t, "IdentityFile ~/.ssh/id_rsa_gitlab")
		})
	})

	// Test cases for AssertConfigNotContains with non-existing content
	t.Run("AssertConfigNotContains with non-existing content", func(t *testing.T) {
		// These should not panic
		assert.NotPanics(t, func() {
			testConfig.AssertConfigNotContains(t, "Host bitbucket.org")
		})
		assert.NotPanics(t, func() {
			testConfig.AssertConfigNotContains(t, "IdentityFile ~/.ssh/id_rsa_bitbucket")
		})
	})

	// Test with empty config file
	t.Run("Empty config file", func(t *testing.T) {
		// Create an empty config file
		emptyConfigFile := filepath.Join(tempDir, "empty_config")
		err := os.WriteFile(emptyConfigFile, []byte{}, 0644)
		require.NoError(t, err, "Failed to write empty config file")

		emptyConfig := &ssh.TestConfig{
			TempDir:    tempDir,
			ConfigFile: emptyConfigFile,
		}

		// Should not find any content in empty file
		assert.NotPanics(t, func() {
			emptyConfig.AssertConfigNotContains(t, "Host github.com")
		})
	})

	// Test with modified config
	t.Run("Config modification", func(t *testing.T) {
		// Create a new config file for this test
		modifiedConfigFile := filepath.Join(tempDir, "modified_config")
		err := os.WriteFile(modifiedConfigFile, []byte(configContent), 0644)
		require.NoError(t, err, "Failed to write modified config file")

		modifiedConfig := &ssh.TestConfig{
			TempDir:    tempDir,
			ConfigFile: modifiedConfigFile,
		}

		// Verify initial content
		assert.NotPanics(t, func() {
			modifiedConfig.AssertConfigContains(t, "Host github.com")
		})
		assert.NotPanics(t, func() {
			modifiedConfig.AssertConfigNotContains(t, "Host bitbucket.org")
		})

		// Modify the config file
		newContent := configContent + "\nHost bitbucket.org\n  User git\n  IdentityFile ~/.ssh/id_rsa_bitbucket\n"
		err = os.WriteFile(modifiedConfigFile, []byte(newContent), 0644)
		require.NoError(t, err, "Failed to update modified config file")

		// Verify new content
		assert.NotPanics(t, func() {
			modifiedConfig.AssertConfigContains(t, "Host bitbucket.org")
		})
	})
}
