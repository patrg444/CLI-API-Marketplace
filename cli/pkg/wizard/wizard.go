package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// APITemplate represents a template for API creation
type APITemplate struct {
	ID          string
	Name        string
	Description string
	Runtime     string
	Category    string
	Features    []string
}

// WizardConfig holds the configuration collected from the wizard
type WizardConfig struct {
	APIName     string
	Template    APITemplate
	Runtime     string
	Description string
	Features    []string
}

var (
	// Color functions for better UX
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
	
	// Available templates
	templates = []APITemplate{
		{
			ID:          "basic-rest",
			Name:        "Basic REST API",
			Description: "Simple REST API with CRUD operations",
			Runtime:     "python3.9",
			Category:    "Web API",
			Features:    []string{"REST endpoints", "JSON responses", "Basic validation"},
		},
		{
			ID:          "crud-database",
			Name:        "CRUD with Database",
			Description: "REST API with database operations (PostgreSQL)",
			Runtime:     "python3.9",
			Category:    "Database API",
			Features:    []string{"Database models", "CRUD operations", "Data validation", "Migrations"},
		},
		{
			ID:          "webhook-receiver",
			Name:        "Webhook Receiver",
			Description: "API for receiving and processing webhooks",
			Runtime:     "python3.9",
			Category:    "Integration",
			Features:    []string{"Webhook validation", "Event processing", "Queue integration"},
		},
		{
			ID:          "ml-model-serving",
			Name:        "ML Model Serving",
			Description: "Serve machine learning models via REST API",
			Runtime:     "python3.9",
			Category:    "Machine Learning",
			Features:    []string{"Model loading", "Prediction endpoints", "Input validation", "Batch processing"},
		},
		{
			ID:          "data-processing",
			Name:        "Data Processing API",
			Description: "API for data transformation and processing",
			Runtime:     "python3.9",
			Category:    "Data Processing",
			Features:    []string{"File upload", "Data transformation", "Export formats", "Async processing"},
		},
		{
			ID:          "auth-service",
			Name:        "Authentication Service",
			Description: "User authentication and authorization API",
			Runtime:     "python3.9",
			Category:    "Authentication",
			Features:    []string{"User registration", "JWT tokens", "Password reset", "Role-based access"},
		},
		{
			ID:          "graphql-api",
			Name:        "GraphQL API",
			Description: "GraphQL API with schema and resolvers",
			Runtime:     "python3.9",
			Category:    "GraphQL",
			Features:    []string{"GraphQL schema", "Query resolvers", "Mutations", "Subscriptions"},
		},
		{
			ID:          "microservice",
			Name:        "Microservice Template",
			Description: "Production-ready microservice with monitoring",
			Runtime:     "python3.9",
			Category:    "Microservice",
			Features:    []string{"Health checks", "Metrics", "Logging", "Circuit breaker"},
		},
	}
)

// RunInteractiveWizard runs the interactive setup wizard
func RunInteractiveWizard() (*WizardConfig, error) {
	config := &WizardConfig{}
	
	printWelcome()
	
	// Step 1: Get API name
	apiName, err := promptAPIName()
	if err != nil {
		return nil, err
	}
	config.APIName = apiName
	
	// Step 2: Choose template
	template, err := promptTemplate()
	if err != nil {
		return nil, err
	}
	config.Template = template
	config.Runtime = template.Runtime
	
	// Step 3: Get description
	description, err := promptDescription(apiName)
	if err != nil {
		return nil, err
	}
	config.Description = description
	
	// Step 4: Choose runtime (if different from template default)
	runtime, err := promptRuntime(template.Runtime)
	if err != nil {
		return nil, err
	}
	config.Runtime = runtime
	
	// Step 5: Additional features
	features, err := promptAdditionalFeatures()
	if err != nil {
		return nil, err
	}
	config.Features = features
	
	// Step 6: Show summary and confirm
	confirmed, err := showSummaryAndConfirm(config)
	if err != nil {
		return nil, err
	}
	
	if !confirmed {
		fmt.Println(yellow("Setup cancelled."))
		return nil, fmt.Errorf("setup cancelled by user")
	}
	
	return config, nil
}

func printWelcome() {
	fmt.Println()
	fmt.Println(bold(cyan("🚀 Welcome to API-Direct Interactive Setup!")))
	fmt.Println()
	fmt.Println("This wizard will help you create a new API project in minutes.")
	fmt.Println("You can always customize the generated code later.")
	fmt.Println()
}

