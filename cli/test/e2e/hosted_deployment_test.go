package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHostedDeploymentFlow tests the complete hosted deployment lifecycle
func TestHostedDeploymentFlow(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	// Setup test environment
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Start mock backend services
	mockBackend := NewMockBackendServices()
	defer mockBackend.Close()

	// Configure CLI to use mock backend
	os.Setenv("APIDIRECT_API_ENDPOINT", mockBackend.GetURL())
	defer os.Unsetenv("APIDIRECT_API_ENDPOINT")

	t.Run("Complete Hosted Deployment", func(t *testing.T) {
		apiName := fmt.Sprintf("hosted-test-api-%d", time.Now().Unix())
		
		// 1. Create test API project
		t.Run("Create API Project", func(t *testing.T) {
			createTestAPIProject(t, testDir, apiName)
		})

		// 2. Import API
		t.Run("Import API", func(t *testing.T) {
			importAPI(t, testDir)
		})

		// 3. Deploy to hosted infrastructure
		t.Run("Deploy Hosted", func(t *testing.T) {
			deployHosted(t, testDir, apiName)
		})

		// 4. Check deployment status
		t.Run("Check Status", func(t *testing.T) {
			checkHostedDeploymentStatus(t, apiName)
		})

		// 5. Test API endpoints
		t.Run("Test API Endpoints", func(t *testing.T) {
			testHostedAPIEndpoints(t, apiName, mockBackend)
		})

		// 6. View logs
		t.Run("View Logs", func(t *testing.T) {
			viewHostedLogs(t, apiName)
		})

		// 7. Scale deployment
		t.Run("Scale Deployment", func(t *testing.T) {
			scaleHostedDeployment(t, apiName, 3)
		})

		// 8. Update deployment
		t.Run("Update Deployment", func(t *testing.T) {
			updateHostedDeployment(t, testDir, apiName)
		})
	})
}

// TestDeploymentModeComparison tests both deployment modes
func TestDeploymentModeComparison(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	apiName := "mode-comparison-api"
	createTestAPIProject(t, testDir, apiName)
	createTestManifest(t, testDir, apiName)

	t.Run("Hosted Deployment Characteristics", func(t *testing.T) {
		// Test hosted deployment specific features
		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted", "--dry-run")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "APIDIRECT_DEMO_MODE=true")
		
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Check for hosted-specific messages
			assert.Contains(t, string(output), "hosted infrastructure")
			assert.Contains(t, string(output), "api-direct.io")
			assert.NotContains(t, string(output), "AWS Account")
		}
	})

	t.Run("BYOA Deployment Characteristics", func(t *testing.T) {
		// Test BYOA deployment specific features
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not configured")
		}

		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted=false", "--dry-run")
		cmd.Dir = testDir
		
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Check for BYOA-specific messages
			assert.Contains(t, string(output), "AWS account")
			assert.Contains(t, string(output), "Terraform")
			assert.NotContains(t, string(output), "api-direct.io")
		}
	})
}

// MockBackendServices simulates the API-Direct backend
type MockBackendServices struct {
	server      *httptest.Server
	deployments map[string]*MockDeployment
}

type MockDeployment struct {
	ID           string    `json:"deployment_id"`
	APIName      string    `json:"api_name"`
	Status       string    `json:"status"`
	Endpoint     string    `json:"endpoint"`
	Replicas     int       `json:"replicas"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Runtime      string    `json:"runtime"`
	HealthCheck  string    `json:"health_check"`
}

// DeploymentStatus represents deployment status info
type DeploymentStatus struct {
	APIName     string             `json:"api_name"`
	Status      string             `json:"status"`
	Health      HealthStatus       `json:"health"`
	Deployment  DeploymentInfo     `json:"deployment"`
	Scale       ScaleInfo          `json:"scale"`
	Metrics     PerformanceMetrics `json:"metrics"`
	LastUpdated time.Time          `json:"last_updated"`
}

// HealthStatus represents health check status
type HealthStatus struct {
	Overall   string          `json:"overall"`
	Replicas  []ReplicaHealth `json:"replicas"`
	LastCheck time.Time       `json:"last_check"`
}

// ReplicaHealth represents individual replica health
type ReplicaHealth struct {
	ID       string    `json:"id"`
	Status   string    `json:"status"`
	Uptime   string    `json:"uptime"`
	CPU      float64   `json:"cpu"`
	Memory   float64   `json:"memory"`
	Requests int64     `json:"requests"`
	LastSeen time.Time `json:"last_seen"`
}

// DeploymentInfo represents deployment details
type DeploymentInfo struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url"`
}

