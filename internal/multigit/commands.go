package multigit

import (
	_ "fmt"
	"github.com/cpuix/multigit/internal/ssh"
)

// createSSHKey is a wrapper around ssh.CreateSSHKey
func createSSHKey(accountName, accountEmail, passphrase string, keyType ssh.KeyType) error {
	return SSHClient.CreateSSHKey(accountName, accountEmail, passphrase, keyType)
}

// CreateSSHKeyForTesting exposes createSSHKey for testing
func CreateSSHKeyForTesting(accountName, accountEmail, passphrase string, keyType ssh.KeyType) error {
	return createSSHKey(accountName, accountEmail, passphrase, keyType)
}

// addSSHKeyToAgent is a wrapper around ssh.AddSSHKeyToAgent
func addSSHKeyToAgent(accountName string) error {
	return SSHClient.AddSSHKeyToAgent(accountName)
}

// AddSSHKeyToAgentForTesting exposes addSSHKeyToAgent for testing
func AddSSHKeyToAgentForTesting(accountName string) error {
	return addSSHKeyToAgent(accountName)
}

// addSSHConfigEntry is a wrapper around ssh.AddSSHConfigEntry
func addSSHConfigEntry(accountName string) error {
	return SSHClient.AddSSHConfigEntry(accountName)
}

// AddSSHConfigEntryForTesting exposes addSSHConfigEntry for testing
func AddSSHConfigEntryForTesting(accountName string) error {
	return addSSHConfigEntry(accountName)
}
