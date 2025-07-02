package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInitPythonProject(t *testing.T) {
	tests := []struct {
		name      string
		apiName   string
		runtime   string
		wantErr   bool
		validate  func(*testing.T, string)
	}{
		{
			name:    "create valid Python project structure",
			apiName: "test-python-api",
			runtime: "python3.11",
			wantErr: false,
			validate: func(t *testing.T, projectPath string) {
				// Verify directory structure
				assert.DirExists(t, projectPath)
				assert.DirExists(t, filepath.Join(projectPath, "tests"))
				
				// Verify essential files exist
				assert.FileExists(t, filepath.Join(projectPath, "apidirect.yaml"))
				assert.FileExists(t, filepath.Join(projectPath, "main.py"))
				assert.FileExists(t, filepath.Join(projectPath, "requirements.txt"))
				assert.FileExists(t, filepath.Join(projectPath, ".gitignore"))
				assert.FileExists(t, filepath.Join(projectPath, "README.md"))
				assert.FileExists(t, filepath.Join(projectPath, "tests", "__init__.py"))
				assert.FileExists(t, filepath.Join(projectPath, "tests", "test_main.py"))
				
				// Verify apidirect.yaml is valid YAML and contains required fields
				configPath := filepath.Join(projectPath, "apidirect.yaml")
				configData, err := os.ReadFile(configPath)
				require.NoError(t, err)
				
				var config map[string]interface{}
				err = yaml.Unmarshal(configData, &config)
				require.NoError(t, err, "apidirect.yaml should be valid YAML")
				
				assert.Equal(t, "test-python-api", config["name"])
				assert.Equal(t, "python3.11", config["runtime"])
				assert.NotNil(t, config["endpoints"], "Config should have endpoints")
				
				// Verify main.py has valid Python code
				mainPath := filepath.Join(projectPath, "main.py")
				mainData, err := os.ReadFile(mainPath)
				require.NoError(t, err)
				mainContent := string(mainData)
				
				// Check for essential Python function definitions
				assert.Contains(t, mainContent, "def hello_world", "Should have hello_world function")
				assert.Contains(t, mainContent, "def hello_name", "Should have hello_name function")
				assert.Contains(t, mainContent, "def process_data", "Should have process_data function")
				assert.Contains(t, mainContent, "import json", "Should import json")
				assert.Contains(t, mainContent, "statusCode", "Should return HTTP status codes")
				
				// Verify test file has valid test code
				testPath := filepath.Join(projectPath, "tests", "test_main.py")
				testData, err := os.ReadFile(testPath)
				require.NoError(t, err)
				testContent := string(testData)
				
				assert.Contains(t, testContent, "import unittest", "Tests should use unittest")
				assert.Contains(t, testContent, "from main import", "Tests should import from main")
				assert.Contains(t, testContent, "class Test", "Should have test class")
				assert.Contains(t, testContent, "def test_", "Should have test methods")
				
				// Verify .gitignore has Python-specific ignores
				gitignorePath := filepath.Join(projectPath, ".gitignore")
				gitignoreData, err := os.ReadFile(gitignorePath)
				require.NoError(t, err)
				gitignoreContent := string(gitignoreData)
				
				assert.Contains(t, gitignoreContent, "__pycache__", "Should ignore Python cache")
				assert.Contains(t, gitignoreContent, ".env", "Should ignore env files")
				assert.Contains(t, gitignoreContent, "venv/", "Should ignore virtual environments")
			},
		},
		{
			name:    "create project with special characters in name",
			apiName: "my-awesome_api.v2",
			runtime: "python3.9",
			wantErr: false,
			validate: func(t *testing.T, projectPath string) {
				assert.DirExists(t, projectPath)
				
				// Verify config has the exact name
				configPath := filepath.Join(projectPath, "apidirect.yaml")
				configData, err := os.ReadFile(configPath)
				require.NoError(t, err)
				
				var config map[string]interface{}
				err = yaml.Unmarshal(configData, &config)
				require.NoError(t, err)
				
				assert.Equal(t, "my-awesome_api.v2", config["name"])
				assert.Equal(t, "python3.9", config["runtime"])
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
			err = InitPythonProject(tt.apiName, tt.runtime)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			if tt.validate != nil {
				tt.validate(t, tt.apiName)
			}
		})
	}
}

