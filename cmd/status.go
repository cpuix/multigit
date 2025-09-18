package cmd

import (
	"fmt"

	"github.com/cpuix/multigit/internal/multigit"
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
			color.Yellow("No active GitHub account. Use 'multigit use <account>' to set an active account.")
			return nil
		}

		// Print active account info
		fmt.Println(color.GreenString("Active GitHub account:"))
		fmt.Printf("  Name:  %s\n", color.CyanString(activeAccountName))
		fmt.Printf("  Email: %s\n", color.CyanString(account.Email))

		// Check if SSH key exists
		keyPath, candidates, err := findSSHKeyPath(activeAccountName)
		switch {
		case err != nil:
			fmt.Printf("  SSH Key: %s\n", color.YellowString(fmt.Sprintf("error checking SSH key: %v", err)))
		case keyPath != "":
			fmt.Printf("  SSH Key: %s\n", keyPath)
		case len(candidates) > 0:
			fmt.Printf("  SSH Key: %s\n", color.YellowString(fmt.Sprintf("%s (not found)", candidates[0])))
		default:
			fmt.Printf("  SSH Key: %s\n", color.YellowString("(not found)"))
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
