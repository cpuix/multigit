package cmd

import (
	"fmt"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/cpuix/multigit/internal/ssh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured GitHub accounts",
	Long:  `List all configured GitHub accounts and show which one is currently active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := multigit.LoadConfig()

		if len(config.Accounts) == 0 {
			fmt.Println("No accounts configured. Use 'multigit create' to add an account.")
			return nil
		}

		// Get active account info if any
		activeAccountName, activeAccount, _ := multigit.GetActiveAccount()

		// Print header
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		fmt.Println(headerFmt("Configured GitHub Accounts:"))
		fmt.Println()

		// Print each account
		for name, account := range config.Accounts {
			// Check if this is the active account
			isActive := activeAccountName == name

			// Format account info
			status := ""
			if isActive {
				status = color.GreenString("✓")
			} else {
				status = " "
			}

			// Print account details
			fmt.Printf("%s %s\n", status, color.CyanString(name))
			fmt.Printf("  Email: %s\n", account.Email)

			// Show SSH key info if available
			keyInfo, err := ssh.ResolveKeyPath(name)
			if err != nil {
				fmt.Printf("  SSH Key: %s\n", color.RedString("failed to resolve key path: %v", err))
			} else if keyInfo.Exists {
				fmt.Printf("  SSH Key: %s (%s)\n", keyInfo.Path, keyInfo.Type)
			} else {
				fmt.Printf("  SSH Key: %s (not found)\n", color.YellowString(keyInfo.Path))
			}

			// Show active status
			if isActive {
				fmt.Printf("  Status: %s\n", color.GreenString("Active"))
			}

			fmt.Println()
		}

		// Print usage instructions
		if activeAccountName != "" {
			fmt.Printf("Active account: %s <%s>\n",
				color.CyanString(activeAccountName),
				color.CyanString(activeAccount.Email))
		} else {
			fmt.Println("No active account. Use 'multigit use <account>' to set an active account.")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