func TestInitNodeProject(t *testing.T) {
	tests := []struct {
		name      string
		apiName   string
		runtime   string
		wantErr   bool
		validate  func(*testing.T, string)
	}{
		{
			name:    "create valid Node.js project structure",
			apiName: "test-node-api",
			runtime: "node18",
			wantErr: false,
			validate: func(t *testing.T, projectPath string) {
				// Verify directory structure
				assert.DirExists(t, projectPath)
				assert.DirExists(t, filepath.Join(projectPath, "tests"))
				
				// Verify essential files exist
				assert.FileExists(t, filepath.Join(projectPath, "apidirect.yaml"))
				assert.FileExists(t, filepath.Join(projectPath, "main.js"))
				assert.FileExists(t, filepath.Join(projectPath, "package.json"))
				assert.FileExists(t, filepath.Join(projectPath, ".gitignore"))
				assert.FileExists(t, filepath.Join(projectPath, "README.md"))
				assert.FileExists(t, filepath.Join(projectPath, "tests", "main.test.js"))
				
				// Verify package.json is valid JSON
				packagePath := filepath.Join(projectPath, "package.json")
				packageData, err := os.ReadFile(packagePath)
				require.NoError(t, err)
				
				var packageJSON map[string]interface{}
				err = json.Unmarshal(packageData, &packageJSON)
				require.NoError(t, err, "package.json should be valid JSON")
				
				assert.Equal(t, "test-node-api", packageJSON["name"])
				assert.Equal(t, "1.0.0", packageJSON["version"])
				assert.Equal(t, "main.js", packageJSON["main"])
				assert.NotNil(t, packageJSON["scripts"], "Should have scripts section")
				
				// Verify main.js has valid JavaScript code
				mainPath := filepath.Join(projectPath, "main.js")
				mainData, err := os.ReadFile(mainPath)
				require.NoError(t, err)
				mainContent := string(mainData)
				
				// Check for essential exports
				assert.Contains(t, mainContent, "exports.helloWorld", "Should export helloWorld")
				assert.Contains(t, mainContent, "exports.helloName", "Should export helloName")
				assert.Contains(t, mainContent, "exports.processData", "Should export processData")
				assert.Contains(t, mainContent, "statusCode: 200", "Should return status codes")
				assert.Contains(t, mainContent, "JSON.parse", "Should handle JSON")
				
				// Verify test file has valid test structure
				testPath := filepath.Join(projectPath, "tests", "main.test.js")
				testData, err := os.ReadFile(testPath)
				require.NoError(t, err)
				testContent := string(testData)
				
				assert.Contains(t, testContent, "describe(", "Should use describe blocks")
				assert.Contains(t, testContent, "it(", "Should have test cases")
				assert.Contains(t, testContent, "expect(", "Should have assertions")
				assert.Contains(t, testContent, "require('../main')", "Should import main module")
				
				// Verify .gitignore has Node-specific ignores
				gitignorePath := filepath.Join(projectPath, ".gitignore")
				gitignoreData, err := os.ReadFile(gitignorePath)
				require.NoError(t, err)
				gitignoreContent := string(gitignoreData)
				
				assert.Contains(t, gitignoreContent, "node_modules", "Should ignore node_modules")
				assert.Contains(t, gitignoreContent, ".env", "Should ignore env files")
				assert.Contains(t, gitignoreContent, "npm-debug.log", "Should ignore npm logs")
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
			err = InitNodeProject(tt.apiName, tt.runtime)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			if tt.validate != nil {
				tt.validate(t, tt.apiName)
			}
		})
	}
}

