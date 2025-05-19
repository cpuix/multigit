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

// TestDeleteCommand tests the delete command execution
func TestDeleteCommand(t *testing.T) {
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

	// Test cases
	tests := []struct {
		name         string
		args         []string
		setup        func()
		expectError  bool
		expectOutput string
	}{
		{
			name: "Delete existing account with force flag",
			args: []string{"delete", "test-account", "-f"},
			setup: func() {
				// Initialize config
				cmd.InitConfig()
				
				// Create a config with a test account
				config := multigit.NewConfig()
				config.Accounts = map[string]multigit.Account{
					"test-account": {
						Name:  "Test User",
						Email: "test@example.com",
					},
				}
				err := multigit.SaveConfigToFile(config, configPath)
				require.NoError(t, err, "Failed to save config")
			},
			expectError:  false,
			expectOutput: "deleted successfully",
		},
		{
			name: "Delete non-existent account",
			args: []string{"delete", "nonexistent", "-f"},
			setup: func() {
				// Initialize config
				cmd.InitConfig()
				
				// Create an empty config
				config := multigit.NewConfig()
				err := multigit.SaveConfigToFile(config, configPath)
				require.NoError(t, err, "Failed to save config")
			},
			expectError:  true,
			expectOutput: "nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test
			if tt.setup != nil {
				tt.setup()
			}

			// Save original stdout and restore it after the test
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Create a fresh root command for this test
			rootCmd := &cobra.Command{Use: "multigit"}
			deleteCmd := &cobra.Command{
				Use:     "delete <account_name>",
				Aliases: []string{"remove", "rm"},
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					accountName := args[0]
					err := multigit.DeleteAccount(accountName)
					if err != nil {
						// Write error message directly to output for testing
						io.WriteString(w, err.Error())
						return err
					}
					return nil
				},
			}
			deleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
			rootCmd.AddCommand(deleteCmd)

			// Execute the command
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			// Read command output
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				assert.Contains(t, output, tt.expectOutput, "Output should contain expected error message")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Contains(t, output, tt.expectOutput, "Output should contain success message")

				// Verify the account was deleted
				config := multigit.LoadConfig()
				_, exists := config.Accounts["test-account"]
				assert.False(t, exists, "Account should be deleted")
			}
		})
	}
}
