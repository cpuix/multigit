package multigit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfigFromFile tests the LoadConfigFromFile function
func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	t.Run("Load valid config file", func(t *testing.T) {
		// Create a test config file
		configPath := filepath.Join(tempDir, "valid_config.json")
		testConfig := multigit.Config{
			Accounts: map[string]multigit.Account{
				"github": {
					Name:  "github",
					Email: "test@example.com",
				},
			},
			ActiveAccount: "github",
			Profiles: map[string]multigit.Profile{
				"default": {
					Name: "default",
					Accounts: map[string]bool{
						"github": true,
					},
				},
			},
			ActiveProfile: "default",
		}

		// Write the config to file
		configData, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err, "Failed to marshal test config")
		err = os.WriteFile(configPath, configData, 0600)
		require.NoError(t, err, "Failed to write test config file")

		// Load the config
		loadedConfig, err := multigit.LoadConfigFromFile(configPath)
		require.NoError(t, err, "LoadConfigFromFile should not return an error for valid config")
		require.NotNil(t, loadedConfig, "Loaded config should not be nil")

		// Verify the loaded config matches the test config
		assert.Equal(t, testConfig.ActiveAccount, loadedConfig.ActiveAccount)
		assert.Equal(t, testConfig.ActiveProfile, loadedConfig.ActiveProfile)
		assert.Equal(t, testConfig.Accounts["github"].Name, loadedConfig.Accounts["github"].Name)
		assert.Equal(t, testConfig.Accounts["github"].Email, loadedConfig.Accounts["github"].Email)
		assert.Equal(t, testConfig.Profiles["default"].Name, loadedConfig.Profiles["default"].Name)
		assert.Equal(t, testConfig.Profiles["default"].Accounts["github"], loadedConfig.Profiles["default"].Accounts["github"])
	})

	t.Run("Load non-existent config file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non_existent.json")
		
		// Ensure the file doesn't exist
		_, err := os.Stat(nonExistentPath)
		require.True(t, os.IsNotExist(err), "Test file should not exist")

		// Try to load the non-existent config
		loadedConfig, err := multigit.LoadConfigFromFile(nonExistentPath)
		require.Error(t, err, "LoadConfigFromFile should return an error for non-existent file")
		require.Nil(t, loadedConfig, "Loaded config should be nil for non-existent file")
	})

	t.Run("Load invalid JSON config file", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid_config.json")
		
		// Write invalid JSON to the file
		err := os.WriteFile(invalidConfigPath, []byte("this is not valid JSON"), 0600)
		require.NoError(t, err, "Failed to write invalid config file")

		// Try to load the invalid config
		loadedConfig, err := multigit.LoadConfigFromFile(invalidConfigPath)
		require.Error(t, err, "LoadConfigFromFile should return an error for invalid JSON")
		require.Nil(t, loadedConfig, "Loaded config should be nil for invalid JSON")
	})
}

