package cmd_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSSH is a mock implementation of the SSH interface
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

func TestCreateAccount(t *testing.T) {
	// Create a mock SSH implementation
	mockSSH := new(MockSSH)

	// Setup test config
	tempDir := t.TempDir()
	cfgFile := filepath.Join(tempDir, "config.json")

	// Set config file path
	oldCfg := cmd.CfgFile
	cmd.CfgFile = cfgFile
	defer func() { cmd.CfgFile = oldCfg }()

	// Initialize config with a new empty config
	config := multigit.NewConfig()
	if err := multigit.SaveConfig(config); err != nil {
		t.Fatalf("Failed to initialize test config: %v", err)
	}

	// Initialize the command
	cmd.InitConfig()

	// Save original SSH client
	oldSSHClient := multigit.SSHClient
	// Replace with our mock
	multigit.SSHClient = mockSSH
	// Restore original client after test
	defer func() { multigit.SSHClient = oldSSHClient }()

	// Test cases
	tests := []struct {
		name        string
		accountName string
		email       string
		passphrase  string
		setup       func()
		mockSetup   func(*MockSSH)
		expectError bool
		errContains string
	}{
		// Success case - new account with no passphrase
		{
			name:        "Add new account",
			accountName: "new-account",
			email:       "new@example.com",
			passphrase:  "",
			setup:       func() {},
			mockSetup: func(m *MockSSH) {
				m.On("CreateSSHKey", "new-account", "new@example.com", "", ssh.KeyTypeED25519).Return(nil)
				m.On("AddSSHKeyToAgent", "new-account").Return(nil)
				m.On("AddSSHConfigEntry", "new-account").Return(nil)
			},
			expectError: false,
		},
		// Success case - new account with passphrase
		{
			name:        "Add new account with passphrase",
			accountName: "secure-account",
			email:       "secure@example.com",
			passphrase:  "my-secure-passphrase",
			setup:       func() {},
			mockSetup: func(m *MockSSH) {
				m.On("CreateSSHKey", "secure-account", "secure@example.com", "my-secure-passphrase", ssh.KeyTypeED25519).Return(nil)
				m.On("AddSSHKeyToAgent", "secure-account").Return(nil)
				m.On("AddSSHConfigEntry", "secure-account").Return(nil)
			},
			expectError: false,
		},
		// Error case - duplicate account
		{
			name:        "Add duplicate account",
			accountName: "test-account",
			email:       "duplicate@example.com",
			setup: func() {
				// Add a test account first
				config := multigit.NewConfig()
				config.Accounts = map[string]multigit.Account{
					"test-account": {
						Name:  "test-account",
						Email: "test@example.com",
					},
				}
				require.NoError(t, multigit.SaveConfig(config), "Failed to setup test account")
			},
			mockSetup:   func(m *MockSSH) {},
			expectError: true,
			errContains: "already exists",
		},
		// Error case - SSH key creation fails
		{
			name:        "SSH key creation fails",
			accountName: "fail-key-account",
			email:       "fail-key@example.com",
			setup:       func() {},
			mockSetup: func(m *MockSSH) {
				m.On("CreateSSHKey", "fail-key-account", "fail-key@example.com", "", ssh.KeyTypeED25519).
					Return(fmt.Errorf("ssh key creation failed"))
			},
			expectError: true,
			errContains: "failed to create SSH key",
		},
		// Error case - Add SSH key to agent fails
		{
			name:        "Add SSH key to agent fails",
			accountName: "fail-agent-account",
			email:       "fail-agent@example.com",
			setup:       func() {},
			mockSetup: func(m *MockSSH) {
				m.On("CreateSSHKey", "fail-agent-account", "fail-agent@example.com", "", ssh.KeyTypeED25519).Return(nil)
				m.On("AddSSHKeyToAgent", "fail-agent-account").
					Return(fmt.Errorf("failed to add to agent"))
			},
			expectError: true,
			errContains: "failed to add SSH key to agent",
		},
		// Error case - Add SSH config entry fails
		{
			name:        "Add SSH config entry fails",
			accountName: "fail-config-account",
			email:       "fail-config@example.com",
			setup:       func() {},
			mockSetup: func(m *MockSSH) {
				m.On("CreateSSHKey", "fail-config-account", "fail-config@example.com", "", ssh.KeyTypeED25519).Return(nil)
				m.On("AddSSHKeyToAgent", "fail-config-account").Return(nil)
				m.On("AddSSHConfigEntry", "fail-config-account").
					Return(fmt.Errorf("failed to add config entry"))
			},
			expectError: true,
			errContains: "failed to add SSH config entry",
		},
		// Error case - Save config fails
		// Note: This test is commented out because we can't easily mock the SaveConfig function
		// without refactoring the code to use dependency injection for the SaveConfig function.
		// To properly test this, we would need to modify the CreateAccount function to accept
		// a saveConfigFunc parameter, similar to how it accepts an SSHClient.
		// {
		// 	name:        "Save config fails",
		// 	accountName: "fail-save-account",
		// 	email:       "fail-save@example.com",
		// 	setup:       func() {},
		// 	mockSetup: func(m *MockSSH) {
		// 		m.On("CreateSSHKey", "fail-save-account", "fail-save@example.com", "").Return(nil)
		// 		m.On("AddSSHKeyToAgent", "fail-save-account").Return(nil)
		// 		m.On("AddSSHConfigEntry", "fail-save-account").Return(nil)
		// 	},
		// 	expectError: true,
		// 	errContains: "failed to save config",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockSSH.ExpectedCalls = nil
			mockSSH.Calls = nil

			// Setup test
			tt.setup()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSSH)
			}

			// Replace the SSH client with our mock for this test
			oldSSHClient := multigit.SSHClient
			multigit.SSHClient = mockSSH
			// Restore the original client after the test
			defer func() { multigit.SSHClient = oldSSHClient }()

			// Call the function
			err := multigit.CreateAccount(tt.accountName, tt.email, tt.passphrase, ssh.KeyTypeED25519)

			// Verify the result
			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
				}

				// If we expected an error, verify no account was created
				if tt.name != "Add duplicate account" { // Skip for duplicate account test
					cfg := multigit.LoadConfig()
					_, exists := cfg.Accounts[tt.accountName]
					assert.False(t, exists, "Account should not exist in config on error")
				}
			} else {
				assert.NoError(t, err)

				// Verify the account was created in the config
				cfg := multigit.LoadConfig()

				// Verify the account exists in the config
				account, exists := cfg.Accounts[tt.accountName]
				assert.True(t, exists, "Account should exist in config")
				assert.Equal(t, tt.accountName, account.Name, "Account name should match")
				assert.Equal(t, tt.email, account.Email, "Account email should match")
			}

			// Verify all expected calls were made
			mockSSH.AssertExpectations(t)
		})
	}
}
