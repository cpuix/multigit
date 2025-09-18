package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// resolveSSHKeyPath searches for SSH keys associated with an account.
// It returns the first existing key path (preferring RSA, then ED25519),
// a boolean indicating whether a key was found, and the list of candidate
// paths that were checked.
func resolveSSHKeyPath(homeDir, accountName string) (string, bool, []string) {
	candidates := []string{
		filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", accountName)),
		filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s", accountName)),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true, candidates
		}
	}

	return candidates[0], false, candidates
}
