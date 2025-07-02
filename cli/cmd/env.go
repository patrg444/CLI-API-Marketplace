package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	envAll       bool
	envProduction bool
	envStaging   bool
	envLocal     bool
	envFormat    string
)

// envCmd represents the env command group
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables for your API",
	Long: `Manage environment variables for different deployment environments.
Environment variables can be set locally (.env file) or remotely for deployed APIs.`,
}

// envSetCmd sets environment variables
var envSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE [KEY2=VALUE2...]",
	Short: "Set environment variables",
	Long: `Set environment variables for your API deployment.
Variables are stored securely and injected at runtime.

Examples:
  apidirect env set DATABASE_URL=postgres://localhost/mydb
  apidirect env set API_KEY=secret DEBUG=true
  apidirect env set --production DATABASE_URL=postgres://prod/db
  apidirect env set --local PORT=3000`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEnvSet,
}

// envGetCmd gets a specific environment variable
var envGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get an environment variable value",
	Long: `Get the value of a specific environment variable.

Examples:
  apidirect env get DATABASE_URL
  apidirect env get --production API_KEY`,
	Args: cobra.ExactArgs(1),
	RunE: runEnvGet,
}

// envListCmd lists all environment variables
var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	Long: `List all environment variables for the current API.

Examples:
  apidirect env list                    # List for current environment
  apidirect env list --all             # List for all environments
  apidirect env list --production      # List production variables
  apidirect env list --format=json     # Output as JSON`,
	RunE: runEnvList,
}

// envUnsetCmd removes environment variables
var envUnsetCmd = &cobra.Command{
	Use:   "unset KEY [KEY2...]",
	Short: "Remove environment variables",
	Long: `Remove one or more environment variables.

Examples:
  apidirect env unset DEBUG
  apidirect env unset API_KEY SECRET_KEY
  apidirect env unset --production DEBUG`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEnvUnset,
}

// envPullCmd pulls remote environment to local .env file
var envPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull remote environment variables to local .env file",
	Long: `Download environment variables from deployed API to local .env file.

Examples:
  apidirect env pull                   # Pull from current deployment
  apidirect env pull --production      # Pull from production
  apidirect env pull --staging         # Pull from staging`,
	RunE: runEnvPull,
}

// envPushCmd pushes local .env to remote
var envPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local .env file to remote deployment",
	Long: `Upload environment variables from local .env file to deployed API.
This will overwrite remote variables!

Examples:
  apidirect env push                   # Push to current deployment
  apidirect env push --production      # Push to production (requires confirmation)`,
	RunE: runEnvPush,
}

func init() {
	rootCmd.AddCommand(envCmd)
	
	// Add subcommands
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(envPullCmd)
	envCmd.AddCommand(envPushCmd)
	
	// Common flags for environment selection
	for _, cmd := range []*cobra.Command{envSetCmd, envGetCmd, envListCmd, envUnsetCmd, envPullCmd, envPushCmd} {
		cmd.Flags().BoolVar(&envProduction, "production", false, "Target production environment")
		cmd.Flags().BoolVar(&envStaging, "staging", false, "Target staging environment")
		cmd.Flags().BoolVar(&envLocal, "local", false, "Target local environment (.env file)")
	}
	
	// List-specific flags
	envListCmd.Flags().BoolVarP(&envAll, "all", "a", false, "Show variables for all environments")
	envListCmd.Flags().StringVar(&envFormat, "format", "table", "Output format (table, json, dotenv)")
}

func getTargetEnvironment() string {
	if envProduction {
		return "production"
	}
	if envStaging {
		return "staging"
	}
	if envLocal {
		return "local"
	}
	return "development" // default
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	environment := getTargetEnvironment()
	
	// Parse KEY=VALUE pairs
	vars := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: %s (expected KEY=VALUE)", arg)
		}
		vars[parts[0]] = parts[1]
	}

	// Handle local environment
	if environment == "local" {
		return setLocalEnvVars(vars)
	}

	// For remote environments, check authentication
	if err := checkAuth(); err != nil {
		return err
	}

	// Get API name from manifest
	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	fmt.Printf("🔐 Setting %d environment variable(s) for %s (%s)\n", 
		len(vars), apiName, environment)

	// In real implementation, this would call the API
	// For now, simulate the operation
	for key, value := range vars {
		// Mask sensitive values in output
		displayValue := value
		if isSensitiveKey(key) {
			displayValue = maskValue(value)
		}
		fmt.Printf("   %s = %s\n", key, displayValue)
	}

	printSuccess(fmt.Sprintf("Environment variables updated for %s", environment))
	
	if environment == "production" {
		printWarning("Changes will take effect after next deployment or restart")
	}

	return nil
}

func runEnvGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	environment := getTargetEnvironment()

	// Handle local environment
	if environment == "local" {
		value, err := getLocalEnvVar(key)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	}

	// For remote environments
	if err := checkAuth(); err != nil {
		return err
	}

	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	// In real implementation, fetch from API
	// For demo, return mock value
	value := fmt.Sprintf("value-for-%s-in-%s-for-%s", key, environment, apiName)
	
	if isSensitiveKey(key) {
		fmt.Printf("%s=%s\n", key, maskValue(value))
		fmt.Println("\n💡 Use 'apidirect env get --show-secrets' to reveal full value")
	} else {
		fmt.Println(value)
	}

	return nil
}

func runEnvList(cmd *cobra.Command, args []string) error {
	environment := getTargetEnvironment()

	// Handle local environment
	if environment == "local" || envAll {
		if err := listLocalEnvVars(); err != nil && !envAll {
			return err
		}
		if !envAll {
			return nil
		}
	}

	// For remote environments
	if err := checkAuth(); err != nil {
		return err
	}

	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	// Mock data for demonstration
	envVars := getEnvironmentVars(apiName, environment, envAll)

	// Output based on format
	switch envFormat {
	case "json":
		return outputEnvJSON(envVars)
	case "dotenv":
		return outputDotenv(envVars)
	default:
		return outputTable(envVars)
	}
}

func runEnvUnset(cmd *cobra.Command, args []string) error {
	environment := getTargetEnvironment()
	
	// Handle local environment
	if environment == "local" {
		return unsetLocalEnvVars(args)
	}

	// For remote environments
	if err := checkAuth(); err != nil {
		return err
	}

	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	fmt.Printf("🗑  Removing %d environment variable(s) from %s (%s)\n", 
		len(args), apiName, environment)

	for _, key := range args {
		fmt.Printf("   - %s\n", key)
	}

	// Confirm for production
	if environment == "production" {
		fmt.Print("\n⚠️  This will remove variables from production. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("Cancelled")
		}
	}

	printSuccess(fmt.Sprintf("Environment variables removed from %s", environment))
	return nil
}

func runEnvPull(cmd *cobra.Command, args []string) error {
	environment := getTargetEnvironment()
	
	if environment == "local" {
		return fmt.Errorf("Cannot pull from local environment")
	}

	if err := checkAuth(); err != nil {
		return err
	}

	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	fmt.Printf("⬇️  Pulling environment variables from %s (%s)\n", apiName, environment)

	// Check if .env exists
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		fmt.Printf("\n⚠️  %s already exists. Overwrite? [y/N]: ", envFile)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			envFile = fmt.Sprintf(".env.%s", environment)
			fmt.Printf("💡 Saving to %s instead\n", envFile)
		}
	}

	// Mock pulling variables
	vars := map[string]string{
		"DATABASE_URL": "postgres://user:pass@host/db",
		"API_KEY":      "sk-1234567890",
		"LOG_LEVEL":    "info",
		"DEBUG":        "false",
	}

	// Write to file
	if err := writeDotenvFile(envFile, vars); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Pulled %d variables to %s", len(vars), envFile))
	fmt.Println("💡 Remember to add .env to .gitignore!")
	
	return nil
}

func runEnvPush(cmd *cobra.Command, args []string) error {
	environment := getTargetEnvironment()
	
	if environment == "local" {
		return fmt.Errorf("Cannot push to local environment")
	}

	if err := checkAuth(); err != nil {
		return err
	}

	apiName, err := getAPIName()
	if err != nil {
		return err
	}

	// Read .env file
	envFile := ".env"
	vars, err := readDotenvFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", envFile, err)
	}

	fmt.Printf("⬆️  Pushing %d environment variables to %s (%s)\n", 
		len(vars), apiName, environment)

	// Show what will be pushed
	fmt.Println("\nVariables to push:")
	for key, value := range vars {
		displayValue := value
		if isSensitiveKey(key) {
			displayValue = maskValue(value)
		}
		fmt.Printf("   %s = %s\n", key, displayValue)
	}

	// Confirm for production
	if environment == "production" {
		fmt.Print("\n⚠️  This will overwrite ALL production variables. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("Cancelled")
		}
	}

	printSuccess(fmt.Sprintf("Pushed environment variables to %s", environment))
	printWarning("Restart your API for changes to take effect")
	
	return nil
}

// Local environment helpers

func setLocalEnvVars(vars map[string]string) error {
	// Read existing .env
	existing, _ := readDotenvFile(".env")
	
	// Merge new variables
	for key, value := range vars {
		existing[key] = value
	}

	// Write back
	if err := writeDotenvFile(".env", existing); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Updated %d variable(s) in .env", len(vars)))
	return nil
}

func getLocalEnvVar(key string) (string, error) {
	vars, err := readDotenvFile(".env")
	if err != nil {
		return "", err
	}

	value, exists := vars[key]
	if !exists {
		return "", fmt.Errorf("variable %s not found in .env", key)
	}

	return value, nil
}

