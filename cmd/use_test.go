package cmd_test

import (
	"fmt"
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
	gitCommands [][]string
	isGitRepo   bool
	t           *testing.T
}

// mockRunGitCommand mocks the RunGitCommand function for testing
func (tuc *testUseCommand) mockRunGitCommand(args ...string) error {
	copiedArgs := append([]string(nil), args...)
	tuc.gitCommands = append(tuc.gitCommands, copiedArgs)
	tuc.t.Logf("git command: %v", args)
	return nil
}

// mockIsGitRepo mocks the IsGitRepo function for testing
func (tuc *testUseCommand) mockIsGitRepo() bool {
	tuc.t.Logf("Checking if current directory is a git repo: %v", tuc.isGitRepo)
	return tuc.isGitRepo
}

// TestUseCommand tests the use command
func TestUseCommand(t *testing.T) {
	tuc := &testUseCommand{
		t:         t,
		isGitRepo: true,
	}

	oldRunGitCommand := cmd.RunGitCommand
	oldIsGitRepo := cmd.IsGitRepo

	cmd.RunGitCommand = tuc.mockRunGitCommand
	cmd.IsGitRepo = tuc.mockIsGitRepo

	t.Cleanup(func() {
		cmd.RunGitCommand = oldRunGitCommand
		cmd.IsGitRepo = oldIsGitRepo
	})

	homeDir := t.TempDir()
	testConfigPath := filepath.Join(homeDir, ".config", "multigit", "config.json")

	require.NoError(t, os.MkdirAll(filepath.Dir(testConfigPath), 0o700))

	os.Setenv("MULTIGIT_CONFIG", testConfigPath)
	t.Cleanup(func() { os.Unsetenv("MULTIGIT_CONFIG") })

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	ed25519KeyPath := filepath.Join(sshDir, "id_ed25519_test-account")
	rsaKeyPath := filepath.Join(sshDir, "id_rsa_test-account")

	baseAccount := multigit.Account{
		Name:  "Test User",
		Email: "test@example.com",
	}

	writeConfig := func(t *testing.T, withAccount bool) {
		t.Helper()
		config := multigit.NewConfig()
		if withAccount {
			config.Accounts["test-account"] = baseAccount
			config.ActiveAccount = "test-account"
		}
		require.NoError(t, multigit.SaveConfig(config))
	}

	removeIfExists := func(t *testing.T, path string) {
		t.Helper()
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			require.NoError(t, err)
		}
	}

	oldSSHAddToAgent := cmd.SSHAddToAgentFunc
	cmd.SSHAddToAgentFunc = func(accountName string) error {
		return nil
	}
	t.Cleanup(func() { cmd.SSHAddToAgentFunc = oldSSHAddToAgent })

	tests := []struct {
		name             string
		args             []string
		setup            func(t *testing.T)
		expectedCommands [][]string
		expectError      bool
		errMsg           string
	}{
		{
			name: "Switch to existing account (global, prefers ed25519)",
			args: []string{"use", "test-account"},
			setup: func(t *testing.T) {
				writeConfig(t, true)
				removeIfExists(t, ed25519KeyPath)
				removeIfExists(t, rsaKeyPath)
				require.NoError(t, os.WriteFile(ed25519KeyPath, []byte("dummy ed key"), 0o600))
				require.NoError(t, os.WriteFile(rsaKeyPath, []byte("dummy rsa key"), 0o600))
			},
			expectedCommands: [][]string{
				{"config", "--global", "user.name", baseAccount.Name},
				{"config", "--global", "user.email", baseAccount.Email},
				{"config", "--global", "url.ssh://git@github.com/.insteadOf", "https://github.com/"},
				{"config", "--global", "push.default", "current"},
				{"config", "--global", "core.sshCommand", fmt.Sprintf("ssh -i %s -F /dev/null", ed25519KeyPath)},
			},
		},
		{
			name: "Switch to existing account with local config prefers ed25519",
			args: []string{"use", "test-account", "--local"},
			setup: func(t *testing.T) {
				writeConfig(t, true)
				removeIfExists(t, ed25519KeyPath)
				removeIfExists(t, rsaKeyPath)
				require.NoError(t, os.WriteFile(ed25519KeyPath, []byte("dummy ed key"), 0o600))
				require.NoError(t, os.WriteFile(rsaKeyPath, []byte("dummy rsa key"), 0o600))
			},
			expectedCommands: [][]string{
				{"config", "user.name", baseAccount.Name},
				{"config", "user.email", baseAccount.Email},
				{"config", "push.default", "current"},
				{"config", "core.sshCommand", fmt.Sprintf("ssh -i %s -F /dev/null", ed25519KeyPath)},
			},
		},
		{
			name: "Switch to existing account falls back to RSA key",
			args: []string{"use", "test-account"},
			setup: func(t *testing.T) {
				writeConfig(t, true)
				removeIfExists(t, ed25519KeyPath)
				removeIfExists(t, rsaKeyPath)
				require.NoError(t, os.WriteFile(rsaKeyPath, []byte("dummy rsa key"), 0o600))
			},
			expectedCommands: [][]string{
				{"config", "--global", "user.name", baseAccount.Name},
				{"config", "--global", "user.email", baseAccount.Email},
				{"config", "--global", "url.ssh://git@github.com/.insteadOf", "https://github.com/"},
				{"config", "--global", "push.default", "current"},
				{"config", "--global", "core.sshCommand", fmt.Sprintf("ssh -i %s -F /dev/null", rsaKeyPath)},
			},
		},
		{
			name: "Non-existent account",
			args: []string{"use", "nonexistent"},
			setup: func(t *testing.T) {
				writeConfig(t, false)
				removeIfExists(t, ed25519KeyPath)
				removeIfExists(t, rsaKeyPath)
			},
			expectError: true,
			errMsg:      "account 'nonexistent' does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			oldDir, err := os.Getwd()
			require.NoError(t, err)
			tempDir := t.TempDir()
			require.NoError(t, os.Chdir(tempDir))
			t.Cleanup(func() { _ = os.Chdir(oldDir) })

			viper.Reset()
			viper.SetConfigFile(testConfigPath)
			require.NoError(t, viper.ReadInConfig())

			tuc.gitCommands = nil

			useCobraCmd, _, err := cmd.RootCmd.Find([]string{"use"})
			require.NoError(t, err)
			require.NoError(t, useCobraCmd.Flags().Set("local", "false"))

			cmd.RootCmd.SetArgs(tt.args)
			err = cmd.RootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err, "Expected an error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
				assert.Empty(t, tuc.gitCommands, "Expected no git commands")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expectedCommands, tuc.gitCommands, "Unexpected git commands executed")
			}
		})
	}
}
