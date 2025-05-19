package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "multigit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "Failed to create config directory")
	configPath := filepath.Join(configDir, "config.json")

	// Save original environment variables
	oldCfgFile := cmd.CfgFile
	oldHome := os.Getenv("HOME")

	// Set the test environment
	os.Setenv("HOME", tempDir)
	cmd.CfgFile = configPath

	// Clean up after the test
	defer func() {
		os.Setenv("HOME", oldHome)
		cmd.CfgFile = oldCfgFile
	}()

	t.Run("NoAccounts", func(t *testing.T) {
		// Save original stdout and restore it after the test
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Create an empty config
		config := multigit.NewConfig()
		err := multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		// Initialize config
		cmd.InitConfig()

		// Execute the list command
		cmd.RootCmd.SetArgs([]string{"list"})
		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		// Read command output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify output contains expected message
		assert.Contains(t, output, "No accounts configured", "Output should indicate no accounts are configured")
	})

	t.Run("WithAccounts", func(t *testing.T) {
		// Save original stdout and restore it after the test
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Create a config with accounts
		config := multigit.NewConfig()
		config.Accounts = map[string]multigit.Account{
			"test-account": {
				Name:  "Test User",
				Email: "test@example.com",
			},
			"another-account": {
				Name:  "Another User",
				Email: "another@example.com",
			},
		}
		config.ActiveAccount = "test-account"
		err := multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		// Initialize config
		cmd.InitConfig()

		// Execute the list command
		cmd.RootCmd.SetArgs([]string{"list"})
		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		// Read command output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify output contains expected account information
		assert.Contains(t, output, "test-account", "Output should contain the test account name")
		assert.Contains(t, output, "another-account", "Output should contain the another account name")
		assert.Contains(t, output, "test@example.com", "Output should contain the test account email")
		assert.Contains(t, output, "another@example.com", "Output should contain the another account email")
		assert.Contains(t, output, "Active account", "Output should indicate which account is active")
	})
}
