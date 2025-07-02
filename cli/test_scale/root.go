package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apidirect",
	Short: "API-Direct CLI - Deploy and manage APIs from your command line",
	Long: color.CyanString(`
   ___   ____  ____     ____  _                __  
  / _ | / __ \/  _/____/ __ \(_)______ _______/ /_ 
 / __ |/ /_/ // // ___/ / / / / __/ _ '/ ___/ __/
/_/ |_/ .___/___/_/  /_____/_/_/  \__,_/\__/\__/  
     /_/                                         v` + Version) + `

API-Direct CLI enables you to:
  • Create and deploy serverless APIs with a single command
  • Manage API versions and environments
  • Publish APIs to the marketplace
  • Monitor API performance and logs

Get started with 'apidirect init' to create your first API.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.apidirect/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Check if version flag is set
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			showVersion()
			os.Exit(0)
		}
	}
	
	// Add commands - removed as they will be added in their respective init() functions
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".apidirect" (without extension).
		configPath := filepath.Join(home, ".apidirect")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configPath, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("APIDIRECT")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Helper functions
func printSuccess(message string) {
	color.Green("✓ %s", message)
}

func printError(message string) {
	color.Red("✗ %s", message)
}

func printInfo(message string) {
	color.Cyan("ℹ %s", message)
}

func printWarning(message string) {
	color.Yellow("⚠ %s", message)
}

// checkAuth checks if the user is authenticated
func checkAuth() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Auth.AccessToken == "" {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	return nil
}
