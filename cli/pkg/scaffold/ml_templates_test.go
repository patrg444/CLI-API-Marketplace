package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMLTemplates(t *testing.T) {
	templates := GetMLTemplates()
	
	// Verify we have the expected number of ML templates
	assert.Len(t, templates, 6)
	
	// Verify all templates are in AI/ML category
	for _, template := range templates {
		assert.Equal(t, "AI/ML", template.Category)
		assert.Equal(t, "python3.9", template.Runtime) // All ML templates use Python
		assert.NotEmpty(t, template.ID)
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Features)
	}
	
	// Verify specific templates exist
	expectedIDs := []string{
		"gpt-wrapper",
		"image-classifier",
		"sentiment-analyzer",
		"embeddings-api",
		"time-series-predictor",
		"document-qa",
	}
	
	actualIDs := make([]string, len(templates))
	for i, template := range templates {
		actualIDs[i] = template.ID
	}
	
	assert.ElementsMatch(t, expectedIDs, actualIDs)
}

func TestGetMLTemplateConfig(t *testing.T) {
	tests := []struct {
		name         string
		template     APITemplate
		apiName      string
		runtime      string
		validateYAML func(*testing.T, string)
	}{
		{
			name: "GPT wrapper config",
			template: APITemplate{
				ID: "gpt-wrapper",
			},
			apiName: "test-gpt-api",
			runtime: "python3.11",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "name: test-gpt-api")
				assert.Contains(t, config, "runtime: python3.11")
				assert.Contains(t, config, "/complete")
				assert.Contains(t, config, "/chat")
				assert.Contains(t, config, "OPENAI_API_KEY")
				assert.Contains(t, config, "pricing:")
				assert.Contains(t, config, "aws:")
				assert.Contains(t, config, "cpu: 1024")
				assert.Contains(t, config, "memory: 2048")
			},
		},
		{
			name: "Image classifier config",
			template: APITemplate{
				ID: "image-classifier",
			},
			apiName: "vision-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "name: vision-api")
				assert.Contains(t, config, "/classify")
				assert.Contains(t, config, "/classify/batch")
				assert.Contains(t, config, "MODEL_NAME")
				assert.Contains(t, config, "g4dn.xlarge") // GPU instance
				assert.Contains(t, config, "gpu: 1")
			},
		},
		{
			name: "Sentiment analyzer config",
			template: APITemplate{
				ID: "sentiment-analyzer",
			},
			apiName: "sentiment-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "/analyze")
				assert.Contains(t, config, "/emotions")
				assert.Contains(t, config, "EMOTION_MODEL")
				assert.Contains(t, config, "Multi-language")
			},
		},
		{
			name: "Embeddings API config",
			template: APITemplate{
				ID: "embeddings-api",
			},
			apiName: "embeddings-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "/embed")
				assert.Contains(t, config, "/similarity")
				assert.Contains(t, config, "/search")
				assert.Contains(t, config, "VECTOR_DIM")
			},
		},
		{
			name: "Time series predictor config",
			template: APITemplate{
				ID: "time-series-predictor",
			},
			apiName: "forecast-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "/forecast")
				assert.Contains(t, config, "/detect-anomalies")
				assert.Contains(t, config, "DEFAULT_PERIODS")
				assert.Contains(t, config, "Prophet forecasting")
			},
		},
		{
			name: "Document QA config",
			template: APITemplate{
				ID: "document-qa",
			},
			apiName: "doc-qa-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				assert.Contains(t, config, "/upload")
				assert.Contains(t, config, "/ask")
				assert.Contains(t, config, "/documents/{id}")
				assert.Contains(t, config, "QA_MODEL")
				assert.Contains(t, config, "MAX_DOCUMENT_SIZE")
			},
		},
		{
			name: "Unknown template falls back to default",
			template: APITemplate{
				ID: "unknown-template",
			},
			apiName: "test-api",
			runtime: "python3.9",
			validateYAML: func(t *testing.T, config string) {
				// Should fall back to getPythonConfigTemplate
				assert.Contains(t, config, "API-Direct Configuration")
				assert.Contains(t, config, "name: test-api")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getMLTemplateConfig(tt.apiName, tt.runtime, tt.template)
			assert.NotEmpty(t, config)
			if tt.validateYAML != nil {
				tt.validateYAML(t, config)
			}
		})
	}
}

