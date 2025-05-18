package multigit

import (
	_ "fmt"
	"github.com/cpuix/multigit/internal/ssh"
)

func createSSHKey(accountName, accountEmail, passphrase string, keyType ssh.KeyType) {
	ssh.CreateSSHKey(accountName, accountEmail, passphrase, keyType)
}

func addSSHKeyToAgent(accountName string) {
	ssh.AddSSHKeyToAgent(accountName)
}

func addSSHConfigEntry(accountName string) {
	ssh.AddSSHConfigEntry(accountName)
}