func TestInitPythonProjectWithTemplate(t *testing.T) {
	tests := []struct {
		name      string
		apiName   string
		runtime   string
		template  APITemplate
		features  []string
		wantErr   bool
		validate  func(*testing.T, string)
	}{
		{
			name:    "create project with Docker support",
			apiName: "docker-api",
			runtime: "python3.11",
			template: APITemplate{
				ID:          "basic",
				Name:        "Basic API",
				Description: "Basic serverless API",
				Category:    "web",
			},
			features: []string{"Docker support"},
			wantErr:  false,
			validate: func(t *testing.T, projectPath string) {
				// Should have Dockerfile
				assert.FileExists(t, filepath.Join(projectPath, "Dockerfile"))
				assert.FileExists(t, filepath.Join(projectPath, ".dockerignore"))
				
				// Verify Dockerfile content
				dockerPath := filepath.Join(projectPath, "Dockerfile")
				dockerData, err := os.ReadFile(dockerPath)
				require.NoError(t, err)
				dockerContent := string(dockerData)
				
				assert.Contains(t, dockerContent, "FROM python:3.11", "Should use correct Python version")
				assert.Contains(t, dockerContent, "WORKDIR /app", "Should set working directory")
				assert.Contains(t, dockerContent, "COPY requirements.txt", "Should copy requirements")
				assert.Contains(t, dockerContent, "RUN pip install", "Should install dependencies")
				assert.Contains(t, dockerContent, "CMD", "Should have CMD instruction")
			},
		},
		{
			name:    "create project with CI/CD",
			apiName: "cicd-api",
			runtime: "python3.11",
			template: APITemplate{
				ID:       "basic",
				Name:     "Basic API",
				Category: "web",
			},
			features: []string{"CI/CD"},
			wantErr:  false,
			validate: func(t *testing.T, projectPath string) {
				// Should have GitHub Actions workflow
				workflowPath := filepath.Join(projectPath, ".github", "workflows", "deploy.yml")
				assert.FileExists(t, workflowPath)
				
				// Verify workflow content
				workflowData, err := os.ReadFile(workflowPath)
				require.NoError(t, err)
				workflowContent := string(workflowData)
				
				assert.Contains(t, workflowContent, "name: Deploy", "Should have workflow name")
				assert.Contains(t, workflowContent, "on:", "Should have triggers")
				assert.Contains(t, workflowContent, "apidirect deploy", "Should deploy using CLI")
			},
		},
		{
			name:    "create project with API documentation",
			apiName: "docs-api",
			runtime: "python3.11",
			template: APITemplate{
				ID:       "basic",
				Name:     "Basic API",
				Category: "web",
			},
			features: []string{"API documentation"},
			wantErr:  false,
			validate: func(t *testing.T, projectPath string) {
				// Should have docs directory
				assert.DirExists(t, filepath.Join(projectPath, "docs"))
				
				// Should have API documentation file
				docsPath := filepath.Join(projectPath, "docs", "api.md")
				assert.FileExists(t, docsPath)
			},
		},
		{
			name:    "create CRUD database template",
			apiName: "crud-api",
			runtime: "python3.11",
			template: APITemplate{
				ID:          "crud-database",
				Name:        "CRUD Database API",
				Description: "API with database operations",
				Category:    "web",
			},
			features: []string{},
			wantErr:  false,
			validate: func(t *testing.T, projectPath string) {
				// Check requirements.txt has database dependencies
				reqPath := filepath.Join(projectPath, "requirements.txt")
				reqData, err := os.ReadFile(reqPath)
				require.NoError(t, err)
				reqContent := string(reqData)
				
				assert.Contains(t, reqContent, "sqlalchemy", "Should have SQLAlchemy")
				assert.Contains(t, reqContent, "psycopg2-binary", "Should have PostgreSQL driver")
				
				// Check main.py has database code
				mainPath := filepath.Join(projectPath, "main.py")
				mainData, err := os.ReadFile(mainPath)
				require.NoError(t, err)
				mainContent := string(mainData)
				
				assert.Contains(t, mainContent, "database", "Should have database operations")
				assert.Contains(t, strings.ToLower(mainContent), "crud", "Should have CRUD operations")
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
			err = InitPythonProjectWithTemplate(tt.apiName, tt.runtime, tt.template, tt.features)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			if tt.validate != nil {
				tt.validate(t, tt.apiName)
			}
		})
	}
}

