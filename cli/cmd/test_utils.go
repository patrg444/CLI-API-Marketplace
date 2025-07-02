package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	
	"github.com/api-direct/cli/pkg/config"
)

// Test variables for mocking
var (
	httpClient HTTPClient = http.DefaultClient
	stdin      io.Reader  = os.Stdin
)

// HTTPClient interface for mocking HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Helper function to make authenticated requests (placeholder for tests)
func makeAuthenticatedRequest(method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	
	// Add authentication header if available
	cfg, err := config.Load()
	if err == nil && cfg.Auth.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	}
	
	// Add content type for POST requests
	if method == "POST" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// In tests, we'll use the mocked httpClient if available
	if httpClient != nil {
		return httpClient.Do(req)
	}
	
	// Otherwise use default client
	return http.DefaultClient.Do(req)
}

// Helper function to handle error responses
func handleErrorResponse(resp *http.Response) error {
	// Try to parse JSON error response
	var errResp struct {
		Error string `json:"error"`
		Message string `json:"message"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		if errResp.Message != "" {
			return fmt.Errorf("%s", errResp.Message)
		}
	}
	
	// Fallback to status code
	return fmt.Errorf("API error: status %d", resp.StatusCode)
}

// Helper function to confirm actions
func confirmAction(prompt string) bool {
	// In tests, this will read from mocked stdin
	var response string
	fmt.Fscanln(stdin, &response)
	return strings.ToLower(response) == "y"
}