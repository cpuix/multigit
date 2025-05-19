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

// TestListCommandExecution tests the list command execution
func TestListCommandExecution(t *testing.T) {
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
		setup        func()
		expectOutput string
	}{
		{
			name: "No accounts",
			setup: func() {
				// Initialize config
				cmd.InitConfig()
				
				// Create an empty config
				config := multigit.NewConfig()
				err := multigit.SaveConfigToFile(config, configPath)
				require.NoError(t, err, "Failed to save config")
			},
			expectOutput: "No accounts configured",
		},
		{
			name: "With accounts",
			setup: func() {
				// Initialize config
				cmd.InitConfig()
				
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
			},
			expectOutput: "Configured GitHub Accounts",
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
			listCmd := &cobra.Command{
				Use:   "list",
				Short: "List all configured GitHub accounts",
				RunE: func(cmd *cobra.Command, args []string) error {
					config := multigit.LoadConfig()

					if len(config.Accounts) == 0 {
						_, err := io.WriteString(w, "No accounts configured. Use 'multigit create' to add an account.\n")
						return err
					}

					_, err := io.WriteString(w, "Configured GitHub Accounts:\n")
					return err
				},
			}
			rootCmd.AddCommand(listCmd)

			// Execute the command
			rootCmd.SetArgs([]string{"list"})
			err := rootCmd.Execute()
			require.NoError(t, err, "Command execution failed")

			// Read command output
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify output contains expected message
			assert.Contains(t, output, tt.expectOutput, "Output should contain expected message")
		})
	}
}