func TestGetProjectDirs(t *testing.T) {
	tests := []struct {
		name     string
		template APITemplate
		features []string
		validate func(*testing.T, []string)
	}{
		{
			name: "basic directories",
			template: APITemplate{
				Category: "web",
			},
			features: []string{},
			validate: func(t *testing.T, dirs []string) {
				assert.Contains(t, dirs, "", "Should have root directory")
				assert.Contains(t, dirs, "tests", "Should have tests directory")
				assert.Len(t, dirs, 2, "Should have exactly 2 directories")
			},
		},
		{
			name: "with CI/CD feature",
			template: APITemplate{
				Category: "web",
			},
			features: []string{"CI/CD"},
			validate: func(t *testing.T, dirs []string) {
				assert.Contains(t, dirs, ".github", "Should have .github directory")
				assert.Contains(t, dirs, ".github/workflows", "Should have workflows directory")
			},
		},
		{
			name: "with API documentation feature",
			template: APITemplate{
				Category: "web",
			},
			features: []string{"API documentation"},
			validate: func(t *testing.T, dirs []string) {
				assert.Contains(t, dirs, "docs", "Should have docs directory")
			},
		},
		{
			name: "with multiple features",
			template: APITemplate{
				Category: "web",
			},
			features: []string{"CI/CD", "API documentation", "Docker support"},
			validate: func(t *testing.T, dirs []string) {
				assert.Contains(t, dirs, "")
				assert.Contains(t, dirs, "tests")
				assert.Contains(t, dirs, ".github")
				assert.Contains(t, dirs, ".github/workflows")
				assert.Contains(t, dirs, "docs")
				// Docker support doesn't add directories, just files
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := getProjectDirs(tt.template, tt.features)
			if tt.validate != nil {
				tt.validate(t, dirs)
			}
		})
	}
}

func TestTemplateFileGeneration(t *testing.T) {
	// Test that template functions generate valid content
	t.Run("Python config template", func(t *testing.T) {
		config := getPythonConfigTemplate("test-api", "python3.11")
		
		// Parse as YAML to ensure it's valid
		var parsed map[string]interface{}
		err := yaml.Unmarshal([]byte(config), &parsed)
		assert.NoError(t, err, "Config should be valid YAML")
		
		assert.Equal(t, "test-api", parsed["name"])
		assert.Equal(t, "python3.11", parsed["runtime"])
		assert.NotNil(t, parsed["endpoints"])
	})
	
	t.Run("Node.js package.json template", func(t *testing.T) {
		pkg := getNodePackageTemplate("test-api")
		
		// Parse as JSON to ensure it's valid
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(pkg), &parsed)
		assert.NoError(t, err, "package.json should be valid JSON")
		
		assert.Equal(t, "test-api", parsed["name"])
		assert.Equal(t, "1.0.0", parsed["version"])
		assert.NotNil(t, parsed["scripts"])
		
		scripts, ok := parsed["scripts"].(map[string]interface{})
		assert.True(t, ok, "Scripts should be a map")
		assert.Contains(t, scripts, "test", "Should have test script")
	})
	
	t.Run("Python Dockerfile", func(t *testing.T) {
		dockerfile := getPythonDockerfile("python3.11")
		
		// Check essential Dockerfile instructions
		assert.Contains(t, dockerfile, "FROM python:3.11", "Should have FROM instruction")
		assert.Contains(t, dockerfile, "WORKDIR", "Should set working directory")
		assert.Contains(t, dockerfile, "COPY", "Should copy files")
		assert.Contains(t, dockerfile, "RUN", "Should run commands")
		assert.Contains(t, dockerfile, "CMD", "Should have CMD instruction")
		assert.Contains(t, dockerfile, "EXPOSE", "Should expose port")
	})
	
	t.Run("GitHub Actions workflow", func(t *testing.T) {
		workflow := getGitHubActionsWorkflow()
		
		// Check it's valid YAML
		var parsed map[string]interface{}
		err := yaml.Unmarshal([]byte(workflow), &parsed)
		assert.NoError(t, err, "Workflow should be valid YAML")
		
		assert.NotNil(t, parsed["name"], "Should have workflow name")
		assert.NotNil(t, parsed["on"], "Should have triggers")
		assert.NotNil(t, parsed["jobs"], "Should have jobs")
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("directory creation fails", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)
		
		// Create a file where directory should be
		err = os.WriteFile("blocked-api", []byte("block"), 0644)
		require.NoError(t, err)
		
		// Try to init project - should fail
		err = InitPythonProject("blocked-api", "python3.11")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
	
	t.Run("file creation fails", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)
		
		// Create directory
		err = os.MkdirAll("test-api", 0755)
		require.NoError(t, err)
		
		// Create a directory where file should be
		err = os.MkdirAll("test-api/main.py", 0755)
		require.NoError(t, err)
		
		// Try to init - should fail when trying to write main.py
		err = InitPythonProject("test-api", "python3.11")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")
	})
}