func TestGetMLTemplateMain(t *testing.T) {
	tests := []struct {
		name         string
		template     APITemplate
		validateCode func(*testing.T, string)
	}{
		{
			name: "GPT wrapper main",
			template: APITemplate{
				ID: "gpt-wrapper",
			},
			validateCode: func(t *testing.T, code string) {
				// Check imports
				assert.Contains(t, code, "import openai")
				assert.Contains(t, code, "import redis")
				
				// Check functions
				assert.Contains(t, code, "def complete_text")
				assert.Contains(t, code, "def chat_completion")
				assert.Contains(t, code, "def health_check")
				
				// Check caching logic
				assert.Contains(t, code, "_generate_cache_key")
				assert.Contains(t, code, "_get_cached_response")
				assert.Contains(t, code, "_cache_response")
				
				// Check error handling
				assert.Contains(t, code, "openai.OpenAIError")
			},
		},
		{
			name: "Image classifier main",
			template: APITemplate{
				ID: "image-classifier",
			},
			validateCode: func(t *testing.T, code string) {
				// Check imports
				assert.Contains(t, code, "from PIL import Image")
				assert.Contains(t, code, "from transformers import pipeline")
				assert.Contains(t, code, "import torch")
				
				// Check functions
				assert.Contains(t, code, "def classify_image")
				assert.Contains(t, code, "def classify_batch")
				assert.Contains(t, code, "def list_models")
				
				// Check image handling
				assert.Contains(t, code, "_decode_image")
				assert.Contains(t, code, "base64.b64decode")
			},
		},
		{
			name: "Sentiment analyzer main",
			template: APITemplate{
				ID: "sentiment-analyzer",
			},
			validateCode: func(t *testing.T, code string) {
				// Check model loading
				assert.Contains(t, code, "sentiment_analyzer = pipeline")
				assert.Contains(t, code, "emotion_analyzer = pipeline")
				
				// Check functions
				assert.Contains(t, code, "def analyze_sentiment")
				assert.Contains(t, code, "def analyze_batch")
				assert.Contains(t, code, "def detect_emotions")
				
				// Check sentiment normalization
				assert.Contains(t, code, "if label in ['label_0', 'negative']")
			},
		},
		{
			name: "Unknown template falls back",
			template: APITemplate{
				ID: "unknown",
			},
			validateCode: func(t *testing.T, code string) {
				// Should get default Python template
				assert.NotEmpty(t, code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := getMLTemplateMain(tt.template)
			assert.NotEmpty(t, code)
			if tt.validateCode != nil {
				tt.validateCode(t, code)
			}
		})
	}
}

func TestGetMLTemplateRequirements(t *testing.T) {
	tests := []struct {
		name         string
		template     APITemplate
		validateReqs func(*testing.T, string)
	}{
		{
			name: "GPT wrapper requirements",
			template: APITemplate{
				ID: "gpt-wrapper",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "openai==")
				assert.Contains(t, reqs, "redis==")
				assert.Contains(t, reqs, "pydantic==")
				assert.Contains(t, reqs, "pytest==")
			},
		},
		{
			name: "Image classifier requirements",
			template: APITemplate{
				ID: "image-classifier",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "transformers==")
				assert.Contains(t, reqs, "torch==")
				assert.Contains(t, reqs, "torchvision==")
				assert.Contains(t, reqs, "Pillow==")
				assert.Contains(t, reqs, "opencv-python-headless==")
			},
		},
		{
			name: "Sentiment analyzer requirements",
			template: APITemplate{
				ID: "sentiment-analyzer",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "transformers==")
				assert.Contains(t, reqs, "torch==")
				assert.Contains(t, reqs, "nltk==")
				assert.Contains(t, reqs, "spacy==")
			},
		},
		{
			name: "Embeddings API requirements",
			template: APITemplate{
				ID: "embeddings-api",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "sentence-transformers==")
				assert.Contains(t, reqs, "faiss-cpu==")
				assert.Contains(t, reqs, "scipy==")
			},
		},
		{
			name: "Time series requirements",
			template: APITemplate{
				ID: "time-series-predictor",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "prophet==")
				assert.Contains(t, reqs, "pandas==")
				assert.Contains(t, reqs, "statsmodels==")
				assert.Contains(t, reqs, "scikit-learn==")
			},
		},
		{
			name: "Document QA requirements",
			template: APITemplate{
				ID: "document-qa",
			},
			validateReqs: func(t *testing.T, reqs string) {
				assert.Contains(t, reqs, "PyPDF2==")
				assert.Contains(t, reqs, "python-docx==")
				assert.Contains(t, reqs, "markdown==")
				assert.Contains(t, reqs, "faiss-cpu==")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := getMLTemplateRequirements(tt.template)
			assert.NotEmpty(t, reqs)
			if tt.validateReqs != nil {
				tt.validateReqs(t, reqs)
			}
		})
	}
}

