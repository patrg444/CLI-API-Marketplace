package orchestrator

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock AWS functions for testing
func mockAWSFunctions(t *testing.T) func() {
	// Save original environment
	originalPath := os.Getenv("PATH")
	originalAWSMock := os.Getenv("MOCK_AWS_FOR_TESTS")
	
	// Set mock environment
	os.Setenv("MOCK_AWS_FOR_TESTS", "true")
	
	// Create mock aws command
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "aws")
	
	scriptContent := `#!/bin/bash
case "$1" in
	"sts")
		if [[ "$2" == "get-caller-identity" ]]; then
			echo '{"Account":"123456789012","UserId":"AIDAI23HXD3MBVANCL4X6","Arn":"arn:aws:iam::123456789012:user/testuser"}'
		fi
		;;
	"s3")
		if [[ "$2" == "mb" ]]; then
			echo "make_bucket: $3"
		elif [[ "$2" == "ls" ]]; then
			echo ""  # Empty bucket
		fi
		;;
	"dynamodb")
		if [[ "$2" == "create-table" ]]; then
			echo '{"TableDescription":{"TableStatus":"CREATING"}}'
		fi
		;;
	"configure")
		if [[ "$2" == "get" && "$3" == "region" ]]; then
			echo "us-east-1"
		fi
		;;
esac
exit 0
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	// Update PATH
	os.Setenv("PATH", mockDir+":"+originalPath)
	
	// Return cleanup function
	return func() {
		os.Setenv("PATH", originalPath)
		if originalAWSMock == "" {
			os.Unsetenv("MOCK_AWS_FOR_TESTS")
		} else {
			os.Setenv("MOCK_AWS_FOR_TESTS", originalAWSMock)
		}
	}
}

func setupTestConfig(t *testing.T) func() {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	
	// Create config directory
	configDir := filepath.Join(testDir, ".apidirect")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	// Create default config
	cfg := &config.Config{
		User: config.UserConfig{
			Email: "test@example.com",
		},
		Deployments: make(map[string]interface{}),
	}
	err = config.SaveConfig(cfg)
	require.NoError(t, err)
	
	// Return cleanup function
	return func() {
		os.Setenv("HOME", originalHome)
	}
}

func TestNewBYOADeployment(t *testing.T) {
	tests := []struct {
		name        string
		apiName     string
		manifest    *manifest.Manifest
		setupFunc   func(*testing.T) func()
		wantErr     bool
		errContains string
		validate    func(*testing.T, *BYOADeployment)
	}{
		{
			name:    "successful creation",
			apiName: "test-api",
			manifest: &manifest.Manifest{
				Name:        "test-api",
				Port:        8080,
				HealthCheck: "/health",
			},
			setupFunc: func(t *testing.T) func() {
				cleanup1 := mockAWSFunctions(t)
				cleanup2 := setupTestConfig(t)
				return func() {
					cleanup1()
					cleanup2()
				}
			},
			wantErr: false,
			validate: func(t *testing.T, d *BYOADeployment) {
				assert.Equal(t, "test-api", d.APIName)
				assert.Equal(t, "123456789012", d.AWSAccountID)
				assert.Equal(t, "us-east-1", d.AWSRegion)
				assert.Equal(t, "prod", d.Environment) // Default environment
				assert.Contains(t, d.WorkDir, "apidirect-deploy-test-api")
			},
		},
		{
			name:    "default environment",
			apiName: "prod-api",
			manifest: &manifest.Manifest{
				Name: "prod-api",
				Port: 3000,
			},
			setupFunc: func(t *testing.T) func() {
				cleanup1 := mockAWSFunctions(t)
				cleanup2 := setupTestConfig(t)
				return func() {
					cleanup1()
					cleanup2()
				}
			},
			wantErr: false,
			validate: func(t *testing.T, d *BYOADeployment) {
				assert.Equal(t, "prod", d.Environment)
			},
		},
		{
			name:    "AWS error",
			apiName: "fail-api",
			manifest: &manifest.Manifest{
				Name: "fail-api",
			},
			setupFunc: func(t *testing.T) func() {
				// Save PATH and set invalid to make AWS fail
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "failed to get AWS account info",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			deployment, err := NewBYOADeployment(tt.apiName, tt.manifest)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, deployment)
				if tt.validate != nil {
					tt.validate(t, deployment)
				}
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	tests := []struct {
		name        string
		deployment  *BYOADeployment
		setupFunc   func(*testing.T, *BYOADeployment) func()
		wantErr     bool
		errContains string
		checkFunc   func(*testing.T, *BYOADeployment)
	}{
		{
			name: "successful prepare",
			deployment: &BYOADeployment{
				APIName:     "test-api",
				WorkDir:     filepath.Join(t.TempDir(), "work"),
				Environment: "dev",
				StateBackend: StateBackend{
					Bucket:   "test-bucket",
					Key:      "test/key",
					Region:   "us-east-1",
					DynamoDB: "test-table",
				},
			},
			setupFunc: func(t *testing.T, d *BYOADeployment) func() {
				cleanup := mockAWSFunctions(t)
				
				// Create mock modules directory
				modulesDir := d.getModulesPath()
				err := os.MkdirAll(modulesDir, 0755)
				require.NoError(t, err)
				
				// Create some test files
				mainTf := filepath.Join(modulesDir, "main.tf")
				err = os.WriteFile(mainTf, []byte("# Main config"), 0644)
				require.NoError(t, err)
				
				return cleanup
			},
			wantErr: false,
			checkFunc: func(t *testing.T, d *BYOADeployment) {
				// Check working directory created
				assert.DirExists(t, d.WorkDir)
				
				// Check backend.tf created
				backendFile := filepath.Join(d.WorkDir, "backend.tf")
				assert.FileExists(t, backendFile)
				
				// Check backend content
				content, err := os.ReadFile(backendFile)
				require.NoError(t, err)
				assert.Contains(t, string(content), "test-bucket")
				assert.Contains(t, string(content), "test/key")
			},
		},
		{
			name: "invalid work directory",
			deployment: &BYOADeployment{
				WorkDir: "/invalid\x00path",
			},
			wantErr:     true,
			errContains: "failed to create working directory",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t, tt.deployment)
			}
			defer cleanup()
			
			err := tt.deployment.Prepare()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, tt.deployment)
				}
			}
		})
	}
}

func TestGetTerraformVars(t *testing.T) {
	tests := []struct {
		name       string
		deployment *BYOADeployment
		setupFunc  func(*testing.T) func()
		validate   func(*testing.T, map[string]interface{})
	}{
		{
			name: "basic vars",
			deployment: &BYOADeployment{
				APIName:      "test-api",
				Environment:  "staging",
				AWSRegion:    "us-west-2",
				AWSAccountID: "999888777666",
				Manifest: &manifest.Manifest{
					Port:        3000,
					HealthCheck: "/api/health",
				},
			},
			setupFunc: setupTestConfig,
			validate: func(t *testing.T, vars map[string]interface{}) {
				assert.Equal(t, "test-api", vars["project_name"])
				assert.Equal(t, "staging", vars["environment"])
				assert.Equal(t, "us-west-2", vars["aws_region"])
				assert.Equal(t, 3000, vars["container_port"])
				assert.Equal(t, "/api/health", vars["health_check_path"])
			},
		},
		{
			name: "with resources and scaling",
			deployment: &BYOADeployment{
				APIName: "scaled-api",
				Manifest: &manifest.Manifest{
					Resources: &manifest.ResourceLimits{
						CPU:    "500m",
						Memory: "1Gi",
					},
					Scaling: &manifest.ScalingConfig{
						Min: 2,
						Max: 20,
					},
				},
			},
			setupFunc: setupTestConfig,
			validate: func(t *testing.T, vars map[string]interface{}) {
				assert.Equal(t, 256, vars["cpu"]) // Simplified parsing
				assert.Equal(t, 512, vars["memory"])
				assert.Equal(t, 2, vars["min_capacity"])
				assert.Equal(t, 20, vars["max_capacity"])
			},
		},
		{
			name: "with environment variables",
			deployment: &BYOADeployment{
				APIName: "env-api",
				Manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"API_KEY", "SECRET_KEY"},
						Optional: map[string]string{"DEBUG": "false"},
					},
				},
			},
			setupFunc: setupTestConfig,
			validate: func(t *testing.T, vars map[string]interface{}) {
				envVars, ok := vars["environment_variables"].(map[string]string)
				assert.True(t, ok)
				assert.Equal(t, "PLACEHOLDER_API_KEY", envVars["API_KEY"])
				assert.Equal(t, "PLACEHOLDER_SECRET_KEY", envVars["SECRET_KEY"])
			},
		},
		{
			name: "database detection",
			deployment: &BYOADeployment{
				APIName: "db-api",
				Manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"DATABASE_URL", "DB_PASSWORD"},
					},
				},
			},
			setupFunc: setupTestConfig,
			validate: func(t *testing.T, vars map[string]interface{}) {
				assert.Equal(t, true, vars["enable_database"])
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			vars := tt.deployment.getTerraformVars()
			
			if tt.validate != nil {
				tt.validate(t, vars)
			}
			
			// Common validations
			tags, ok := vars["tags"].(map[string]string)
			assert.True(t, ok)
			assert.Equal(t, "API-Direct", tags["ManagedBy"])
			assert.Equal(t, "CLI", tags["DeployedBy"])
		})
	}
}

func TestPlan(t *testing.T) {
	deployment := &BYOADeployment{
		APIName:      "test-api",
		WorkDir:      t.TempDir(),
		OutputWriter: &bytes.Buffer{},
		Manifest: &manifest.Manifest{
			Port: 8080,
		},
	}
	
	// Mock terraform command
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "terraform")
	
	scriptContent := `#!/bin/bash
