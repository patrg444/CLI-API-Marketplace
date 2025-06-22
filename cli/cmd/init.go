package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/api-direct/cli/pkg/scaffold"
	"github.com/api-direct/cli/pkg/wizard"
	"github.com/spf13/cobra"
)

var (
	initRuntime     string
	interactive bool
	template    string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [api-name]",
	Short: "Initialize a new API project",
	Long: `Initialize a new API project with boilerplate code and configuration.
This command creates a new directory with the specified name and sets up
the basic structure for your API.

Use --interactive for a guided setup experience with templates and features.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Interactive mode
		if interactive || len(args) == 0 {
			return runInteractiveInit()
		}

		// Non-interactive mode (existing behavior)
		apiName := args[0]
		return runStandardInit(apiName)
	},
}

func runInteractiveInit() error {
	config, err := wizard.RunInteractiveWizard()
	if err != nil {
		return err
	}

	// Create project directory
	if err := os.MkdirAll(config.APIName, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize project based on template and initRuntime
	var initErr error
	switch {
	case strings.HasPrefix(config.Runtime, "python"):
		scaffoldTemplate := scaffold.APITemplate{
			ID:          config.Template.ID,
			Name:        config.Template.Name,
			Description: config.Template.Description,
			Runtime:     config.Template.Runtime,
			Category:    config.Template.Category,
			Features:    config.Template.Features,
		}
		initErr = scaffold.InitPythonProjectWithTemplate(config.APIName, config.Runtime, scaffoldTemplate, config.Features)
	case strings.HasPrefix(config.Runtime, "nodejs"):
		scaffoldTemplate := scaffold.APITemplate{
			ID:          config.Template.ID,
			Name:        config.Template.Name,
			Description: config.Template.Description,
			Runtime:     config.Template.Runtime,
			Category:    config.Template.Category,
			Features:    config.Template.Features,
		}
		initErr = scaffold.InitNodeProjectWithTemplate(config.APIName, config.Runtime, scaffoldTemplate, config.Features)
	default:
		initErr = fmt.Errorf("unsupported initRuntime: %s", config.Runtime)
	}

	if initErr != nil {
		// Clean up on error
		os.RemoveAll(config.APIName)
		return fmt.Errorf("failed to initialize project: %w", initErr)
	}

	// Success message
	printSuccess(fmt.Sprintf("🎉 API project '%s' created successfully!", config.APIName))
	fmt.Printf("📁 Template: %s\n", config.Template.Name)
	fmt.Printf("🐍 Runtime: %s\n", config.Runtime)
	
	if len(config.Features) > 0 {
		fmt.Printf("✨ Features: %s\n", strings.Join(config.Features, ", "))
	}
	
	fmt.Println("\n🚀 Next steps:")
	fmt.Printf("  1. cd %s\n", config.APIName)
	fmt.Println("  2. Review the generated code and configuration")
	fmt.Println("  3. Customize your API logic")
	fmt.Println("  4. Test locally with: apidirect run")
	fmt.Println("  5. Deploy with: apidirect deploy")
	fmt.Println("  6. Publish to marketplace: apidirect publish")
	
	return nil
}

func runStandardInit(apiName string) error {
	// Validate API name
	if !isValidAPIName(apiName) {
		return fmt.Errorf("invalid API name: %s. Use only lowercase letters, numbers, and hyphens", apiName)
	}

	// Check if directory already exists
	if _, err := os.Stat(apiName); err == nil {
		return fmt.Errorf("directory %s already exists", apiName)
	}

	// Handle template flag
	var selectedTemplate wizard.APITemplate
	if template != "" {
		var found bool
		selectedTemplate, found = wizard.GetTemplateByID(template)
		if !found {
			fmt.Println("Available templates:")
			for _, t := range wizard.ListTemplates() {
				fmt.Printf("  %s - %s\n", t.ID, t.Name)
			}
			return fmt.Errorf("invalid template: %s", template)
		}
		if initRuntime == "" {
			initRuntime = selectedTemplate.Runtime
		}
	}

	// Validate initRuntime
	validRuntimes := []string{"python3.9", "python3.10", "python3.11", "nodejs18", "nodejs20"}
	if initRuntime == "" {
		initRuntime = "python3.9" // Default initRuntime
	}
	
	initRuntimeValid := false
	for _, r := range validRuntimes {
		if r == initRuntime {
			initRuntimeValid = true
			break
		}
	}
	
	if !initRuntimeValid {
		return fmt.Errorf("invalid initRuntime: %s. Valid options are: %s", initRuntime, strings.Join(validRuntimes, ", "))
	}

	printInfo(fmt.Sprintf("Creating new API project: %s", apiName))
	printInfo(fmt.Sprintf("Runtime: %s", initRuntime))
	if template != "" {
		printInfo(fmt.Sprintf("Template: %s", selectedTemplate.Name))
	}

	// Create project directory
	if err := os.MkdirAll(apiName, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize project based on initRuntime
	var err error
	if template != "" {
		// Use template-based initialization
		switch {
		case strings.HasPrefix(initRuntime, "python"):
			scaffoldTemplate := scaffold.APITemplate{
				ID:          selectedTemplate.ID,
				Name:        selectedTemplate.Name,
				Description: selectedTemplate.Description,
				Runtime:     selectedTemplate.Runtime,
				Category:    selectedTemplate.Category,
				Features:    selectedTemplate.Features,
			}
			err = scaffold.InitPythonProjectWithTemplate(apiName, initRuntime, scaffoldTemplate, []string{})
		case strings.HasPrefix(initRuntime, "nodejs"):
			scaffoldTemplate := scaffold.APITemplate{
				ID:          selectedTemplate.ID,
				Name:        selectedTemplate.Name,
				Description: selectedTemplate.Description,
				Runtime:     selectedTemplate.Runtime,
				Category:    selectedTemplate.Category,
				Features:    selectedTemplate.Features,
			}
			err = scaffold.InitNodeProjectWithTemplate(apiName, initRuntime, scaffoldTemplate, []string{})
		default:
			err = fmt.Errorf("unsupported initRuntime: %s", initRuntime)
		}
	} else {
		// Use standard initialization
		switch {
		case strings.HasPrefix(initRuntime, "python"):
			err = scaffold.InitPythonProject(apiName, initRuntime)
		case strings.HasPrefix(initRuntime, "nodejs"):
			err = scaffold.InitNodeProject(apiName, initRuntime)
		default:
			err = fmt.Errorf("unsupported initRuntime: %s", initRuntime)
		}
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
	fmt.Println("  3. Implement your API logic")
	fmt.Println("  4. Test locally with: apidirect run")
	fmt.Println("  5. Deploy with: apidirect deploy")
	
	return nil
}

func init() {
	initCmd.Flags().StringVarP(&initRuntime, "initRuntime", "r", "", "Runtime for the API (e.g., python3.9, nodejs18)")
	initCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run interactive setup wizard")
	initCmd.Flags().StringVarP(&template, "template", "t", "", "Template to use (e.g., basic-rest, crud-database)")
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

// Duplicate command definitions removed - these are now in their respective files
