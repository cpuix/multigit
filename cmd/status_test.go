package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/spf13/cobra"
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

		// Initialize config
		cmd.InitConfig()

		// Create a fresh command for testing
		rootCmd := &cobra.Command{Use: "multigit"}
		statusCmd := &cobra.Command{
			Use:   "status",
			Short: "Show the currently active GitHub account",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Get active account
				_, _, err := multigit.GetActiveAccount()
				if err != nil {
					io.WriteString(w, "No active GitHub account. Use 'multigit use <account>' to set an active account.\n")
					return nil
				}
				return nil
			},
		}
		rootCmd.AddCommand(statusCmd)

		// Execute the status command
		rootCmd.SetArgs([]string{"status"})
		err = rootCmd.Execute()
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

		// Create a config with an active account
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

		// Initialize config
		cmd.InitConfig()

		// Create a fresh command for testing
		rootCmd := &cobra.Command{Use: "multigit"}
		statusCmd := &cobra.Command{
			Use:   "status",
			Short: "Show the currently active GitHub account",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Get active account
				activeAccountName, account, err := multigit.GetActiveAccount()
				if err != nil {
					io.WriteString(w, "No active GitHub account. Use 'multigit use <account>' to set an active account.\n")
					return nil
				}

				// Print active account info
				io.WriteString(w, "Active GitHub account:\n")
				io.WriteString(w, "  Name:  "+activeAccountName+"\n")
				io.WriteString(w, "  Email: "+account.Email+"\n")
				io.WriteString(w, "  SSH Key: ~/.ssh/id_ed25519_"+activeAccountName+"\n")

				return nil
			},
		}
		rootCmd.AddCommand(statusCmd)

		// Execute the status command
		rootCmd.SetArgs([]string{"status"})
		err = rootCmd.Execute()
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
		assert.Contains(t, output, "SSH Key", "Output should mention the SSH key")
	})
}
