package multigit_test

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Alias the package for easier access to unexported types
type (
	Account = multigit.Account
	Profile = multigit.Profile
	Config  = multigit.Config
)

// MockSSH is a mock implementation of the SSHOperations interface
type MockSSH struct {
	mock.Mock
}

func (m *MockSSH) CreateSSHKey(accountName, email, passphrase string) error {
	args := m.Called(accountName, email, passphrase)
	return args.Error(0)
}

func (m *MockSSH) AddSSHKeyToAgent(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}

func (m *MockSSH) AddSSHConfigEntry(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}

func (m *MockSSH) DeleteSSHKey(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}

func (m *MockSSH) RemoveSSHConfigEntry(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}

// setupTestEnvironment sets up a test environment with a temporary directory
// and returns the path to the temp directory, the config path, and a cleanup function
func setupTestEnvironment(t *testing.T) (string, string, func()) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Set up the test environment variables
	oldHome := os.Getenv("HOME")
	oldConfigPath := os.Getenv("MULTIGIT_CONFIG")

	// Set the HOME directory to the temp directory
	os.Setenv("HOME", tempDir)

	// Create the config directory
	configDir := filepath.Join(tempDir, ".multigit")
	err := os.MkdirAll(configDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Set the config path
	configPath := filepath.Join(configDir, "config.json")
	os.Setenv("MULTIGIT_CONFIG", configPath)

	// Create an empty config file
	config := Config{
		Accounts: make(map[string]Account),
		Profiles: make(map[string]Profile),
	}

	// Write the initial config
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Return the tempDir, configPath, and cleanup function
	return tempDir, configPath, func() {
		// Restore the original environment variables
		os.Setenv("HOME", oldHome)
		if oldConfigPath == "" {
			os.Unsetenv("MULTIGIT_CONFIG")
		} else {
			os.Setenv("MULTIGIT_CONFIG", oldConfigPath)
		}
	}
}

// testConfig holds the in-memory config for testing
var testConfig *Config

// saveTestConfig is a helper function to save the config in tests
func saveTestConfig(c Config) error {
	log.Printf("saveTestConfig called with config: %+v", c)
	
	// Create a deep copy to avoid modifying the original
	configCopy := Config{
		Accounts:      make(map[string]Account),
		ActiveAccount: c.ActiveAccount,
		Profiles:      make(map[string]Profile),
		ActiveProfile: c.ActiveProfile,
	}

	// Copy accounts
	for k, v := range c.Accounts {
		log.Printf("Copying account %s: %+v", k, v)
		configCopy.Accounts[k] = v
	}

	// Copy profiles
	for k, v := range c.Profiles {
		log.Printf("Copying profile %s: %+v", k, v)
		profileCopy := Profile{
			Name:     v.Name,
			Accounts: make(map[string]bool),
		}
		for acc := range v.Accounts {
			log.Printf("Adding account %s to profile %s", acc, k)
			profileCopy.Accounts[acc] = true
		}
		configCopy.Profiles[k] = profileCopy
	}

	testConfig = &configCopy
	log.Printf("saveTestConfig: Updated testConfig to: %+v", *testConfig)
	log.Printf("testConfig accounts: %+v", testConfig.Accounts)
	return nil
}

// loadTestConfig is a helper function to load the config in tests
func loadTestConfig() Config {
	if testConfig == nil {
		log.Println("loadTestConfig: testConfig is nil, returning empty config")
		return Config{
			Accounts: make(map[string]Account),
			Profiles: make(map[string]Profile),
		}
	}
	log.Printf("loadTestConfig: returning config with %d accounts: %+v", len(testConfig.Accounts), *testConfig)
	return *testConfig
}

// withMockedConfig runs the test function with a clean test environment
func withMockedConfig(t *testing.T, testFunc func()) {
	t.Logf("Entering withMockedConfig")
	// Save original save function
	oldSaveConfig := multigit.SaveConfig

	// Set up test config if it's nil
	if testConfig == nil {
		testConfig = &multigit.Config{
			Accounts: make(map[string]multigit.Account),
			Profiles: make(map[string]multigit.Profile),
		}
		t.Logf("Initialized new testConfig in withMockedConfig")
	}
	t.Logf("Current testConfig in withMockedConfig: %+v", *testConfig)

	// Set up mock save function that updates our test config
	multigit.SaveConfig = func(c multigit.Config) error {
		t.Logf("Mock SaveConfig called with config: %+v", c)
		t.Logf("Before update - testConfig: %+v", *testConfig)
		*testConfig = c
		t.Logf("After update - testConfig: %+v", *testConfig)
		t.Logf("testConfig accounts after update: %+v", testConfig.Accounts)
		return nil
	}

	// Run the test
	testFunc()
	t.Logf("Exiting withMockedConfig, testConfig: %+v", *testConfig)

	// Restore original save function
	multigit.SaveConfig = oldSaveConfig
}