func TestInitMLProject(t *testing.T) {
	tests := []struct {
		name        string
		apiName     string
		runtime     string
		template    APITemplate
		wantErr     bool
		validateDir func(*testing.T, string)
	}{
		{
			name:    "create GPT wrapper project",
			apiName: "gpt-api",
			runtime: "python3.9",
			template: APITemplate{
				ID:          "gpt-wrapper",
				Name:        "GPT Wrapper API",
				Description: "GPT wrapper with caching",
				Category:    "AI/ML",
				Features:    []string{"Caching", "Rate limiting"},
			},
			wantErr: false,
			validateDir: func(t *testing.T, projectPath string) {
				// Check directories
				assert.DirExists(t, projectPath)
				assert.DirExists(t, filepath.Join(projectPath, "tests"))
				assert.DirExists(t, filepath.Join(projectPath, "models"))
				assert.DirExists(t, filepath.Join(projectPath, "data"))
				
				// Check files
				assert.FileExists(t, filepath.Join(projectPath, "apidirect.yaml"))
				assert.FileExists(t, filepath.Join(projectPath, "main.py"))
				assert.FileExists(t, filepath.Join(projectPath, "requirements.txt"))
				assert.FileExists(t, filepath.Join(projectPath, "README.md"))
				assert.FileExists(t, filepath.Join(projectPath, ".gitignore"))
				assert.FileExists(t, filepath.Join(projectPath, "tests", "__init__.py"))
				assert.FileExists(t, filepath.Join(projectPath, "tests", "test_main.py"))
				assert.FileExists(t, filepath.Join(projectPath, "data", ".gitkeep"))
				assert.FileExists(t, filepath.Join(projectPath, "models", ".gitkeep"))
				
				// Validate content
				mainContent, err := os.ReadFile(filepath.Join(projectPath, "main.py"))
				require.NoError(t, err)
				assert.Contains(t, string(mainContent), "import openai")
				assert.Contains(t, string(mainContent), "def complete_text")
				
				configContent, err := os.ReadFile(filepath.Join(projectPath, "apidirect.yaml"))
				require.NoError(t, err)
				assert.Contains(t, string(configContent), "name: gpt-api")
				assert.Contains(t, string(configContent), "runtime: python3.9")
				
				readmeContent, err := os.ReadFile(filepath.Join(projectPath, "README.md"))
				require.NoError(t, err)
				assert.Contains(t, string(readmeContent), "# gpt-api")
				assert.Contains(t, string(readmeContent), "GPT Wrapper API")
			},
		},
		{
			name:    "create image classifier project",
			apiName: "vision-api",
			runtime: "python3.11",
			template: APITemplate{
				ID:          "image-classifier",
				Name:        "Image Classification API",
				Description: "Computer vision API",
				Category:    "AI/ML",
				Features:    []string{"GPU support", "Batch processing"},
			},
			wantErr: false,
			validateDir: func(t *testing.T, projectPath string) {
				// Check ML-specific directories
				assert.DirExists(t, filepath.Join(projectPath, "models"))
				assert.DirExists(t, filepath.Join(projectPath, "data"))
				
				// Check main.py has correct template
				mainContent, err := os.ReadFile(filepath.Join(projectPath, "main.py"))
				require.NoError(t, err)
				assert.Contains(t, string(mainContent), "from PIL import Image")
				assert.Contains(t, string(mainContent), "def classify_image")
				
				// Check requirements
				reqContent, err := os.ReadFile(filepath.Join(projectPath, "requirements.txt"))
				require.NoError(t, err)
				assert.Contains(t, string(reqContent), "transformers")
				assert.Contains(t, string(reqContent), "torch")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			
			err := os.Chdir(tempDir)
			require.NoError(t, err)
			
			// Run initialization
			err = InitMLProject(tt.apiName, tt.runtime, tt.template)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateDir != nil {
					tt.validateDir(t, tt.apiName)
				}
			}
		})
	}
}

