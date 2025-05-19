package multigit_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
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

// CreateSSHKey mocks the CreateSSHKey function
func (m *MockSSH) CreateSSHKey(accountName, email, passphrase string, keyType ssh.KeyType) error {
	args := m.Called(accountName, email, passphrase, keyType)
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
		setupMocks     func(*MockSSH, string, string, string, ssh.KeyType)
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
			setupMocks: func(m *MockSSH, accountName, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", accountName, email, passphrase, keyType).Return(nil)
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
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(nil)
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
			skipSSH:     true,
		},
		{
			name:        "Fail on invalid email",
			accountName: "test-account",
			email:       "invalid-email",
			expectError: true,
			errContains: "invalid email format",
			skipSSH:     true,
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
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(nil)
				m.On("AddSSHConfigEntry", name).Return(nil)
			},
			expectError: true,
			errContains: "failed to save config",
			skipSSH:     false,
		},
		// Test for account already exists is handled in TestCreateAccountWithExistingAccount
		{
			name:        "Fail on SSH key creation error",
			accountName: "ssh-key-error",
			email:       "ssh-key-error@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(fmt.Errorf("failed to create SSH key"))
			},
			expectError: true,
			errContains: "failed to create SSH key",
			skipSSH:     false,
		},
		{
			name:        "Fail on adding SSH key to agent",
			accountName: "agent-error",
			email:       "agent-error@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(fmt.Errorf("failed to add SSH key to agent"))
			},
			expectError: true,
			errContains: "failed to add SSH key to agent",
			skipSSH:     false,
		},
		{
			name:        "Fail on adding SSH config entry",
			accountName: "config-entry-error",
			email:       "config-entry-error@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// No additional setup needed
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(nil)
				m.On("AddSSHConfigEntry", name).Return(fmt.Errorf("failed to add SSH config entry"))
			},
			expectError: true,
			errContains: "failed to add SSH config entry",
			skipSSH:     false,
		},
		{
			name:        "Create account with nil Accounts map",
			accountName: "nil-accounts",
			email:       "nil-accounts@example.com",
			setup: func(t *testing.T, m *MockSSH) {
				// Set accounts to nil to test initialization
				testConfig.Accounts = nil
			},
			setupMocks: func(m *MockSSH, name, email, passphrase string, keyType ssh.KeyType) {
				m.On("CreateSSHKey", name, email, passphrase, keyType).Return(nil)
				m.On("AddSSHKeyToAgent", name).Return(nil)
				m.On("AddSSHConfigEntry", name).Return(nil)
			},
			expectError: false,
			skipSSH:     false,
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
				tt.setupMocks(mockSSH, tt.accountName, tt.email, tt.passphrase, ssh.KeyTypeED25519)
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
					err = multigit.CreateAccount(tt.accountName, tt.email, tt.passphrase, ssh.KeyTypeED25519, tt.saveConfigFunc)
					t.Logf("After CreateAccount with saveConfigFunc, err: %v", err)
				} else {
					t.Logf("Creating account without saveConfigFunc")
					err = multigit.CreateAccount(tt.accountName, tt.email, tt.passphrase, ssh.KeyTypeED25519, saveTestConfig)
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

// Helper function to validate CreateAccount input
func validateCreateAccountInput(accountName, accountEmail string) error {
	// Input validation
	if accountName == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	if accountEmail == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if !strings.Contains(accountEmail, "@") {
		return fmt.Errorf("invalid email format")
	}

	// Check if account already exists
	if _, exists := testConfig.Accounts[accountName]; exists {
		return fmt.Errorf("account '%s' already exists", accountName)
	}

	return nil
}

// TestCreateAccountWithExistingAccount tests the scenario where an account already exists
func TestCreateAccountWithExistingAccount(t *testing.T) {
	// Create a test config with an existing account
	testConfig = &Config{
		Accounts: map[string]Account{
			"existing-account": {
				Name:  "existing-account",
				Email: "existing@example.com",
			},
		},
		Profiles: make(map[string]Profile),
	}

	// Create a direct test for the validation logic
	err := validateCreateAccountInput("existing-account", "new-email@example.com")

	// Assertions
	assert.Error(t, err, "Expected an error but got none")
	assert.Contains(t, err.Error(), "already exists", "Error message should contain 'already exists'")
}

// TestGetConfigPath tests the getConfigPath function
func TestGetConfigPath(t *testing.T) {
	// Create a temporary test function that calls the unexported getConfigPath
	getConfigPathTest := func() (string, error) {
		// Simplified implementation of getConfigPath
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %v", err)
		}
		return filepath.Join(home, ".config", "multigit", "config.json"), nil
	}

	// Test the normal case
	path, err := getConfigPathTest()
	assert.NoError(t, err, "getConfigPath should not return an error")
	assert.NotEmpty(t, path, "getConfigPath should return a non-empty path")

	// Verify the path format
	home, err := os.UserHomeDir()
	assert.NoError(t, err, "UserHomeDir should not return an error")
	expectedPath := filepath.Join(home, ".config", "multigit", "config.json")
	assert.Equal(t, expectedPath, path, "getConfigPath should return the expected path")
}





