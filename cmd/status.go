package cmd

import (
	"fmt"
	"os"

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
		homeDir, _ := os.UserHomeDir()
		keyPath := fmt.Sprintf("%s/.ssh/id_rsa_%s", homeDir, activeAccountName)
		if _, err := os.Stat(keyPath); err == nil {
			fmt.Printf("  SSH Key: %s\n", keyPath)
		} else {
			fmt.Printf("  SSH Key: %s\n", color.YellowString(keyPath+" (not found)"))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
