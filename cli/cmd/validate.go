package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/api-direct/cli/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	manifestPath string
	validateVerbose      bool
	dryRun       bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate an apidirect.yaml manifest file",
	Long: `Validate an apidirect.yaml manifest file to ensure it has correct syntax
and all required fields. This command checks:

- YAML syntax validity
- Required fields presence
- Field value formats
- File references existence
- Port and resource configurations

Run this after editing your manifest to catch errors before deployment.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	
	validateCmd.Flags().StringVarP(&manifestPath, "file", "f", "", "Path to manifest file (default: apidirect.yaml)")
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "Show detailed validation information")
	validateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show deployment preview")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Determine project path
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	
	// Change to project directory for file validation
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)
	
	// Find manifest file
	var manifestFile string
	if manifestPath != "" {
		manifestFile = manifestPath
	} else {
		manifestFile, err = manifest.FindManifest(".")
		if err != nil {
			return fmt.Errorf("no manifest file found. Create one with 'apidirect import' or 'apidirect init'")
		}
	}
	
	fmt.Printf("🔍 Validating %s...\n\n", manifestFile)
	
	// Load and parse manifest
	m, err := manifest.Load(manifestFile)
	if err != nil {
		printValidationError("YAML Parsing", err.Error())
		
		// Provide helpful error messages
		if strings.Contains(err.Error(), "yaml:") {
			fmt.Println("\n💡 Common YAML issues:")
			fmt.Println("  - Check indentation (use spaces, not tabs)")
			fmt.Println("  - Ensure proper quoting of strings with special characters")
			fmt.Println("  - Verify colons are followed by a space")
		}
		
		return fmt.Errorf("validation failed")
	}
	
	// Perform validation checks
	validationResults := performValidation(m)
	
	// Display results
	hasErrors := displayValidationResults(validationResults, verbose)
	
	if hasErrors {
		fmt.Println("\n❌ Validation failed")
		fmt.Println("📝 Fix the issues above and run 'apidirect validate' again")
		return fmt.Errorf("validation failed")
	}
	
	// Show deployment preview if requested
	if dryRun {
		fmt.Println("\n" + strings.Repeat("─", 50))
		showDeploymentPreview(m)
	}
	
	fmt.Println("\n✅ Validation passed!")
	fmt.Println("🚀 Ready to deploy with: apidirect deploy")
	
	return nil
}

type validationCheck struct {
	name    string
	status  string // "pass", "fail", "warning"
	message string
	details []string
}

func performValidation(m *manifest.Manifest) []validationCheck {
	checks := []validationCheck{}
	
	// YAML syntax (already passed if we get here)
	checks = append(checks, validationCheck{
		name:    "YAML syntax",
		status:  "pass",
		message: "Valid",
	})
	
	// Required fields
	requiredCheck := validationCheck{
		name:    "Required fields",
		status:  "pass",
		message: "All present",
		details: []string{},
	}
	
	if m.Name == "" {
		requiredCheck.status = "fail"
		requiredCheck.details = append(requiredCheck.details, "Missing 'name' field")
	}
	if m.Runtime == "" {
		requiredCheck.status = "fail"
		requiredCheck.details = append(requiredCheck.details, "Missing 'runtime' field")
	}
	if m.StartCommand == "" {
		requiredCheck.status = "fail"
		requiredCheck.details = append(requiredCheck.details, "Missing 'start_command' field")
	}
	if m.Port == 0 {
		requiredCheck.status = "fail"
		requiredCheck.details = append(requiredCheck.details, "Missing 'port' field")
	}
	
	if requiredCheck.status == "fail" {
		requiredCheck.message = fmt.Sprintf("%d fields missing", len(requiredCheck.details))
	}
	checks = append(checks, requiredCheck)
	
	// Field formats
	formatCheck := validationCheck{
		name:    "Field formats",
		status:  "pass",
		message: "Valid",
		details: []string{},
	}
	
	// Validate name format
	if m.Name != "" && !isValidProjectName(m.Name) {
		formatCheck.status = "fail"
		formatCheck.details = append(formatCheck.details, 
			fmt.Sprintf("Invalid name '%s': use only lowercase letters, numbers, and hyphens", m.Name))
	}
	
	// Validate port
	if m.Port != 0 && (m.Port < 1 || m.Port > 65535) {
		formatCheck.status = "fail"
		formatCheck.details = append(formatCheck.details, 
			fmt.Sprintf("Invalid port %d: must be between 1-65535", m.Port))
	}
	
	if formatCheck.status == "fail" {
		formatCheck.message = fmt.Sprintf("%d format errors", len(formatCheck.details))
	}
	checks = append(checks, formatCheck)
	
	// File references
	fileCheck := validationCheck{
		name:    "File references",
		status:  "pass",
		message: "All files exist",
		details: []string{},
	}
	
	// Check main file
	if m.Files.Main != "" && !fileExists(m.Files.Main) {
		fileCheck.status = "warning"
		fileCheck.details = append(fileCheck.details, 
			fmt.Sprintf("Main file not found: %s", m.Files.Main))
	}
	
	// Check requirements file
	if m.Files.Requirements != "" && !fileExists(m.Files.Requirements) {
		fileCheck.status = "warning"
		fileCheck.details = append(fileCheck.details, 
			fmt.Sprintf("Requirements file not found: %s", m.Files.Requirements))
	}
	
	// Check Dockerfile if specified
	if m.Files.Dockerfile != "" && !fileExists(m.Files.Dockerfile) {
		fileCheck.status = "warning"
		fileCheck.details = append(fileCheck.details, 
			fmt.Sprintf("Dockerfile not found: %s", m.Files.Dockerfile))
	} else if m.Files.Dockerfile == "" {
		fileCheck.details = append(fileCheck.details, 
			"No Dockerfile specified (will auto-generate during deploy)")
	}
	
	if fileCheck.status == "warning" {
		fileCheck.message = fmt.Sprintf("%d files missing", len(fileCheck.details))
	}
	checks = append(checks, fileCheck)
	
	// Start command validation
	cmdCheck := validationCheck{
		name:    "Start command",
		status:  "pass",
		message: "Looks good",
		details: []string{},
	}
	
	// Basic sanity checks
	if m.StartCommand != "" {
		if m.Files.Main != "" && !strings.Contains(m.StartCommand, filepath.Base(m.Files.Main)) &&
		   !strings.Contains(m.StartCommand, strings.TrimSuffix(filepath.Base(m.Files.Main), filepath.Ext(m.Files.Main))) {
			cmdCheck.status = "warning"
			cmdCheck.details = append(cmdCheck.details, 
				fmt.Sprintf("Start command doesn't reference main file (%s)", m.Files.Main))
		}
		
		// Check port in command matches manifest port
		portStr := fmt.Sprintf("%d", m.Port)
		if !strings.Contains(m.StartCommand, portStr) &&
		   !strings.Contains(m.StartCommand, "$PORT") &&
		   !strings.Contains(m.StartCommand, "${PORT}") {
			cmdCheck.details = append(cmdCheck.details, 
				fmt.Sprintf("Port %d not found in start command - ensure your app binds to this port", m.Port))
		}
	}
	
	if cmdCheck.status == "warning" {
		cmdCheck.message = "Review recommended"
	}
	checks = append(checks, cmdCheck)
	
	// Environment variables
	if len(m.Env.Required) > 0 {
		envCheck := validationCheck{
			name:    "Environment",
			status:  "warning",
			message: fmt.Sprintf("%d required variables", len(m.Env.Required)),
			details: []string{
				"Required variables: " + strings.Join(m.Env.Required, ", "),
				"These must be set during deployment",
			},
		}
		checks = append(checks, envCheck)
	}
	
	return checks
}

func displayValidationResults(checks []validationCheck, verbose bool) bool {
	hasErrors := false
	
	for _, check := range checks {
		// Display check result
		symbol := "✅"
		if check.status == "fail" {
			symbol = "❌"
			hasErrors = true
		} else if check.status == "warning" {
			symbol = "⚠️ "
		}
		
		fmt.Printf("%s %s: %s\n", symbol, check.name, check.message)
		
		// Show details if failed or verbose mode
		if (check.status != "pass" || verbose) && len(check.details) > 0 {
			for _, detail := range check.details {
				fmt.Printf("   ↳ %s\n", detail)
			}
		}
	}
	
	return hasErrors
}

func printValidationError(check, message string) {
	fmt.Printf("❌ %s: Failed\n", check)
	fmt.Printf("   ↳ %s\n", message)
}

func showDeploymentPreview(m *manifest.Manifest) {
	fmt.Println("🔍 Deployment Preview (DRY RUN - nothing will be deployed)")
	fmt.Println()
	
	fmt.Println("📦 Container configuration:")
	fmt.Printf("   - Runtime: %s\n", m.Runtime)
	fmt.Printf("   - Start command: %s\n", m.StartCommand)
	fmt.Printf("   - Exposed port: %d\n", m.Port)
	
	if m.Files.Requirements != "" {
		fmt.Printf("   - Dependencies: %s\n", m.Files.Requirements)
	}
	
	if m.Files.Dockerfile != "" {
		fmt.Printf("   - Custom Dockerfile: %s\n", m.Files.Dockerfile)
	} else {
		fmt.Println("   - Dockerfile: Auto-generated")
	}
	
	fmt.Println("\n🌐 Deployment settings:")
	fmt.Printf("   - API name: %s\n", m.Name)
	fmt.Printf("   - URL: https://%s-[random].api-direct.io\n", m.Name)
	fmt.Printf("   - Health check: %s\n", m.HealthCheck)
	
	if m.Scaling != nil {
		fmt.Printf("   - Auto-scaling: %d-%d instances (target CPU: %d%%)\n", 
			m.Scaling.Min, m.Scaling.Max, m.Scaling.TargetCPU)
	} else {
		fmt.Println("   - Auto-scaling: 1-10 instances (default)")
	}
	
	if m.Resources != nil {
		fmt.Printf("   - Resources: %s memory, %s CPU\n", 
			m.Resources.Memory, m.Resources.CPU)
	} else {
		fmt.Println("   - Resources: 512Mi memory, 250m CPU (default)")
	}
	
	if len(m.Endpoints) > 0 {
		fmt.Printf("\n📍 %d endpoints detected:\n", len(m.Endpoints))
		for i, endpoint := range m.Endpoints {
			if i < 5 {
				fmt.Printf("   - %s\n", endpoint)
			}
		}
		if len(m.Endpoints) > 5 {
			fmt.Printf("   ... and %d more\n", len(m.Endpoints)-5)
		}
	}
	
	fmt.Println("\n💰 Estimated cost: $0.20-0.50/month for 10K requests")
}

func isValidProjectName(name string) bool {
	// Same validation as in manifest package
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	
	// Check valid characters and format
	for i, char := range name {
		if !((char >= 'a' && char <= 'z') || 
		     (char >= '0' && char <= '9') || 
		     char == '-') {
			return false
		}
		
		// Must start with letter
		if i == 0 && !(char >= 'a' && char <= 'z') {
			return false
		}
		
		// Cannot end with hyphen
		if i == len(name)-1 && char == '-' {
			return false
		}
	}
	
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}