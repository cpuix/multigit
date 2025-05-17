package multigit

import (
	"encoding/json"
	"fmt"
	"strings"
	"log"
	"os"
	"path/filepath"

	"github.com/cpuix/multigit/internal/ssh"
)

// Account represents a GitHub account with its associated information
type Account struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Profile represents a collection of accounts
type Profile struct {
	Name     string          `json:"name"`
	Accounts map[string]bool `json:"accounts"` // map[accountName]enabled
}

// Config represents the multigit configuration
type Config struct {
	Accounts      map[string]Account `json:"accounts"`
	ActiveAccount string             `json:"active_account"`
	Profiles      map[string]Profile `json:"profiles"`
	ActiveProfile string             `json:"active_profile"`
}

// SSHClient is the interface for SSH operations
var SSHClient ssh.SSHOperations = &ssh.DefaultSSH{}

// CreateAccount creates a new GitHub account with SSH key and configures it
func CreateAccount(accountName, accountEmail, passphrase string, saveConfigFunc ...func(Config) error) error {
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
	config := LoadConfig()
	if _, exists := config.Accounts[accountName]; exists {
		return fmt.Errorf("account '%s' already exists", accountName)
	}

	// Create SSH key pair
	if err := SSHClient.CreateSSHKey(accountName, accountEmail, passphrase); err != nil {
		return fmt.Errorf("failed to create SSH key: %w", err)
	}

	// Add SSH key to agent
	if err := SSHClient.AddSSHKeyToAgent(accountName); err != nil {
		return fmt.Errorf("failed to add SSH key to agent: %w", err)
	}

	// Add SSH config entry
	if err := SSHClient.AddSSHConfigEntry(accountName); err != nil {
		return fmt.Errorf("failed to add SSH config entry: %w", err)
	}

	// Add or update account in config
	if config.Accounts == nil {
		config.Accounts = make(map[string]Account)
	}

	config.Accounts[accountName] = Account{
		Name:  accountName,
		Email: accountEmail,
	}

	// Save config
	saveFunc := SaveConfig
	if len(saveConfigFunc) > 0 {
		saveFunc = saveConfigFunc[0]
	}
	if err := saveFunc(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Account '%s' created successfully\n", accountName)
	return nil
}

// DeleteAccount deletes a GitHub account and its associated SSH keys and config
func DeleteAccount(accountName string) error {
	// Load config
	config := LoadConfig()

	// Check if account exists
	if _, exists := config.Accounts[accountName]; !exists {
		return fmt.Errorf("account '%s' does not exist", accountName)
	}

	// Remove SSH key pair
	if err := SSHClient.DeleteSSHKey(accountName); err != nil {
		log.Printf("Warning: Failed to delete SSH key: %v", err)
	}

	// Remove SSH config entry
	if err := SSHClient.RemoveSSHConfigEntry(accountName); err != nil {
		log.Printf("Warning: Failed to remove SSH config entry: %v", err)
	}

	// Remove account from config
	delete(config.Accounts, accountName)

	// If the deleted account was active, clear the active account
	if config.ActiveAccount == accountName {
		config.ActiveAccount = ""
	}

	// Save updated config
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Account '%s' deleted successfully\n", accountName)
	return nil
}

// NewConfig creates a new empty configuration
func NewConfig() Config {
	return Config{
		Accounts: make(map[string]Account),
		Profiles: make(map[string]Profile),
	}
}

// getConfigPath returns the path to the multigit config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return filepath.Join(home, ".config", "multigit", "config.json"), nil
}

// LoadConfigFromFile loads the configuration from a specific file
func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadConfig loads the multigit configuration from disk
func LoadConfig() Config {
	configPath, err := getConfigPath()
	if err != nil {
		log.Printf("Error getting config path: %v\n", err)
		return NewConfig()
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return NewConfig()
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config file: %v\n", err)
		return NewConfig()
	}

	// Unmarshal config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Error parsing config file: %v\n", err)
		return NewConfig()
	}

	// Initialize maps if they are nil
	if config.Accounts == nil {
		config.Accounts = make(map[string]Account)
	}
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	return config
}

// SaveConfig saves the multigit configuration to the default location
var SaveConfig = func(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	return SaveConfigToFile(config, configPath)
}

// SaveConfigToFile saves the configuration to a specific file
func SaveConfigToFile(config Config, filePath string) error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(filePath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Ensure active profile exists if set
	if config.ActiveProfile != "" {
		if _, exists := config.Profiles[config.ActiveProfile]; !exists {
			config.ActiveProfile = ""
		}
	}

	// Ensure active account exists if set
	if config.ActiveAccount != "" {
		if _, exists := config.Accounts[config.ActiveAccount]; !exists {
			config.ActiveAccount = ""
		}
	}

	// Marshal config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetActiveAccount returns the currently active account
func GetActiveAccount() (string, *Account, error) {
	config := LoadConfig()
	if config.ActiveAccount == "" {
		return "", nil, fmt.Errorf("no active account")
	}

	account, exists := config.Accounts[config.ActiveAccount]
	if !exists {
		return "", nil, fmt.Errorf("active account '%s' not found in config", config.ActiveAccount)
	}

	return config.ActiveAccount, &account, nil
}
