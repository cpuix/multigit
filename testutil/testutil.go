package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/stretchr/testify/require"
)

// SetupTestConfig creates a test configuration file and initializes the config
func SetupTestConfig(t *testing.T, tempDir string) {
	t.Helper()

	// Create test config
	config := multigit.NewConfig()
	config.Accounts["test-account"] = multigit.Account{
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Save config
	cfgPath := filepath.Join(tempDir, "config.json")
	err := multigit.SaveConfigToFile(config, cfgPath)
	require.NoError(t, err, "Failed to save test config")

	// Set config path
	cmd.CfgFile = cfgPath
	cmd.InitConfig()
}

// SetupGitRepo creates a test git repository in a temporary directory
func SetupGitRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err, "Failed to create .git directory")

	return tempDir
}

// Chdir changes the current working directory and returns a cleanup function
func Chdir(t *testing.T, dir string) func() {
	t.Helper()

	oldDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	err = os.Chdir(dir)
	require.NoError(t, err, "Failed to change directory")

	return func() {
		err := os.Chdir(oldDir)
		require.NoError(t, err, "Failed to restore working directory")
	}
}