case "$1" in
	"init")
		echo "Terraform initialized"
		exit 0
		;;
	"plan")
		echo "Creating terraform plan..."
		# Create plan file
		for arg in "$@"; do
			if [[ $arg == -out=* ]]; then
				planfile="${arg#-out=}"
				echo "plan data" > "$planfile"
			fi
		done
		exit 0
		;;
esac
exit 1
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)
	
	// Execute plan
	err = deployment.Plan()
	assert.NoError(t, err)
	
	// Check output
	output := deployment.OutputWriter.(*bytes.Buffer).String()
	assert.Contains(t, output, "Initializing Terraform")
	assert.Contains(t, output, "Creating deployment plan")
	assert.Contains(t, output, "Deployment plan created successfully")
	
	// Check plan file created
	planFile := filepath.Join(deployment.WorkDir, "tfplan")
	assert.FileExists(t, planFile)
}

func TestDeploy(t *testing.T) {
	deployment := &BYOADeployment{
		APIName:      "test-api",
		Environment:  "test",
		AWSRegion:    "us-east-1",
		AWSAccountID: "123456789012",
		WorkDir:      t.TempDir(),
		OutputWriter: &bytes.Buffer{},
		Manifest: &manifest.Manifest{
			Port: 8080,
		},
	}
	
	// Setup config
	cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Create plan file
	planFile := filepath.Join(deployment.WorkDir, "tfplan")
	err := os.WriteFile(planFile, []byte("mock plan"), 0644)
	require.NoError(t, err)
	
	// Mock terraform command
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "terraform")
	
	scriptContent := `#!/bin/bash
case "$1" in
	"apply")
		echo "Applying terraform plan..."
		echo "Resources created"
		exit 0
		;;
	"output")
		if [[ "$2" == "-json" ]]; then
			echo '{"api_url":{"value":"https://test-api.example.com"},"load_balancer_dns":{"value":"test-lb.elb.amazonaws.com"}}'
		fi
		exit 0
		;;
esac
exit 1
`
	
	err = os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)
	
	// Execute deploy
	result, err := deployment.Deploy()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Validate result
	assert.Equal(t, "https://test-api.example.com", result.APIURL)
	assert.Equal(t, "test-lb.elb.amazonaws.com", result.LoadBalancerDNS)
	assert.Equal(t, "us-east-1", result.AWSRegion)
	assert.Equal(t, "123456789012", result.AWSAccountID)
	assert.Contains(t, result.DeploymentID, "test-api-test-123456789012")
	
	// Check output
	output := deployment.OutputWriter.(*bytes.Buffer).String()
	assert.Contains(t, output, "Deploying infrastructure")
	assert.Contains(t, output, "Retrieving deployment details")
	
	// Check deployment saved to config
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	
	deployInfo, exists := cfg.Deployments["test-api"]
	assert.True(t, exists)
	
	info := deployInfo.(map[string]interface{})
	assert.Equal(t, "byoa", info["type"])
	assert.Equal(t, "123456789012", info["aws_account"])
	assert.Equal(t, "https://test-api.example.com", info["api_url"])
}