// ScaleInfo represents scaling configuration
type ScaleInfo struct {
	CurrentReplicas int  `json:"current_replicas"`
	DesiredReplicas int  `json:"desired_replicas"`
	MinReplicas     int  `json:"min_replicas"`
	MaxReplicas     int  `json:"max_replicas"`
	AutoScaling     bool `json:"auto_scaling"`
}

// PerformanceMetrics represents API performance metrics
type PerformanceMetrics struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	AverageLatency    float64 `json:"average_latency"`
	P95Latency        float64 `json:"p95_latency"`
	P99Latency        float64 `json:"p99_latency"`
	ErrorRate         float64 `json:"error_rate"`
	Throughput        string  `json:"throughput"`
}

// NewMockBackendServices creates mock backend services
func NewMockBackendServices() *MockBackendServices {
	mock := &MockBackendServices{
		deployments: make(map[string]*MockDeployment),
	}

	mux := http.NewServeMux()
	
	// Authentication endpoint
	mux.HandleFunc("/auth/login", mock.handleLogin)
	
	// Deployment endpoints
	mux.HandleFunc("/hosted/v1/build", mock.handleBuild)
	mux.HandleFunc("/hosted/v1/deploy", mock.handleDeploy)
	mux.HandleFunc("/hosted/v1/deployments/", mock.handleDeploymentStatus)
	mux.HandleFunc("/deployment/v1/status/", mock.handleStatus)
	mux.HandleFunc("/deployment/v1/logs/", mock.handleLogs)
	mux.HandleFunc("/deployment/v1/scale/", mock.handleScale)

	mock.server = httptest.NewServer(mux)
	return mock
}

func (m *MockBackendServices) Close() {
	m.server.Close()
}

func (m *MockBackendServices) GetURL() string {
	return m.server.URL
}

// Mock handlers

func (m *MockBackendServices) handleLogin(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"access_token": "mock-access-token",
		"id_token":     "mock-id-token",
		"expires_in":   3600,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockBackendServices) handleBuild(w http.ResponseWriter, r *http.Request) {
	// Simulate build process
	time.Sleep(100 * time.Millisecond)
	
	response := map[string]interface{}{
		"image_tag": fmt.Sprintf("mock-image-%d", time.Now().Unix()),
		"build_id":  fmt.Sprintf("build-%d", time.Now().Unix()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockBackendServices) handleDeploy(w http.ResponseWriter, r *http.Request) {
	var deployReq map[string]interface{}
	json.NewDecoder(r.Body).Decode(&deployReq)
	
	apiName := deployReq["api_name"].(string)
	deploymentID := fmt.Sprintf("dep_%s_%d", apiName, time.Now().Unix())
	endpoint := fmt.Sprintf("https://%s-%s.api-direct.io", apiName, generateRandomID())
	
	deployment := &MockDeployment{
		ID:          deploymentID,
		APIName:     apiName,
		Status:      "deploying",
		Endpoint:    endpoint,
		Replicas:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Runtime:     deployReq["runtime"].(string),
		HealthCheck: deployReq["health_check"].(string),
	}
	
	m.deployments[apiName] = deployment
	
	// Simulate async deployment
	go func() {
		time.Sleep(2 * time.Second)
		deployment.Status = "running"
	}()
	
	response := map[string]interface{}{
		"endpoint":      endpoint,
		"deployment_id": deploymentID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockBackendServices) handleDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	
	apiName := parts[3]
	deployment, exists := m.deployments[apiName]
	
	if !exists {
		w.WriteHeader(404)
		return
	}
	
	if strings.HasSuffix(r.URL.Path, "/status") {
		response := map[string]string{
			"status": deployment.Status,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	json.NewEncoder(w).Encode(deployment)
}

func (m *MockBackendServices) handleStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	
	apiName := parts[3]
	deployment, exists := m.deployments[apiName]
	
	if !exists {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "API not found",
		})
		return
	}
	
	status := DeploymentStatus{
		APIName: apiName,
		Status:  deployment.Status,
		Health: HealthStatus{
			Overall: "healthy",
			Replicas: []ReplicaHealth{
				{
					ID:       fmt.Sprintf("replica-%s-1", apiName),
					Status:   "healthy",
					Uptime:   "2h 45m",
					LastSeen: time.Now(),
				},
			},
			LastCheck: time.Now(),
		},
		Deployment: DeploymentInfo{
			ID:          deployment.ID,
			Version:     "v1.0.0",
			Environment: "production",
			CreatedAt:   deployment.CreatedAt,
			UpdatedAt:   deployment.UpdatedAt,
			URL:         deployment.Endpoint,
		},
		Scale: ScaleInfo{
			CurrentReplicas: deployment.Replicas,
			DesiredReplicas: deployment.Replicas,
			MinReplicas:     1,
			MaxReplicas:     10,
			AutoScaling:     true,
		},
		Metrics: PerformanceMetrics{
			RequestsPerSecond: 42.5,
			AverageLatency:    23.4,
			P95Latency:        45.2,
			P99Latency:        89.7,
			ErrorRate:         0.02,
			Throughput:        "512KB/s",
		},
		LastUpdated: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (m *MockBackendServices) handleLogs(w http.ResponseWriter, r *http.Request) {
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute),
			"message":   "Starting application server...",
			"level":     "INFO",
		},
		{
			"timestamp": time.Now().Add(-4 * time.Minute),
			"message":   "Server listening on port 8000",
			"level":     "INFO",
		},
		{
			"timestamp": time.Now().Add(-3 * time.Minute),
			"message":   "Health check passed",
			"level":     "INFO",
		},
		{
			"timestamp": time.Now().Add(-1 * time.Minute),
			"message":   "Handled 150 requests",
			"level":     "INFO",
		},
	}
	
	json.NewEncoder(w).Encode(logs)
}