// TestSaveConfigToFile tests the SaveConfigToFile function
func TestSaveConfigToFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	t.Run("Save config to new file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "new_config.json")
		
		// Create a test config
		testConfig := multigit.Config{
			Accounts: map[string]multigit.Account{
				"github": {
					Name:  "github",
					Email: "test@example.com",
				},
			},
			ActiveAccount: "github",
			Profiles: map[string]multigit.Profile{
				"default": {
					Name: "default",
					Accounts: map[string]bool{
						"github": true,
					},
				},
			},
			ActiveProfile: "default",
		}

		// Save the config
		err := multigit.SaveConfigToFile(testConfig, configPath)
		require.NoError(t, err, "SaveConfigToFile should not return an error")

		// Verify the file was created
		_, err = os.Stat(configPath)
		require.NoError(t, err, "Config file should exist after saving")

		// Read the file and verify its contents
		data, err := os.ReadFile(configPath)
		require.NoError(t, err, "Failed to read saved config file")

		var savedConfig multigit.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err, "Failed to unmarshal saved config")

		assert.Equal(t, testConfig.ActiveAccount, savedConfig.ActiveAccount)
		assert.Equal(t, testConfig.ActiveProfile, savedConfig.ActiveProfile)
		assert.Equal(t, testConfig.Accounts["github"].Name, savedConfig.Accounts["github"].Name)
		assert.Equal(t, testConfig.Accounts["github"].Email, savedConfig.Accounts["github"].Email)
		assert.Equal(t, testConfig.Profiles["default"].Name, savedConfig.Profiles["default"].Name)
		assert.Equal(t, testConfig.Profiles["default"].Accounts["github"], savedConfig.Profiles["default"].Accounts["github"])
	})

	t.Run("Save config with invalid active references", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "invalid_refs_config.json")
		
		// Create a test config with invalid active references
		testConfig := multigit.Config{
			Accounts:      map[string]multigit.Account{},
			ActiveAccount: "non_existent_account",
			Profiles:      map[string]multigit.Profile{},
			ActiveProfile: "non_existent_profile",
		}

		// Save the config
		err := multigit.SaveConfigToFile(testConfig, configPath)
		require.NoError(t, err, "SaveConfigToFile should not return an error")

		// Read the file and verify its contents
		data, err := os.ReadFile(configPath)
		require.NoError(t, err, "Failed to read saved config file")

		var savedConfig multigit.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err, "Failed to unmarshal saved config")

		// The active references should be cleared
		assert.Empty(t, savedConfig.ActiveAccount, "ActiveAccount should be cleared for non-existent account")
		assert.Empty(t, savedConfig.ActiveProfile, "ActiveProfile should be cleared for non-existent profile")
	})

	t.Run("Save config to directory with no write permission", func(t *testing.T) {
		// Skip this test on Windows as permissions work differently
		if os.Getenv("OS") == "Windows_NT" {
			t.Skip("Skipping permission test on Windows")
		}

		// Create a directory with no write permission
		noWriteDir := filepath.Join(tempDir, "no_write_dir")
		err := os.MkdirAll(noWriteDir, 0500) // read and execute, but no write
		require.NoError(t, err, "Failed to create directory with no write permission")

		// Try to save config to a file in the no-write directory
		configPath := filepath.Join(noWriteDir, "config.json")
		testConfig := multigit.Config{
			Accounts: map[string]multigit.Account{},
			Profiles: map[string]multigit.Profile{},
		}

		// This should fail because we can't write to the directory
		err = multigit.SaveConfigToFile(testConfig, configPath)
		require.Error(t, err, "SaveConfigToFile should return an error for directory with no write permission")
	})
}