func TestGetMLTemplateReadme(t *testing.T) {
	template := APITemplate{
		ID:          "test-ml-template",
		Name:        "Test ML Template",
		Description: "A test ML template for testing",
		Category:    "AI/ML",
		Features:    []string{"Feature 1", "Feature 2", "Feature 3"},
	}
	
	readme := getMLTemplateReadme("test-ml-api", template)
	
	// Verify structure
	assert.Contains(t, readme, "# test-ml-api")
	assert.Contains(t, readme, template.Description)
	assert.Contains(t, readme, "**Template:** Test ML Template")
	assert.Contains(t, readme, "**Category:** AI/ML")
	assert.Contains(t, readme, "**Runtime:** Python 3.9")
	
	// Verify features
	assert.Contains(t, readme, "## 🚀 Features")
	assert.Contains(t, readme, "Feature 1")
	assert.Contains(t, readme, "Feature 2")
	assert.Contains(t, readme, "Feature 3")
	
	// Verify quick start
	assert.Contains(t, readme, "## 🔧 Quick Start")
	assert.Contains(t, readme, "pip install -r requirements.txt")
	assert.Contains(t, readme, "apidirect run")
	assert.Contains(t, readme, "apidirect deploy")
	assert.Contains(t, readme, "apidirect publish test-ml-api")
	
	// Verify sections
	assert.Contains(t, readme, "## 📊 API Endpoints")
	assert.Contains(t, readme, "## 💰 Pricing Suggestions")
	assert.Contains(t, readme, "## 🔗 Resources")
	assert.Contains(t, readme, "## 🆘 Support")
}