func TestCleanup(t *testing.T) {
	tests := []struct {
		name       string
		deployment *BYOADeployment
		setupFunc  func(*testing.T, *BYOADeployment)
		wantErr    bool
	}{
		{
			name: "cleanup valid directory",
			deployment: &BYOADeployment{
				WorkDir: filepath.Join(t.TempDir(), "apidirect-deploy-test-123"),
			},
			setupFunc: func(t *testing.T, d *BYOADeployment) {
				err := os.MkdirAll(d.WorkDir, 0755)
				require.NoError(t, err)
				
				// Create some files
				testFile := filepath.Join(d.WorkDir, "test.tf")
				err = os.WriteFile(testFile, []byte("test"), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "skip cleanup for non-apidirect directory",
			deployment: &BYOADeployment{
				WorkDir: "/tmp/important-dir",
			},
			wantErr: false,
		},
		{
			name: "empty workdir",
			deployment: &BYOADeployment{
				WorkDir: "",
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t, tt.deployment)
			}
			
			// Check directory exists before cleanup (if applicable)
			var existedBefore bool
			if tt.deployment.WorkDir != "" && strings.Contains(tt.deployment.WorkDir, "apidirect-deploy") {
				_, err := os.Stat(tt.deployment.WorkDir)
				existedBefore = err == nil
			}
			
			err := tt.deployment.Cleanup()
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify cleanup
				if existedBefore && strings.Contains(tt.deployment.WorkDir, "apidirect-deploy") {
					_, err := os.Stat(tt.deployment.WorkDir)
					assert.True(t, os.IsNotExist(err), "Directory should be removed")
				}
			}
		})
	}
}

