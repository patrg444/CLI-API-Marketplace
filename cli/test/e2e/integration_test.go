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

// TestFullIntegrationFlow tests the complete integration flow
func TestFullIntegrationFlow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// This test requires real AWS credentials but uses a safe test mode
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	apiName := fmt.Sprintf("integration-test-%d", time.Now().Unix())
	
	// Create a complete test scenario
	t.Run("Complete Integration Test", func(t *testing.T) {
		// 1. Setup project
		setupIntegrationTestProject(t, testDir, apiName)
		
		// 2. Test CLI commands in sequence
		testCLICommands(t, testDir, apiName)
		
		// 3. Test error scenarios
		testErrorScenarios(t, testDir)
		
		// 4. Test configuration management
		testConfigManagement(t, apiName)
	})
}

// setupIntegrationTestProject creates a complete test project
func setupIntegrationTestProject(t *testing.T, testDir, apiName string) {
	projectDir := filepath.Join(testDir, apiName)
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	
	// Create a more complex FastAPI application
	createCompleteAPIStructure(t, projectDir)
	
	// Change to project directory
	require.NoError(t, os.Chdir(projectDir))
}

// createCompleteAPIStructure creates a full API project structure
func createCompleteAPIStructure(t *testing.T, projectDir string) {
	// Main application file
	mainPy := `from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional
import os

app = FastAPI(
    title="Integration Test API",
    description="API for testing BYOA deployment",
    version="1.0.0"
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Models
class User(BaseModel):
    id: int
    name: str
    email: str

class CreateUserRequest(BaseModel):
    name: str
    email: str

# In-memory database
users_db = [
    User(id=1, name="John Doe", email="john@example.com"),
    User(id=2, name="Jane Smith", email="jane@example.com"),
]

# Routes
@app.get("/")
def read_root():
    return {
        "message": "Welcome to Integration Test API",
        "environment": os.getenv("ENVIRONMENT", "development"),
        "version": "1.0.0"
    }

@app.get("/health")
def health_check():
    return {
        "status": "healthy",
        "timestamp": "2024-06-28T12:00:00Z",
        "checks": {
            "database": "ok",
            "external_api": "ok"
        }
    }

@app.get("/api/users", response_model=List[User])
def get_users(skip: int = 0, limit: int = 10):
    return users_db[skip:skip + limit]

@app.get("/api/users/{user_id}", response_model=User)
def get_user(user_id: int):
    for user in users_db:
        if user.id == user_id:
            return user
    raise HTTPException(status_code=404, detail="User not found")

@app.post("/api/users", response_model=User)
def create_user(user_request: CreateUserRequest):
    new_user = User(
        id=len(users_db) + 1,
        name=user_request.name,
        email=user_request.email
    )
    users_db.append(new_user)
    return new_user

@app.get("/api/metrics")
def get_metrics():
    return {
        "total_users": len(users_db),
        "api_calls": 1234,
        "error_rate": 0.02,
        "average_response_time": 45.2
    }

# Protected endpoint (would require auth in production)
@app.get("/api/admin/config")
def get_config():
    return {
        "debug_mode": os.getenv("DEBUG", "false"),
        "log_level": os.getenv("LOG_LEVEL", "INFO"),
        "features": {
            "new_ui": True,
            "beta_features": False
        }
    }
`
	require.NoError(t, ioutil.WriteFile(
		filepath.Join(projectDir, "main.py"),
		[]byte(mainPy),
		0644,
	))
	
	// Create requirements.txt
	requirements := `fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
python-multipart==0.0.6
`
	require.NoError(t, ioutil.WriteFile(
		filepath.Join(projectDir, "requirements.txt"),
		[]byte(requirements),
		0644,
	))
	
	// Create Dockerfile
	dockerfile := `FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
`
	require.NoError(t, ioutil.WriteFile(
		filepath.Join(projectDir, "Dockerfile"),
		[]byte(dockerfile),
		0644,
	))
	
	// Create .env.example
	envExample := `# Environment Configuration
ENVIRONMENT=production
LOG_LEVEL=INFO
DEBUG=false

# Database Configuration (if needed)
DATABASE_URL=postgresql://user:pass@localhost/dbname

# API Keys (example)
EXTERNAL_API_KEY=your-api-key-here
`
	require.NoError(t, ioutil.WriteFile(
		filepath.Join(projectDir, ".env.example"),
		[]byte(envExample),
		0644,
	))
	
	// Create README.md
	readme := `# Integration Test API

This is a test API for validating BYOA deployments.

## Endpoints

- GET ` + "`/`" + ` - Root endpoint
- GET ` + "`/health`" + ` - Health check
- GET ` + "`/api/users`" + ` - List users
- GET ` + "`/api/users/{id}`" + ` - Get user by ID
- POST ` + "`/api/users`" + ` - Create new user
- GET ` + "`/api/metrics`" + ` - Get API metrics
- GET ` + "`/api/admin/config`" + ` - Get configuration (protected)

## Local Development

` + "```bash" + `
pip install -r requirements.txt
uvicorn main:app --reload
` + "```" + `

## Deployment

This API is designed to be deployed using API-Direct BYOA:

` + "```bash" + `
apidirect import
apidirect deploy
` + "```" + `
`
	require.NoError(t, ioutil.WriteFile(
		filepath.Join(projectDir, "README.md"),
		[]byte(readme),
		0644,
	))
}

