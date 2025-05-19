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

// TestFullWorkflow tests the full workflow from creating an account to using it
func TestFullWorkflow(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "multigit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "Failed to create config directory")
	configPath := filepath.Join(configDir, "config.json")

	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	err = os.MkdirAll(sshDir, 0700)
	require.NoError(t, err, "Failed to create .ssh directory")

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

	// Create a mock SSH implementation
	mockSSH := new(MockSSH)
	
	// Save original SSH client
	oldSSHClient := multigit.SSHClient
	// Replace with our mock
	multigit.SSHClient = mockSSH
	// Restore original client after test
	defer func() { multigit.SSHClient = oldSSHClient }()

	// Initialize config
	cmd.InitConfig()

	// Setup mock expectations for create command
	mockSSH.On("CreateSSHKey", "test-account", "test@example.com", "", ssh.KeyTypeED25519).Return(nil)
	mockSSH.On("AddSSHKeyToAgent", "test-account").Return(nil)
	mockSSH.On("AddSSHConfigEntry", "test-account").Return(nil)

	// Step 1: Create a new account
	t.Run("Step1_CreateAccount", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute create command
		cmd.RootCmd.SetArgs([]string{"create", "test-account", "test@example.com"})
		err := cmd.RootCmd.Execute()
		require.NoError(t, err, "Create command failed")

		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify that the account was created in config
		config := multigit.LoadConfig()
		account, exists := config.Accounts["test-account"]
		assert.True(t, exists, "Account should exist in config")
		assert.Equal(t, "test@example.com", account.Email, "Account email should match")
	})

	// For this integration test, we'll skip the actual git command execution
	// by relying on the fact that our test environment doesn't have git configured
	// This is a simplified approach for testing the command interaction flow

	// Step 2: Use the created account
	t.Run("Step2_UseAccount", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute use command
		cmd.RootCmd.SetArgs([]string{"use", "test-account"})
		err := cmd.RootCmd.Execute()
		require.NoError(t, err, "Use command failed")

		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify that the account is active in config
		config := multigit.LoadConfig()
		assert.Equal(t, "test-account", config.ActiveAccount, "Account should be active")
	})

	// Step 3: Check status
	t.Run("Step3_CheckStatus", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute status command
		cmd.RootCmd.SetArgs([]string{"status"})
		err := cmd.RootCmd.Execute()
		require.NoError(t, err, "Status command failed")

		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify that the account is active in config
		config := multigit.LoadConfig()
		assert.Equal(t, "test-account", config.ActiveAccount, "Account should be active")
		account, exists := config.Accounts["test-account"]
		assert.True(t, exists, "Account should exist in config")
		assert.Equal(t, "test@example.com", account.Email, "Account email should match")
	})

	// Step 4: Delete the account
	t.Run("Step4_DeleteAccount", func(t *testing.T) {
		// Setup mock expectations for delete command
		mockSSH.On("DeleteSSHKey", "test-account").Return(nil)
		mockSSH.On("RemoveSSHConfigEntry", "test-account").Return(nil)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute delete command with force flag
		cmd.RootCmd.SetArgs([]string{"delete", "test-account", "-f"})
		err := cmd.RootCmd.Execute()
		require.NoError(t, err, "Delete command failed")

		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify account was deleted from config
		config := multigit.LoadConfig()
		_, exists := config.Accounts["test-account"]
		assert.False(t, exists, "Account should be deleted from config")
		assert.Equal(t, "", config.ActiveAccount, "No account should be active")
	})

	// Verify all expected calls were made
	mockSSH.AssertExpectations(t)
}

// TestErrorConditions tests various error conditions and edge cases
func TestErrorConditions(t *testing.T) {
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

	// Create a mock SSH implementation
	mockSSH := new(MockSSH)
	
	// Save original SSH client
	oldSSHClient := multigit.SSHClient
	// Replace with our mock
	multigit.SSHClient = mockSSH
	// Restore original client after test
	defer func() { multigit.SSHClient = oldSSHClient }()

	// Initialize config
	cmd.InitConfig()

	// Test case 1: Create account with invalid email
	t.Run("CreateAccountWithInvalidEmail", func(t *testing.T) {
		// Initialize config
		cmd.InitConfig()
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute create command with invalid email
		cmd.RootCmd.SetArgs([]string{"create", "test-account", "invalid-email"})
		err := cmd.RootCmd.Execute()
		
		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify error
		assert.Error(t, err, "Command should fail with invalid email")
		// The actual error might be different depending on the validation implementation
		// Just check that there's an error and it's related to the email
		assert.Contains(t, err.Error(), "email", "Error should mention email")
	})

	// Test case 2: Use non-existent account
	t.Run("UseNonExistentAccount", func(t *testing.T) {
		// Initialize config
		cmd.InitConfig()
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute use command with non-existent account
		cmd.RootCmd.SetArgs([]string{"use", "non-existent-account"})
		err := cmd.RootCmd.Execute()
		
		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify error
		assert.Error(t, err, "Command should fail with non-existent account")
		// Check the error message directly
		assert.Contains(t, err.Error(), "does not exist", "Error should indicate account doesn't exist")
	})

	// Test case 3: Delete non-existent account
	t.Run("DeleteNonExistentAccount", func(t *testing.T) {
		// Initialize config
		cmd.InitConfig()
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Execute delete command with non-existent account
		cmd.RootCmd.SetArgs([]string{"delete", "non-existent-account", "-f"})
		err := cmd.RootCmd.Execute()
		
		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify error
		assert.Error(t, err, "Command should fail with non-existent account")
		// Check the error message directly
		assert.Contains(t, err.Error(), "does not exist", "Error should indicate account doesn't exist")
	})

	// Test case 4: Create duplicate account
	t.Run("CreateDuplicateAccount", func(t *testing.T) {
		// Initialize config
		cmd.InitConfig()
		
		// Setup mock expectations for first create
		mockSSH.On("CreateSSHKey", "duplicate-account", "test@example.com", "", ssh.KeyTypeED25519).Return(nil)
		mockSSH.On("AddSSHKeyToAgent", "duplicate-account").Return(nil)
		mockSSH.On("AddSSHConfigEntry", "duplicate-account").Return(nil)

		// Create account first
		cmd.RootCmd.SetArgs([]string{"create", "duplicate-account", "test@example.com"})
		err := cmd.RootCmd.Execute()
		require.NoError(t, err, "First create should succeed")

		// Capture stdout for second create
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		// Try to create duplicate account
		cmd.RootCmd.SetArgs([]string{"create", "duplicate-account", "another@example.com"})
		err = cmd.RootCmd.Execute()
		
		// Read command output (for debugging purposes)
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		_ = buf.String() // Discard output, we're checking the error message directly

		// Verify error
		assert.Error(t, err, "Command should fail with duplicate account")
		// Check the error message directly
		assert.Contains(t, err.Error(), "already exists", "Error should indicate account already exists")
	})

	// Verify all expected calls were made
	mockSSH.AssertExpectations(t)
}
