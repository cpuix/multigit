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
var SSHCreateFunc = func(accountName, email, passphrase string, keyType ssh.KeyType) error {
	return ssh.CreateSSHKey(accountName, email, passphrase, keyType)
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
	Args: cobra.ExactArgs(1),
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
		if err := setGitConfig(accountName, account, useLocal); err != nil {
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
func setGitConfig(accountName string, account multigit.Account, local bool) error {
	scope := "--global"
	if local {
		if !IsGitRepo() {
			return fmt.Errorf("--local flag requires running inside a git repository")
		}
		scope = "--local"
	}

	if err := RunGitCommand("config", scope, "user.name", account.Name); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}

	if err := RunGitCommand("config", scope, "user.email", account.Email); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}

	if !local {
		if err := RunGitCommand("config", "--global", "url.ssh://git@github.com/.insteadOf", "https://github.com/"); err != nil {
			return fmt.Errorf("failed to set git URL rewrite: %w", err)
		}
		if err := RunGitCommand("config", "--global", "push.default", "current"); err != nil {
			return fmt.Errorf("failed to set git push.default: %w", err)
		}
	}

	keyInfo, err := ssh.ResolveKeyPath(accountName)
	if err != nil {
		return fmt.Errorf("failed to determine SSH key path: %w", err)
	}
	if !keyInfo.Exists {
		return fmt.Errorf("ssh key for account '%s' not found at %s", account.Name, keyInfo.Path)
	}

	sshCommand := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -F /dev/null", keyInfo.Path)
	if err := RunGitCommand("config", scope, "core.sshCommand", sshCommand); err != nil {
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