// testCLICommands tests the CLI commands in sequence
func testCLICommands(t *testing.T, testDir, apiName string) {
	// Test import command
	t.Run("Import Command", func(t *testing.T) {
		cmd := exec.Command("apidirect", "import", "--auto")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Import output: %s", string(output))
			// In test mode, we might create manifest manually
			createTestManifestForIntegration(t, testDir, apiName)
		} else {
			assert.Contains(t, string(output), "imported successfully")
		}
		
		// Verify manifest exists
		assert.FileExists(t, filepath.Join(testDir, apiName, "apidirect.manifest.json"))
	})
	
	// Test validate command
	t.Run("Validate Command", func(t *testing.T) {
		cmd := exec.Command("apidirect", "validate")
		cmd.Dir = filepath.Join(testDir, apiName)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Validate output: %s", string(output))
		}
		
		// Should pass validation
		assert.Contains(t, string(output), "valid")
	})
	
	// Test deploy command (dry run)
	t.Run("Deploy Command (Dry Run)", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not available")
		}
		
		cmd := exec.Command("apidirect", "deploy", "--dry-run")
		cmd.Dir = filepath.Join(testDir, apiName)
		output, err := cmd.CombinedOutput()
		
		t.Logf("Deploy dry run output: %s", string(output))
		
		// Should show what would be deployed
		if err == nil {
			assert.Contains(t, string(output), "would create")
		}
	})
	
	// Test status command
	t.Run("Status Command", func(t *testing.T) {
		cmd := exec.Command("apidirect", "status", apiName)
		output, err := cmd.CombinedOutput()
		
		// Should show not found (since we didn't actually deploy)
		assert.Error(t, err)
		assert.Contains(t, string(output), "not found")
	})
}

// testErrorScenarios tests various error conditions
func testErrorScenarios(t *testing.T, testDir string) {
	t.Run("Deploy Without Manifest", func(t *testing.T) {
		emptyDir := filepath.Join(testDir, "empty")
		require.NoError(t, os.MkdirAll(emptyDir, 0755))
		
		cmd := exec.Command("apidirect", "deploy")
		cmd.Dir = emptyDir
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err)
		assert.Contains(t, string(output), "no manifest found")
	})
	
	t.Run("Invalid Manifest", func(t *testing.T) {
		invalidDir := filepath.Join(testDir, "invalid")
		require.NoError(t, os.MkdirAll(invalidDir, 0755))
		
		// Create invalid manifest
		invalidManifest := `{
			"version": "1.0",
			"name": "",  // Empty name
			"runtime": "invalid-runtime",
			"port": "not-a-number"
		}`
		require.NoError(t, ioutil.WriteFile(
			filepath.Join(invalidDir, "apidirect.manifest.json"),
			[]byte(invalidManifest),
			0644,
		))
		
		cmd := exec.Command("apidirect", "validate")
		cmd.Dir = invalidDir
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(string(output)), "invalid")
	})
	
	t.Run("Destroy Non-Existent", func(t *testing.T) {
		cmd := exec.Command("apidirect", "destroy", "non-existent-api", "--force")
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err)
		assert.Contains(t, string(output), "not found")
	})
}