// TestLoadConfig tests the LoadConfig function
// Note: This test is more complex because LoadConfig uses the user's home directory
// We'll need to mock or override this behavior
func TestLoadConfig(t *testing.T) {
	// Save the original HOME environment variable
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create a temporary directory to use as HOME
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)

	// Create the config directory
	configDir := filepath.Join(tempHome, ".config", "multigit")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err, "Failed to create config directory")

	// Create a test config file
	configPath := filepath.Join(configDir, "config.json")
	testConfig := multigit.Config{
		Accounts: map[string]multigit.Account{
			"github": {
				Name:  "github",
				Email: "test@example.com",
			},
		},
		ActiveAccount: "github",
		Profiles: map[string]multigit.Profile{
			"default": {
				Name: "default",
				Accounts: map[string]bool{
					"github": true,
				},
			},
		},
		ActiveProfile: "default",
	}

	// Write the config to file
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal test config")
	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err, "Failed to write test config file")

	// Load the config
	loadedConfig := multigit.LoadConfig()

	// Verify the loaded config matches the test config
	assert.Equal(t, testConfig.ActiveAccount, loadedConfig.ActiveAccount)
	assert.Equal(t, testConfig.ActiveProfile, loadedConfig.ActiveProfile)
	assert.Equal(t, testConfig.Accounts["github"].Name, loadedConfig.Accounts["github"].Name)
	assert.Equal(t, testConfig.Accounts["github"].Email, loadedConfig.Accounts["github"].Email)
	assert.Equal(t, testConfig.Profiles["default"].Name, loadedConfig.Profiles["default"].Name)
	assert.Equal(t, testConfig.Profiles["default"].Accounts["github"], loadedConfig.Profiles["default"].Accounts["github"])

	// Test loading non-existent config
	t.Run("Load non-existent config", func(t *testing.T) {
		// Create a new temporary home directory
		newTempHome := t.TempDir()
		os.Setenv("HOME", newTempHome)

		// Load the config (should return a new empty config)
		loadedConfig := multigit.LoadConfig()

		// Verify it's a new empty config
		assert.Empty(t, loadedConfig.ActiveAccount)
		assert.Empty(t, loadedConfig.ActiveProfile)
		assert.NotNil(t, loadedConfig.Accounts)
		assert.NotNil(t, loadedConfig.Profiles)
		assert.Empty(t, loadedConfig.Accounts)
		assert.Empty(t, loadedConfig.Profiles)
	})

	// Test loading invalid JSON config
	t.Run("Load invalid JSON config", func(t *testing.T) {
		// Create a new temporary home directory
		newTempHome := t.TempDir()
		os.Setenv("HOME", newTempHome)

		// Create the config directory
		newConfigDir := filepath.Join(newTempHome, ".config", "multigit")
		err := os.MkdirAll(newConfigDir, 0700)
		require.NoError(t, err, "Failed to create config directory")

		// Create an invalid config file
		newConfigPath := filepath.Join(newConfigDir, "config.json")
		err = os.WriteFile(newConfigPath, []byte("this is not valid JSON"), 0600)
		require.NoError(t, err, "Failed to write invalid config file")

		// Load the config (should return a new empty config)
		loadedConfig := multigit.LoadConfig()

		// Verify it's a new empty config
		assert.Empty(t, loadedConfig.ActiveAccount)
		assert.Empty(t, loadedConfig.ActiveProfile)
		assert.NotNil(t, loadedConfig.Accounts)
		assert.NotNil(t, loadedConfig.Profiles)
		assert.Empty(t, loadedConfig.Accounts)
		assert.Empty(t, loadedConfig.Profiles)
	})
}

// TestSaveConfig tests the SaveConfig function
func TestSaveConfig(t *testing.T) {
	// Save the original HOME environment variable
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create a temporary directory to use as HOME
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)

	// Create a test config
	testConfig := multigit.Config{
		Accounts: map[string]multigit.Account{
			"github": {
				Name:  "github",
				Email: "test@example.com",
			},
		},
		ActiveAccount: "github",
		Profiles: map[string]multigit.Profile{
			"default": {
				Name: "default",
				Accounts: map[string]bool{
					"github": true,
				},
			},
		},
		ActiveProfile: "default",
	}

	// Save the config
	err := multigit.SaveConfig(testConfig)
	require.NoError(t, err, "SaveConfig should not return an error")

	// Verify the config file was created
	configPath := filepath.Join(tempHome, ".config", "multigit", "config.json")
	_, err = os.Stat(configPath)
	require.NoError(t, err, "Config file should exist after saving")

	// Read the file and verify its contents
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "Failed to read saved config file")

	var savedConfig multigit.Config
	err = json.Unmarshal(data, &savedConfig)
	require.NoError(t, err, "Failed to unmarshal saved config")

	assert.Equal(t, testConfig.ActiveAccount, savedConfig.ActiveAccount)
	assert.Equal(t, testConfig.ActiveProfile, savedConfig.ActiveProfile)
	assert.Equal(t, testConfig.Accounts["github"].Name, savedConfig.Accounts["github"].Name)
	assert.Equal(t, testConfig.Accounts["github"].Email, savedConfig.Accounts["github"].Email)
	assert.Equal(t, testConfig.Profiles["default"].Name, savedConfig.Profiles["default"].Name)
	assert.Equal(t, testConfig.Profiles["default"].Accounts["github"], savedConfig.Profiles["default"].Accounts["github"])
}
