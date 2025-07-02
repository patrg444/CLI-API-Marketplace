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

func TestStatusCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		configFile   string
		manifest     string
		mockAPI      func(*testing.T) *httptest.Server
		setup        func(*testing.T) string
		validate     func(*testing.T, string)
		wantErr      bool
	}{
		{
			name: "status for hosted deployment",
			configFile: `{
  "auth": {
    "access_token": "test-token",
    "refresh_token": "refresh-token",
    "expires_at": 9999999999
  },
  "api_url": "{{API_URL}}",
  "deployments": {
    "test-api": {
      "mode": "hosted",
      "id": "dep-123",
      "url": "https://test-api.api-direct.io"
    }
  }
}`,
			manifest: `name: test-api
runtime: python3.11
port: 8080
`,
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/deployments/dep-123":
						if r.Header.Get("Authorization") != "Bearer test-token" {
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
						resp := map[string]interface{}{
							"id":     "dep-123",
							"name":   "test-api",
							"status": "running",
							"url":    "https://test-api.api-direct.io",
							"instances": []map[string]interface{}{
								{
									"id":       "inst-1",
									"status":   "healthy",
									"cpu":      "15%",
									"memory":   "256MB",
									"requests": 1523,
								},
							},
							"created_at": "2024-01-01T00:00:00Z",
							"updated_at": "2024-01-01T12:00:00Z",
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
				assert.Contains(t, output, "test-api")
				assert.Contains(t, output, "Status: running")
				assert.Contains(t, output, "https://test-api.api-direct.io")
				assert.Contains(t, output, "inst-1")
				assert.Contains(t, output, "healthy")
				assert.Contains(t, output, "15%")
				assert.Contains(t, output, "256MB")
			},
		},
		{
			name: "status for BYOA deployment",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "deployments": {
    "test-api": {
      "mode": "byoa",
      "aws_account_id": "123456789012",
      "aws_region": "us-east-1",
      "infrastructure": {
        "alb_dns": "test-api-alb-123.us-east-1.elb.amazonaws.com",
        "ecs_cluster": "test-api-cluster",
        "ecs_service": "test-api-service"
      }
    }
  }
}`,
			manifest: `name: test-api
runtime: python3.11
port: 8080
`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "test-api")
				assert.Contains(t, output, "Mode: BYOA")
				assert.Contains(t, output, "AWS Account: 123456789012")
				assert.Contains(t, output, "Region: us-east-1")
				assert.Contains(t, output, "test-api-alb-123.us-east-1.elb.amazonaws.com")
			},
		},
		{
			name: "status with details flag",
			args: []string{"--details"},
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "api_url": "{{API_URL}}",
  "deployments": {
    "test-api": {
      "mode": "hosted",
      "id": "dep-123"
    }
  }
}`,
			manifest: `name: test-api
runtime: python3.11
`,
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := map[string]interface{}{
						"id":     "dep-123",
						"status": "running",
						"instances": []map[string]interface{}{
							{
								"id":          "inst-1",
								"status":      "healthy",
								"cpu":         "15%",
								"memory":      "256MB",
								"uptime":      "2d 4h",
								"last_deploy": "2024-01-01T12:00:00Z",
							},
						},
						"metrics": map[string]interface{}{
							"requests_total":  15234,
							"requests_per_min": 25,
							"avg_response_time": "124ms",
							"error_rate":       "0.1%",
						},
						"recent_logs": []string{
							"[INFO] Server started on port 8080",
							"[INFO] Connected to database",
							"[INFO] Health check passed",
						},
					}
					json.NewEncoder(w).Encode(resp)
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				// Should show detailed metrics
				assert.Contains(t, output, "Metrics")
				assert.Contains(t, output, "requests_total")
				assert.Contains(t, output, "15234")
				assert.Contains(t, output, "avg_response_time")
				assert.Contains(t, output, "124ms")
				assert.Contains(t, output, "Recent Logs")
				assert.Contains(t, output, "Server started")
			},
		},
		{
			name: "status for specific API",
			args: []string{"my-other-api"},
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "deployments": {
    "test-api": {
      "mode": "hosted",
      "id": "dep-123"
    },
    "my-other-api": {
      "mode": "hosted",
      "id": "dep-456",
      "status": "stopped"
    }
  }
}`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "my-other-api")
				assert.Contains(t, output, "stopped")
				assert.NotContains(t, output, "test-api")
			},
		},
		{
			name: "status with no deployments",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "deployments": {}
}`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "No deployments found")
				assert.Contains(t, output, "apidirect deploy")
			},
		},
		{
			name: "status when not authenticated",
			configFile: `{}`,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "status with API error",
			configFile: `{
  "auth": {
    "access_token": "test-token"
  },
  "api_url": "{{API_URL}}",
  "deployments": {
    "test-api": {
      "mode": "hosted",
      "id": "dep-123"
    }
  }
}`,
			manifest: `name: test-api`,
			mockAPI: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}))
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Error fetching status")
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
			
			// Create config directory and file
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
			
			// Write config file with API URL
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
			
			// Write manifest if provided
			if tt.manifest != "" {
				err = ioutil.WriteFile("apidirect.yaml", []byte(tt.manifest), 0644)
				require.NoError(t, err)
			}
			
			// Create command
			rootCmd := &cobra.Command{
				Use:   "apidirect",
				Short: "API-Direct CLI",
			}
			
			// Mock status command
			testStatusCmd := &cobra.Command{
				Use:   "status [api-name]",
				Short: "Check deployment status",
				RunE: func(cmd *cobra.Command, args []string) error {
					details, _ := cmd.Flags().GetBool("details")
					
					// Check authentication
					configPath := filepath.Join(".apidirect", "config.json")
					configData, err := ioutil.ReadFile(configPath)
					if err != nil {
						return fmt.Errorf("not authenticated")
					}
					
					var config map[string]interface{}
					if err := json.Unmarshal(configData, &config); err != nil {
						return err
					}
					
					auth, ok := config["auth"].(map[string]interface{})
					if !ok || auth["access_token"] == nil {
						return fmt.Errorf("not authenticated. Run 'apidirect auth login'")
					}
					
					deployments, ok := config["deployments"].(map[string]interface{})
					if !ok || len(deployments) == 0 {
						cmd.Println("No deployments found.")
						cmd.Println("💡 Deploy your first API: apidirect deploy")
						return nil
					}
					
					// Filter by API name if provided
					if len(args) > 0 {
						apiName := args[0]
						if dep, ok := deployments[apiName]; ok {
							printDeploymentStatus(cmd, apiName, dep.(map[string]interface{}), details, config)
						} else {
							return fmt.Errorf("deployment not found: %s", apiName)
						}
					} else {
						// Show all deployments
						for name, dep := range deployments {
							printDeploymentStatus(cmd, name, dep.(map[string]interface{}), details, config)
							cmd.Println()
						}
					}
					
					return nil
				},
			}
			
			// Add flags
			testStatusCmd.Flags().BoolP("details", "d", false, "Show detailed status")
			
			// Capture output
			output := &bytes.Buffer{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)
			
			// Add command
			rootCmd.AddCommand(testStatusCmd)
			
			// Execute
			args := append([]string{"status"}, tt.args...)
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

