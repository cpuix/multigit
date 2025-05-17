package cmd

import (
	"fmt"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage git profiles",
	Long:  `Create, list, or delete git profiles that contain multiple account configurations.`,
}

var createProfileCmd = &cobra.Command{
	Use:   "create [profile-name]",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		config := multigit.LoadConfig()

		// Check if profile already exists
		if _, exists := config.Profiles[profileName]; exists {
			return fmt.Errorf("profile '%s' already exists", profileName)
		}

		// Create new profile
		config.Profiles[profileName] = multigit.Profile{
			Name:     profileName,
			Accounts: make(map[string]bool),
		}

		// Save the updated config
		if err := multigit.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save profile: %v", err)
		}

		color.Green("✓ Created profile: %s\n", profileName)
		return nil
	},
}

var listProfilesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := multigit.LoadConfig()

		if len(config.Profiles) == 0 {
			fmt.Println("No profiles found. Create one with 'multigit profile create <name>'")
			return nil
		}

		// Get active profile name
		activeProfile := ""
		if config.ActiveProfile != "" {
			activeProfile = config.ActiveProfile
		}

		// Print profiles
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		fmt.Println(headerFmt("Available Profiles:"))
		fmt.Println()

		for name, profile := range config.Profiles {
			// Check if this is the active profile
			status := " "
			if name == activeProfile {
				status = color.GreenString("✓")
			}

			// Count accounts in this profile
			accountCount := len(profile.Accounts)

			// Print profile info
			fmt.Printf("%s %s\n", status, color.CyanString(name))
			fmt.Printf("  Accounts: %d\n", accountCount)
			fmt.Println()
		}

		if activeProfile != "" {
			fmt.Printf("Active profile: %s\n", color.GreenString(activeProfile))
		} else {
			fmt.Println("No active profile. Use 'multigit profile use <name>' to set one.")
		}

		return nil
	},
}

var useProfileCmd = &cobra.Command{
	Use:   "use [profile-name]",
	Short: "Set active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		config := multigit.LoadConfig()

		// Check if profile exists
		if _, exists := config.Profiles[profileName]; !exists {
			return fmt.Errorf("profile '%s' does not exist", profileName)
		}

		// Set active profile
		config.ActiveProfile = profileName

		// Save the updated config
		if err := multigit.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to update active profile: %v", err)
		}

		color.Green("✓ Active profile set to: %s\n", profileName)
		return nil
	},
}

var deleteProfileCmd = &cobra.Command{
	Use:     "delete [profile-name]",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		config := multigit.LoadConfig()

		// Check if profile exists
		if _, exists := config.Profiles[profileName]; !exists {
			return fmt.Errorf("profile '%s' does not exist", profileName)
		}

		// Ask for confirmation
		if !profileForceDelete {
			if !confirm(fmt.Sprintf("Are you sure you want to delete profile '%s'?", profileName)) {
				return nil
			}
		}

		// If this was the active profile, clear active profile
		if config.ActiveProfile == profileName {
			config.ActiveProfile = ""
		}

		// Delete the profile
		delete(config.Profiles, profileName)


		// Save the updated config
		if err := multigit.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to delete profile: %v", err)
		}

		color.Green("✓ Deleted profile: %s\n", profileName)
		return nil
	},
}

var (
	profileForceDelete bool
)

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(createProfileCmd)
	profileCmd.AddCommand(listProfilesCmd)
	profileCmd.AddCommand(useProfileCmd)
	profileCmd.AddCommand(deleteProfileCmd)

	// Add flags
	deleteProfileCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "force deletion without confirmation")
}

// confirm shows a prompt with question and waits for user confirmation
type confirmFunc func(string) bool

var confirm confirmFunc = func(question string) bool {
	var response string
	fmt.Printf("%s [y/N]: ", question)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	return response == "y" || response == "Y"
}
