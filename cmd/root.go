package cmd

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/cpuix/multigit/internal/multigit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// CfgFile holds the path to the config file
	CfgFile string
	// Config holds the application configuration
	Config *multigit.Config
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
    Use:   "multigit",
    Short: "MultiGit is a CLI for managing multiple GitHub accounts",
    Long:  `MultiGit helps you to manage multiple GitHub accounts effortlessly.`,
}

// Execute executes the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
    cobra.OnInitialize(InitConfig)
	RootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is $HOME/.multigit.json)")
}

// InitConfig initializes the configuration
func InitConfig() {
	// Initialize a default config
	Config = &multigit.Config{
		Accounts:      make(map[string]multigit.Account),
		Profiles:      make(map[string]multigit.Profile),
		ActiveAccount: "",
	}

	if CfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(CfgFile)
		
		// Create directory if it doesn't exist
		dir := filepath.Dir(CfgFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
		}
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "config" (without extension)
		configDir := filepath.Join(home, ".config", "multigit")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		// Unmarshal the config into our Config variable
		if err := viper.Unmarshal(Config); err != nil {
			fmt.Println("Error unmarshaling config:", err)
		}
	} else {
		// Create default config if it doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create default
			if CfgFile != "" {
				// Create the config file at the specified path
				if err := viper.WriteConfigAs(CfgFile); err != nil {
					fmt.Println("Error creating config file:", err)
				}
			} else {
				// Use default config path
				if err := multigit.SaveConfig(*Config); err != nil {
					fmt.Println("Error creating default config:", err)
				}
			}
		} else {
			// Config file was found but another error was produced
			fmt.Println("Error reading config file:", err)
		}
	}

	// Ensure accounts and profiles maps are initialized
	if Config.Accounts == nil {
		Config.Accounts = make(map[string]multigit.Account)
	}
	if Config.Profiles == nil {
		Config.Profiles = make(map[string]multigit.Profile)
	}
}