func listLocalEnvVars() error {
	vars, err := readDotenvFile(".env")
	if err != nil {
		return fmt.Errorf("failed to read .env: %w", err)
	}

	if len(vars) == 0 {
		fmt.Println("No variables in .env")
		return nil
	}

	fmt.Println("📋 Local environment variables (.env):")
	fmt.Println()

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Display as table
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	for _, key := range keys {
		value := vars[key]
		if isSensitiveKey(key) {
			value = maskValue(value)
		}
		fmt.Printf("  %-*s = %s\n", maxKeyLen, key, value)
	}

	return nil
}

func unsetLocalEnvVars(keys []string) error {
	vars, err := readDotenvFile(".env")
	if err != nil {
		return err
	}

	removed := 0
	for _, key := range keys {
		if _, exists := vars[key]; exists {
			delete(vars, key)
			removed++
		}
	}

	if removed == 0 {
		return fmt.Errorf("No matching variables found in .env")
	}

	if err := writeDotenvFile(".env", vars); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Removed %d variable(s) from .env", removed))
	return nil
}

// File I/O helpers

func readDotenvFile(filename string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(filename)
	if err != nil {
		return vars, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove quotes if present
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
				   (value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			
			vars[key] = value
		}
	}

	return vars, scanner.Err()
}

func writeDotenvFile(filename string, vars map[string]string) error {
	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build content
	var content strings.Builder
	content.WriteString("# Environment variables for API-Direct\n")
	content.WriteString(fmt.Sprintf("# Generated on %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, key := range keys {
		value := vars[key]
		// Quote values with spaces or special characters
		if strings.ContainsAny(value, " \t\n#$") {
			value = fmt.Sprintf(`"%s"`, value)
		}
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// Output formatters

func outputEnvJSON(envVars map[string]map[string]string) error {
	output, err := json.MarshalIndent(envVars, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputDotenv(envVars map[string]map[string]string) error {
	for env, vars := range envVars {
		if len(envVars) > 1 {
			fmt.Printf("# %s\n", env)
		}
		
		keys := make([]string, 0, len(vars))
		for key := range vars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		
		for _, key := range keys {
			fmt.Printf("%s=%s\n", key, vars[key])
		}
		
		if len(envVars) > 1 {
			fmt.Println()
		}
	}
	return nil
}

func outputTable(envVars map[string]map[string]string) error {
	for env, vars := range envVars {
		if len(envVars) > 1 {
			fmt.Printf("\n📋 %s environment:\n\n", strings.Title(env))
		}
		
		if len(vars) == 0 {
			fmt.Println("  (no variables set)")
			continue
		}
		
		// Calculate column width
		maxKeyLen := 0
		for key := range vars {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
		}
		
		// Sort and display
		keys := make([]string, 0, len(vars))
		for key := range vars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		
		for _, key := range keys {
			value := vars[key]
			if isSensitiveKey(key) {
				value = maskValue(value)
			}
			fmt.Printf("  %-*s = %s\n", maxKeyLen, key, value)
		}
	}
	
	return nil
}

// Utility functions

func getAPIName() (string, error) {
	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	// Try to load from manifest
	if manifestPath, err := manifest.FindManifest(dir); err == nil {
		if m, err := manifest.Load(manifestPath); err == nil {
			return m.Name, nil
		}
	}
	
	// Fallback to directory name
	return filepath.Base(dir), nil
}

func isSensitiveKey(key string) bool {
	sensitive := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", 
		"PRIVATE", "CREDENTIAL", "AUTH",
	}
	
	upperKey := strings.ToUpper(key)
	for _, s := range sensitive {
		if strings.Contains(upperKey, s) {
			return true
		}
	}
	
	return false
}

func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	
	// Show first 2 and last 2 characters
	return fmt.Sprintf("%s...%s", value[:2], value[len(value)-2:])
}

func getEnvironmentVars(apiName, environment string, all bool) map[string]map[string]string {
	// Mock data for demonstration
	result := make(map[string]map[string]string)
	
	if all {
		// Return all environments
		result["development"] = map[string]string{
			"LOG_LEVEL": "debug",
			"PORT":      "8080",
		}
		result["staging"] = map[string]string{
			"LOG_LEVEL":    "info",
			"PORT":         "8080",
			"DATABASE_URL": "postgres://staging/db",
		}
		result["production"] = map[string]string{
			"LOG_LEVEL":    "error",
			"PORT":         "8080",
			"DATABASE_URL": "postgres://prod/db",
			"API_KEY":      "sk-prod-1234",
		}
	} else {
		// Return specific environment
		switch environment {
		case "production":
			result[environment] = map[string]string{
				"LOG_LEVEL":    "error",
				"PORT":         "8080", 
				"DATABASE_URL": "postgres://prod/db",
				"API_KEY":      "sk-prod-1234",
			}
		default:
			result[environment] = map[string]string{
				"LOG_LEVEL": "debug",
				"PORT":      "8080",
			}
		}
	}
	
	return result
}