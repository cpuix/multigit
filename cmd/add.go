package cmd

import (
	"fmt"
	"strings"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/spf13/cobra"
)

var passphrase string

var createCmd = &cobra.Command{
	Use:   "create <account_name> <account_email>",
	Short: "Create a new GitHub account SSH key",
	Long: `Create a new GitHub account SSH key and add it to the SSH agent and config file.
This will:
1. Generate a new 4096-bit RSA key pair
2. Add the key to the SSH agent
3. Update the SSH config file
4. Save the account information to the multigit config`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]
		accountEmail := args[1]

		// Validate email format (basic check)
		if len(accountEmail) < 3 || !strings.Contains(accountEmail, "@") {
			return fmt.Errorf("invalid email format")
		}

		// Create the account
		if err := multigit.CreateAccount(accountName, accountEmail, passphrase, ssh.KeyTypeED25519); err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}

		fmt.Printf("\n🎉 Account '%s' has been created and configured successfully!\n", accountName)
		fmt.Println("\nTo use this account, run:")
		fmt.Printf("  multigit use %s\n", accountName)
		fmt.Println("\nTo add this key to your GitHub account, copy the public key above")
		fmt.Println("and add it at: https://github.com/settings/ssh/new")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Passphrase for the SSH key (recommended for security)")
}
