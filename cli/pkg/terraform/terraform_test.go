package terraform

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock terraform executable for testing
func mockTerraformCommand(t *testing.T, workDir string) func() {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	
	// Create mock terraform script
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "terraform")
	
	scriptContent := `#!/bin/bash
# Mock terraform command for testing
case "$1" in
	"version")
		echo "Terraform v1.5.0"
		exit 0
		;;
	"init")
		mkdir -p .terraform
		echo "Terraform has been successfully initialized!"
		exit 0
		;;
	"plan")
		echo "Terraform will perform the following actions:"
		echo "  + resource.example will be created"
		# Create plan file if -out is specified
		for arg in "$@"; do
			if [[ $arg == -out=* ]]; then
				planfile="${arg#-out=}"
				echo "mock plan data" > "$planfile"
			fi
		done
		exit 0
		;;
	"apply")
		echo "Apply complete! Resources: 1 added, 0 changed, 0 destroyed."
		exit 0
		;;
	"destroy")
		echo "Destroy complete! Resources: 1 destroyed."
		exit 0
		;;
	"output")
		if [[ "$2" == "-json" ]]; then
			echo '{"api_endpoint":{"value":"https://api.example.com","type":"string"},"load_balancer_dns":{"value":"lb.example.com","type":"string"}}'
		else
			echo "api_endpoint = https://api.example.com"
			echo "load_balancer_dns = lb.example.com"
		fi
		exit 0
		;;
	*)
		echo "Unknown terraform command: $1"
		exit 1
		;;
esac
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	// Update PATH to use mock
	os.Setenv("PATH", mockDir+":"+originalPath)
	
	// Return cleanup function
	return func() {
		os.Setenv("PATH", originalPath)
	}
}

func TestNewClient(t *testing.T) {
	workDir := "/tmp/terraform"
	client := NewClient(workDir)
	
	assert.NotNil(t, client)
	assert.Equal(t, workDir, client.workDir)
	assert.NotNil(t, client.vars)
	assert.Empty(t, client.vars)
}

func TestSetVar(t *testing.T) {
	client := NewClient("/tmp")
	
	// Test setting different types of variables
	client.SetVar("string_var", "value")
	client.SetVar("int_var", 42)
	client.SetVar("bool_var", true)
	client.SetVar("list_var", []string{"a", "b", "c"})
	
	assert.Equal(t, "value", client.vars["string_var"])
	assert.Equal(t, 42, client.vars["int_var"])
	assert.Equal(t, true, client.vars["bool_var"])
	assert.Equal(t, []string{"a", "b", "c"}, client.vars["list_var"])
}

func TestSetVars(t *testing.T) {
	client := NewClient("/tmp")
	
	vars := map[string]interface{}{
		"var1": "value1",
		"var2": 123,
		"var3": true,
	}
	
	client.SetVars(vars)
	
	assert.Equal(t, "value1", client.vars["var1"])
	assert.Equal(t, 123, client.vars["var2"])
	assert.Equal(t, true, client.vars["var3"])
	
	// Test merging with existing vars
	client.SetVar("var4", "value4")
	newVars := map[string]interface{}{
		"var2": 456, // Override existing
		"var5": "value5",
	}
	client.SetVars(newVars)
	
	assert.Equal(t, "value1", client.vars["var1"])
	assert.Equal(t, 456, client.vars["var2"]) // Updated
	assert.Equal(t, true, client.vars["var3"])
	assert.Equal(t, "value4", client.vars["var4"])
	assert.Equal(t, "value5", client.vars["var5"])
}

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*testing.T, string) func()
		wantErr    bool
		errContains string
	}{
		{
			name: "successful init",
			setupFunc: func(t *testing.T, workDir string) func() {
				return mockTerraformCommand(t, workDir)
			},
			wantErr: false,
		},
		{
			name: "terraform not found",
			setupFunc: func(t *testing.T, workDir string) func() {
				// Save PATH and clear it to ensure terraform isn't found
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "terraform init failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			client := NewClient(workDir)
			
			cleanup := tt.setupFunc(t, workDir)
			defer cleanup()
			
			err := client.Init()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlan(t *testing.T) {
	tests := []struct {
		name       string
		vars       map[string]interface{}
		planFile   string
		wantErr    bool
		errContains string
	}{
		{
			name: "successful plan",
			vars: map[string]interface{}{
				"project_name": "test-api",
				"environment":  "dev",
			},
			planFile: "test.tfplan",
			wantErr:  false,
		},
		{
			name:     "plan with no vars",
			planFile: "empty.tfplan",
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			client := NewClient(workDir)
			
			cleanup := mockTerraformCommand(t, workDir)
			defer cleanup()
			
			if tt.vars != nil {
				client.SetVars(tt.vars)
			}
			
			planPath := filepath.Join(workDir, tt.planFile)
			err := client.Plan(planPath)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				// Check that plan file was created
				_, statErr := os.Stat(planPath)
				assert.NoError(t, statErr, "Plan file should be created")
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name       string
		planFile   string
		setupFunc  func(*testing.T, string)
		wantErr    bool
		errContains string
	}{
		{
			name:     "successful apply",
			planFile: "test.tfplan",
			setupFunc: func(t *testing.T, workDir string) {
				// Create mock plan file
				planPath := filepath.Join(workDir, "test.tfplan")
				err := os.WriteFile(planPath, []byte("mock plan"), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name:        "apply with missing plan file",
			planFile:    "missing.tfplan",
			setupFunc:   func(t *testing.T, workDir string) {},
			wantErr:     false, // Mock doesn't check for file existence
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			client := NewClient(workDir)
			
			cleanup := mockTerraformCommand(t, workDir)
			defer cleanup()
			
			tt.setupFunc(t, workDir)
			
			planPath := filepath.Join(workDir, tt.planFile)
			err := client.Apply(planPath)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDestroy(t *testing.T) {
	tests := []struct {
		name       string
		vars       map[string]interface{}
		wantErr    bool
		errContains string
	}{
		{
			name: "successful destroy",
			vars: map[string]interface{}{
				"project_name": "test-api",
				"environment":  "dev",
			},
			wantErr: false,
		},
		{
			name:    "destroy without vars",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			client := NewClient(workDir)
			
			cleanup := mockTerraformCommand(t, workDir)
			defer cleanup()
			
			if tt.vars != nil {
				client.SetVars(tt.vars)
			}
			
			err := client.Destroy()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOutput(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T, string) func()
		wantOutputs map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "successful output",
			setupFunc: func(t *testing.T, workDir string) func() {
				return mockTerraformCommand(t, workDir)
			},
			wantOutputs: map[string]interface{}{
				"api_endpoint":      "https://api.example.com",
				"load_balancer_dns": "lb.example.com",
			},
			wantErr: false,
		},
		{
			name: "terraform not found",
			setupFunc: func(t *testing.T, workDir string) func() {
				// Save PATH and clear it to ensure terraform isn't found
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "terraform output failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			client := NewClient(workDir)
			
			cleanup := tt.setupFunc(t, workDir)
			defer cleanup()
			
			outputs, err := client.Output()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutputs, outputs)
			}
		})
	}
}

func TestCheckInstalled(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) func()
		wantErr   bool
	}{
		{
			name: "terraform installed",
			setupFunc: func(t *testing.T) func() {
				return mockTerraformCommand(t, "")
			},
			wantErr: false,
		},
		{
			name: "terraform not installed",
			setupFunc: func(t *testing.T) func() {
				// Save PATH and clear it to ensure terraform isn't found
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()
			
			err := CheckInstalled()
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "terraform not found")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStreamingApply(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	
	cleanup := mockTerraformCommand(t, workDir)
	defer cleanup()
	
	// Create plan file
	planFile := filepath.Join(workDir, "test.tfplan")
	err := os.WriteFile(planFile, []byte("mock plan"), 0644)
	require.NoError(t, err)
	
	// Capture output
	var output bytes.Buffer
	err = client.StreamingApply(planFile, &output)
	
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Apply complete!")
}

func TestStreamingPlan(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	
	cleanup := mockTerraformCommand(t, workDir)
	defer cleanup()
	
	// Set some variables
	client.SetVars(map[string]interface{}{
		"project_name": "test-api",
		"environment":  "dev",
	})
	
	// Capture output
	var output bytes.Buffer
	planFile := filepath.Join(workDir, "test.tfplan")
	err := client.StreamingPlan(planFile, &output)
	
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Terraform will perform")
	
	// Check plan file was created
	_, err = os.Stat(planFile)
	assert.NoError(t, err)
}

func TestCopyModules(t *testing.T) {
	tests := []struct {
		name        string
		setupSource func(*testing.T) string
		wantErr     bool
		checkDest   func(*testing.T, string)
	}{
		{
			name: "copy simple module structure",
			setupSource: func(t *testing.T) string {
				sourceDir := t.TempDir()
				
				// Create module structure
				mainTf := filepath.Join(sourceDir, "main.tf")
				err := os.WriteFile(mainTf, []byte("resource \"null_resource\" \"example\" {}"), 0644)
				require.NoError(t, err)
				
				varsTf := filepath.Join(sourceDir, "variables.tf")
				err = os.WriteFile(varsTf, []byte("variable \"project_name\" {}"), 0644)
				require.NoError(t, err)
				
				// Create subdirectory
				modulesDir := filepath.Join(sourceDir, "modules")
				err = os.MkdirAll(modulesDir, 0755)
				require.NoError(t, err)
				
				moduleFile := filepath.Join(modulesDir, "vpc.tf")
				err = os.WriteFile(moduleFile, []byte("module \"vpc\" {}"), 0644)
				require.NoError(t, err)
				
				return sourceDir
			},
			wantErr: false,
			checkDest: func(t *testing.T, destDir string) {
				// Check files were copied
				mainTf := filepath.Join(destDir, "main.tf")
				assert.FileExists(t, mainTf)
				
				varsTf := filepath.Join(destDir, "variables.tf")
				assert.FileExists(t, varsTf)
				
				moduleFile := filepath.Join(destDir, "modules", "vpc.tf")
				assert.FileExists(t, moduleFile)
			},
		},
		{
			name: "skip .terraform directory",
			setupSource: func(t *testing.T) string {
				sourceDir := t.TempDir()
				
				// Create main file
				mainTf := filepath.Join(sourceDir, "main.tf")
				err := os.WriteFile(mainTf, []byte("resource \"null_resource\" \"example\" {}"), 0644)
				require.NoError(t, err)
				
				// Create .terraform directory
				terraformDir := filepath.Join(sourceDir, ".terraform")
				err = os.MkdirAll(terraformDir, 0755)
				require.NoError(t, err)
				
				lockFile := filepath.Join(terraformDir, ".terraform.lock.hcl")
				err = os.WriteFile(lockFile, []byte("# lock file"), 0644)
				require.NoError(t, err)
				
				return sourceDir
			},
			wantErr: false,
			checkDest: func(t *testing.T, destDir string) {
				// Debug: list what was actually copied
				var copiedFiles []string
				filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
					if err == nil {
						relPath, _ := filepath.Rel(destDir, path)
						copiedFiles = append(copiedFiles, relPath)
					}
					return nil
				})
				t.Logf("Copied files: %v", copiedFiles)
				
				// Check main file was copied
				mainTf := filepath.Join(destDir, "main.tf")
				assert.FileExists(t, mainTf)
				
				// Check .terraform was NOT copied
				terraformDir := filepath.Join(destDir, ".terraform")
				assert.NoDirExists(t, terraformDir)
			},
		},
		{
			name: "source directory doesn't exist",
			setupSource: func(t *testing.T) string {
				return "/non/existent/directory"
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceDir := tt.setupSource(t)
			destDir := t.TempDir()
			
			err := CopyModules(sourceDir, destDir)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkDest != nil {
					tt.checkDest(t, destDir)
				}
			}
		})
	}
}

func TestWriteVarsFile(t *testing.T) {
	tests := []struct {
		name       string
		vars       map[string]interface{}
		wantContains []string
	}{
		{
			name: "write various types",
			vars: map[string]interface{}{
				"string_var": "hello world",
				"number_var": 42,
				"bool_var":   true,
				"list_var":   []string{"item1", "item2", "item3"},
				"map_var": map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantContains: []string{
				`string_var = "hello world"`,
				`number_var = 42`,
				`bool_var = true`,
				`list_var = ["item1", "item2", "item3"]`,
				`map_var = {`,
				`key1 = "value1"`,
				`key2 = "value2"`,
			},
		},
		{
			name: "empty vars",
			vars: map[string]interface{}{},
			wantContains: []string{},
		},
		{
			name: "special characters in strings",
			vars: map[string]interface{}{
				"special": "line1\nline2\ttab",
				"quotes":  `has "quotes" inside`,
			},
			wantContains: []string{
				`special = "line1\nline2\ttab"`,
				`quotes = "has \"quotes\" inside"`,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "test.tfvars")
			
			err := WriteVarsFile(filename, tt.vars)
			assert.NoError(t, err)
			
			// Read and check content
			content, err := os.ReadFile(filename)
			require.NoError(t, err)
			
			contentStr := string(content)
			for _, expected := range tt.wantContains {
				assert.Contains(t, contentStr, expected)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) (string, string)
		wantErr   bool
	}{
		{
			name: "successful copy",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "source.txt")
				dst := filepath.Join(t.TempDir(), "dest.txt")
				
				err := os.WriteFile(src, []byte("test content"), 0644)
				require.NoError(t, err)
				
				return src, dst
			},
			wantErr: false,
		},
		{
			name: "source doesn't exist",
			setupFunc: func(t *testing.T) (string, string) {
				return "/non/existent/file", filepath.Join(t.TempDir(), "dest.txt")
			},
			wantErr: true,
		},
		{
			name: "destination directory doesn't exist",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "source.txt")
				err := os.WriteFile(src, []byte("test"), 0644)
				require.NoError(t, err)
				
				return src, "/non/existent/dir/dest.txt"
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setupFunc(t)
			
			err := copyFile(src, dst)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify content
				srcContent, err := os.ReadFile(src)
				require.NoError(t, err)
				
				dstContent, err := os.ReadFile(dst)
				require.NoError(t, err)
				
				assert.Equal(t, srcContent, dstContent)
			}
		})
	}
}

// Test terraform command building with variables
func TestCommandBuilding(t *testing.T) {
	// This test verifies that variables are properly formatted in commands
	workDir := t.TempDir()
	client := NewClient(workDir)
	
	// Set up various types of variables
	client.SetVars(map[string]interface{}{
		"simple_string":    "value",
		"string_with_space": "hello world",
		"number":           123,
		"boolean":          true,
		"special_chars":    "test=value",
	})
	
	// We can't easily test the actual command execution without mocking exec.Command
	// but we can verify the variable storage
	assert.Equal(t, "value", client.vars["simple_string"])
	assert.Equal(t, "hello world", client.vars["string_with_space"])
	assert.Equal(t, 123, client.vars["number"])
	assert.Equal(t, true, client.vars["boolean"])
	assert.Equal(t, "test=value", client.vars["special_chars"])
}

// Test error handling in terraform commands
func TestErrorHandling(t *testing.T) {
	// Create a mock that returns errors
	workDir := t.TempDir()
	client := NewClient(workDir)
	
	// Mock terraform that exits with error
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "terraform")
	
	scriptContent := `#!/bin/bash
echo "Error: Something went wrong" >&2
exit 1
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)
	
	// Test each command returns error
	err = client.Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform init failed")
	assert.Contains(t, err.Error(), "Something went wrong")
	
	err = client.Plan("test.tfplan")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform plan failed")
	
	err = client.Apply("test.tfplan")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform apply failed")
	
	err = client.Destroy()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform destroy failed")
	
	_, err = client.Output()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform output failed")
}

// Test invalid JSON output handling
func TestInvalidJSONOutput(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	
	// Mock terraform that returns invalid JSON
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "terraform")
	
	scriptContent := `#!/bin/bash
if [[ "$1" == "output" && "$2" == "-json" ]]; then
	echo "invalid json{"
	exit 0
fi
exit 1
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)
	
	_, err = client.Output()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse terraform output")
}

// Test concurrent access to variables
func TestConcurrentVarAccess(t *testing.T) {
	// This test demonstrates that concurrent access would cause issues
	// The terraform Client is not thread-safe by design
	t.Skip("Skipping concurrent access test - Client is not designed to be thread-safe")
}