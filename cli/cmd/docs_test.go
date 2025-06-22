package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDocsGenerateCommand(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create test manifest
	manifest := map[string]interface{}{
		"name":        "test-api",
		"description": "Test API for documentation",
		"version":     "1.0.0",
		"framework":   "express",
		"language":    "javascript",
	}
	
	manifestData, _ := yaml.Marshal(manifest)
	manifestPath := filepath.Join(tempDir, "apidirect.yaml")
	os.WriteFile(manifestPath, manifestData, 0644)
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)
	
	tests := []struct {
		name       string
		format     string
		wantFiles  []string
		wantErr    bool
	}{
		{
			name:   "generate openapi",
			format: "openapi",
			wantFiles: []string{
				"docs/openapi.yaml",
			},
		},
		{
			name:   "generate markdown",
			format: "markdown",
			wantFiles: []string{
				"docs/API_DOCUMENTATION.md",
			},
		},
		{
			name:   "generate html",
			format: "html",
			wantFiles: []string{
				"docs/index.html",
				"docs/openapi.yaml",
			},
		},
		{
			name:   "generate postman",
			format: "postman",
			wantFiles: []string{
				"docs/test-api_postman_collection.json",
			},
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean docs directory
			os.RemoveAll(filepath.Join(tempDir, "docs"))
			
			// Set flags
			docsFormat = tt.format
			docsOutput = "./docs"
			
			err := runGenerateDocs(nil, []string{"test-api"})
			
			if (err != nil) != tt.wantErr {
				t.Errorf("runGenerateDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Check files were created
				for _, file := range tt.wantFiles {
					path := filepath.Join(tempDir, file)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("expected file %s was not created", file)
					}
				}
			}
		})
	}
}

func TestDocsPublishCommand(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		// Check method and path
		if r.Method != "POST" || !strings.Contains(r.URL.Path, "/docs/publish") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Return success response
		response := map[string]interface{}{
			"url":    "https://docs.apidirect.io/test-api",
			"status": "published",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override API endpoint
	os.Setenv("APIDIRECT_API_ENDPOINT", server.URL)
	defer os.Unsetenv("APIDIRECT_API_ENDPOINT")
	
	// Set auth token
	os.Setenv("APIDIRECT_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("APIDIRECT_AUTH_TOKEN")
	
	cmd := &cobra.Command{}
	err := runPublishDocs(cmd, []string{"test-api"})
	
	if err != nil {
		t.Errorf("runPublishDocs() failed: %v", err)
	}
}

func TestGenerateOpenAPIDoc(t *testing.T) {
	tempDir := t.TempDir()
	docsOutput = tempDir
	
	manifest := map[string]interface{}{
		"name":        "test-api",
		"description": "Test API",
		"version":     "1.0.0",
	}
	
	endpoints := []map[string]interface{}{
		{
			"path":        "/users",
			"method":      "GET",
			"description": "List users",
			"parameters": []map[string]interface{}{
				{"name": "page", "type": "integer", "required": false},
			},
		},
	}
	
	err := generateOpenAPIDoc("test-api", manifest, endpoints)
	if err != nil {
		t.Fatalf("generateOpenAPIDoc() failed: %v", err)
	}
	
	// Check file was created
	openAPIPath := filepath.Join(tempDir, "openapi.yaml")
	data, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("failed to read openapi.yaml: %v", err)
	}
	
	// Parse and validate
	var openapi map[string]interface{}
	if err := yaml.Unmarshal(data, &openapi); err != nil {
		t.Fatalf("failed to parse openapi.yaml: %v", err)
	}
	
	// Check structure
	if openapi["openapi"] != "3.0.0" {
		t.Error("expected OpenAPI version 3.0.0")
	}
	
	info := openapi["info"].(map[string]interface{})
	if info["title"] != "test-api" {
		t.Error("expected API title to be test-api")
	}
	
	paths := openapi["paths"].(map[string]interface{})
	if _, ok := paths["/users"]; !ok {
		t.Error("expected /users path in OpenAPI spec")
	}
}

func TestGenerateMarkdownDoc(t *testing.T) {
	tempDir := t.TempDir()
	docsOutput = tempDir
	
	manifest := map[string]interface{}{
		"name":        "test-api",
		"description": "Test API for markdown generation",
	}
	
	endpoints := []map[string]interface{}{
		{
			"path":        "/users",
			"method":      "GET",
			"description": "List all users",
			"parameters": []map[string]interface{}{
				{
					"name":        "limit",
					"type":        "integer",
					"required":    false,
					"description": "Number of results",
				},
			},
		},
		{
			"path":        "/users/{id}",
			"method":      "GET",
			"description": "Get user by ID",
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"type":        "string",
					"required":    true,
					"description": "User ID",
				},
			},
		},
	}
	
	err := generateMarkdownDoc("test-api", manifest, endpoints)
	if err != nil {
		t.Fatalf("generateMarkdownDoc() failed: %v", err)
	}
	
	// Check file was created
	mdPath := filepath.Join(tempDir, "API_DOCUMENTATION.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read markdown file: %v", err)
	}
	
	content := string(data)
	
	// Check content
	expectedStrings := []string{
		"# test-api API Documentation",
		"Test API for markdown generation",
		"## Table of Contents",
		"## Endpoints",
		"GET /users",
		"List all users",
		"| limit | integer | false |",
		"curl -X GET",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("markdown missing expected string: %q", expected)
		}
	}
}

func TestGeneratePostmanCollection(t *testing.T) {
	tempDir := t.TempDir()
	docsOutput = tempDir
	
	manifest := map[string]interface{}{
		"name":        "test-api",
		"description": "Test API collection",
	}
	
	endpoints := []map[string]interface{}{
		{
			"path":        "/users",
			"method":      "GET",
			"description": "List users",
		},
		{
			"path":        "/users/{id}",
			"method":      "PUT",
			"description": "Update user",
		},
	}
	
	err := generatePostmanCollection("test-api", manifest, endpoints)
	if err != nil {
		t.Fatalf("generatePostmanCollection() failed: %v", err)
	}
	
	// Check file was created
	collectionPath := filepath.Join(tempDir, "test-api_postman_collection.json")
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		t.Fatalf("failed to read Postman collection: %v", err)
	}
	
	// Parse and validate
	var collection map[string]interface{}
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("failed to parse Postman collection: %v", err)
	}
	
	// Check structure
	info := collection["info"].(map[string]interface{})
	if info["name"] != "test-api API" {
		t.Error("expected collection name to be 'test-api API'")
	}
	
	items := collection["item"].([]interface{})
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	
	// Check variables
	variables := collection["variable"].([]interface{})
	if len(variables) < 2 {
		t.Error("expected at least 2 variables (baseUrl and apiKey)")
	}
}