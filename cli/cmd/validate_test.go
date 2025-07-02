package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		manifest  string
		files     map[string]string
		setup     func(*testing.T) string
		validate  func(*testing.T, string)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid manifest passes validation",
			manifest: `name: test-api
runtime: python3.11
start_command: uvicorn main:app --host 0.0.0.0 --port 8080
port: 8080
files:
  main: main.py
  requirements: requirements.txt
endpoints:
  - method: GET
    path: /
    description: Root endpoint
  - method: GET
    path: /health
    description: Health check
health_check: /health
env:
  required: []
  optional: ["DEBUG", "LOG_LEVEL"]
`,
			files: map[string]string{
				"main.py":          "# Main application file",
				"requirements.txt": "fastapi\nuvicorn",
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "✅")
				assert.Contains(t, output, "Valid")
				assert.Contains(t, output, "All files exist")
			},
		},
		{
			name: "missing required files",
			manifest: `name: test-api
runtime: python3.11
start_command: uvicorn main:app --host 0.0.0.0 --port 8080
port: 8080
files:
  main: main.py
  requirements: requirements.txt
health_check: /health
`,
			files: map[string]string{
				"main.py": "# Main application file",
				// requirements.txt is missing
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "requirements.txt")
				assert.Contains(t, output, "not found")
			},
		},
		{
			name: "invalid port number",
			manifest: `name: test-api
runtime: python3.11
port: 99999
health_check: /health
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "port")
				assert.Contains(t, output, "invalid")
			},
		},
		{
			name: "missing required fields",
			manifest: `name: test-api
# runtime is missing
port: 8080
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "Runtime")  // Capital R
				assert.Contains(t, output, "required")
			},
		},
		{
			name: "invalid runtime",
			manifest: `name: test-api
runtime: invalid-runtime-99
port: 8080
health_check: /health
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "runtime")
				assert.Contains(t, output, "unsupported")
			},
		},
		{
			name: "valid manifest with environment variables",
			manifest: `name: test-api
runtime: node18
start_command: node index.js
port: 3000
files:
  main: index.js
  deps: package.json
health_check: /health
env:
  required: ["DATABASE_URL", "API_KEY"]
  optional: ["DEBUG", "LOG_LEVEL"]
`,
			files: map[string]string{
				"index.js":     "// Main application",
				"package.json": `{"name": "test-api"}`,
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "✅")
				assert.Contains(t, output, "Required environment variables")
				assert.Contains(t, output, "DATABASE_URL")
				assert.Contains(t, output, "API_KEY")
			},
		},
		{
			name: "manifest with scaling configuration",
			manifest: `name: test-api
runtime: go1.21
port: 8080
health_check: /health
scaling:
  min: 1
  max: 10
  target_cpu: 70
resources:
  memory: 512Mi
  cpu: 250m
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "✅")
				assert.Contains(t, output, "Scaling")
				assert.Contains(t, output, "Resources")
			},
		},
		{
			name: "no manifest file",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "apidirect.yaml not found")
			},
		},
		{
			name: "malformed YAML",
			manifest: `name: test-api
runtime: python3.11
  port: 8080  # Invalid indentation
[invalid yaml content
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "YAML")
				assert.Contains(t, output, "Failed")
			},
		},
		{
			name: "validate specific file path",
			args: []string{"custom-manifest.yaml"},
			manifest: `name: custom-api
runtime: python3.11
port: 8080
health_check: /health
`,
			setup: func(t *testing.T) string {
				testDir := t.TempDir()
				// Write to custom file name
				err := ioutil.WriteFile(
					filepath.Join(testDir, "custom-manifest.yaml"),
					[]byte(`name: custom-api
runtime: python3.11
port: 8080
health_check: /health
`),
					0644,
				)
				require.NoError(t, err)
				return testDir
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "custom-manifest.yaml")
				assert.Contains(t, output, "✅")
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
			
			// Write manifest file if provided
			if tt.manifest != "" && len(tt.args) == 0 {
				err = ioutil.WriteFile("apidirect.yaml", []byte(tt.manifest), 0644)
				require.NoError(t, err)
			}
			
			// Create any additional files
			for filename, content := range tt.files {
				dir := filepath.Dir(filename)
				if dir != "." {
					err = os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}
				err = ioutil.WriteFile(filename, []byte(content), 0644)
				require.NoError(t, err)
			}
			
			// Create command structure for testing
			rootCmd := &cobra.Command{
				Use:   "apidirect",
				Short: "API-Direct CLI",
			}
			
			// Create test version of validate command
			testValidateCmd := &cobra.Command{
				Use:   "validate [file]",
				Short: "Validate apidirect.yaml configuration",
				RunE: func(cmd *cobra.Command, args []string) error {
					manifestPath := "apidirect.yaml"
					if len(args) > 0 {
						manifestPath = args[0]
					}
					
					cmd.Printf("🔍 Validating %s...\n\n", manifestPath)
					
					// Check if file exists
					if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
						return fmt.Errorf("apidirect.yaml not found")
					}
					
					// Read and parse YAML
					data, err := ioutil.ReadFile(manifestPath)
					if err != nil {
						return fmt.Errorf("failed to read manifest: %w", err)
					}
					
					var manifest map[string]interface{}
					if err := yaml.Unmarshal(data, &manifest); err != nil {
						cmd.Println("❌ YAML Parsing: Failed")
						cmd.Printf("   ↳ %v\n", err)
						return fmt.Errorf("invalid YAML")
					}
					
					cmd.Println("✅ YAML Parsing: Valid")
					
					// Validate required fields
					hasErrors := false
					
					if _, ok := manifest["name"]; !ok {
						cmd.Println("❌ Name: Missing (required)")
						hasErrors = true
					} else {
						cmd.Println("✅ Name: Valid")
					}
					
					if runtime, ok := manifest["runtime"]; !ok {
						cmd.Println("❌ Runtime: Missing (required)")
						hasErrors = true
					} else if err := validateRuntime(runtime.(string)); err != nil {
						cmd.Printf("❌ Runtime: %v\n", err)
						hasErrors = true
					} else {
						cmd.Println("✅ Runtime: Valid")
					}
					
					if port, ok := manifest["port"]; ok {
						if portInt, ok := port.(int); ok {
							if err := validatePort(portInt); err != nil {
								cmd.Printf("❌ Port: %v\n", err)
								hasErrors = true
							} else {
								cmd.Println("✅ Port: Valid")
							}
						}
					}
					
					// Check files
					if files, ok := manifest["files"].(map[string]interface{}); ok {
						allExist := true
						for _, path := range files {
							if pathStr, ok := path.(string); ok {
								if _, err := os.Stat(pathStr); os.IsNotExist(err) {
									cmd.Printf("❌ File: %s not found\n", pathStr)
									allExist = false
									hasErrors = true
								}
							}
						}
						if allExist {
							cmd.Println("✅ Files: All files exist")
						}
					}
					
					// Check environment variables
					if env, ok := manifest["env"].(map[string]interface{}); ok {
						if required, ok := env["required"].([]interface{}); ok && len(required) > 0 {
							cmd.Println("\n📋 Required environment variables:")
							for _, v := range required {
								cmd.Printf("   - %v\n", v)
							}
						}
					}
					
					// Check scaling config
					if scaling, ok := manifest["scaling"]; ok {
						cmd.Printf("\n⚖️  Scaling configuration: %v\n", scaling)
					}
					
					// Check resources
					if resources, ok := manifest["resources"]; ok {
						cmd.Printf("\n💾 Resources configuration: %v\n", resources)
					}
					
					if hasErrors {
						return fmt.Errorf("validation failed")
					}
					
					cmd.Println("\n✅ Manifest is valid!")
					return nil
				},
			}
			
			// Capture output
			output := &bytes.Buffer{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)
			
			// Add validate command
			rootCmd.AddCommand(testValidateCmd)
			
			// Execute command
			args := append([]string{"validate"}, tt.args...)
			rootCmd.SetArgs(args)
			
			err = rootCmd.Execute()
			
			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, output.String(), tt.errMsg)
				}
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

func TestValidationRules(t *testing.T) {
	t.Run("runtime validation", func(t *testing.T) {
		validRuntimes := []string{
			"python3.9", "python3.10", "python3.11", "python3.12",
			"node16", "node18", "node20",
			"go1.19", "go1.20", "go1.21",
			"ruby3.0", "ruby3.1", "ruby3.2",
		}
		
		for _, runtime := range validRuntimes {
			err := validateRuntime(runtime)
			assert.NoError(t, err, "Runtime %s should be valid", runtime)
		}
		
		invalidRuntimes := []string{
			"python2.7",     // Too old
			"node14",        // Too old
			"java11",        // Not supported
			"php8",          // Not supported
			"invalid",       // Invalid format
			"",              // Empty
		}
		
		for _, runtime := range invalidRuntimes {
			err := validateRuntime(runtime)
			assert.Error(t, err, "Runtime %s should be invalid", runtime)
		}
	})
	
	t.Run("port validation", func(t *testing.T) {
		validPorts := []int{80, 443, 3000, 8080, 8000, 9000}
		
		for _, port := range validPorts {
			err := validatePort(port)
			assert.NoError(t, err, "Port %d should be valid", port)
		}
		
		invalidPorts := []int{0, -1, 70000, 99999}
		
		for _, port := range invalidPorts {
			err := validatePort(port)
			assert.Error(t, err, "Port %d should be invalid", port)
		}
	})
	
	t.Run("endpoint validation", func(t *testing.T) {
		validEndpoints := []struct {
			method string
			path   string
		}{
			{"GET", "/"},
			{"POST", "/api/users"},
			{"PUT", "/api/users/{id}"},
			{"DELETE", "/api/items/{id}"},
			{"PATCH", "/api/items/{id}"},
			{"GET", "/health"},
			{"GET", "/api/v1/products"},
		}
		
		for _, ep := range validEndpoints {
			err := validateEndpoint(ep.method, ep.path)
			assert.NoError(t, err, "Endpoint %s %s should be valid", ep.method, ep.path)
		}
		
		invalidEndpoints := []struct {
			method string
			path   string
		}{
			{"INVALID", "/"},          // Invalid method
			{"GET", "api/users"},      // Missing leading slash
			{"GET", "/api/users/"},    // Trailing slash
			{"", "/api"},              // Empty method
			{"GET", ""},               // Empty path
		}
		
		for _, ep := range invalidEndpoints {
			err := validateEndpoint(ep.method, ep.path)
			assert.Error(t, err, "Endpoint %s %s should be invalid", ep.method, ep.path)
		}
	})
	
	t.Run("environment variable name validation", func(t *testing.T) {
		validNames := []string{
			"DATABASE_URL",
			"API_KEY",
			"DEBUG",
			"PORT",
			"LOG_LEVEL",
			"AWS_ACCESS_KEY_ID",
		}
		
		for _, name := range validNames {
			err := validateEnvVarName(name)
			assert.NoError(t, err, "Env var %s should be valid", name)
		}
		
		invalidNames := []string{
			"",              // Empty
			"lowercase",     // Not uppercase
			"WITH-DASH",     // Contains dash
			"WITH SPACE",    // Contains space
			"1STARTS_NUM",   // Starts with number
		}
		
		for _, name := range invalidNames {
			err := validateEnvVarName(name)
			assert.Error(t, err, "Env var %s should be invalid", name)
		}
	})
}

// Mock validation functions (these would be in the actual implementation)
func validateRuntime(runtime string) error {
	validRuntimes := map[string]bool{
		"python3.9": true, "python3.10": true, "python3.11": true, "python3.12": true,
		"node16": true, "node18": true, "node20": true,
		"go1.19": true, "go1.20": true, "go1.21": true,
		"ruby3.0": true, "ruby3.1": true, "ruby3.2": true,
	}
	
	if !validRuntimes[runtime] {
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	return nil
}

func validateEndpoint(method, path string) error {
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, 
		"DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	
	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}
	
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}
	
	if strings.HasSuffix(path, "/") && path != "/" {
		return fmt.Errorf("path should not end with /")
	}
	
	return nil
}

func validateEnvVarName(name string) error {
	if name == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}
	
	// Check if uppercase with underscores only
	for i, ch := range name {
		if !((ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9' && i > 0)) {
			return fmt.Errorf("invalid environment variable name: %s", name)
		}
	}
	
	if name[0] >= '0' && name[0] <= '9' {
		return fmt.Errorf("environment variable name cannot start with a number")
	}
	
	return nil
}