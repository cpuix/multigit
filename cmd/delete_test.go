package cmd_test

import (
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
	cfgFile := filepath.Join(tempDir, "config.json")
	
	// Set config file path
	oldCfg := cmd.CfgFile
	cmd.CfgFile = cfgFile
	
	// Initialize config
	cmd.InitConfig()
	
	// Setup test account
	config := multigit.LoadConfig()
	setupTestAccount(t, &config)
	require.NoError(t, multigit.SaveConfig(config), "Failed to save test config")
	
	// Return cleanup function
	return cfgFile, func() {
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
