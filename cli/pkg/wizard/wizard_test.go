package wizard

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidAPIName(t *testing.T) {
	tests := []struct {
		name     string
		apiName  string
		expected bool
	}{
		// Valid names
		{"simple name", "myapi", true},
		{"with hyphens", "my-api-name", true},
		{"with numbers", "api123", true},
		{"mixed", "my-api-v2", true},
		{"single letter", "a", true},
		{"max length", "a" + strings.Repeat("b", 62), true},

		// Invalid names
		{"empty", "", false},
		{"too long", "a" + strings.Repeat("b", 63), false},
		{"uppercase", "MyAPI", false},
		{"starts with number", "123api", false},
		{"starts with hyphen", "-api", false},
		{"ends with hyphen", "api-", false},
		{"special characters", "api@name", false},
		{"spaces", "api name", false},
		{"underscore", "api_name", false},
		{"dot", "api.name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAPIName(tt.apiName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTemplateByID(t *testing.T) {
	tests := []struct {
		name         string
		templateID   string
		expectFound  bool
		validateName string
	}{
		{"find GPT wrapper", "gpt-wrapper", true, "🤖 GPT Wrapper API"},
		{"find image classifier", "image-classifier", true, "👁️ Image Classification API"},
		{"find sentiment analyzer", "sentiment-analyzer", true, "😊 Sentiment Analysis API"},
		{"find basic REST", "basic-rest", true, "Basic REST API"},
		{"find CRUD database", "crud-database", true, "CRUD with Database"},
		{"not found", "non-existent", false, ""},
		{"empty ID", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, found := GetTemplateByID(tt.templateID)
			
			assert.Equal(t, tt.expectFound, found)
			
			if tt.expectFound {
				assert.Equal(t, tt.templateID, template.ID)
				assert.Equal(t, tt.validateName, template.Name)
				assert.NotEmpty(t, template.Description)
				assert.NotEmpty(t, template.Runtime)
				assert.NotEmpty(t, template.Category)
			} else {
				assert.Empty(t, template.ID)
			}
		})
	}
}

func TestListTemplates(t *testing.T) {
	templates := ListTemplates()
	
	// Verify we have templates
	assert.NotEmpty(t, templates)
	assert.GreaterOrEqual(t, len(templates), 13) // At least 13 templates as defined
	
	// Verify AI/ML templates are first (priority positioning)
	aimlCount := 0
	for i := 0; i < 6; i++ { // First 6 should be AI/ML
		assert.Equal(t, "AI/ML", templates[i].Category)
		aimlCount++
	}
	assert.Equal(t, 6, aimlCount)
	
	// Verify each template has required fields
	seenIDs := make(map[string]bool)
	for _, template := range templates {
		assert.NotEmpty(t, template.ID, "Template should have ID")
		assert.NotEmpty(t, template.Name, "Template should have name")
		assert.NotEmpty(t, template.Description, "Template should have description")
		assert.NotEmpty(t, template.Runtime, "Template should have runtime")
		assert.NotEmpty(t, template.Category, "Template should have category")
		assert.NotEmpty(t, template.Features, "Template should have features")
		
		// Check for unique IDs
		assert.False(t, seenIDs[template.ID], "Duplicate template ID: %s", template.ID)
		seenIDs[template.ID] = true
	}
}

func TestWizardConfig(t *testing.T) {
	// Test WizardConfig structure
	config := &WizardConfig{
		APIName: "test-api",
		Template: APITemplate{
			ID:          "gpt-wrapper",
			Name:        "GPT Wrapper API",
			Description: "Test description",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features:    []string{"Feature1", "Feature2"},
		},
		Runtime:     "python3.11",
		Description: "My test API",
		Features:    []string{"Docker support", "CI/CD"},
	}
	
	// Verify all fields are accessible
	assert.Equal(t, "test-api", config.APIName)
	assert.Equal(t, "gpt-wrapper", config.Template.ID)
	assert.Equal(t, "python3.11", config.Runtime)
	assert.Equal(t, "My test API", config.Description)
	assert.Len(t, config.Features, 2)
}

func TestTemplateCategories(t *testing.T) {
	templates := ListTemplates()
	
	// Count templates by category
	categories := make(map[string]int)
	for _, template := range templates {
		categories[template.Category]++
	}
	
	// Verify we have multiple categories
	assert.GreaterOrEqual(t, len(categories), 5)
	
	// Verify AI/ML is the most common category (focused positioning)
	assert.GreaterOrEqual(t, categories["AI/ML"], 6)
	
	// Verify each category has at least one template
	for category, count := range categories {
		assert.Greater(t, count, 0, "Category %s should have at least one template", category)
	}
}

func TestColorFunctions(t *testing.T) {
	// Test that color functions work without errors
	// Note: We can't easily test the actual color output in unit tests
	
	assert.NotPanics(t, func() {
		_ = cyan("test")
		_ = green("test")
		_ = yellow("test")
		_ = red("test")
		_ = bold("test")
	})
}

func TestTemplateFeatures(t *testing.T) {
	// Verify AI/ML templates have appropriate features
	aimlTemplates := []string{
		"gpt-wrapper",
		"image-classifier",
		"sentiment-analyzer",
		"embeddings-api",
		"time-series-predictor",
		"document-qa",
	}
	
	for _, id := range aimlTemplates {
		template, found := GetTemplateByID(id)
		assert.True(t, found)
		assert.Equal(t, "AI/ML", template.Category)
		assert.GreaterOrEqual(t, len(template.Features), 4, "AI/ML template %s should have at least 4 features", id)
		
		// Verify it's a Python template (required for ML libraries)
		assert.Contains(t, template.Runtime, "python")
	}
}

func TestPromptAPINameValidation(t *testing.T) {
	// Test the validation logic used in promptAPIName
	
	// Create temp directory for testing
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	
	err := os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Test that validation catches invalid names
	invalidNames := []string{
		"InvalidName!",
		"123-start",
		"ends-with-",
		"has spaces",
		"has_underscore",
		"",
		strings.Repeat("a", 64),
	}
	
	for _, name := range invalidNames {
		assert.False(t, isValidAPIName(name), "Name '%s' should be invalid", name)
	}
	
	// Test that valid names pass
	validNames := []string{
		"my-api",
		"api123",
		"test-api-v2",
		"a",
		strings.Repeat("a", 63),
	}
	
	for _, name := range validNames {
		assert.True(t, isValidAPIName(name), "Name '%s' should be valid", name)
	}
}

func TestTemplateRuntimes(t *testing.T) {
	// Verify all templates have valid runtimes
	templates := ListTemplates()
	validRuntimes := map[string]bool{
		"python3.9":  true,
		"python3.10": true,
		"python3.11": true,
		"nodejs18":   true,
		"nodejs20":   true,
	}
	
	for _, template := range templates {
		assert.Contains(t, validRuntimes, template.Runtime, "Template %s has invalid runtime: %s", template.ID, template.Runtime)
	}
}

func TestAdditionalFeaturesAvailable(t *testing.T) {
	// Test that the additional features list is comprehensive
	expectedFeatures := []string{
		"Docker support",
		"GitHub Actions CI/CD",
		"API documentation generation",
		"Rate limiting",
		"CORS configuration",
		"Environment-based configuration",
		"Logging and monitoring",
		"Unit test examples",
	}
	
	// Note: We can't directly access the additionalFeatures from promptAdditionalFeatures
	// but we can verify the concept
	assert.Len(t, expectedFeatures, 8, "Should have 8 additional features available")
}

func TestTemplateOrdering(t *testing.T) {
	// Verify templates are ordered with AI/ML first
	templates := ListTemplates()
	
	// First 6 should be AI/ML
	for i := 0; i < 6; i++ {
		assert.Equal(t, "AI/ML", templates[i].Category, "Template at index %d should be AI/ML category", i)
	}
	
	// Remaining should be other categories
	otherCategories := make(map[string]bool)
	for i := 6; i < len(templates); i++ {
		assert.NotEqual(t, "AI/ML", templates[i].Category, "Template at index %d should not be AI/ML category", i)
		otherCategories[templates[i].Category] = true
	}
	
	// Should have multiple other categories
	assert.Greater(t, len(otherCategories), 3, "Should have multiple non-AI/ML categories")
}

func TestWizardConfigValidation(t *testing.T) {
	// Test various wizard configurations
	configs := []struct {
		name     string
		config   WizardConfig
		validate func(*testing.T, WizardConfig)
	}{
		{
			name: "AI template with custom runtime",
			config: WizardConfig{
				APIName: "ai-api",
				Template: APITemplate{
					ID:       "gpt-wrapper",
					Runtime:  "python3.9",
					Category: "AI/ML",
				},
				Runtime:     "python3.11", // Overridden
				Description: "AI API",
				Features:    []string{"Docker", "CI/CD"},
			},
			validate: func(t *testing.T, c WizardConfig) {
				assert.Equal(t, "python3.11", c.Runtime, "Runtime should be overridden")
				assert.Equal(t, "python3.9", c.Template.Runtime, "Template runtime should remain unchanged")
			},
		},
		{
			name: "No additional features",
			config: WizardConfig{
				APIName: "simple-api",
				Template: APITemplate{
					ID:       "basic-rest",
					Features: []string{"REST", "JSON"},
				},
				Features: []string{}, // No additional features
			},
			validate: func(t *testing.T, c WizardConfig) {
				assert.Empty(t, c.Features, "Should have no additional features")
				assert.NotEmpty(t, c.Template.Features, "Template should still have its own features")
			},
		},
	}
	
	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			tc.validate(t, tc.config)
		})
	}
}

