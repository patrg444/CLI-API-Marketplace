package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) func() {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	
	// Create config directory
	configDir := filepath.Join(testDir, ".apidirect")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	// Return cleanup function
	return func() {
		os.Setenv("HOME", originalHome)
	}
}

func TestGetToken(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T)
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid token",
			setupConfig: func(t *testing.T) {
				cfg := &config.Config{
					Auth: config.AuthConfig{
						AccessToken: "valid-token-123",
						ExpiresAt:   time.Now().Add(time.Hour),
					},
				}
				err := config.SaveConfig(cfg)
				require.NoError(t, err)
			},
			wantToken: "valid-token-123",
			wantErr:   false,
		},
		{
			name: "no token - not authenticated",
			setupConfig: func(t *testing.T) {
				cfg := config.DefaultConfig()
				err := config.SaveConfig(cfg)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "not authenticated",
		},
		{
			name: "expired token",
			setupConfig: func(t *testing.T) {
				cfg := &config.Config{
					Auth: config.AuthConfig{
						AccessToken: "expired-token",
						ExpiresAt:   time.Now().Add(-time.Hour), // Expired
					},
				}
				err := config.SaveConfig(cfg)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "token expired",
		},
		{
			name: "token with no expiry",
			setupConfig: func(t *testing.T) {
				cfg := &config.Config{
					Auth: config.AuthConfig{
						AccessToken: "no-expiry-token",
						// ExpiresAt is zero value
					},
				}
				err := config.SaveConfig(cfg)
				require.NoError(t, err)
			},
			wantToken: "no-expiry-token",
			wantErr:   false,
		},
		{
			name: "config error",
			setupConfig: func(t *testing.T) {
				// Create invalid config file
				home := os.Getenv("HOME")
				configPath := filepath.Join(home, ".apidirect", "config.json")
				err := ioutil.WriteFile(configPath, []byte("invalid json"), 0600)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "failed to parse config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestConfig(t)
			defer cleanup()
			
			// Setup config for this test
			tt.setupConfig(t)
			
			// Get token
			token, err := GetToken()
			
			// Check results
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestMakeAuthenticatedRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           []byte
		token          string
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		validateReq    func(t *testing.T, r *http.Request)
		validateResp   func(t *testing.T, resp *http.Response)
		wantErr        bool
	}{
		{
			name:   "GET request with token",
			method: "GET",
			token:  "test-token",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			},
			validateReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "GET", r.Method)
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name:   "POST request with body",
			method: "POST",
			body:   []byte(`{"name": "test"}`),
			token:  "post-token",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				body, _ := ioutil.ReadAll(r.Body)
				assert.Equal(t, `{"name": "test"}`, string(body))
				w.WriteHeader(http.StatusCreated)
			},
			validateReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "Bearer post-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "POST", r.Method)
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			},
		},
		{
			name:   "PUT request with empty token",
			method: "PUT",
			token:  "",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			validateReq: func(t *testing.T, r *http.Request) {
				// HTTP might trim trailing spaces in headers
				auth := r.Header.Get("Authorization")
				assert.True(t, auth == "Bearer" || auth == "Bearer ", "Authorization header should be 'Bearer' or 'Bearer ', got: %q", auth)
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name:   "DELETE request",
			method: "DELETE",
			token:  "delete-token",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			validateReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Empty(t, r.Header.Get("Content-Type")) // No body, no content type
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			var capturedReq *http.Request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				tt.mockHandler(w, r)
			}))
			defer server.Close()
			
			// Make request
			resp, err := MakeAuthenticatedRequest(tt.method, server.URL, tt.token, tt.body)
			
			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()
			
			// Validate request
			if tt.validateReq != nil && capturedReq != nil {
				tt.validateReq(t, capturedReq)
			}
			
			// Validate response
			if tt.validateResp != nil {
				tt.validateResp(t, resp)
			}
		})
	}
}

func TestMakeAuthenticatedRequestErrors(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		url     string
		wantErr bool
	}{
		{
			name:    "invalid URL",
			method:  "GET",
			url:     "://invalid-url",
			wantErr: true,
		},
		{
			name:    "invalid method",
			method:  "INVALID METHOD",
			url:     "http://example.com",
			wantErr: true,
		},
		{
			name:    "unreachable server",
			method:  "GET",
			url:     "http://localhost:99999", // Invalid port
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MakeAuthenticatedRequest(tt.method, tt.url, "token", nil)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestMakeAuthenticatedRequestTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Delay to test timeout handling
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Make request (should complete within default 30s timeout)
	resp, err := MakeAuthenticatedRequest("GET", server.URL, "token", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestMakeAuthenticatedRequestHeaders(t *testing.T) {
	// Test that headers are properly set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back headers for validation
		for key, values := range r.Header {
			for _, value := range values {
				w.Header().Add("Echo-"+key, value)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Test with body (should set Content-Type)
	resp, err := MakeAuthenticatedRequest("POST", server.URL, "test-token", []byte("test body"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	
	// Check echoed headers
	assert.Equal(t, "Bearer test-token", resp.Header.Get("Echo-Authorization"))
	assert.Equal(t, "application/json", resp.Header.Get("Echo-Content-Type"))
	resp.Body.Close()
	
	// Test without body (should not set Content-Type)
	resp2, err := MakeAuthenticatedRequest("GET", server.URL, "test-token", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp2)
	
	assert.Equal(t, "Bearer test-token", resp2.Header.Get("Echo-Authorization"))
	assert.Empty(t, resp2.Header.Get("Echo-Content-Type"))
	resp2.Body.Close()
}