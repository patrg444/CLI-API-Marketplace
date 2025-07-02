package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// mockHTTPClient is a mock HTTP client for testing
type mockHTTPClient struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	statusCode int
	body       interface{}
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	
	// Debug: print the requested key
	// fmt.Printf("Mock: Requested key: %s\n", key)
	
	// Try with query params first
	if resp, ok := m.responses[key]; ok {
		if resp.err != nil {
			return nil, resp.err
		}
		
		body, _ := json.Marshal(resp.body)
		return &http.Response{
			StatusCode: resp.statusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}
	
	// Try without query params for backward compatibility
	key = req.Method + " " + req.URL.Path
	if resp, ok := m.responses[key]; ok {
		if resp.err != nil {
			return nil, resp.err
		}
		
		body, _ := json.Marshal(resp.body)
		return &http.Response{
			StatusCode: resp.statusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}
	
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("Not found")),
	}, nil
}

// setupTestAuth sets up authentication for tests
func setupTestAuth(t *testing.T) func() {
	// Set up authentication token
	oldToken := os.Getenv("APIDIRECT_AUTH_TOKEN")
	os.Setenv("APIDIRECT_AUTH_TOKEN", "test-token")
	
	// Create temp HOME to ensure config can be created
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	
	// Return cleanup function
	return func() {
		os.Setenv("APIDIRECT_AUTH_TOKEN", oldToken)
		os.Setenv("HOME", oldHome)
	}
}