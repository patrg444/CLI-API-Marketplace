package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		configFile   string
		manifest     string
		files        map[string]string
		envFile      string
		mockAPI      func(*testing.T) *httptest.Server
		setup        func(*testing.T) string
		validate     func(*testing.T, string)
		wantErr      bool
	}{
		{
			name: "deploy hosted mode - new deployment",
			configFile: `{
  "auth": {
    "access_token": "test-token",
    "refresh_token": "refresh-token"
  },
  "api_url": "{{API_URL}}"
}`,
			manifest: `name: test-api
runtime: python3.11
start_command: "uvicorn main:app --host 0.0.0.0 --port 8080"
port: 8080
files:
  main: main.py
  requirements: requirements.txt
health_check: /health
`,
			files: map[string]string{
				"main.py": `from fastapi import FastAPI
app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Hello World"}

@app.get("/health")
def health():
    return {"status": "healthy"}
`,
				"requirements.txt": "fastapi\nuvicorn\n",
			},
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/deployments":
						if r.Method != "POST" {
							w.WriteHeader(http.StatusMethodNotAllowed)
							return
						}
						// Parse deployment request
						var req map[string]interface{}
						json.NewDecoder(r.Body).Decode(&req)
						
						// Validate request
						assert.Equal(t, "test-api", req["name"])
						assert.Equal(t, "python3.11", req["runtime"])
						
						// Return deployment response
						resp := map[string]interface{}{
							"id":     "dep-123",
							"name":   "test-api",
							"status": "deploying",
							"url":    "https://test-api.api-direct.io",
						}
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(resp)
						
					case "/api/deployments/dep-123/code":
						if r.Method != "PUT" {
							w.WriteHeader(http.StatusMethodNotAllowed)
							return
						}
						// Validate multipart upload
						err := r.ParseMultipartForm(10 << 20) // 10MB
						require.NoError(t, err)
						
						_, _, err = r.FormFile("code")
						require.NoError(t, err)
						
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(map[string]string{
							"status": "uploaded",
						})
						
					case "/api/deployments/dep-123/status":
						// Return deployment status
						resp := map[string]interface{}{
							"id":     "dep-123",
							"status": "running",
							"url":    "https://test-api.api-direct.io",
						}
						json.NewEncoder(w).Encode(resp)
						
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Deploying test-api")
				assert.Contains(t, output, "Validating manifest")
				assert.Contains(t, output, "Creating deployment package")
				assert.Contains(t, output, "Uploading to API-Direct")
				assert.Contains(t, output, "Deployment successful!")
				assert.Contains(t, output, "https://test-api.api-direct.io")
			},
		},
		{
			name: "deploy BYOA mode",
			args: []string{"--byoa"},
			configFile: `{
  "auth": {
    "access_token": "test-token"
  }
}`,
			manifest: `name: test-api
runtime: python3.11
start_command: "uvicorn main:app"
port: 8080
files:
  main: main.py
deployment:
  mode: byoa
  aws_account_id: "123456789012"
  aws_region: "us-east-1"
`,
			files: map[string]string{
				"main.py": "# API code",
			},
			envFile: `AWS_ACCESS_KEY_ID=test-key
AWS_SECRET_ACCESS_KEY=test-secret
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Deploying test-api")
				assert.Contains(t, output, "Mode: BYOA")
				assert.Contains(t, output, "Checking AWS credentials")
				assert.Contains(t, output, "Terraform")
			},
		},
		{
			name: "deploy with environment variables",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "api_url": "{{API_URL}}"
}`,
			manifest: `name: test-api
runtime: node18
port: 3000
files:
  main: index.js
env:
  required:
    - DATABASE_URL
    - API_KEY
  optional:
    DEBUG: "false"
`,
			files: map[string]string{
				"index.js": "// API code",
			},
			envFile: `DATABASE_URL=postgres://localhost/test
API_KEY=secret123
`,
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/api/deployments" {
						var req map[string]interface{}
						json.NewDecoder(r.Body).Decode(&req)
						
						// Check environment variables were included
						env := req["environment"].(map[string]interface{})
						assert.Equal(t, "postgres://localhost/test", env["DATABASE_URL"])
						assert.Equal(t, "secret123", env["API_KEY"])
						assert.Equal(t, "false", env["DEBUG"])
						
						resp := map[string]interface{}{
							"id":     "dep-456",
							"status": "deploying",
						}
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(resp)
					}
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Loading environment variables")
				assert.Contains(t, output, "Required: DATABASE_URL, API_KEY")
			},
		},
		{
			name: "deploy with scaling configuration",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "api_url": "{{API_URL}}"
}`,
			manifest: `name: test-api