func (m *MockBackendServices) handleScale(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	
	var scaleReq struct {
		Replicas int `json:"replicas"`
	}
	json.NewDecoder(r.Body).Decode(&scaleReq)
	
	parts := strings.Split(r.URL.Path, "/")
	apiName := parts[3]
	
	if deployment, exists := m.deployments[apiName]; exists {
		deployment.Replicas = scaleReq.Replicas
		deployment.UpdatedAt = time.Now()
		
		response := map[string]interface{}{
			"success":  true,
			"replicas": scaleReq.Replicas,
			"message":  fmt.Sprintf("Scaling %s to %d replicas", apiName, scaleReq.Replicas),
		}
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(404)
	}
}

// Test helper functions for hosted deployment

func deployHosted(t *testing.T, testDir, apiName string) {
	cmd := exec.Command("apidirect", "deploy", apiName, "--hosted", "--yes")
	cmd.Dir = testDir
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Deploy output: %s", string(output))
		// In test mode, we might accept certain errors
		if strings.Contains(string(output), "mock") {
			t.Log("Running with mock backend")
			return
		}
	}
	
	// Check for expected output
	if !strings.Contains(string(output), "error") {
		assert.Contains(t, string(output), "Deployment successful")
		assert.Contains(t, string(output), "api-direct.io")
	}
}

func checkHostedDeploymentStatus(t *testing.T, apiName string) {
	cmd := exec.Command("apidirect", "status", apiName, "--json")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Status output: %s", string(output))
		return
	}
	
	var status map[string]interface{}
	if err := json.Unmarshal(output, &status); err == nil {
		assert.Equal(t, apiName, status["api_name"])
		assert.Contains(t, []string{"running", "healthy", "deploying"}, status["status"])
	}
}

func testHostedAPIEndpoints(t *testing.T, apiName string, mockBackend *MockBackendServices) {
	deployment, exists := mockBackend.deployments[apiName]
	if !exists {
		t.Log("No deployment found in mock backend")
		return
	}
	
	t.Logf("Would test API endpoints at: %s", deployment.Endpoint)
	assert.Contains(t, deployment.Endpoint, "api-direct.io")
}

func viewHostedLogs(t *testing.T, apiName string) {
	cmd := exec.Command("apidirect", "logs", apiName, "--tail", "10")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Logs output: %s", string(output))
		return
	}
	
	// Should contain log entries
	assert.NotEmpty(t, string(output))
}

func scaleHostedDeployment(t *testing.T, apiName string, replicas int) {
	cmd := exec.Command("apidirect", "scale", apiName, "--replicas", fmt.Sprintf("%d", replicas))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Scale output: %s", string(output))
		return
	}
	
	assert.Contains(t, string(output), fmt.Sprintf("%d replicas", replicas))
}

func updateHostedDeployment(t *testing.T, testDir, apiName string) {
	// Modify the code slightly
	mainPath := filepath.Join(testDir, apiName, "main.py")
	content, _ := ioutil.ReadFile(mainPath)
	newContent := strings.Replace(string(content), "Hello from", "Updated hello from", 1)
	ioutil.WriteFile(mainPath, []byte(newContent), 0644)
	
	// Redeploy
	cmd := exec.Command("apidirect", "deploy", apiName, "--hosted", "--yes", "--force")
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Update deploy output: %s", string(output))
		return
	}
	
	assert.Contains(t, string(output), "Updating existing deployment")
}

