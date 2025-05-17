package cmd

import (
	"fmt"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <account_name>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a GitHub account",
	Long: `Delete a GitHub account and its associated SSH keys and config.
This will:
1. Remove the SSH key from the SSH agent
2. Delete the SSH key files
3. Remove the SSH config entry
4. Remove the account from the multigit config`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]

		// Confirm before deleting
		if !forceDelete {
			fmt.Printf("⚠️  WARNING: This will permanently delete the account '%s' and its SSH keys.\n", accountName)
			fmt.Print("Are you sure you want to continue? [y/N] ")

			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || (response != "y" && response != "Y") {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		// Delete the account
		if err := multigit.DeleteAccount(accountName); err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
		}

		fmt.Printf("✅ Account '%s' has been deleted successfully\n", accountName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
}