func TestHelperMethods(t *testing.T) {
	t.Run("shouldEnableDatabase", func(t *testing.T) {
		tests := []struct {
			name     string
			manifest *manifest.Manifest
			want     bool
		}{
			{
				name: "with DATABASE_URL",
				manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"DATABASE_URL", "API_KEY"},
					},
				},
				want: true,
			},
			{
				name: "with DB_ prefix",
				manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"DB_HOST", "DB_PASSWORD"},
					},
				},
				want: true,
			},
			{
				name: "no database vars",
				manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"API_KEY", "SECRET"},
					},
				},
				want: false,
			},
			{
				name: "case insensitive",
				manifest: &manifest.Manifest{
					Env: manifest.EnvironmentVars{
						Required: []string{"database_connection"},
					},
				},
				want: true,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				d := &BYOADeployment{
					Manifest: tt.manifest,
				}
				assert.Equal(t, tt.want, d.shouldEnableDatabase())
			})
		}
	})
	
	t.Run("getContainerImage", func(t *testing.T) {
		d := &BYOADeployment{
			APIName: "my-api",
		}
		image := d.getContainerImage()
		assert.Equal(t, "my-api:latest", image)
	})
	
	t.Run("resource defaults", func(t *testing.T) {
		d := &BYOADeployment{
			Manifest: &manifest.Manifest{},
		}
		
		assert.Equal(t, 256, d.getCPUValue())
		assert.Equal(t, 512, d.getMemoryValue())
		assert.Equal(t, 1, d.getMinCapacity())
		assert.Equal(t, 10, d.getMaxCapacity())
	})
}

func TestDeploymentResult(t *testing.T) {
	result := &DeploymentResult{
		APIURL:          "https://api.example.com",
		LoadBalancerDNS: "lb.example.com",
		DeploymentID:    "test-123",
		AWSRegion:       "us-east-1",
		AWSAccountID:    "123456789012",
		Timestamp:       time.Now().Format(time.RFC3339),
	}
	
	// Test JSON serialization
	data, err := json.Marshal(result)
	assert.NoError(t, err)
	
	var decoded DeploymentResult
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	
	assert.Equal(t, result.APIURL, decoded.APIURL)
	assert.Equal(t, result.LoadBalancerDNS, decoded.LoadBalancerDNS)
	assert.Equal(t, result.DeploymentID, decoded.DeploymentID)
}

func TestStateBackend(t *testing.T) {
	backend := StateBackend{
		Bucket:   "my-state-bucket",
		Key:      "path/to/state.tfstate",
		Region:   "us-west-2",
		DynamoDB: "my-lock-table",
	}
	
	// Verify fields
	assert.Equal(t, "my-state-bucket", backend.Bucket)
	assert.Equal(t, "path/to/state.tfstate", backend.Key)
	assert.Equal(t, "us-west-2", backend.Region)
	assert.Equal(t, "my-lock-table", backend.DynamoDB)
}

func TestCreateBackendConfig(t *testing.T) {
	deployment := &BYOADeployment{
		WorkDir: t.TempDir(),
		StateBackend: StateBackend{
			Bucket:   "test-state-bucket",
			Key:      "deployments/test-api/terraform.tfstate",
			Region:   "eu-west-1",
			DynamoDB: "terraform-locks",
		},
	}
	
	err := deployment.createBackendConfig()
	assert.NoError(t, err)
	
	// Read and validate backend config
	backendFile := filepath.Join(deployment.WorkDir, "backend.tf")
	content, err := os.ReadFile(backendFile)
	require.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, `bucket         = "test-state-bucket"`)
	assert.Contains(t, contentStr, `key            = "deployments/test-api/terraform.tfstate"`)
	assert.Contains(t, contentStr, `region         = "eu-west-1"`)
	assert.Contains(t, contentStr, `dynamodb_table = "terraform-locks"`)
	assert.Contains(t, contentStr, `encrypt        = true`)
}

func TestOutputWriter(t *testing.T) {
	// Test with custom output writer
	var buf bytes.Buffer
	deployment := &BYOADeployment{
		OutputWriter: &buf,
	}
	
	// Write some output
	io.WriteString(deployment.OutputWriter, "Test output")
	
	assert.Equal(t, "Test output", buf.String())
	
	// Test default output writer
	deployment2 := &BYOADeployment{}
	assert.Nil(t, deployment2.OutputWriter)
}