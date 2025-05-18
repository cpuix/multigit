package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testUseCommand is a test helper that sets up the test environment for the use command
type testUseCommand struct {
	gitCommands []string
	isGitRepo   bool
	t           *testing.T
}

// mockRunGitCommand mocks the RunGitCommand function for testing
func (tuc *testUseCommand) mockRunGitCommand(args ...string) error {
	tuc.gitCommands = append(tuc.gitCommands, args[0])
	tuc.t.Logf("git command: %v", args)
	return nil
}

// mockIsGitRepo mocks the IsGitRepo function for testing
func (tuc *testUseCommand) mockIsGitRepo() bool {
	tuc.t.Logf("Checking if current directory is a git repo: %v", tuc.isGitRepo)
	return tuc.isGitRepo
}

// save original functions for restoration
var originalRunGitCommand = cmd.RunGitCommand
var originalIsGitRepo = cmd.IsGitRepo

// TestUseCommand tests the use command
func TestUseCommand(t *testing.T) {
	// Create test helper
	tuc := &testUseCommand{
		t:         t,
		isGitRepo: true, // Default to true for tests
	}

	// Save original functions
	oldRunGitCommand := cmd.RunGitCommand
	oldIsGitRepo := cmd.IsGitRepo

	// Replace with our mocks
	cmd.RunGitCommand = tuc.mockRunGitCommand
	cmd.IsGitRepo = tuc.mockIsGitRepo

	// Restore original functions after test
	t.Cleanup(func() {
		cmd.RunGitCommand = oldRunGitCommand
		cmd.IsGitRepo = oldIsGitRepo
	})

	// Setup test environment
	tempDir := t.TempDir()
	testConfigPath := filepath.Join(tempDir, "config.json")

	// Set up test config
	os.Setenv("MULTIGIT_CONFIG", testConfigPath)
	t.Cleanup(func() { os.Unsetenv("MULTIGIT_CONFIG") })

	// Mock home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700))

	// Create a test SSH key
	testSSHKeyPath := filepath.Join(sshDir, "id_rsa_test-account")
	require.NoError(t, os.WriteFile(testSSHKeyPath, []byte("dummy key"), 0600))

	// Initialize a new config
	config := &multigit.Config{
		Accounts: make(map[string]multigit.Account),
	}

	// Create a test account
	config.Accounts["test-account"] = multigit.Account{
		Name:  "Test User",
		Email: "test@example.com",
	}
	config.ActiveAccount = "test-account"

	// Save the config
	viper.SetConfigFile(testConfigPath)
	viper.Set("accounts", config.Accounts)
	viper.Set("active_account", config.ActiveAccount)
	require.NoError(t, viper.WriteConfig())

	// Mock SSH functions
	oldSSHAddToAgent := cmd.SSHAddToAgentFunc
	cmd.SSHAddToAgentFunc = func(accountName string) error {
		// Skip actual SSH agent operations in tests
		return nil
	}
	t.Cleanup(func() { cmd.SSHAddToAgentFunc = oldSSHAddToAgent })

	// Test cases
	tests := []struct {
		name        string
		args        []string
		setup       func()
		expectError bool
		errMsg      string
	}{
		{
			name: "Switch to existing account",
			args: []string{"use", "test-account"},
			setup: func() {
				// Ensure the test account exists in the config
				config := multigit.LoadConfig()
				config.Accounts["test-account"] = multigit.Account{
					Name:  "Test User",
					Email: "test@example.com",
				}
				require.NoError(t, multigit.SaveConfig(config))
			},
			expectError: false,
		},
		{
			name: "Non-existent account",
			args: []string{"use", "nonexistent"},
			setup: func() {
				// Ensure the test account does not exist in the config
				config := multigit.LoadConfig()
				delete(config.Accounts, "nonexistent")
				require.NoError(t, multigit.SaveConfig(config))
			},
			expectError: true,
			errMsg:      "account 'nonexistent' does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test setup if provided
			if tt.setup != nil {
				tt.setup()
			}

			// Save and restore working directory
			oldDir, _ := os.Getwd()
			t.Cleanup(func() { os.Chdir(oldDir) })

			// Change to a temporary directory
			tempDir := t.TempDir()
			os.Chdir(tempDir)

			// Reset viper for each test case
			viper.Reset()
			viper.SetConfigFile(testConfigPath)
			require.NoError(t, viper.ReadInConfig())

			// Reset command tracking
			tuc.gitCommands = nil

			// Execute the command with the test arguments
			cmd.RootCmd.SetArgs(tt.args)
			err := cmd.RootCmd.Execute()

			t.Logf("Git commands executed: %v", tuc.gitCommands)

			// Verify the correct git commands were called
			if !tt.expectError {
				// Verify that git config was called with the correct arguments
				assert.Greater(t, len(tuc.gitCommands), 0, "Expected at least one git command to be executed")
				// Add more specific assertions about the git commands if needed
			}

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected an error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error")
			}
		})
	}
}
