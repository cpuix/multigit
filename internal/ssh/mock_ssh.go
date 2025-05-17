package ssh

import "github.com/stretchr/testify/mock"

// MockSSH is a mock implementation of SSH operations
type MockSSH struct {
	mock.Mock
}

// CreateSSHKey mocks the CreateSSHKey function
func (m *MockSSH) CreateSSHKey(accountName, email, passphrase string) error {
	args := m.Called(accountName, email, passphrase)
	return args.Error(0)
}

// AddSSHKeyToAgent mocks the AddSSHKeyToAgent function
func (m *MockSSH) AddSSHKeyToAgent(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}

// AddSSHConfigEntry mocks the AddSSHConfigEntry function
func (m *MockSSH) AddSSHConfigEntry(accountName string) error {
	args := m.Called(accountName)
	return args.Error(0)
}
