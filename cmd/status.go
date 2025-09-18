package cmd

import (
	"fmt"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the currently active GitHub account",
	Long:  `Display information about the currently active GitHub account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get active account
		activeAccountName, account, err := multigit.GetActiveAccount()
		if err != nil {
			fmt.Println(color.YellowString("No active GitHub account. Use 'multigit use <account>' to set an active account."))
			return nil
		}

		// Print active account info
		fmt.Println(color.GreenString("Active GitHub account:"))
		fmt.Printf("  Name:  %s\n", color.CyanString(activeAccountName))
		fmt.Printf("  Email: %s\n", color.CyanString(account.Email))

		// Check if SSH key exists
		keyInfo, err := ssh.ResolveKeyPath(activeAccountName)
		if err != nil {
			fmt.Printf("  SSH Key: %s\n", color.RedString("failed to resolve key path: %v", err))
		} else if keyInfo.Exists {
			fmt.Printf("  SSH Key: %s (%s)\n", keyInfo.Path, keyInfo.Type)
		} else {
			fmt.Printf("  SSH Key: %s\n", color.YellowString(keyInfo.Path+" (not found)"))
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
