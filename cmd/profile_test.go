package cmd_test

import (
	"testing"

	"github.com/cpuix/multigit/cmd"
	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/testutil"
	"github.com/stretchr/testify/assert"
)

// save original confirm function
var originalConfirm = cmd.Confirm

// mockConfirm is a mock for the confirm function that always returns true
func mockConfirm(string) bool {
	return true
}

func TestProfileCommand(t *testing.T) {
	tempDir := t.TempDir()
	testutil.SetupTestConfig(t, tempDir)

	// Mock the confirm function to always return true
	cmd.Confirm = mockConfirm
	// Restore the original confirm function after the test
	t.Cleanup(func() {
		cmd.Confirm = originalConfirm
	})

	tests := []struct {
		name        string
		args        []string
		expectError bool
		setup       func()
		verify      func(t *testing.T)
	}{
		{
			name:        "Create profile",
			args:        []string{"profile", "create", "test-profile"},
			expectError: false,
			verify: func(t *testing.T) {
				config := multigit.LoadConfig()
				_, exists := config.Profiles["test-profile"]
				assert.True(t, exists, "Profile should exist in config")
			},
		},
		{
			name:        "Create duplicate profile",
			args:        []string{"profile", "create", "test-profile"},
			expectError: true,
			setup: func() {
				// Create the profile first
				config := multigit.LoadConfig()
				config.Profiles["test-profile"] = multigit.Profile{
					Name:     "test-profile",
					Accounts: make(map[string]bool),
				}
				multigit.SaveConfig(config)
			},
		},
		{
			name:        "List profiles",
			args:        []string{"profile", "list"},
			expectError: false,
		},
		{
			name:        "Use profile",
			args:        []string{"profile", "use", "test-profile"},
			expectError: false,
			setup: func() {
				// Create the profile first
				config := multigit.LoadConfig()
				config.Profiles["test-profile"] = multigit.Profile{
					Name:     "test-profile",
					Accounts: make(map[string]bool),
				}
				multigit.SaveConfig(config)
			},
			verify: func(t *testing.T) {
				config := multigit.LoadConfig()
				assert.Equal(t, "test-profile", config.ActiveProfile)
			},
		},
		{
			name:        "Delete profile",
			args:        []string{"profile", "delete", "test-profile"},
			expectError: false,
			setup: func() {
				// Create the profile first
				config := multigit.LoadConfig()
				config.Profiles["test-profile"] = multigit.Profile{
					Name:     "test-profile",
					Accounts: make(map[string]bool),
				}
				multigit.SaveConfig(config)
			},
			verify: func(t *testing.T) {
				config := multigit.LoadConfig()
				_, exists := config.Profiles["test-profile"]
				assert.False(t, exists, "Profile should be deleted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			// Execute command
			cmd.RootCmd.SetArgs(tt.args)
			err := cmd.RootCmd.Execute()

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}