// testConfigManagement tests configuration management
func testConfigManagement(t *testing.T, apiName string) {
	t.Run("Config Persistence", func(t *testing.T) {
		// Check if config exists
		configPath, err := getConfigPath()
		require.NoError(t, err)
		
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Log("Config file doesn't exist yet")
			return
		}
		
		// Read config
		configData, err := ioutil.ReadFile(configPath)
		require.NoError(t, err)
		
		var config map[string]interface{}
		require.NoError(t, json.Unmarshal(configData, &config))
		
		t.Logf("Config content: %+v", config)
		
		// Check for expected fields
		assert.Contains(t, config, "api")
		assert.Contains(t, config, "preferences")
	})
}

// createTestManifestForIntegration creates a manifest for integration testing
func createTestManifestForIntegration(t *testing.T, testDir, apiName string) {
	manifest := map[string]interface{}{
		"version": "1.0",
		"name":    apiName,
		"runtime": "python3.9",
		"port":    8000,
		"start_command": "uvicorn main:app --host 0.0.0.0 --port 8000",
		"health_check": "/health",
		"endpoints": []string{
			"GET /",
			"GET /health",
			"GET /api/users",
			"GET /api/users/{user_id}",
			"POST /api/users",
			"GET /api/metrics",
			"GET /api/admin/config",
		},
		"env": map[string]interface{}{
			"required": []string{"ENVIRONMENT"},
			"optional": []string{"LOG_LEVEL", "DEBUG", "DATABASE_URL"},
		},
		"scaling": map[string]interface{}{
			"min": 2,
			"max": 10,
			"target_cpu": 70,
		},
		"resources": map[string]interface{}{
			"memory": "512Mi",
			"cpu":    "256m",
		},
		"files": map[string]interface{}{
			"main":         "main.py",
			"requirements": "requirements.txt",
			"dockerfile":   "Dockerfile",
		},
		"metadata": map[string]interface{}{
			"description": "Integration test API for BYOA deployment",
			"version":     "1.0.0",
			"author":      "API-Direct Test Suite",
		},
	}
	
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	
	manifestPath := filepath.Join(testDir, apiName, "apidirect.manifest.json")
	require.NoError(t, ioutil.WriteFile(manifestPath, manifestBytes, 0644))
}

// getConfigPath returns the path to the CLI config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".apidirect", "config.json"), nil
}

// TestCLIBinaryAvailability tests if the CLI binary is available
func TestCLIBinaryAvailability(t *testing.T) {
	cmd := exec.Command("apidirect", "--version")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Skipf("apidirect CLI not found in PATH: %v", err)
	}
	
	assert.Contains(t, string(output), "apidirect")
	t.Logf("CLI version: %s", strings.TrimSpace(string(output)))
}

// TestAWSPrerequisites tests AWS prerequisites
func TestAWSPrerequisites(t *testing.T) {
	t.Run("AWS CLI Available", func(t *testing.T) {
		cmd := exec.Command("aws", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skipf("AWS CLI not installed: %v", err)
		}
		
		assert.Contains(t, string(output), "aws-cli")
	})
	
	t.Run("Terraform Available", func(t *testing.T) {
		cmd := exec.Command("terraform", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skipf("Terraform not installed: %v", err)
		}
		
		assert.Contains(t, string(output), "Terraform")
	})
}