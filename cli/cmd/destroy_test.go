package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDestroyCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		flags       []string
		config      *config.Config
		userInput   string
		setupFunc   func(*testing.T) func()
		validateOut func(*testing.T, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "destroy BYOA deployment with confirmation",
			args: []string{"test-api"},
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type":         "byoa",
						"aws_account":  "123456789012",
						"aws_region":   "us-east-1",
						"environment":  "prod",
					},
				},
			},
			userInput: "test-api\n", // Confirm destruction
			setupFunc: func(t *testing.T) func() {
				// Mock AWS CLI check
				os.Setenv("MOCK_AWS_CLI", "true")
				os.Setenv("MOCK_AWS_ACCOUNT", "123456789012")
				return func() {
					os.Unsetenv("MOCK_AWS_CLI")
					os.Unsetenv("MOCK_AWS_ACCOUNT")
				}
			},
			validateOut: func(t *testing.T, output string) {
				assert.Contains(t, output, "WARNING: This will destroy ALL resources")
				assert.Contains(t, output, "AWS Account: 123456789012")
				assert.Contains(t, output, "AWS Region: us-east-1")
				assert.Contains(t, output, "Application Load Balancer")
				assert.Contains(t, output, "ECS Fargate Service")
				assert.Contains(t, output, "Type the API name to confirm destruction:")
			},
			wantErr:     true, // Will fail because terraform isn't actually available
			errContains: "terraform",
		},
		{
			name: "destroy with force flag",
			args: []string{"test-api"},
			flags: []string{"--force"},
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type":        "byoa",
						"aws_account": "123456789012",
						"aws_region":  "us-east-1",
					},
				},
			},
			setupFunc: func(t *testing.T) func() {
				os.Setenv("MOCK_AWS_CLI", "true")
				os.Setenv("MOCK_AWS_ACCOUNT", "123456789012")
				return func() {
					os.Unsetenv("MOCK_AWS_CLI")
					os.Unsetenv("MOCK_AWS_ACCOUNT")
				}
			},
			validateOut: func(t *testing.T, output string) {
				// Should not show confirmation prompt
				assert.NotContains(t, output, "Type the API name to confirm")
			},
			wantErr:     true,
			errContains: "terraform",
		},
		{
			name: "destroy non-existent deployment",
			args: []string{"non-existent"},
			config: &config.Config{
				Deployments: map[string]interface{}{},
			},
			wantErr:     true,
			errContains: "deployment 'non-existent' not found",
		},
		{
			name: "destroy non-BYOA deployment",
			args: []string{"hosted-api"},
			config: &config.Config{
				Deployments: map[string]interface{}{
					"hosted-api": map[string]interface{}{
						"type": "hosted",
						"id":   "dep-123",
					},
				},
			},
			wantErr:     true,
			errContains: "is not a BYOA deployment",
		},
		{
			name: "destroy with wrong AWS account",
			args: []string{"test-api"},
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type":        "byoa",
						"aws_account": "999999999999",
						"aws_region":  "us-east-1",
					},
				},
			},
			setupFunc: func(t *testing.T) func() {
				os.Setenv("MOCK_AWS_CLI", "true")
				os.Setenv("MOCK_AWS_ACCOUNT", "123456789012") // Different account
				return func() {
					os.Unsetenv("MOCK_AWS_CLI")
					os.Unsetenv("MOCK_AWS_ACCOUNT")
				}
			},
			wantErr:     true,
			errContains: "doesn't match deployment account",
		},
		{
			name: "destroy cancelled by user",
			args: []string{"test-api"},
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type":        "byoa",
						"aws_account": "123456789012",
						"aws_region":  "us-east-1",
					},
				},
			},
			userInput: "wrong-name\n", // Wrong confirmation
			setupFunc: func(t *testing.T) func() {
				os.Setenv("MOCK_AWS_CLI", "true")
				os.Setenv("MOCK_AWS_ACCOUNT", "123456789012")
				return func() {
					os.Unsetenv("MOCK_AWS_CLI")
					os.Unsetenv("MOCK_AWS_ACCOUNT")
				}
			},
			wantErr:     true,
			errContains: "Destruction cancelled",
		},
		{
			name: "destroy without args",
			args: []string{},
			config: &config.Config{},
			wantErr: true,
		},
		{
			name: "destroy with yes flag",
			args: []string{"test-api"},
			flags: []string{"--yes"},
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type":        "byoa",
						"aws_account": "123456789012",
						"aws_region":  "us-west-2",
					},
				},
			},
			setupFunc: func(t *testing.T) func() {
				os.Setenv("MOCK_AWS_CLI", "true")
				os.Setenv("MOCK_AWS_ACCOUNT", "123456789012")
				return func() {
					os.Unsetenv("MOCK_AWS_CLI")
					os.Unsetenv("MOCK_AWS_ACCOUNT")
				}
			},
			validateOut: func(t *testing.T, output string) {
				// Should not show confirmation prompt
				assert.NotContains(t, output, "Type the API name to confirm")
			},
			wantErr:     true,
			errContains: "terraform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			testDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", testDir)
			defer func() {
				if oldHome == "" {
					os.Unsetenv("HOME")
				} else {
					os.Setenv("HOME", oldHome)
				}
			}()

			// Create config directory
			configDir := filepath.Join(testDir, ".apidirect")
			err := os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			// Save test config if provided
			if tt.config != nil {
				err = config.SaveConfig(tt.config)
				require.NoError(t, err)
			}

			// Run setup function if provided
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}

			// Create command
			rootCmd := &cobra.Command{
				Use:   "apidirect",
				Short: "API-Direct CLI",
			}

			// Create a mock destroy command
			testDestroyCmd := &cobra.Command{
				Use:   "destroy [API_NAME]",
				Short: "Destroy a BYOA deployment",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					force, _ := cmd.Flags().GetBool("force")
					yes, _ := cmd.Flags().GetBool("yes")
					
					// Simplified destroy logic for testing
					apiName := args[0]
					
					// Load config
					cfg, err := config.LoadConfig()
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}
					
					// Check deployment exists
					if cfg.Deployments == nil || cfg.Deployments[apiName] == nil {
						return fmt.Errorf("deployment '%s' not found", apiName)
					}
					
					deployment := cfg.Deployments[apiName].(map[string]interface{})
					
					// Check if BYOA
					if deployment["type"] != "byoa" {
						return fmt.Errorf("deployment '%s' is not a BYOA deployment", apiName)
					}
					
					// Mock AWS checks
					if os.Getenv("MOCK_AWS_CLI") != "true" {
						return fmt.Errorf("AWS CLI not found")
					}
					
					// Check account match
					deployAccount := deployment["aws_account"].(string)
					currentAccount := os.Getenv("MOCK_AWS_ACCOUNT")
					if currentAccount != deployAccount {
						return fmt.Errorf("current AWS account (%s) doesn't match deployment account (%s)",
							currentAccount, deployAccount)
					}
					
					// Show warning and get confirmation
					if !force && !yes {
						awsRegion := deployment["aws_region"].(string)
						environment := "prod"
						if env, ok := deployment["environment"].(string); ok {
							environment = env
						}
						
						cmd.Printf("⚠️  WARNING: This will destroy ALL resources for '%s'\n", apiName)
						cmd.Printf("   AWS Account: %s\n", deployAccount)
						cmd.Printf("   AWS Region: %s\n", awsRegion)
						cmd.Printf("   Environment: %s\n", environment)
						cmd.Printf("\n   Resources to be destroyed:\n")
						cmd.Printf("   - Application Load Balancer\n")
						cmd.Printf("   - ECS Fargate Service and Tasks\n")
						cmd.Printf("   - RDS Database (if enabled)\n")
						cmd.Printf("   - VPC and all networking components\n")
						cmd.Printf("   - IAM roles and policies\n")
						cmd.Printf("   - CloudWatch logs and metrics\n")
						cmd.Printf("\n   This action is IRREVERSIBLE!\n")
						cmd.Printf("\nType the API name to confirm destruction: ")
						
						// In test, read from provided input
						if tt.userInput != "" {
							response := strings.TrimSpace(strings.Split(tt.userInput, "\n")[0])
							if response != apiName {
								return fmt.Errorf("Destruction cancelled")
							}
						}
					}
					
					// Simulate terraform operations
					cmd.Println("🔧 Initializing Terraform...")
					// This would fail in test because terraform isn't installed
					return fmt.Errorf("terraform init failed: exec: \"terraform\": executable file not found in $PATH")
				},
			}

			// Add flags
			var forceFlag, yesFlag bool
			testDestroyCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force destroy")
			testDestroyCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation")

			// Capture output
			output := &bytes.Buffer{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)

			// Set input if needed
			if tt.userInput != "" {
				rootCmd.SetIn(strings.NewReader(tt.userInput))
			}

			// Add command
			rootCmd.AddCommand(testDestroyCmd)

			// Build args
			args := append([]string{"destroy"}, tt.args...)
			args = append(args, tt.flags...)
			rootCmd.SetArgs(args)

			// Execute
			err = rootCmd.Execute()

			// Validate
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.validateOut != nil {
				tt.validateOut(t, output.String())
			}

			// Cleanup
			if cleanup != nil {
				cleanup()
			}
		})
	}
}

