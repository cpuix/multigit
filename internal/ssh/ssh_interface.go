package ssh

// KeyType represents the type of SSH key
type KeyType string

const (
	// KeyTypeRSA represents RSA key type
	KeyTypeRSA KeyType = "rsa"
	// KeyTypeED25519 represents ED25519 key type
	KeyTypeED25519 KeyType = "ed25519"
)

// SSHOperations defines the interface for SSH-related operations
type SSHOperations interface {
	CreateSSHKey(accountName, email, passphrase string, keyType KeyType) error
	AddSSHKeyToAgent(accountName string) error
	AddSSHConfigEntry(accountName string) error
	DeleteSSHKey(accountName string) error
	RemoveSSHConfigEntry(accountName string) error
}

// DefaultSSH implements the default SSH operations
type DefaultSSH struct{}

// CreateSSHKey creates a new SSH key pair
func (d *DefaultSSH) CreateSSHKey(accountName, email, passphrase string, keyType KeyType) error {
	return CreateSSHKey(accountName, email, passphrase, keyType)
}

// AddSSHKeyToAgent adds the SSH key to the SSH agent
func (d *DefaultSSH) AddSSHKeyToAgent(accountName string) error {
	return AddSSHKeyToAgent(accountName)
}

// AddSSHConfigEntry adds an entry to the SSH config file
func (d *DefaultSSH) AddSSHConfigEntry(accountName string) error {
	return AddSSHConfigEntry(accountName)
}

// DeleteSSHKey deletes the SSH key pair
func (d *DefaultSSH) DeleteSSHKey(accountName string) error {
	return DeleteSSHKey(accountName)
}

// RemoveSSHConfigEntry removes the SSH config entry
func (d *DefaultSSH) RemoveSSHConfigEntry(accountName string) error {
	return RemoveSSHConfigEntry(accountName)
}