func TestCreateAccount(t *testing.T) {
	// Initialize test config
	testConfig = &Config{
		Accounts: make(map[string]Account),
		Profiles: make(map[string]Profile),
	}

	tests := []struct {
		name           string
		accountName    string
		email          string
		passphrase     string
		saveConfigFunc func(Config) error
		expectError    bool
		errContains    string
		setupMocks     func(*MockSSH, string, string, string)
		setup          func(*testing.T, *MockSSH)
		skipSSH        bool
	}{
		{
			name:        "Create account successfully",
			accountName: "test-account",
			email:       "test@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			saveConfigFunc: func(c Config) error {
				t.Logf("Custom saveConfigFunc called with config: %+v", c)
				testConfig = &c
				t.Logf("Updated testConfig: %+v", *testConfig)
				return nil
			},
			setupMocks: func(m *MockSSH, accountName, email, passphrase string) {
				m.On("CreateSSHKey", accountName, email, passphrase).Return(nil)
				m.On("AddSSHKeyToAgent", accountName).Return(nil)
				m.On("AddSSHConfigEntry", accountName).Return(nil)
			},
			skipSSH: false,
		},
		{
			name:        "Create account with passphrase",
			accountName: "test-account-pass",
			email:       "test-pass@example.com",
			passphrase:  "securepass",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			saveConfigFunc: func(c Config) error {
				t.Logf("Custom saveConfigFunc called with config: %+v", c)
				testConfig = &c
				t.Logf("Updated testConfig: %+v", *testConfig)
				return nil
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string) {
				m.On("CreateSSHKey", name, email, passphrase).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(nil)
				m.On("AddSSHConfigEntry", name).Return(nil)
			},
			skipSSH: false,
		},
		{
			name:        "Fail on empty account name",
			accountName: "",
			email:       "test@example.com",
			expectError: true,
			errContains: "account name cannot be empty",
			skipSSH:    true,
		},
		{
			name:        "Fail on invalid email",
			accountName: "test-account",
			email:       "invalid-email",
			expectError: true,
			errContains: "invalid email format",
			skipSSH:    true,
		},
		{
			name:        "Fail on SaveConfig error",
			accountName: "test-account-error",
			email:       "error@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			saveConfigFunc: func(c Config) error {
				return errors.New("failed to save config")
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string) {
				m.On("CreateSSHKey", name, email, passphrase).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(nil)
				m.On("AddSSHConfigEntry", name).Return(nil)
			},
			expectError: true,
			errContains: "failed to save config",
			skipSSH: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			_, configPath, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Set the config path for this test
			os.Setenv("MULTIGIT_CONFIG", configPath)

			// Initialize mock SSH
			mockSSH := new(MockSSH)

			// Setup test-specific mocks if not skipped
			if !tt.skipSSH && tt.setupMocks != nil {
				tt.setupMocks(mockSSH, tt.accountName, tt.email, tt.passphrase)
			}

			// Initialize test config before each test case
			testConfig = &Config{
				Accounts: make(map[string]Account),
				Profiles: make(map[string]Profile),
			}
			t.Logf("Initialized testConfig: %+v", *testConfig)

			// Use our mocked config for the test
			withMockedConfig(t, func() {
				// Log the address of the current SaveConfig function
				t.Logf("Original SaveConfig function address: %p", multigit.SaveConfig)
				
				// Override SaveConfig for test
				oldSaveConfig := multigit.SaveConfig
				multigit.SaveConfig = func(c Config) error {
					t.Logf("Mock SaveConfig function called, address: %p", oldSaveConfig)
					if tt.saveConfigFunc != nil {
						return tt.saveConfigFunc(c)
					}
					return saveTestConfig(c)
				}
				t.Logf("New SaveConfig function address: %p", multigit.SaveConfig)
				defer func() { multigit.SaveConfig = oldSaveConfig }()

				// Override SSHClient for test if not skipped
				oldSSHClient := multigit.SSHClient
				if !tt.skipSSH {
					multigit.SSHClient = mockSSH
				}
				defer func() { multigit.SSHClient = oldSSHClient }()

				// Run any additional test setup
				if tt.setup != nil {
					tt.setup(t, mockSSH)
				}

				// Call the function with saveConfigFunc if provided
				var err error
				if tt.saveConfigFunc != nil {
					t.Logf("Creating account with saveConfigFunc")
					err = multigit.CreateAccount(tt.accountName, tt.email, tt.passphrase, tt.saveConfigFunc)
					t.Logf("After CreateAccount with saveConfigFunc, err: %v", err)
				} else {
					t.Logf("Creating account without saveConfigFunc")
					err = multigit.CreateAccount(tt.accountName, tt.email, tt.passphrase, saveTestConfig)
					t.Logf("After CreateAccount with saveTestConfig, err: %v", err)
				}

				// Assertions
				if tt.expectError {
					assert.Error(t, err, "Expected an error but got none")
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains, "Error message should contain the expected text")
					}
				} else {
					assert.NoError(t, err, "Expected no error but got one: %v", err)
					// Verify the account was created in testConfig
					t.Logf("testConfig after save: %+v", *testConfig)
					t.Logf("Looking for account: %s in testConfig", tt.accountName)
					account, exists := testConfig.Accounts[tt.accountName]
					if !exists {
						t.Logf("Accounts in testConfig: %+v", testConfig.Accounts)
					}
					assert.True(t, exists, "Account should exist in testConfig")
					assert.Equal(t, tt.accountName, account.Name, "Account name should match")
					assert.Equal(t, tt.email, account.Email, "Account email should match")
				}

				// Verify all expected mock calls were made if not skipped
				if !tt.skipSSH {
					mockSSH.AssertExpectations(t)
				}
			})
		})
	}
}

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
			Profiles:      make(map[string]multigit.Profile),
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