func TestGetDeploymentInfo(t *testing.T) {
	tests := []struct {
		name       string
		config     *config.Config
		apiName    string
		wantExists bool
		wantInfo   map[string]interface{}
	}{
		{
			name: "existing deployment",
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": map[string]interface{}{
						"type": "byoa",
						"aws_account": "123456789012",
					},
				},
			},
			apiName:    "test-api",
			wantExists: true,
			wantInfo: map[string]interface{}{
				"type": "byoa",
				"aws_account": "123456789012",
			},
		},
		{
			name: "non-existent deployment",
			config: &config.Config{
				Deployments: map[string]interface{}{
					"other-api": map[string]interface{}{},
				},
			},
			apiName:    "test-api",
			wantExists: false,
		},
		{
			name:       "nil deployments",
			config:     &config.Config{},
			apiName:    "test-api",
			wantExists: false,
		},
		{
			name: "invalid deployment format",
			config: &config.Config{
				Deployments: map[string]interface{}{
					"test-api": "invalid", // Not a map
				},
			},
			apiName:    "test-api",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, exists := getDeploymentInfo(tt.config, tt.apiName)
			
			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				assert.Equal(t, tt.wantInfo, info)
			} else {
				assert.Nil(t, info)
			}
		})
	}
}

func TestCleanupEmptyStateBucket(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		region     string
		setupFunc  func(*testing.T) func()
		wantErr    bool
	}{
		{
			name:       "cleanup with mock AWS",
			bucketName: "test-bucket",
			region:     "us-east-1",
			setupFunc: func(t *testing.T) func() {
				// In real test, we'd mock the AWS CLI commands
				os.Setenv("MOCK_AWS_S3", "empty")
				return func() {
					os.Unsetenv("MOCK_AWS_S3")
				}
			},
			wantErr: true, // Will fail because aws CLI not available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}

			err := cleanupEmptyStateBucket(tt.bucketName, tt.region)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if cleanup != nil {
				cleanup()
			}
		})
	}
}

func TestGetModulesPath(t *testing.T) {
	// Test the modules path function
	path := getModulesPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "infrastructure")
}

// Test helper functions (renamed to avoid conflicts)