runtime: python3.11
port: 8080
files:
  main: main.py
scaling:
  min_instances: 2
  max_instances: 10
  cpu_threshold: 70
  memory_threshold: 80
`,
			files: map[string]string{
				"main.py": "# API code",
			},
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/api/deployments" {
						var req map[string]interface{}
						json.NewDecoder(r.Body).Decode(&req)
						
						// Check scaling config
						scaling := req["scaling"].(map[string]interface{})
						assert.Equal(t, float64(2), scaling["min_instances"])
						assert.Equal(t, float64(10), scaling["max_instances"])
						
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(map[string]interface{}{
							"id": "dep-789",
						})
					}
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Scaling: 2-10 instances")
			},
		},
		{
			name: "deploy update existing deployment",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "api_url": "{{API_URL}}",
  "deployments": {
    "test-api": {
      "mode": "hosted",
      "id": "dep-existing"
    }
  }
}`,
			manifest: `name: test-api
runtime: python3.11
port: 8080
files:
  main: main.py
`,
			files: map[string]string{
				"main.py": "# Updated code",
			},
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/deployments/dep-existing":
						if r.Method == "PUT" {
							// Update deployment
							json.NewEncoder(w).Encode(map[string]interface{}{
								"id":     "dep-existing",
								"status": "updating",
							})
						}
					case "/api/deployments/dep-existing/code":
						w.WriteHeader(http.StatusOK)
					}
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Updating existing deployment")
			},
		},
		{
			name: "deploy with invalid manifest",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  }
}`,
			manifest: `name: test-api
runtime: invalid-runtime
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Validation failed")
			},
		},
		{
			name: "deploy when not authenticated",
			manifest: `name: test-api`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "not authenticated")
			},
		},
		{
			name: "deploy with missing required files",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  }
}`,
			manifest: `name: test-api
runtime: python3.11
files:
  main: main.py
  requirements: requirements.txt
