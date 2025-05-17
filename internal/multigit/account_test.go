package multigit_test

import (
	"os"
	"testing"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountManagement(t *testing.T) {
	tempDir := t.TempDir()
	
	// Override the config directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Test account details
	accountName := "test-account"
	accountEmail := "test@example.com"

	t.Run("CreateAndLoadConfig", func(t *testing.T) {
		// Create a new config
		config := multigit.Config{
			Accounts: map[string]multigit.Account{
				accountName: {
					Name:  accountName,
					Email: accountEmail,
				},
			},
			ActiveAccount: accountName,
		}

		// Save config
		err := multigit.SaveConfig(config)
		require.NoError(t, err, "Failed to save config")

		// Load config
		loadedConfig := multigit.LoadConfig()
		assert.Equal(t, config, loadedConfig, "Loaded config should match saved config")
	})

	t.Run("GetActiveAccount", func(t *testing.T) {
		// Set up test config
		config := multigit.Config{
			Accounts: map[string]multigit.Account{
				accountName: {
					Name:  accountName,
					Email: accountEmail,
				},
			},
			ActiveAccount: accountName,
		}
		multigit.SaveConfig(config)

		// Test getting active account
		name, account, err := multigit.GetActiveAccount()
		require.NoError(t, err, "Failed to get active account")
		assert.Equal(t, accountName, name, "Active account name should match")
		assert.Equal(t, accountEmail, account.Email, "Active account email should match")
	})

	t.Run("GetActiveAccount_NoneActive", func(t *testing.T) {
		// Set up test config with no active account
		config := multigit.Config{
			Accounts: map[string]multigit.Account{
				accountName: {
					Name:  accountName,
					Email: accountEmail,
				},
			},
			ActiveAccount: "",
		}
		multigit.SaveConfig(config)

		// Test getting active account when none is set
		_, _, err := multigit.GetActiveAccount()
		assert.Error(t, err, "Should return error when no active account")
	})
}