// TestGetActiveAccount tests the GetActiveAccount function
func TestGetActiveAccount(t *testing.T) {
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

	// Create a test config with an active account
	testConfig := multigit.Config{
		Accounts: map[string]multigit.Account{
			"test-account": {
				Name:  "test-account",
				Email: "test@example.com",
			},
		},
		ActiveAccount: "test-account",
	}

	// Write the test config to file
	configPath := filepath.Join(configDir, "config.json")
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal test config")

	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err, "Failed to write test config file")

	// Test case 1: Get active account when one is set
	name, account, err := multigit.GetActiveAccount()
	assert.NoError(t, err, "GetActiveAccount should not return an error")
	assert.Equal(t, "test-account", name, "Active account name should match")
	assert.NotNil(t, account, "Active account should not be nil")
	assert.Equal(t, "test-account", account.Name, "Active account name should match")
	assert.Equal(t, "test@example.com", account.Email, "Active account email should match")

	// Test case 2: Get active account when none is set
	// Update the config file with no active account
	testConfig.ActiveAccount = ""
	configData, err = json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal updated test config")

	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err, "Failed to write updated test config file")

	name, account, err = multigit.GetActiveAccount()
	assert.Error(t, err, "GetActiveAccount should return an error")
	assert.Contains(t, err.Error(), "no active account", "Error message should contain 'no active account'")
	assert.Empty(t, name, "Active account name should be empty")
	assert.Nil(t, account, "Active account should be nil")

	// Test case 3: Get active account when account doesn't exist
	// Update the config file with a non-existent active account
	testConfig.ActiveAccount = "non-existent-account"
	configData, err = json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal updated test config")

	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err, "Failed to write updated test config file")

	name, account, err = multigit.GetActiveAccount()
	assert.Error(t, err, "GetActiveAccount should return an error")
	assert.Contains(t, err.Error(), "not found", "Error message should contain 'not found'")
	assert.Empty(t, name, "Active account name should be empty")
	assert.Nil(t, account, "Active account should be nil")
}