func promptAPIName() (string, error) {
	fmt.Print(bold("📝 What's your API name? "))
	fmt.Print(cyan("(lowercase, hyphens allowed): "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	apiName := strings.TrimSpace(input)
	
	// Validate API name
	if !isValidAPIName(apiName) {
		fmt.Println(red("❌ Invalid API name. Use only lowercase letters, numbers, and hyphens."))
		return promptAPIName() // Retry
	}
	
	// Check if directory exists
	if _, err := os.Stat(apiName); err == nil {
		fmt.Printf(red("❌ Directory '%s' already exists.\n"), apiName)
		return promptAPIName() // Retry
	}
	
	fmt.Printf(green("✅ API name: %s\n\n"), apiName)
	return apiName, nil
}

func promptTemplate() (APITemplate, error) {
	fmt.Println(bold("🎨 Choose a template for your API:"))
	fmt.Println()
	
	// Display templates
	for i, template := range templates {
		fmt.Printf("%s %d. %s\n", cyan("▶"), i+1, bold(template.Name))
		fmt.Printf("   %s\n", template.Description)
		fmt.Printf("   %s Category: %s | Runtime: %s\n", 
			yellow("ℹ"), template.Category, template.Runtime)
		fmt.Printf("   Features: %s\n", strings.Join(template.Features, ", "))
		fmt.Println()
	}
	
	fmt.Print(bold("Enter your choice (1-" + strconv.Itoa(len(templates)) + "): "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return APITemplate{}, err
	}
	
	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(templates) {
		fmt.Println(red("❌ Invalid choice. Please enter a number between 1 and " + strconv.Itoa(len(templates))))
		return promptTemplate() // Retry
	}
	
	selectedTemplate := templates[choice-1]
	fmt.Printf(green("✅ Selected: %s\n\n"), selectedTemplate.Name)
	
	return selectedTemplate, nil
}

func promptDescription(apiName string) (string, error) {
	fmt.Printf(bold("📄 Brief description for %s "), apiName)
	fmt.Print(cyan("(optional, press Enter to skip): "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	description := strings.TrimSpace(input)
	if description == "" {
		description = fmt.Sprintf("API project: %s", apiName)
	}
	
	fmt.Printf(green("✅ Description: %s\n\n"), description)
	return description, nil
}

func promptRuntime(defaultRuntime string) (string, error) {
	runtimes := []string{"python3.9", "python3.10", "python3.11", "nodejs18", "nodejs20"}
	
	fmt.Printf(bold("🐍 Choose runtime "))
	fmt.Printf(cyan("(default: %s): "), defaultRuntime)
	fmt.Println()
	
	for i, runtime := range runtimes {
		marker := " "
		if runtime == defaultRuntime {
			marker = cyan("▶")
		}
		fmt.Printf("%s %d. %s", marker, i+1, runtime)
		if runtime == defaultRuntime {
			fmt.Print(green(" (recommended)"))
		}
		fmt.Println()
	}
	
	fmt.Print(bold("\nEnter your choice (1-" + strconv.Itoa(len(runtimes)) + ") or press Enter for default: "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Printf(green("✅ Runtime: %s (default)\n\n"), defaultRuntime)
		return defaultRuntime, nil
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(runtimes) {
		fmt.Println(red("❌ Invalid choice. Using default runtime."))
		return defaultRuntime, nil
	}
	
	selectedRuntime := runtimes[choice-1]
	fmt.Printf(green("✅ Runtime: %s\n\n"), selectedRuntime)
	
	return selectedRuntime, nil
}

func promptAdditionalFeatures() ([]string, error) {
	additionalFeatures := []string{
		"Docker support",
		"GitHub Actions CI/CD",
		"API documentation generation",
		"Rate limiting",
		"CORS configuration",
		"Environment-based configuration",
		"Logging and monitoring",
		"Unit test examples",
	}
	
	fmt.Println(bold("🔧 Additional features (optional):"))
	fmt.Println(cyan("Select features to include (comma-separated numbers, or press Enter to skip):"))
	fmt.Println()
	
	for i, feature := range additionalFeatures {
		fmt.Printf("  %d. %s\n", i+1, feature)
	}
	
	fmt.Print(bold("\nYour choice: "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println(green("✅ No additional features selected\n"))
		return []string{}, nil
	}
	
	// Parse comma-separated choices
	choices := strings.Split(input, ",")
	selectedFeatures := []string{}
	
	for _, choice := range choices {
		choice = strings.TrimSpace(choice)
		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(additionalFeatures) {
			fmt.Printf(yellow("⚠️  Skipping invalid choice: %s\n"), choice)
			continue
		}
		selectedFeatures = append(selectedFeatures, additionalFeatures[index-1])
	}
	
	if len(selectedFeatures) > 0 {
		fmt.Printf(green("✅ Additional features: %s\n\n"), strings.Join(selectedFeatures, ", "))
	} else {
		fmt.Println(green("✅ No additional features selected\n"))
	}
	
	return selectedFeatures, nil
}

func showSummaryAndConfirm(config *WizardConfig) (bool, error) {
	fmt.Println(bold(cyan("📋 Project Summary:")))
	fmt.Println()
	fmt.Printf("  %s %s\n", bold("API Name:"), config.APIName)
	fmt.Printf("  %s %s\n", bold("Template:"), config.Template.Name)
	fmt.Printf("  %s %s\n", bold("Runtime:"), config.Runtime)
	fmt.Printf("  %s %s\n", bold("Description:"), config.Description)
	
	if len(config.Features) > 0 {
		fmt.Printf("  %s %s\n", bold("Features:"), strings.Join(config.Features, ", "))
	}
	
	fmt.Println()
	fmt.Printf("  %s %s\n", bold("Template Features:"), strings.Join(config.Template.Features, ", "))
	fmt.Println()
	
	fmt.Print(bold("🚀 Create this API project? "))
	fmt.Print(cyan("(y/N): "))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	
	response := strings.ToLower(strings.TrimSpace(input))
	confirmed := response == "y" || response == "yes"
	
	if confirmed {
		fmt.Println(green("✅ Creating project...\n"))
	}
	
	return confirmed, nil
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

// GetTemplateByID returns a template by its ID
func GetTemplateByID(id string) (APITemplate, bool) {
	for _, template := range templates {
		if template.ID == id {
			return template, true
		}
	}
	return APITemplate{}, false
}

// ListTemplates returns all available templates
func ListTemplates() []APITemplate {
	return templates
}
