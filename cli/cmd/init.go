package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/api-direct/cli/pkg/scaffold"
	"github.com/spf13/cobra"
)

var (
	runtime string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [api-name]",
	Short: "Initialize a new API project",
	Long: `Initialize a new API project with boilerplate code and configuration.
This command creates a new directory with the specified name and sets up
the basic structure for your API.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiName := args[0]
		
		// Validate API name
		if !isValidAPIName(apiName) {
			return fmt.Errorf("invalid API name: %s. Use only lowercase letters, numbers, and hyphens", apiName)
		}

		// Check if directory already exists
		if _, err := os.Stat(apiName); err == nil {
			return fmt.Errorf("directory %s already exists", apiName)
		}

		// Validate runtime
		validRuntimes := []string{"python3.9", "python3.10", "python3.11", "nodejs18", "nodejs20"}
		if runtime == "" {
			runtime = "python3.9" // Default runtime
		}
		
		runtimeValid := false
		for _, r := range validRuntimes {
			if r == runtime {
				runtimeValid = true
				break
			}
		}
		
		if !runtimeValid {
			return fmt.Errorf("invalid runtime: %s. Valid options are: %s", runtime, strings.Join(validRuntimes, ", "))
		}

		printInfo(fmt.Sprintf("Creating new API project: %s", apiName))
		printInfo(fmt.Sprintf("Runtime: %s", runtime))

		// Create project directory
		if err := os.MkdirAll(apiName, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Initialize project based on runtime
		var err error
		switch {
		case strings.HasPrefix(runtime, "python"):
			err = scaffold.InitPythonProject(apiName, runtime)
		case strings.HasPrefix(runtime, "nodejs"):
			err = scaffold.InitNodeProject(apiName, runtime)
		default:
			err = fmt.Errorf("unsupported runtime: %s", runtime)
		}

		if err != nil {
			// Clean up on error
			os.RemoveAll(apiName)
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		// Success message
		printSuccess(fmt.Sprintf("API project '%s' created successfully!", apiName))
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. cd %s\n", apiName)
		fmt.Println("  2. Review and edit apidirect.yaml")
		fmt.Println("  3. Implement your API logic in main.py")
		fmt.Println("  4. Test locally with: apidirect run")
		fmt.Println("  5. Deploy with: apidirect deploy")
		
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "Runtime for the API (e.g., python3.9, nodejs18)")
}

// isValidAPIName checks if the API name is valid
func isValidAPIName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	
	// Must start with a letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	
	// Only lowercase letters, numbers, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	
	// Cannot end with a hyphen
	if name[len(name)-1] == '-' {
		return false
	}
	
	return true
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage API-Direct configuration",
	Long:  `View and manage API-Direct configuration settings.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation for getting config values
		printInfo("Config get command not yet implemented")
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation for setting config values
		printInfo("Config set command not yet implemented")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long:  `Manage environment variables for your API.`,
}

var envSetCmd = &cobra.Command{
	Use:   "set [KEY=VALUE]",
	Short: "Set an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation for setting env variables
		printInfo("Env set command not yet implemented")
		return nil
	},
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation for listing env variables
		printInfo("Env list command not yet implemented")
		return nil
	},
}

func init() {
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envListCmd)
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your API to API-Direct",
	Long: `Deploy your API to the API-Direct platform. This command packages your code,
uploads it to the cloud, and makes it available at a public endpoint.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkAuth()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if we're in an API project directory
		if _, err := os.Stat("apidirect.yaml"); os.IsNotExist(err) {
			return fmt.Errorf("not in an API project directory. Run this command from a directory containing apidirect.yaml")
		}

		printInfo("Deploy command not yet fully implemented")
		printWarning("This is a placeholder for the deployment functionality")
		
		return nil
	},
}

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [api-name]",
	Short: "Stream logs from your deployed API",
	Long:  `Stream real-time logs from your deployed API or view recent log entries.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkAuth()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		apiName := args[0]
		printInfo(fmt.Sprintf("Streaming logs for API: %s", apiName))
		printWarning("Logs command not yet implemented")
		return nil
	},
}

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish [api-name]",
	Short: "Publish your API to the marketplace",
	Long:  `Publish your deployed API to the API-Direct marketplace, making it discoverable by other users.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkAuth()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		apiName := args[0]
		printInfo(fmt.Sprintf("Publishing API: %s", apiName))
		printWarning("Publish command not yet implemented")
		return nil
	},
}

// unpublishCmd represents the unpublish command
var unpublishCmd = &cobra.Command{
	Use:   "unpublish [api-name]",
	Short: "Remove your API from the marketplace",
	Long:  `Remove your API from the API-Direct marketplace. The API will still be deployed and accessible via its direct URL.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkAuth()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		apiName := args[0]
		printInfo(fmt.Sprintf("Unpublishing API: %s", apiName))
		printWarning("Unpublish command not yet implemented")
		return nil
	},
}
