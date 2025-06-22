package auth

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/api-direct/cli/pkg/config"
)

// GetToken returns the current authentication token
func GetToken() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}
	
	if cfg.Auth.AccessToken == "" {
		return "", fmt.Errorf("not authenticated")
	}
	
	// Check if token is expired
	if !cfg.Auth.ExpiresAt.IsZero() && time.Now().After(cfg.Auth.ExpiresAt) {
		return "", fmt.Errorf("token expired")
	}
	
	return cfg.Auth.AccessToken, nil
}

// MakeAuthenticatedRequest makes an HTTP request with authentication
func MakeAuthenticatedRequest(method, url, token string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	return client.Do(req)
}