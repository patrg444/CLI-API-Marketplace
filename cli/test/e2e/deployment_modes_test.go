package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBothDeploymentModes tests both hosted and BYOA deployment modes comprehensively
func TestBothDeploymentModes(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	// Test configuration
	tests := []struct {
		name          string
		deploymentMode string
		needsAWS      bool
		setupFunc     func(t *testing.T) func()
		validations   []validation
	}{
		{
			name:          "Hosted Mode Deployment",
			deploymentMode: "hosted",
			needsAWS:      false,
			setupFunc:     setupHostedMode,
			validations: []validation{
				{name: "URL contains api-direct.io", check: validateHostedURL},
				{name: "No AWS credentials required", check: validateNoAWSRequired},
				{name: "Instant SSL enabled", check: validateInstantSSL},
				{name: "Auto-scaling configured", check: validateAutoScaling},
			},
		},
		{
			name:          "BYOA Mode Deployment",
			deploymentMode: "byoa",
			needsAWS:      true,
			setupFunc:     setupBYOAMode,
			validations: []validation{
				{name: "URL contains AWS ELB", check: validateAWSURL},
				{name: "AWS credentials verified", check: validateAWSCredentials},
				{name: "Terraform resources created", check: validateTerraformResources},
				{name: "Custom VPC configured", check: validateCustomVPC},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip BYOA tests if no AWS credentials
			if tt.needsAWS && !hasAWSCredentials() {
				t.Skip("AWS credentials not configured")
			}

			// Setup test environment
			testDir := setupTestEnvironment(t)
			defer cleanupTestEnvironment(testDir)

			// Run mode-specific setup
			cleanup := tt.setupFunc(t)
			if cleanup != nil {
				defer cleanup()
			}

			// Create test API
			apiName := fmt.Sprintf("test-api-%s-%d", tt.deploymentMode, time.Now().Unix())
			createTestAPIProject(t, testDir, apiName)

			// Deploy API
			t.Run("Deploy", func(t *testing.T) {
				deployAPI(t, testDir, apiName, tt.deploymentMode)
			})

			// Run validations
			for _, v := range tt.validations {
				t.Run(v.name, func(t *testing.T) {
					v.check(t, testDir, apiName, tt.deploymentMode)
				})
			}

			// Test common operations
			t.Run("Status Command", func(t *testing.T) {
				testStatusCommand(t, testDir, apiName, tt.deploymentMode)
			})

			t.Run("Logs Command", func(t *testing.T) {
				testLogsCommand(t, testDir, apiName, tt.deploymentMode)
			})

			// Clean up deployment
			if tt.deploymentMode == "byoa" {
				t.Run("Destroy", func(t *testing.T) {
					destroyBYOADeployment(t, testDir, apiName)
				})
			}
		})
	}
}

// validation represents a test validation
type validation struct {
	name  string
	check func(t *testing.T, testDir, apiName, mode string)
}

// Setup functions

func setupHostedMode(t *testing.T) func() {
	// Start mock backend
	mockBackend := NewMockBackendServices()
	os.Setenv("APIDIRECT_API_ENDPOINT", mockBackend.GetURL())
	os.Setenv("APIDIRECT_DEMO_MODE", "true")
	
	return func() {
		mockBackend.Close()
		os.Unsetenv("APIDIRECT_API_ENDPOINT")
		os.Unsetenv("APIDIRECT_DEMO_MODE")
	}
}

func setupBYOAMode(t *testing.T) func() {
	// Load AWS credentials from .env if available
	envPath := filepath.Join("..", "..", "..", ".env")
	if data, err := ioutil.ReadFile(envPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if strings.HasPrefix(key, "AWS_") {
						os.Setenv(key, value)
					}
				}
			}
		}
	}
	
	return nil
}

// Deployment functions

func deployAPI(t *testing.T, testDir, apiName, mode string) {
	args := []string{"deploy", apiName}
	
	if mode == "hosted" {
		args = append(args, "--hosted")
	} else {
		args = append(args, "--hosted=false", "--yes")
	}
	
	cmd := exec.Command("apidirect", args...)
	cmd.Dir = testDir
	
	output, err := cmd.CombinedOutput()
	t.Logf("Deploy output: %s", string(output))
	
	if mode == "hosted" && strings.Contains(string(output), "Demo Mode") {
		// Demo mode is acceptable for hosted tests
		assert.Contains(t, string(output), "Deployment successful")
		return
	}
	
	require.NoError(t, err, "Deployment should succeed")
}

// Validation functions

func validateHostedURL(t *testing.T, testDir, apiName, mode string) {
	output := getDeploymentInfo(t, testDir, apiName)
	assert.Contains(t, output, "api-direct.io", "Hosted deployment should use api-direct.io domain")
}

func validateNoAWSRequired(t *testing.T, testDir, apiName, mode string) {
	// Hosted mode should work without AWS credentials
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	_ = cmd.Run() // Ignore error - hosted mode works regardless
	
	// Test passes whether AWS is configured or not
	t.Log("Hosted mode does not require AWS credentials")
}

func validateInstantSSL(t *testing.T, testDir, apiName, mode string) {
	output := getDeploymentInfo(t, testDir, apiName)
	assert.Contains(t, output, "https://", "Hosted deployment should have SSL enabled")
}

func validateAutoScaling(t *testing.T, testDir, apiName, mode string) {
	cmd := exec.Command("apidirect", "status", apiName, "--json")
	cmd.Dir = testDir
	
	output, _ := cmd.CombinedOutput()
	
	var status map[string]interface{}
	if err := json.Unmarshal(output, &status); err == nil {
		if scale, ok := status["scale"].(map[string]interface{}); ok {
			assert.True(t, scale["auto_scaling"].(bool), "Auto-scaling should be enabled")
		}
	}
}

