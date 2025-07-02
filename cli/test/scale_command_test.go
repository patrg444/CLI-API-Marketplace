package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock structures matching scale.go
type DeploymentScale struct {
	CurrentReplicas int                `json:"current_replicas"`
	DesiredReplicas int                `json:"desired_replicas"`
	MinReplicas     int                `json:"min_replicas"`
	MaxReplicas     int                `json:"max_replicas"`
	AutoScaling     bool               `json:"auto_scaling"`
	CPUThreshold    int                `json:"cpu_threshold"`
	Resources       ResourceAllocation `json:"resources"`
	Status          string             `json:"status"`
	LastScaled      time.Time          `json:"last_scaled"`
}

type ResourceAllocation struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type ScaleRequest struct {
	Replicas    *int                `json:"replicas,omitempty"`
	MinReplicas *int                `json:"min_replicas,omitempty"`
	MaxReplicas *int                `json:"max_replicas,omitempty"`
	AutoScaling *bool               `json:"auto_scaling,omitempty"`
	CPUTarget   *int                `json:"cpu_target,omitempty"`
	Resources   *ResourceAllocation `json:"resources,omitempty"`
}

func TestScaleAPIEndpoints(t *testing.T) {
	tests := []struct {
		name         string
		apiName      string
		request      *ScaleRequest
		mockResponse func(w http.ResponseWriter, r *http.Request)
		wantErr      bool
	}{
		{
			name:    "successful scale to replicas",
			apiName: "test-api",
			request: &ScaleRequest{
				Replicas: intPtr(5),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					scale := DeploymentScale{
						CurrentReplicas: 3,
						DesiredReplicas: 3,
						Status:          "healthy",
					}
					json.NewEncoder(w).Encode(scale)
				} else if r.Method == "PUT" {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
					
					var req ScaleRequest
					json.NewDecoder(r.Body).Decode(&req)
					assert.NotNil(t, req.Replicas)
					assert.Equal(t, 5, *req.Replicas)
					
					scale := DeploymentScale{
						CurrentReplicas: 3,
						DesiredReplicas: 5,
						Status:          "scaling",
					}
					json.NewEncoder(w).Encode(scale)
				}
			},
			wantErr: false,
		},
		{
			name:    "enable auto-scaling",
			apiName: "test-api",
			request: &ScaleRequest{
				AutoScaling: boolPtr(true),
				MinReplicas: intPtr(2),
				MaxReplicas: intPtr(10),
				CPUTarget:   intPtr(70),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					scale := DeploymentScale{
						CurrentReplicas: 1,
						AutoScaling:     false,
					}
					json.NewEncoder(w).Encode(scale)
				} else if r.Method == "PUT" {
					var req ScaleRequest
					json.NewDecoder(r.Body).Decode(&req)
					assert.True(t, *req.AutoScaling)
					assert.Equal(t, 2, *req.MinReplicas)
					assert.Equal(t, 10, *req.MaxReplicas)
					assert.Equal(t, 70, *req.CPUTarget)
					
					scale := DeploymentScale{
						CurrentReplicas: 2,
						MinReplicas:     2,
						MaxReplicas:     10,
						AutoScaling:     true,
						CPUThreshold:    70,
					}
					json.NewEncoder(w).Encode(scale)
				}
			},
			wantErr: false,
		},
		{
			name:    "scale to zero",
			apiName: "test-api",
			request: &ScaleRequest{
				Replicas: intPtr(0),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					scale := DeploymentScale{CurrentReplicas: 3}
					json.NewEncoder(w).Encode(scale)
				} else if r.Method == "PUT" {
					scale := DeploymentScale{
						CurrentReplicas: 0,
						DesiredReplicas: 0,
					}
					json.NewEncoder(w).Encode(scale)
				}
			},
			wantErr: false,
		},
		{
			name:    "update resources",
			apiName: "test-api",
			request: &ScaleRequest{
				Resources: &ResourceAllocation{
					Memory: "2Gi",
				},
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					scale := DeploymentScale{
						Resources: ResourceAllocation{
							Memory: "1Gi",
							CPU:    "500m",
						},
					}
					json.NewEncoder(w).Encode(scale)
				} else if r.Method == "PUT" {
					var req ScaleRequest
					json.NewDecoder(r.Body).Decode(&req)
					assert.NotNil(t, req.Resources)
					assert.Equal(t, "2Gi", req.Resources.Memory)
					
					scale := DeploymentScale{
						Resources: ResourceAllocation{
							Memory: "2Gi",
							CPU:    "500m",
						},
					}
					json.NewEncoder(w).Encode(scale)
				}
			},
			wantErr: false,
		},
		{
			name:    "server error",
			apiName: "test-api",
			request: &ScaleRequest{
				Replicas: intPtr(5),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					scale := DeploymentScale{CurrentReplicas: 3}
					json.NewEncoder(w).Encode(scale)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal server error"))
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			testDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", testDir)
			defer os.Setenv("HOME", oldHome)

			// Create config directory
			configDir := filepath.Join(testDir, ".apidirect")
			err := os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.mockResponse(w, r)
			}))
			defer server.Close()

			// Save test config
			cfg := &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				API: config.APIConfig{
					BaseURL: server.URL,
				},
			}
			err = config.SaveConfig(cfg)
			require.NoError(t, err)

			// Test GET endpoint
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/deployment/v1/scale/%s", server.URL, tt.apiName), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			if !tt.wantErr {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}

			// Test PUT endpoint
			if tt.request != nil {
				body, err := json.Marshal(tt.request)
				require.NoError(t, err)

				req, err = http.NewRequest("PUT", fmt.Sprintf("%s/deployment/v1/scale/%s", server.URL, tt.apiName), bytes.NewBuffer(body))
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
				req.Header.Set("Content-Type", "application/json")

				resp, err = client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				if tt.wantErr {
					assert.NotEqual(t, http.StatusOK, resp.StatusCode)
				} else {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}

func TestValidateScaleInputs(t *testing.T) {
	tests := []struct {
		name        string
		replicas    int
		min         int
		max         int
		cpu         int
		memory      string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid inputs",
			replicas: 5,
			min:      2,
			max:      10,
			cpu:      70,
			memory:   "1Gi",
			wantErr:  false,
		},
		{
			name:        "negative replicas",
			replicas:    -1,
			wantErr:     true,
			errContains: "negative",
		},
		{
			name:        "min greater than max",
			min:         10,
			max:         5,
			wantErr:     true,
			errContains: "greater than max",
		},
		{
			name:        "invalid CPU",
			cpu:         150,
			wantErr:     true,
			errContains: "between 1 and 100",
		},
		{
			name:        "invalid memory",
			memory:      "invalid",
			wantErr:     true,
			errContains: "invalid memory format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation logic
			err := validateInputs(tt.replicas, tt.min, tt.max, tt.cpu, tt.memory)
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

func TestResourceStringValidation(t *testing.T) {
	validFormats := []string{
		"512Mi", "1Gi", "2.5Gi", "1024Ki", "100m", "1000", "256", "1.5Ti", "500k",
	}
	
	invalidFormats := []string{
		"invalid", "Mi512", "", "-100Mi",
	}
	
	for _, format := range validFormats {
		t.Run("valid_"+format, func(t *testing.T) {
			assert.True(t, isValidResourceString(format))
		})
	}
	
	for _, format := range invalidFormats {
		t.Run("invalid_"+format, func(t *testing.T) {
			assert.False(t, isValidResourceString(format))
		})
	}
}

func TestScaleRequestBuilding(t *testing.T) {
	t.Run("scale to zero", func(t *testing.T) {
		req := buildScaleRequest(true, false, 0, 0, 0, 0, "", nil)
		assert.NotNil(t, req.Replicas)
		assert.Equal(t, 0, *req.Replicas)
	})
	
	t.Run("scale down", func(t *testing.T) {
		current := &DeploymentScale{MinReplicas: 2}
		req := buildScaleRequest(false, true, 0, 0, 0, 0, "", current)
		assert.NotNil(t, req.Replicas)
		assert.Equal(t, 2, *req.Replicas)
	})
	
	t.Run("enable auto-scaling", func(t *testing.T) {
		req := buildScaleRequest(false, false, 0, 2, 10, 70, "", nil)
		assert.NotNil(t, req.MinReplicas)
		assert.Equal(t, 2, *req.MinReplicas)
		assert.NotNil(t, req.MaxReplicas)
		assert.Equal(t, 10, *req.MaxReplicas)
		assert.NotNil(t, req.CPUTarget)
		assert.Equal(t, 70, *req.CPUTarget)
	})
	
	t.Run("set replicas", func(t *testing.T) {
		req := buildScaleRequest(false, false, 5, 0, 0, 0, "", nil)
		assert.NotNil(t, req.Replicas)
		assert.Equal(t, 5, *req.Replicas)
		assert.NotNil(t, req.AutoScaling)
		assert.False(t, *req.AutoScaling)
	})
	
	t.Run("update memory", func(t *testing.T) {
		req := buildScaleRequest(false, false, 0, 0, 0, 0, "2Gi", nil)
		assert.NotNil(t, req.Resources)
		assert.Equal(t, "2Gi", req.Resources.Memory)
	})
}

// Helper functions

func validateInputs(replicas, min, max, cpu int, memory string) error {
	if replicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}
	
	if min < 0 || max < 0 {
		return fmt.Errorf("min/max replicas cannot be negative")
	}
	
	if min > 0 && max > 0 && min > max {
		return fmt.Errorf("min replicas cannot be greater than max replicas")
	}
	
	if cpu != 0 && (cpu < 1 || cpu > 100) {
		return fmt.Errorf("CPU threshold must be between 1 and 100")
	}
	
	if memory != "" && !isValidResourceString(memory) {
		return fmt.Errorf("invalid memory format")
	}
	
	return nil
}

func isValidResourceString(resource string) bool {
	validSuffixes := []string{"Ki", "Mi", "Gi", "Ti", "m", "k", "M", "G", "T"}
	
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(resource, suffix) {
			numberPart := strings.TrimSuffix(resource, suffix)
			var f float64
			_, err := fmt.Sscanf(numberPart, "%f", &f)
			return err == nil && f >= 0
		}
	}
	
	// Plain number
	var f float64
	_, err := fmt.Sscanf(resource, "%f", &f)
	return err == nil && f >= 0
}

func buildScaleRequest(toZero, down bool, replicas, min, max, cpu int, memory string, current *DeploymentScale) *ScaleRequest {
	req := &ScaleRequest{}
	
	if toZero {
		req.Replicas = intPtr(0)
		return req
	}
	
	if down && current != nil {
		minRep := current.MinReplicas
		if minRep == 0 {
			minRep = 1
		}
		req.Replicas = intPtr(minRep)
		return req
	}
	
	if replicas > 0 {
		req.Replicas = intPtr(replicas)
		req.AutoScaling = boolPtr(false)
	}
	
	if min > 0 {
		req.MinReplicas = intPtr(min)
	}
	
	if max > 0 {
		req.MaxReplicas = intPtr(max)
	}
	
	if cpu > 0 {
		req.CPUTarget = intPtr(cpu)
	}
	
	if memory != "" {
		req.Resources = &ResourceAllocation{Memory: memory}
	}
	
	return req
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}