package multigit_test

import (
	"testing"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
)

// TestCreateSSHKey tests the createSSHKey command handler
func TestCreateSSHKey(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"
	accountEmail := "test@example.com"
	passphrase := "test-passphrase"
	keyType := ssh.KeyTypeED25519

	// The mock should expect CreateSSHKey to be called with the specified parameters
	mockSSH.On("CreateSSHKey", accountName, accountEmail, passphrase, keyType).Return(nil)

	// Call the function being tested
	err := multigit.CreateSSHKeyForTesting(accountName, accountEmail, passphrase, keyType)
	
	// Verify the error is nil
	assert.NoError(t, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}

// TestAddSSHKeyToAgent tests the addSSHKeyToAgent command handler
func TestAddSSHKeyToAgent(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"

	// The mock should expect AddSSHKeyToAgent to be called with the specified parameters
	mockSSH.On("AddSSHKeyToAgent", accountName).Return(nil)

	// Call the function being tested
	err := multigit.AddSSHKeyToAgentForTesting(accountName)
	
	// Verify the error is nil
	assert.NoError(t, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}

// TestAddSSHConfigEntry tests the addSSHConfigEntry command handler
func TestAddSSHConfigEntry(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"

	// The mock should expect AddSSHConfigEntry to be called with the specified parameters
	mockSSH.On("AddSSHConfigEntry", accountName).Return(nil)

	// Call the function being tested
	err := multigit.AddSSHConfigEntryForTesting(accountName)
	
	// Verify the error is nil
	assert.NoError(t, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}

// TestCreateSSHKeyError tests the createSSHKey command handler with an error
func TestCreateSSHKeyError(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"
	accountEmail := "test@example.com"
	passphrase := "test-passphrase"
	keyType := ssh.KeyTypeED25519

	// The mock should expect CreateSSHKey to be called and return an error
	mockError := assert.AnError
	mockSSH.On("CreateSSHKey", accountName, accountEmail, passphrase, keyType).Return(mockError)

	// Call the function being tested
	err := multigit.CreateSSHKeyForTesting(accountName, accountEmail, passphrase, keyType)
	
	// Verify the error is returned
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}

// TestAddSSHKeyToAgentError tests the addSSHKeyToAgent command handler with an error
func TestAddSSHKeyToAgentError(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"

	// The mock should expect AddSSHKeyToAgent to be called and return an error
	mockError := assert.AnError
	mockSSH.On("AddSSHKeyToAgent", accountName).Return(mockError)

	// Call the function being tested
	err := multigit.AddSSHKeyToAgentForTesting(accountName)
	
	// Verify the error is returned
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}

// TestAddSSHConfigEntryError tests the addSSHConfigEntry command handler with an error
func TestAddSSHConfigEntryError(t *testing.T) {
	// Create a mock SSH client
	mockSSH := new(MockSSH)
	
	// Save the original SSH client and restore it after the test
	originalSSH := multigit.SSHClient
	multigit.SSHClient = mockSSH
	defer func() {
		multigit.SSHClient = originalSSH
	}()

	// Set up the mock expectations
	accountName := "test-account"

	// The mock should expect AddSSHConfigEntry to be called and return an error
	mockError := assert.AnError
	mockSSH.On("AddSSHConfigEntry", accountName).Return(mockError)

	// Call the function being tested
	err := multigit.AddSSHConfigEntryForTesting(accountName)
	
	// Verify the error is returned
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	
	// Verify that the mock was called as expected
	mockSSH.AssertExpectations(t)
}