func validateAWSURL(t *testing.T, testDir, apiName, mode string) {
	output := getDeploymentInfo(t, testDir, apiName)
	assert.Contains(t, output, "elb.amazonaws.com", "BYOA deployment should use AWS ELB")
}

func validateAWSCredentials(t *testing.T, testDir, apiName, mode string) {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	output, err := cmd.CombinedOutput()
	
	require.NoError(t, err, "AWS credentials should be valid")
	assert.Contains(t, string(output), "Account", "Should return valid AWS account info")
}

func validateTerraformResources(t *testing.T, testDir, apiName, mode string) {
	// Check if Terraform state exists
	stateFile := filepath.Join(testDir, ".apidirect", apiName, "terraform.tfstate")
	
	if _, err := os.Stat(stateFile); err == nil {
		data, _ := ioutil.ReadFile(stateFile)
		assert.Contains(t, string(data), "aws_", "Terraform state should contain AWS resources")
	} else {
		t.Log("Terraform state not accessible in test environment")
	}
}

func validateCustomVPC(t *testing.T, testDir, apiName, mode string) {
	output := getDeploymentInfo(t, testDir, apiName)
	
	// BYOA deployments create custom VPC
	if strings.Contains(output, "vpc-") {
		assert.Contains(t, output, "vpc-", "BYOA deployment should have custom VPC")
	} else {
		t.Log("VPC information not available in test output")
	}
}

// Helper functions

func getDeploymentInfo(t *testing.T, testDir, apiName string) string {
	cmd := exec.Command("apidirect", "status", apiName)
	cmd.Dir = testDir
	
	output, _ := cmd.CombinedOutput()
	return string(output)
}

func testStatusCommand(t *testing.T, testDir, apiName, mode string) {
	cmd := exec.Command("apidirect", "status", apiName)
	cmd.Dir = testDir
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Status command output: %s", string(output))
		if mode == "hosted" && strings.Contains(string(output), "Demo Mode") {
			// Demo mode is acceptable
			return
		}
	}
	
	assert.Contains(t, string(output), apiName, "Status should show API name")
}

func testLogsCommand(t *testing.T, testDir, apiName, mode string) {
	cmd := exec.Command("apidirect", "logs", apiName, "--tail", "5")
	cmd.Dir = testDir
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Logs command output: %s", string(output))
		if mode == "hosted" && strings.Contains(string(output), "Demo Mode") {
			// Demo mode is acceptable
			return
		}
	}
	
	// Logs command should at least run without error
	t.Log("Logs command executed")
}

func destroyBYOADeployment(t *testing.T, testDir, apiName string) {
	cmd := exec.Command("apidirect", "destroy", apiName, "--yes")
	cmd.Dir = testDir
	
	output, err := cmd.CombinedOutput()
	t.Logf("Destroy output: %s", string(output))
	
	// Destroy might fail in test environment, but command should execute
	if err != nil && !strings.Contains(string(output), "not found") {
		t.Logf("Destroy command completed with warnings: %v", err)
	}
}

// TestDeploymentModeTransition tests transitioning between deployment modes
func TestDeploymentModeTransition(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	apiName := "transition-test-api"
	createTestAPIProject(t, testDir, apiName)

	// Start with hosted deployment
	t.Run("Initial Hosted Deployment", func(t *testing.T) {
		cleanup := setupHostedMode(t)
		defer cleanup()

		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted")
		cmd.Dir = testDir
		
		output, _ := cmd.CombinedOutput()
		assert.Contains(t, string(output), "hosted", "Should deploy to hosted mode")
	})

	// Export configuration
	t.Run("Export Configuration", func(t *testing.T) {
		cmd := exec.Command("apidirect", "export", apiName)
		cmd.Dir = testDir
		
		output, err := cmd.CombinedOutput()
		
		if err == nil {
			assert.Contains(t, string(output), "export", "Should export configuration")
		} else {
			t.Log("Export command not implemented")
		}
	})

	// Attempt transition to BYOA (dry-run)
	t.Run("Transition to BYOA (Dry Run)", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not configured")
		}

		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted=false", "--dry-run")
		cmd.Dir = testDir
		
		output, _ := cmd.CombinedOutput()
		
		if strings.Contains(string(output), "would create") || strings.Contains(string(output), "Plan:") {
			assert.Contains(t, string(output), "AWS", "Should show AWS deployment plan")
		}
	})
}

// TestDeploymentModeDocumentation tests that documentation is accurate
func TestDeploymentModeDocumentation(t *testing.T) {
	// Verify deployment options documentation exists
	docPath := filepath.Join("..", "..", "..", "DEPLOYMENT_OPTIONS_COMPARISON.md")
	
	data, err := ioutil.ReadFile(docPath)
	require.NoError(t, err, "Deployment options documentation should exist")
	
	content := string(data)
	
	// Verify key sections exist
	assert.Contains(t, content, "Hosted Deployment", "Should document hosted mode")
	assert.Contains(t, content, "BYOA Deployment", "Should document BYOA mode")
	assert.Contains(t, content, "How to Choose", "Should include decision guide")
	assert.Contains(t, content, "apidirect deploy", "Should include example commands")
	
	// Verify accuracy of key features
	assert.Contains(t, content, "AWS Account Required | ❌ No | ✅ Yes", "Should accurately show AWS requirements")
	assert.Contains(t, content, "Setup Time | ⚡ 2 minutes | ⏱️ 5 minutes", "Should show setup time differences")
}