func TestGetMLTemplateTests(t *testing.T) {
	tests := []struct {
		name         string
		template     APITemplate
		validateTest func(*testing.T, string)
	}{
		{
			name: "GPT wrapper tests",
			template: APITemplate{
				ID: "gpt-wrapper",
			},
			validateTest: func(t *testing.T, testCode string) {
				assert.Contains(t, testCode, "Tests for GPT Wrapper API")
				assert.Contains(t, testCode, "import unittest")
				assert.Contains(t, testCode, "from main import complete_text, chat_completion, health_check")
				assert.Contains(t, testCode, "class TestGPTWrapperAPI")
				assert.Contains(t, testCode, "test_complete_text_success")
				assert.Contains(t, testCode, "test_complete_text_missing_prompt")
				assert.Contains(t, testCode, "@patch('main.openai')")
			},
		},
		{
			name: "Image classifier tests",
			template: APITemplate{
				ID: "image-classifier",
			},
			validateTest: func(t *testing.T, testCode string) {
				assert.Contains(t, testCode, "Tests for Image Classification API")
				assert.Contains(t, testCode, "from PIL import Image")
				assert.Contains(t, testCode, "from main import classify_image, health_check")
				assert.Contains(t, testCode, "class TestImageClassifierAPI")
				assert.Contains(t, testCode, "create_test_image_data")
				assert.Contains(t, testCode, "test_classify_image_success")
				assert.Contains(t, testCode, "@patch('main.classifier')")
			},
		},
		{
			name: "Default test template",
			template: APITemplate{
				ID: "unknown-template",
			},
			validateTest: func(t *testing.T, testCode string) {
				// Should get default Python test template
				assert.NotEmpty(t, testCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCode := getMLTemplateTests(tt.template)
			assert.NotEmpty(t, testCode)
			if tt.validateTest != nil {
				tt.validateTest(t, testCode)
			}
		})
	}
}

func TestMLTemplateIntegration(t *testing.T) {
	// Test that ML templates integrate properly with the scaffold system
	mlTemplates := GetMLTemplates()
	
	for _, template := range mlTemplates {
		t.Run(template.ID, func(t *testing.T) {
			// Each template should generate valid config
			config := getMLTemplateConfig("test-api", "python3.9", template)
			assert.NotEmpty(t, config)
			assert.Contains(t, config, "name: test-api")
			assert.Contains(t, config, "endpoints:")
			
			// Each template should generate valid main code
			mainCode := getMLTemplateMain(template)
			assert.NotEmpty(t, mainCode)
			assert.Contains(t, mainCode, "def ")
			assert.Contains(t, mainCode, "import ")
			
			// Each template should generate valid requirements
			reqs := getMLTemplateRequirements(template)
			assert.NotEmpty(t, reqs)
			assert.Contains(t, reqs, "==") // Version pinning
			
			// Each template should generate valid tests
			tests := getMLTemplateTests(template)
			assert.NotEmpty(t, tests)
			assert.Contains(t, tests, "unittest")
		})
	}
}

func TestMLTemplatePricing(t *testing.T) {
	// Test that all ML templates include pricing suggestions
	mlTemplates := GetMLTemplates()
	
	for _, template := range mlTemplates {
		t.Run(template.ID, func(t *testing.T) {
			config := getMLTemplateConfig("test-api", "python3.9", template)
			
			// All ML templates should have pricing section
			assert.Contains(t, config, "pricing:")
			assert.Contains(t, config, "free_tier:")
			assert.Contains(t, config, "tiers:")
			
			// Should have at least 2 pricing tiers
			assert.Contains(t, config, "- name:")
			
			// Pricing should be appropriate for the template type
			switch template.ID {
			case "gpt-wrapper":
				assert.Contains(t, config, "price_per_1k:")
			case "image-classifier":
				assert.Contains(t, config, "price_per_image:")
			case "time-series-predictor":
				assert.Contains(t, config, "price_per_forecast:")
			case "document-qa":
				assert.Contains(t, config, "price_per_query:")
			}
		})
	}
}

func TestMLTemplateAWSConfig(t *testing.T) {
	// Test that ML templates have appropriate AWS configurations
	tests := []struct {
		templateID   string
		expectGPU    bool
		minMemory    int
		instanceType string
	}{
		{"gpt-wrapper", false, 2048, "t3.large"},
		{"image-classifier", true, 16384, "g4dn.xlarge"},
		{"sentiment-analyzer", false, 8192, "t3.xlarge"},
		{"embeddings-api", false, 4096, "t3.xlarge"},
		{"time-series-predictor", false, 4096, "t3.xlarge"},
		{"document-qa", false, 8192, "t3.2xlarge"},
	}
	
	for _, tt := range tests {
		t.Run(tt.templateID, func(t *testing.T) {
			template := APITemplate{ID: tt.templateID}
			config := getMLTemplateConfig("test-api", "python3.9", template)
			
			// Check AWS section exists
			assert.Contains(t, config, "aws:")
			assert.Contains(t, config, "cpu:")
			assert.Contains(t, config, "memory:")
			assert.Contains(t, config, "instance_type:")
			
			// Check specific requirements
			if tt.expectGPU {
				assert.Contains(t, config, "gpu:")
			}
			assert.Contains(t, config, tt.instanceType)
		})
	}
}

// Benchmark tests
func BenchmarkGetMLTemplates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetMLTemplates()
	}
}

func BenchmarkGetMLTemplateConfig(b *testing.B) {
	template := APITemplate{ID: "gpt-wrapper"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getMLTemplateConfig("test-api", "python3.9", template)
	}
}

func BenchmarkGetMLTemplateMain(b *testing.B) {
	template := APITemplate{ID: "image-classifier"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getMLTemplateMain(template)
	}
}