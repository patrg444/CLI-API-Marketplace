package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	
	// In tests, we'll use the mocked httpClient
	return httpClient.Do(req)
}

// Helper function to handle error responses
func handleErrorResponse(resp *http.Response) error {
	// Simplified version for tests
	return fmt.Errorf("API error: status %d", resp.StatusCode)
}

// Helper function to confirm actions
func confirmAction(prompt string) bool {
	// In tests, this will read from mocked stdin
	var response string
	fmt.Fscanln(stdin, &response)
	return strings.ToLower(response) == "y"
}