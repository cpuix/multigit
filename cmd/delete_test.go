package cmd_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAccount(t *testing.T, config *multigit.Config) {
	t.Helper()
	if config.Accounts == nil {
		config.Accounts = make(map[string]multigit.Account)
	}
	config.Accounts["test-account"] = multigit.Account{
		Name:  "test-account",
		Email: "test@example.com",
	}
}

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp dir for test config
	tempDir := t.TempDir()
	cfgFile := filepath.Join(tempDir, ".config", "multigit", "config.json")

	// Set config file path and HOME
	oldCfg := cmd.CfgFile
	oldHome := os.Getenv("HOME")
	cmd.CfgFile = cfgFile
	require.NoError(t, os.Setenv("HOME", tempDir))

	// Initialize config
	cmd.InitConfig()

	// Setup test account
	config := multigit.LoadConfig()
	setupTestAccount(t, &config)
	require.NoError(t, multigit.SaveConfig(config), "Failed to save test config")

	// Return cleanup function
	return cfgFile, func() {
		if oldHome == "" {
			require.NoError(t, os.Unsetenv("HOME"))
		} else {
			require.NoError(t, os.Setenv("HOME", oldHome))
		}
		cmd.CfgFile = oldCfg
	}
}

func TestDeleteAccount(t *testing.T) {
	// Setup test config
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Test cases
	tests := []struct {
		name        string
		accountName string
		setup       func()
		expectError bool
	}{
		{
			name:        "Delete existing account",
			accountName: "test-account",
			setup: func() {
				config := multigit.LoadConfig()
				setupTestAccount(t, &config)
				require.NoError(t, multigit.SaveConfig(config), "Failed to setup test account")
			},
			expectError: false,
		},
		{
			name:        "Delete non-existent account",
			accountName: "nonexistent",
			setup:       func() {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			// Execute the delete function directly
			err := multigit.DeleteAccount(tt.accountName)

			// Verify the results
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")

				// Verify the account was deleted
				config := multigit.LoadConfig()
				_, exists := config.Accounts[tt.accountName]
				assert.False(t, exists, "Account should be deleted")
			}
		})
	}
}

func TestDeleteCommandForceSkipsPrompt(t *testing.T) {
	// Setup test config and ensure test account exists
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	config := multigit.LoadConfig()
	setupTestAccount(t, &config)
	require.NoError(t, multigit.SaveConfig(config), "Failed to setup test account")

	// Capture stdout to verify prompt is skipped
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Execute delete command with force flag
	oldStdin := os.Stdin
	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)
	require.NoError(t, stdinWriter.Close())
	os.Stdin = stdinReader
	defer func() {
		os.Stdin = oldStdin
		stdinReader.Close()
	}()
	cmd.RootCmd.SetArgs([]string{"delete", "test-account", "--force"})
	defer cmd.RootCmd.SetArgs([]string{})

	err = cmd.RootCmd.Execute()
	require.NoError(t, err, "Delete command with --force should succeed")

	require.NoError(t, w.Close())
	output, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	outStr := string(output)
	assert.NotContains(t, outStr, "Are you sure you want to continue?", "Force flag should skip confirmation prompt")
	assert.Contains(t, outStr, "deleted successfully", "Delete command should report success")

	// Ensure the account was deleted
	config = multigit.LoadConfig()
	_, exists := config.Accounts["test-account"]
	assert.False(t, exists, "Account should be deleted when using --force")
}