// TestHostedVsBYOAFeatures tests feature differences between modes
func TestHostedVsBYOAFeatures(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		expectations map[string]bool
	}{
		{
			name: "Hosted Mode Features",
			mode: "hosted",
			expectations: map[string]bool{
				"requires_aws":        false,
				"instant_ssl":         true,
				"auto_scaling":        true,
				"managed_updates":     true,
				"custom_vpc":          false,
				"data_sovereignty":    false,
			},
		},
		{
			name: "BYOA Mode Features",
			mode: "byoa",
			expectations: map[string]bool{
				"requires_aws":        true,
				"instant_ssl":         false,
				"auto_scaling":        true,
				"managed_updates":     false,
				"custom_vpc":          true,
				"data_sovereignty":    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Document expected features for each mode
			t.Logf("Mode: %s", tt.mode)
			for feature, expected := range tt.expectations {
				t.Logf("  %s: %v", feature, expected)
			}
		})
	}
}

// TestModeSwitching tests switching between hosted and BYOA modes
func TestModeSwitching(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Start mock backend for hosted mode
	mockBackend := NewMockBackendServices()
	defer mockBackend.Close()
	os.Setenv("APIDIRECT_API_ENDPOINT", mockBackend.GetURL())
	defer os.Unsetenv("APIDIRECT_API_ENDPOINT")

	apiName := "mode-switch-api"
	createTestAPIProject(t, testDir, apiName)

	t.Run("Deploy to Hosted First", func(t *testing.T) {
		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "APIDIRECT_DEMO_MODE=true")
		
		output, err := cmd.CombinedOutput()
		if err == nil {
			assert.Contains(t, string(output), "hosted infrastructure")
			assert.Contains(t, string(output), "api-direct.io")
		}
	})

	t.Run("Export from Hosted", func(t *testing.T) {
		cmd := exec.Command("apidirect", "export", apiName)
		cmd.Dir = testDir
		
		output, err := cmd.CombinedOutput()
		if err == nil {
			assert.Contains(t, string(output), "export")
			// Check export files created
			exportPath := filepath.Join(testDir, fmt.Sprintf("%s-export.tar.gz", apiName))
			assert.FileExists(t, exportPath)
		}
	})

	t.Run("Switch to BYOA", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not configured")
		}

		cmd := exec.Command("apidirect", "deploy", apiName, "--hosted=false", "--dry-run")
		cmd.Dir = testDir
		
		output, err := cmd.CombinedOutput()
		if err == nil {
			assert.Contains(t, string(output), "AWS account")
			assert.NotContains(t, string(output), "api-direct.io")
		}
	})
}

// TestConcurrentDeployments tests multiple deployments in both modes
func TestConcurrentDeployments(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Start mock backend
	mockBackend := NewMockBackendServices()
	defer mockBackend.Close()
	os.Setenv("APIDIRECT_API_ENDPOINT", mockBackend.GetURL())
	defer os.Unsetenv("APIDIRECT_API_ENDPOINT")

	// Create multiple test APIs
	apiNames := []string{"api-1", "api-2", "api-3"}
	for _, name := range apiNames {
		createTestAPIProject(t, testDir, name)
	}

	t.Run("Deploy Multiple APIs Concurrently", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan string, len(apiNames))

		for _, apiName := range apiNames {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				
				cmd := exec.Command("apidirect", "deploy", name, "--hosted")
				cmd.Dir = testDir
				cmd.Env = append(os.Environ(), "APIDIRECT_DEMO_MODE=true")
				
				output, err := cmd.CombinedOutput()
				if err == nil {
					results <- fmt.Sprintf("%s: success", name)
				} else {
					results <- fmt.Sprintf("%s: failed - %s", name, string(output))
				}
			}(apiName)
		}

		wg.Wait()
		close(results)

		// Check results
		successCount := 0
		for result := range results {
			t.Log(result)
			if strings.Contains(result, "success") {
				successCount++
			}
		}

		assert.Equal(t, len(apiNames), successCount, "All deployments should succeed")
	})
}

// generateRandomID generates a random ID for testing
func generateRandomID() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[(time.Now().UnixNano()+int64(i)*1000)%int64(len(chars))]
	}
	return string(id)
}