func TestEndToEndProjectCreation(t *testing.T) {
	// Test complete project creation and validate it works
	t.Run("Python project can be created and has valid structure", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)
		
		// Create project
		err = InitPythonProject("e2e-test", "python3.11")
		require.NoError(t, err)
		
		// Verify we can read and parse all config files
		configPath := filepath.Join("e2e-test", "apidirect.yaml")
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err)
		
		var config map[string]interface{}
		err = yaml.Unmarshal(configData, &config)
		require.NoError(t, err)
		
		// Verify endpoints are properly configured
		endpoints, ok := config["endpoints"].([]interface{})
		assert.True(t, ok, "Endpoints should be a list")
		assert.Greater(t, len(endpoints), 0, "Should have at least one endpoint")
		
		// Verify first endpoint has required fields
		firstEndpoint, ok := endpoints[0].(map[string]interface{})
		if !ok {
			// Try map[interface{}]interface{} as fallback
			if ep, ok2 := endpoints[0].(map[interface{}]interface{}); ok2 {
				// Convert to map[string]interface{}
				firstEndpoint = make(map[string]interface{})
				for k, v := range ep {
					if ks, ok := k.(string); ok {
						firstEndpoint[ks] = v
					}
				}
				ok = true
			}
		}
		assert.True(t, ok, "Endpoint should be a map")
		assert.NotNil(t, firstEndpoint["path"], "Endpoint should have path")
		assert.NotNil(t, firstEndpoint["method"], "Endpoint should have method")
		assert.NotNil(t, firstEndpoint["handler"], "Endpoint should have handler")
		
		// Verify handler references actual function in main.py
		handler, ok := firstEndpoint["handler"].(string)
		assert.True(t, ok, "Handler should be a string")
		
		// Extract function name from handler (e.g., "main.helloWorld" -> "helloWorld")
		parts := strings.Split(handler, ".")
		if len(parts) == 2 {
			funcName := parts[1]
			
			// Verify function exists in main.py
			mainPath := filepath.Join("e2e-test", "main.py")
			mainData, err := os.ReadFile(mainPath)
			require.NoError(t, err)
			
			assert.Contains(t, string(mainData), "def "+funcName, "Handler function should exist in main.py")
		}
	})
}

func TestTemplateReadmeGeneration(t *testing.T) {
	template := APITemplate{
		ID:          "test-template",
		Name:        "Test Template",
		Description: "A comprehensive test template",
		Runtime:     "python3.11",
		Category:    "test",
		Features:    []string{"Feature A", "Feature B"},
	}
	
	readme := getTemplateReadme("awesome-api", "Python", template, []string{"Extra feature"})
	
	// Verify structure
	assert.Contains(t, readme, "# awesome-api", "Should have main heading")
	assert.Contains(t, readme, template.Description, "Should include description")
	assert.Contains(t, readme, "**Template:** Test Template", "Should show template name")
	assert.Contains(t, readme, "**Runtime:** Python", "Should show runtime")
	assert.Contains(t, readme, "**Category:** test", "Should show category")
	
	// Verify features are listed
	assert.Contains(t, readme, "## Features Included", "Should have features section")
	assert.Contains(t, readme, "Extra feature", "Should list extra features")
	assert.Contains(t, readme, "## Template Features", "Should have template features section")
	assert.Contains(t, readme, "Feature A", "Should list template features")
	assert.Contains(t, readme, "Feature B", "Should list template features")
	
	// Verify instructions
	assert.Contains(t, readme, "## Getting Started", "Should have getting started section")
	assert.Contains(t, readme, "apidirect run", "Should mention local testing")
	assert.Contains(t, readme, "apidirect deploy", "Should mention deployment")
	assert.Contains(t, readme, "apidirect publish", "Should mention publishing")
	
	// Verify correct install command for Python
	assert.Contains(t, readme, "pip install -r requirements.txt", "Should have Python install command")
}

func TestGetInstallCommand(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"Node.js", "npm install"},
		{"Node", "npm install"},
		{"NodeJS", "npm install"},
		{"node", "npm install"},
		{"Python", "pip install -r requirements.txt"},
		{"python3.11", "pip install -r requirements.txt"},
		{"Go", "pip install -r requirements.txt"}, // Default fallback
		{"Ruby", "pip install -r requirements.txt"}, // Default fallback
		{"Unknown", "pip install -r requirements.txt"}, // Default fallback
	}
	
	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := getInstallCommand(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}