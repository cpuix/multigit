package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
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
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		config := multigit.NewConfig()
		err := multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		cmd.InitConfig()

		cmd.RootCmd.SetArgs([]string{"status"})
		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "No active GitHub account", "Output should indicate no active account")
	})

	t.Run("WithActiveAccount", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		config := multigit.NewConfig()
		config.Accounts = map[string]multigit.Account{
			"test-account": {
				Name:  "Test User",
				Email: "test@example.com",
			},
		}
		config.ActiveAccount = "test-account"
		err := multigit.SaveConfigToFile(config, configPath)
		require.NoError(t, err, "Failed to save config")

		// Ensure the SSH key path exists so the command reports it as present
		keyInfo, err := ssh.ResolveKeyPath("test-account")
		require.NoError(t, err, "Failed to resolve key path")
		require.NoError(t, os.MkdirAll(filepath.Dir(keyInfo.Path), 0700))
		require.NoError(t, os.WriteFile(keyInfo.Path, []byte("test-key"), 0600))

		cmd.InitConfig()

		cmd.RootCmd.SetArgs([]string{"status"})
		err = cmd.RootCmd.Execute()
		require.NoError(t, err, "Command execution failed")

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Active GitHub account", "Output should indicate an active account")
		assert.Contains(t, output, "test-account", "Output should contain the account name")
		assert.Contains(t, output, "test@example.com", "Output should contain the account email")
		assert.Contains(t, output, keyInfo.Path, "Output should mention the SSH key path")
		assert.Contains(t, output, string(ssh.KeyTypeED25519), "Output should mention the key type")
	})
}
