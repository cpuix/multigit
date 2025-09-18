package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// findSSHKeyPath returns the first existing SSH key path for the account and the list of
// candidate paths that were checked. The ED25519 key naming convention is preferred, with
// RSA used as a fallback for older keys.
func findSSHKeyPath(accountName string) (string, []string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("failed to determine user home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s", accountName)),
		filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName)),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, candidates, nil
		}
	}

	return "", candidates, nil
}