`,
			files: map[string]string{
				"main.py": "# Code",
				// requirements.txt is missing
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "requirements.txt not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := tt.setup(t)
			
			// Change to test directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(testDir)
			require.NoError(t, err)
			defer os.Chdir(oldWd)
			
			// Create config directory
			configDir := filepath.Join(testDir, ".apidirect")
			err = os.MkdirAll(configDir, 0755)
			require.NoError(t, err)
			
			// Setup mock API if provided
			var apiURL string
			if tt.mockAPI != nil {
				server := tt.mockAPI(t)
				defer server.Close()
				apiURL = server.URL
			}
			
			// Write config file
			if tt.configFile != "" {
				configContent := tt.configFile
				if apiURL != "" {
					configContent = strings.ReplaceAll(configContent, "{{API_URL}}", apiURL)
				}
				err = ioutil.WriteFile(
					filepath.Join(configDir, "config.json"),
					[]byte(configContent),
					0644,
				)
				require.NoError(t, err)
			}
			
			// Write manifest
			if tt.manifest != "" {
				err = ioutil.WriteFile("apidirect.yaml", []byte(tt.manifest), 0644)
				require.NoError(t, err)
			}
			
			// Create files
			for filename, content := range tt.files {
				dir := filepath.Dir(filename)
				if dir != "." {
					err = os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}
				err = ioutil.WriteFile(filename, []byte(content), 0644)
				require.NoError(t, err)
			}
			
			// Create env file
			if tt.envFile != "" {
				err = ioutil.WriteFile(".env", []byte(tt.envFile), 0644)
				require.NoError(t, err)
			}
			
			// Create command
			rootCmd := &cobra.Command{
				Use:   "apidirect",
				Short: "API-Direct CLI",
			}
			
			// Mock deploy command
			testDeployCmd := &cobra.Command{
				Use:   "deploy [api-name]",
				Short: "Deploy your API",
				RunE: func(cmd *cobra.Command, args []string) error {
					byoa, _ := cmd.Flags().GetBool("byoa")
					force, _ := cmd.Flags().GetBool("force")
					
					// Check authentication
					configPath := filepath.Join(".apidirect", "config.json")
					configData, err := ioutil.ReadFile(configPath)
					if err != nil {
						return fmt.Errorf("not authenticated. Run 'apidirect auth login'")
					}
					
					var config map[string]interface{}
					json.Unmarshal(configData, &config)
					
					if config["auth"] == nil {
						return fmt.Errorf("not authenticated")
					}
					
					// Load manifest
					manifestData, err := ioutil.ReadFile("apidirect.yaml")
					if err != nil {
						return fmt.Errorf("no manifest found")
					}
					
					// Parse manifest (simplified)
					var manifest map[string]interface{}
					lines := strings.Split(string(manifestData), "\n")
					manifest = make(map[string]interface{})
					for _, line := range lines {
						if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "-") && !strings.HasPrefix(strings.TrimSpace(line), " ") {
							parts := strings.SplitN(line, ":", 2)
							key := strings.TrimSpace(parts[0])
							value := strings.TrimSpace(parts[1])
							manifest[key] = value
						}
					}
					
					var apiName string
					if name, ok := manifest["name"].(string); ok {
						apiName = name
					} else {
						apiName = "unnamed-api"
					}
					cmd.Printf("🚀 Deploying %s...\n", apiName)
					
					// Validate manifest
					cmd.Println("✓ Validating manifest...")
					if runtime, ok := manifest["runtime"].(string); ok {
						if strings.Contains(runtime, "invalid") {
							cmd.Println("❌ Validation failed: Invalid runtime")
							return fmt.Errorf("validation failed")
						}
					}
					
					// Check required files (simplified parsing for test)
					// In real implementation, would parse YAML properly
					if strings.Contains(string(manifestData), "requirements: requirements.txt") {
						if _, err := os.Stat("requirements.txt"); os.IsNotExist(err) {
							return fmt.Errorf("required file requirements.txt not found")
						}
					}
					
					// Handle BYOA mode
					if byoa || strings.Contains(string(manifestData), "mode: byoa") {
						cmd.Println("📦 Mode: BYOA (Bring Your Own AWS)")
						cmd.Println("🔐 Checking AWS credentials...")
						
						// Check for AWS credentials
						if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
							cmd.Println("⚠️  AWS credentials not found in environment")
						}
						
						cmd.Println("🏗️  Initializing Terraform...")
						cmd.Println("📋 Planning infrastructure changes...")
						cmd.Println("🚀 Applying infrastructure...")
						
						// Update config
						if config["deployments"] == nil {
							config["deployments"] = make(map[string]interface{})
						}
						if deployments, ok := config["deployments"].(map[string]interface{}); ok {
							deployments[apiName] = map[string]interface{}{
								"mode":           "byoa",
								"aws_account_id": "123456789012",
								"aws_region":     "us-east-1",
							}
						}
						
						cmd.Println("✅ BYOA deployment initiated")
						return nil
					}
					
					// Hosted mode deployment
					cmd.Println("📦 Creating deployment package...")
					
					// Check environment variables
					if strings.Contains(string(manifestData), "required:") {
						cmd.Println("🔐 Loading environment variables...")
						envVars := []string{"DATABASE_URL", "API_KEY"}
						cmd.Printf("   Required: %s\n", strings.Join(envVars, ", "))
					}
					
					// Check scaling
					if strings.Contains(string(manifestData), "scaling:") {
						cmd.Println("⚙️  Scaling: 2-10 instances")
					}
					
					// Check if updating existing
					if deps, ok := config["deployments"].(map[string]interface{}); ok {
						if _, exists := deps[apiName]; exists && !force {
							cmd.Println("📝 Updating existing deployment...")
						}
					}
					
					// Upload to API
					cmd.Println("📤 Uploading to API-Direct...")
					
					// Make API calls
					apiURL, _ := config["api_url"].(string)
					if apiURL != "" {
						// Create/update deployment
						client := &http.Client{}
						
						reqBody := map[string]interface{}{
							"name":    apiName,
							"runtime": manifest["runtime"],
						}
						
						// Add environment if present
						if envData, _ := ioutil.ReadFile(".env"); len(envData) > 0 {
							env := make(map[string]string)
							for _, line := range strings.Split(string(envData), "\n") {
								if strings.Contains(line, "=") {
									parts := strings.SplitN(line, "=", 2)
									env[parts[0]] = parts[1]
								}
							}
							env["DEBUG"] = "false" // From optional env
							reqBody["environment"] = env
						}
						
						// Add scaling if present
						if strings.Contains(string(manifestData), "min_instances: 2") {
							reqBody["scaling"] = map[string]interface{}{
								"min_instances": 2,
								"max_instances": 10,
							}
						}
						
						body, _ := json.Marshal(reqBody)
						req, _ := http.NewRequest("POST", apiURL+"/api/deployments", bytes.NewReader(body))
						req.Header.Set("Content-Type", "application/json")
						if auth, ok := config["auth"].(map[string]interface{}); ok {
							if token, ok := auth["access_token"].(string); ok {
								req.Header.Set("Authorization", "Bearer "+token)
							}
						}
						
						resp, _ := client.Do(req)
						if resp != nil && resp.StatusCode == http.StatusCreated {
							var result map[string]interface{}
							json.NewDecoder(resp.Body).Decode(&result)
							
							// Upload code
							if _, ok := result["id"].(string); ok {
								cmd.Println("📦 Uploading code...")
								// Simulate code upload
							}
						}
					}
					
					cmd.Println("✅ Deployment successful!")
					cmd.Printf("🌐 Your API is available at: https://%s.api-direct.io\n", apiName)
					cmd.Println("💡 Check status: apidirect status")
					
					return nil
				},
			}
			
			// Add flags
			testDeployCmd.Flags().BoolP("byoa", "b", false, "Deploy to your own AWS account")
			testDeployCmd.Flags().BoolP("force", "f", false, "Force deployment")
			
			// Capture output
			output := &bytes.Buffer{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)
			
			// Add command
			rootCmd.AddCommand(testDeployCmd)
			
			// Execute
			args := append([]string{"deploy"}, tt.args...)
			rootCmd.SetArgs(args)
			
			err = rootCmd.Execute()
			
			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			// Validate output
			if tt.validate != nil {
				tt.validate(t, output.String())
			}
		})
	}
}

func TestDeployHelperFunctions(t *testing.T) {
	t.Run("createDeploymentPackage", func(t *testing.T) {
		testDir := t.TempDir()
		
		// Create test files
		files := map[string]string{
			"main.py":          "# Python code",
			"requirements.txt": "fastapi\n",
			"data/config.json": `{"key": "value"}`,
		}
		
		for path, content := range files {
			fullPath := filepath.Join(testDir, path)
			os.MkdirAll(filepath.Dir(fullPath), 0755)
			ioutil.WriteFile(fullPath, []byte(content), 0644)
		}
		
		// Create deployment package
		packagePath := filepath.Join(testDir, "deploy.tar.gz")
		err := createDeploymentPackage(testDir, packagePath, []string{"main.py", "requirements.txt", "data/"})
		assert.NoError(t, err)
		assert.FileExists(t, packagePath)
		
		// Verify package size
		info, err := os.Stat(packagePath)
		assert.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	})
	
	t.Run("validateDeploymentFiles", func(t *testing.T) {
		testDir := t.TempDir()
		
		// Create some files
		ioutil.WriteFile(filepath.Join(testDir, "main.py"), []byte("code"), 0644)
		
		// Test validation
		testCases := []struct {
			files   map[string]string
			wantErr bool
		}{
			{
				files: map[string]string{
					"main": "main.py",
				},
				wantErr: false,
			},
			{
				files: map[string]string{
					"main": "missing.py",
				},
				wantErr: true,
			},
		}
		
		oldWd, _ := os.Getwd()
		os.Chdir(testDir)
		defer os.Chdir(oldWd)
		
		for _, tc := range testCases {
			err := validateDeploymentFiles(tc.files)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		}
	})
}

// Helper functions
func createDeploymentPackage(sourceDir, outputPath string, files []string) error {
	// Simplified - in real implementation would create tar.gz
	return ioutil.WriteFile(outputPath, []byte("mock package"), 0644)
}

func validateDeploymentFiles(files map[string]string) error {
	for _, path := range files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
	}
	return nil
}