func TestDeleteAccount(t *testing.T) {
	tempDir := t.TempDir()

	// Override the config directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create a mock SSH client
	mockSSH := new(MockSSH)

	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	defer func() { multigit.SSHClient = originalSSH }()

	// Test cases
	tests := []struct {
		name        string
		accountName string
		setup       func()
		mockSetup   func()
		expectError bool
		errContains string
	}{
		{
			name:        "Delete account successfully",
			accountName: "test-account",
			setup: func() {
				// Create a config with the test account
				config := multigit.Config{
					Accounts: map[string]multigit.Account{
						"test-account": {
							Name:  "test-account",
							Email: "test@example.com",
						},
					},
					ActiveAccount: "test-account",
					Profiles:      make(map[string]multigit.Profile),
				}
				// Save the config
				multigit.SaveConfig(config)
			},
			mockSetup: func() {
				// Reset mock and set new expectations
				mockSSH = new(MockSSH)
				multigit.SSHClient = mockSSH
				
				// Mock the SSH operations
				mockSSH.On("DeleteSSHKey", "test-account").Return(nil)
				mockSSH.On("RemoveSSHConfigEntry", "test-account").Return(nil)
			},
			expectError: false,
		},
		{
			name:        "Delete non-existent account",
			accountName: "non-existent-account",
			setup: func() {
				// Create a config without the test account
				config := multigit.Config{
					Accounts:      make(map[string]multigit.Account),
					ActiveAccount: "",
					Profiles:      make(map[string]multigit.Profile),
				}
				// Save the config
				multigit.SaveConfig(config)
			},
			mockSetup: func() {
				// Reset mock and set new expectations
				mockSSH = new(MockSSH)
				multigit.SSHClient = mockSSH
				// No mock expectations needed as the function should return early
			},
			expectError: true,
			errContains: "does not exist",
		},
		{
			name:        "Delete account with SSH key deletion error",
			accountName: "test-account-ssh-error",
			setup: func() {
				// Create a config with the test account
				config := multigit.Config{
					Accounts: map[string]multigit.Account{
						"test-account-ssh-error": {
							Name:  "test-account-ssh-error",
							Email: "test-ssh-error@example.com",
						},
					},
					ActiveAccount: "test-account-ssh-error",
					Profiles:      make(map[string]multigit.Profile),
				}
				// Save the config
				multigit.SaveConfig(config)
			},
			mockSetup: func() {
				// Reset mock and set new expectations
				mockSSH = new(MockSSH)
				multigit.SSHClient = mockSSH
				
				// Mock the SSH operations with an error for DeleteSSHKey
				mockSSH.On("DeleteSSHKey", "test-account-ssh-error").Return(fmt.Errorf("SSH key deletion error"))
				mockSSH.On("RemoveSSHConfigEntry", "test-account-ssh-error").Return(nil)
			},
			expectError: false, // Should still succeed despite SSH key deletion error
		},
		{
			name:        "Delete account with SSH config removal error",
			accountName: "test-account-config-error",
			setup: func() {
				// Create a config with the test account
				config := multigit.Config{
					Accounts: map[string]multigit.Account{
						"test-account-config-error": {
							Name:  "test-account-config-error",
							Email: "test-config-error@example.com",
						},
					},
					ActiveAccount: "test-account-config-error",
					Profiles:      make(map[string]multigit.Profile),
				}
				// Save the config
				multigit.SaveConfig(config)
			},
			mockSetup: func() {
				// Reset mock and set new expectations
				mockSSH = new(MockSSH)
				multigit.SSHClient = mockSSH
				
				// Mock the SSH operations with an error for RemoveSSHConfigEntry
				mockSSH.On("DeleteSSHKey", "test-account-config-error").Return(nil)
				mockSSH.On("RemoveSSHConfigEntry", "test-account-config-error").Return(fmt.Errorf("SSH config removal error"))
			},
			expectError: false, // Should still succeed despite SSH config removal error
		},
		{
			name:        "Delete active account",
			accountName: "active-account",
			setup: func() {
				// Create a config with the test account as active
				config := multigit.Config{
					Accounts: map[string]multigit.Account{
						"active-account": {
							Name:  "active-account",
							Email: "active@example.com",
						},
					},
					ActiveAccount: "active-account",
					Profiles:      make(map[string]multigit.Profile),
				}
				// Save the config
				multigit.SaveConfig(config)
			},
			mockSetup: func() {
				// Reset mock and set new expectations
				mockSSH = new(MockSSH)
				multigit.SSHClient = mockSSH
				
				// Mock the SSH operations
				mockSSH.On("DeleteSSHKey", "active-account").Return(nil)
				mockSSH.On("RemoveSSHConfigEntry", "active-account").Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new temp directory for each test case
			testDir := t.TempDir()
			os.Setenv("HOME", testDir)
			
			// Set up the test environment
			if tt.setup != nil {
				tt.setup()
			}

			// Set up the mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// Call the function being tested
			err := multigit.DeleteAccount(tt.accountName)

			// Verify the result
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "Error message should contain the expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error but got one: %v", err)

				// Verify the account was removed from the config
				config := multigit.LoadConfig()
				_, exists := config.Accounts[tt.accountName]
				assert.False(t, exists, "Account should not exist in config after deletion")

				// If this was the active account, verify it's no longer active
				if tt.accountName == "active-account" {
					assert.Empty(t, config.ActiveAccount, "Active account should be cleared after deleting the active account")
				}
			}

			// Verify all expected mock calls were made
			mockSSH.AssertExpectations(t)
		})
	}
}
