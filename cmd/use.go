package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <account_name>",
	Short: "Switch to the specified GitHub account",
	Long: `Switch to the specified GitHub account by setting up the SSH key and git configuration.
This will:
1. Add the specified account's SSH key to the SSH agent
2. Set the global git user name and email
3. Update the active account in the configuration`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]
		
		// Check if account exists in config
		config := multigit.LoadConfig()
		account, exists := config.Accounts[accountName]
		if !exists {
			return fmt.Errorf("account '%s' does not exist. Use 'multigit create' to add it first", accountName)
		}

		// Add SSH key to agent
		if err := ssh.AddSSHKeyToAgent(accountName); err != nil {
			return fmt.Errorf("failed to add SSH key to agent: %w", err)
		}

		// Set git config
		if err := setGitConfig(account); err != nil {
			return fmt.Errorf("failed to set git config: %w", err)
		}

		// Update active account in config
		config.ActiveAccount = accountName
		if err := multigit.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✅ Switched to account: %s <%s>\n", account.Name, account.Email)
		return nil
	},
}

// setGitConfig sets the global git user name and email
func setGitConfig(account multigit.Account) error {
	// Set git user name
	if err := runGitCommand("config", "--global", "user.name", account.Name); err != nil {
		return fmt.Errorf("failed to set git user.name: %w", err)
	}

	// Set git user email
	if err := runGitCommand("config", "--global", "user.email", account.Email); err != nil {
		return fmt.Errorf("failed to set git user.email: %w", err)
	}

	// Set github.com to use the correct SSH key
	sshCommand := fmt.Sprintf("ssh -i ~/.ssh/id_rsa_%s -F /dev/null", account.Name)
	if err := runGitCommand("config", "--global", "core.sshCommand", sshCommand); err != nil {
		return fmt.Errorf("failed to set git core.sshCommand: %w", err)
	}

	// Set push default to current
	if err := runGitCommand("config", "--global", "push.default", "current"); err != nil {
		return fmt.Errorf("failed to set git push.default: %w", err)
	}

	return nil
}

// runGitCommand is a helper function to run git commands
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(useCmd)
}
