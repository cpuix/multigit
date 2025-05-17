package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/spf13/cobra"
)

// SSHCreateFunc is a function type for creating SSH keys
var SSHCreateFunc = func(accountName, email, passphrase string) error {
	return ssh.CreateSSHKey(accountName, email, passphrase)
}

// SSHAddToAgentFunc is a function type for adding SSH keys to the agent
var SSHAddToAgentFunc = func(accountName string) error {
	return ssh.AddSSHKeyToAgent(accountName)
}

var (
	useLocal bool
	profile  string
)

var useCmd = &cobra.Command{
	Use:   "use <account_name>",
	Short: "Switch to the specified GitHub account",
	Long: `Switch to the specified GitHub account by setting up the SSH key and git configuration.
This will:
1. Add the specified account's SSH key to the SSH agent
2. Set the git user name and email (global or local)
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
		if err := SSHAddToAgentFunc(accountName); err != nil {
			return fmt.Errorf("failed to add SSH key to agent: %w", err)
		}

		// Set git config
		if err := setGitConfig(account, useLocal); err != nil {
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

// setGitConfig sets the git user name and email (global or local)
func setGitConfig(account multigit.Account, local bool) error {
	// Prepare git config args
	configArgs := []string{"config"}
	if !local {
		configArgs = append(configArgs, "--global")
	}

	// Set git config
	if err := RunGitCommand(append(configArgs, "user.name", account.Name)...); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}

	if err := RunGitCommand(append(configArgs, "user.email", account.Email)...); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}

	// Only set URL rewrite for global config
	if !local {
		if err := RunGitCommand(append(configArgs, "url.ssh://git@github.com/.insteadOf", "https://github.com/")...); err != nil {
			return fmt.Errorf("failed to set git URL rewrite: %w", err)
		}
	}

	// Set push default to current
	if err := RunGitCommand("config", "--global", "push.default", "current"); err != nil {
		return fmt.Errorf("failed to set git push.default: %w", err)
	}

	// Set github.com to use the correct SSH key
	sshCommand := fmt.Sprintf("ssh -i ~/.ssh/id_rsa_%s -F /dev/null", account.Name)
	if err := RunGitCommand("config", "--global", "core.sshCommand", sshCommand); err != nil {
		return fmt.Errorf("failed to set git core.sshCommand: %w", err)
	}

	return nil
}

// RunGitCommand is a variable that holds the function to run git commands
var RunGitCommand = func(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsGitRepo is a variable that holds the function to check if current directory is a git repo
var IsGitRepo = func() bool {
	_, err := os.Stat(".git")
	return err == nil || !os.IsNotExist(err)
}

func init() {
	RootCmd.AddCommand(useCmd)
	useCmd.Flags().BoolVarP(&useLocal, "local", "l", false, "Set git config locally (for current repository only)")
	useCmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile to use (optional)")
}