func TestStatusHelperFunctions(t *testing.T) {
	t.Run("formatDeploymentMode", func(t *testing.T) {
		testCases := []struct {
			mode     string
			expected string
		}{
			{"hosted", "Hosted by API-Direct"},
			{"byoa", "BYOA (Your AWS Account)"},
			{"unknown", "Unknown"},
		}
		
		for _, tc := range testCases {
			result := formatDeploymentMode(tc.mode)
			assert.Equal(t, tc.expected, result)
		}
	})
	
	t.Run("formatInstanceStatus", func(t *testing.T) {
		testCases := []struct {
			status   string
			expected string
		}{
			{"healthy", "✅ healthy"},
			{"unhealthy", "❌ unhealthy"},
			{"starting", "🔄 starting"},
			{"stopping", "🛑 stopping"},
			{"unknown", "❓ unknown"},
		}
		
		for _, tc := range testCases {
			result := formatInstanceStatus(tc.status)
			assert.Equal(t, tc.expected, result)
		}
	})
}

// Helper function implementations
func printDeploymentStatus(cmd *cobra.Command, name string, dep map[string]interface{}, details bool, config map[string]interface{}) {
	cmd.Printf("📊 %s\n", name)
	cmd.Println(strings.Repeat("─", 40))
	
	mode, _ := dep["mode"].(string)
	cmd.Printf("Mode: %s\n", formatDeploymentMode(mode))
	
	if mode == "hosted" {
		// Fetch status from API
		apiURL, _ := config["api_url"].(string)
		token, _ := config["auth"].(map[string]interface{})["access_token"].(string)
		depID, _ := dep["id"].(string)
		
		if apiURL != "" && token != "" && depID != "" {
			req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/deployments/%s", apiURL, depID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				cmd.Printf("Error fetching status: %v\n", err)
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				cmd.Println("Error fetching status from API")
				return
			}
			
			var depStatus map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&depStatus)
			
			status, _ := depStatus["status"].(string)
			cmd.Printf("Status: %s\n", status)
			
			if url, ok := depStatus["url"].(string); ok {
				cmd.Printf("URL: %s\n", url)
			}
			
			// Show instances
			if instances, ok := depStatus["instances"].([]interface{}); ok {
				cmd.Println("\nInstances:")
				for _, inst := range instances {
					instance := inst.(map[string]interface{})
					id, _ := instance["id"].(string)
					status, _ := instance["status"].(string)
					cpu, _ := instance["cpu"].(string)
					memory, _ := instance["memory"].(string)
					
					cmd.Printf("  • %s: %s (CPU: %s, Memory: %s)\n",
						id, formatInstanceStatus(status), cpu, memory)
				}
			}
			
			// Show details if requested
			if details {
				if metrics, ok := depStatus["metrics"].(map[string]interface{}); ok {
					cmd.Println("\nMetrics:")
					for key, value := range metrics {
						cmd.Printf("  • %s: %v\n", key, value)
					}
				}
				
				if logs, ok := depStatus["recent_logs"].([]interface{}); ok {
					cmd.Println("\nRecent Logs:")
					for _, log := range logs {
						cmd.Printf("  %s\n", log)
					}
				}
			}
		}
	} else if mode == "byoa" {
		// Show BYOA details
		awsAccount, _ := dep["aws_account_id"].(string)
		awsRegion, _ := dep["aws_region"].(string)
		
		cmd.Printf("AWS Account: %s\n", awsAccount)
		cmd.Printf("Region: %s\n", awsRegion)
		
		if infra, ok := dep["infrastructure"].(map[string]interface{}); ok {
			if albDNS, ok := infra["alb_dns"].(string); ok {
				cmd.Printf("Load Balancer: %s\n", albDNS)
			}
			if cluster, ok := infra["ecs_cluster"].(string); ok {
				cmd.Printf("ECS Cluster: %s\n", cluster)
			}
		}
		
		cmd.Println("\n💡 To check AWS resources directly:")
		cmd.Printf("   aws ecs describe-services --cluster %s --services %s\n",
			dep["infrastructure"].(map[string]interface{})["ecs_cluster"],
			dep["infrastructure"].(map[string]interface{})["ecs_service"])
	}
	
	// Show status if available
	if status, ok := dep["status"].(string); ok {
		cmd.Printf("Status: %s\n", status)
	}
}

func formatDeploymentMode(mode string) string {
	switch mode {
	case "hosted":
		return "Hosted by API-Direct"
	case "byoa":
		return "BYOA (Your AWS Account)"
	default:
		return "Unknown"
	}
}

func formatInstanceStatus(status string) string {
	switch status {
	case "healthy":
		return "✅ healthy"
	case "unhealthy":
		return "❌ unhealthy"
	case "starting":
		return "🔄 starting"
	case "stopping":
		return "🛑 stopping"
	default:
		return "❓ " + status
	}
}