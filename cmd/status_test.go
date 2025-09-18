package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCommand(t *testing.T) {
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

	t.Run("NoActiveAccount", func(t *testing.T) {
		// Save original stdout and restore it after the test
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Create a config with no active account
		config := multigit.NewConfig()
		err := multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		// Initialize config and execute the actual status command
		cmd.InitConfig()
		cmd.RootCmd.SetArgs([]string{"status"})

		oldColorOutput := color.Output
		color.Output = w
		defer func() { color.Output = oldColorOutput }()

		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		// Read command output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify output contains expected message
		assert.Contains(t, output, "No active GitHub account", "Output should indicate no active account")
	})

	t.Run("WithActiveAccount", func(t *testing.T) {
		// Save original stdout and restore it after the test
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Prepare SSH directory with an ED25519 key for the active account
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err, "Failed to create SSH directory")

		edKeyPath := filepath.Join(sshDir, "id_ed25519_test-account")
		err = os.WriteFile(edKeyPath, []byte("dummy"), 0600)
		require.NoError(t, err, "Failed to create ED25519 key file")

		// Create a config with an active account
		config := multigit.NewConfig()
		config.Accounts = map[string]multigit.Account{
			"test-account": {
				Name:  "Test User",
				Email: "test@example.com",
			},
		}
		config.ActiveAccount = "test-account"
		err = multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		// Initialize config and execute the actual status command
		cmd.InitConfig()
		cmd.RootCmd.SetArgs([]string{"status"})

		oldColorOutput := color.Output
		color.Output = w
		defer func() { color.Output = oldColorOutput }()

		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		// Read command output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify output contains expected account information
		assert.Contains(t, output, "Active GitHub account", "Output should indicate an active account")
		assert.Contains(t, output, "test-account", "Output should contain the account name")
		assert.Contains(t, output, "test@example.com", "Output should contain the account email")
		assert.Contains(t, output, edKeyPath, "Output should report the ED25519 SSH key path")
		assert.NotContains(t, output, "not found", "Output should not report missing keys for existing ED25519 key")
	})
}