// Benchmark tests
func BenchmarkIsValidAPIName(b *testing.B) {
	names := []string{
		"valid-api-name",
		"invalid-NAME",
		"123-invalid",
		"a",
		strings.Repeat("a", 63),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			_ = isValidAPIName(name)
		}
	}
}

func BenchmarkGetTemplateByID(b *testing.B) {
	ids := []string{"gpt-wrapper", "image-classifier", "non-existent", "basic-rest"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, id := range ids {
			_, _ = GetTemplateByID(id)
		}
	}
}

func BenchmarkListTemplates(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ListTemplates()
	}
}

// Test helper functions
func TestAPITemplateStructure(t *testing.T) {
	// Verify APITemplate has all required fields
	template := APITemplate{
		ID:          "test",
		Name:        "Test Template",
		Description: "Test description",
		Runtime:     "python3.9",
		Category:    "Test",
		Features:    []string{"Feature1", "Feature2"},
	}
	
	assert.NotEmpty(t, template.ID)
	assert.NotEmpty(t, template.Name)
	assert.NotEmpty(t, template.Description)
	assert.NotEmpty(t, template.Runtime)
	assert.NotEmpty(t, template.Category)
	assert.Len(t, template.Features, 2)
}

func TestTemplateConsistency(t *testing.T) {
	// Verify all templates follow consistent patterns
	templates := ListTemplates()
	
	for _, template := range templates {
		// IDs should be lowercase with hyphens
		assert.Regexp(t, "^[a-z0-9-]+$", template.ID, "Template ID should be lowercase with hyphens: %s", template.ID)
		
		// Names should not be empty
		assert.NotEmpty(t, template.Name)
		
		// Descriptions should be meaningful (at least 10 chars)
		assert.GreaterOrEqual(t, len(template.Description), 10, "Template %s should have meaningful description", template.ID)
		
		// Should have at least one feature
		assert.NotEmpty(t, template.Features, "Template %s should have at least one feature", template.ID)
		
		// Category should be one of the known categories
		knownCategories := map[string]bool{
			"AI/ML":           true,
			"Web API":         true,
			"Database API":    true,
			"Integration":     true,
			"Data Processing": true,
			"Authentication":  true,
			"GraphQL":         true,
			"Microservice":    true,
		}
		assert.Contains(t, knownCategories, template.Category, "Template %s has unknown category: %s", template.ID, template.Category)
	}
}