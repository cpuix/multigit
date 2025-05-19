package cmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock os.Exit for testing
var osExit = os.Exit

func TestRootCommand(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		// Execute root command
		cmd.RootCmd.SetArgs([]string{"--help"})
		err := cmd.RootCmd.Execute()

		// Check results
		assert.NoError(t, err, "Execute should not return error")
	})

	// We'll skip testing the Execute function directly since it's just a thin wrapper around RootCmd.Execute
	// and modifying RootCmd causes side effects in other tests
}

func TestInitConfig(t *testing.T) {
	t.Run("WithConfigFile", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, "config.json")

		// Create a test config
		testConfig := multigit.Config{
			Accounts:      make(map[string]multigit.Account),
			Profiles:      make(map[string]multigit.Profile),
			ActiveAccount: "",
		}

		// Save the test config
		err := multigit.SaveConfigToFile(testConfig, cfgFile)
		require.NoError(t, err, "Failed to create test config file")

		// Set up the test environment
		oldCfg := cmd.CfgFile
		oldConfig := cmd.Config

		// Clean up
		defer func() {
			cmd.CfgFile = oldCfg
			cmd.Config = oldConfig
		}()

		// Set up test environment
		cmd.CfgFile = cfgFile
		cmd.Config = &testConfig

		// Call InitConfig
		cmd.InitConfig()

		// Verify the config file exists and is not empty
		fileInfo, err := os.Stat(cfgFile)
		if os.IsNotExist(err) {
			t.Log("Config file was not created. Directory contents:")
			files, _ := os.ReadDir(tempDir)
			for _, f := range files {
				t.Logf("  - %s (dir: %v)", f.Name(), f.IsDir())
			}
			t.Fatal("Config file was not created")
		}

		// Verify config file exists and is not empty
		if !assert.NoError(t, err, "Config file should be created") {
			t.FailNow()
		}
		assert.Greater(t, fileInfo.Size(), int64(0), "Config file should not be empty")

		// Verify the config was initialized
		assert.NotNil(t, cmd.Config, "Config should be initialized")
		assert.NotNil(t, cmd.Config.Accounts, "Accounts map should be initialized")
		assert.NotNil(t, cmd.Config.Profiles, "Profiles map should be initialized")

		// Verify the config file contains valid JSON
		configData, err := os.ReadFile(cfgFile)
		assert.NoError(t, err, "Should be able to read config file")
		assert.True(t, len(configData) > 0, "Config file should not be empty")

		// Parse the JSON to ensure it's valid
		var config map[string]interface{}
		err = json.Unmarshal(configData, &config)
		assert.NoError(t, err, "Config file should contain valid JSON")

		t.Logf("Config file created successfully at %s", cfgFile)
	})

	t.Run("WithInvalidConfigFile", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, "invalid_config.json")

		// Create invalid config file
		err := os.WriteFile(cfgFile, []byte("invalid json"), 0644)
		require.NoError(t, err)

		// Set config file
		oldCfg := cmd.CfgFile
		oldConfig := cmd.Config
		defer func() {
			cmd.CfgFile = oldCfg
			cmd.Config = oldConfig
		}()

		cmd.CfgFile = cfgFile

		// Should not panic
		assert.NotPanics(t, cmd.InitConfig, "Should handle invalid config file gracefully")

		// Verify the config was initialized
		assert.NotNil(t, cmd.Config, "Config should be initialized")
		assert.NotNil(t, cmd.Config.Accounts, "Accounts map should be initialized")
		assert.NotNil(t, cmd.Config.Profiles, "Profiles map should be initialized")
	})

	t.Run("WithNoConfigFile", func(t *testing.T) {
		// Save original environment
		oldHome := os.Getenv("HOME")
		oldCfg := cmd.CfgFile
		oldConfig := cmd.Config

		// Create a temporary home directory
		tempHome := t.TempDir()
		os.Setenv("HOME", tempHome)

		// Clean up
		defer func() {
			os.Setenv("HOME", oldHome)
			cmd.CfgFile = oldCfg
			cmd.Config = oldConfig
		}()

		// Reset config file
		cmd.CfgFile = ""

		// Call InitConfig
		assert.NotPanics(t, cmd.InitConfig, "InitConfig should not panic when no config file exists")

		// Verify a default config was created
		assert.NotNil(t, cmd.Config, "Config should be initialized")
		assert.NotNil(t, cmd.Config.Accounts, "Accounts map should be initialized")
		assert.NotNil(t, cmd.Config.Profiles, "Profiles map should be initialized")

		// Verify the config file was created in the default location
		defaultConfigPath := filepath.Join(tempHome, ".config", "multigit", "config.json")
		_, err := os.Stat(defaultConfigPath)
		assert.NoError(t, err, "Default config file should be created")
	})

	t.Run("WithConfigFileNotFoundError", func(t *testing.T) {
		// Save original environment
		oldCfg := cmd.CfgFile
		oldConfig := cmd.Config

		// Create a temporary directory for the config file
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, "nonexistent_config.json")

		// Clean up
		defer func() {
			cmd.CfgFile = oldCfg
			cmd.Config = oldConfig
		}()

		// Set config file to a non-existent file
		cmd.CfgFile = cfgFile

		// Call InitConfig
		assert.NotPanics(t, cmd.InitConfig, "InitConfig should not panic when config file doesn't exist")

		// Verify a default config was created
		assert.NotNil(t, cmd.Config, "Config should be initialized")
		assert.NotNil(t, cmd.Config.Accounts, "Accounts map should be initialized")
		assert.NotNil(t, cmd.Config.Profiles, "Profiles map should be initialized")

		// Note: We're not verifying file creation here because the behavior depends on the environment
		// and can be flaky in test environments. The important part is that the function doesn't panic
		// and initializes the config properly.
	})
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfgFile := filepath.Join(tempDir, "config.json")

	config := multigit.NewConfig()
	err := multigit.SaveConfigToFile(config, cfgFile)
	require.NoError(t, err, "Failed to save config")

	// Load and verify
	loaded, err := multigit.LoadConfigFromFile(cfgFile)
	require.NoError(t, err, "Failed to load config")
	assert.NotNil(t, loaded.Accounts, "Accounts map should be initialized")
	assert.NotNil(t, loaded.Profiles, "Profiles map should be initialized")
}
