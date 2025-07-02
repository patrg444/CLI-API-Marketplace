package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCognitoAuth(t *testing.T) {
	tests := []struct {
		name       string
		region     string
		userPoolID string
		clientID   string
		authDomain string
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			region:     "us-east-1",
			userPoolID: "us-east-1_abcdef123",
			clientID:   "1234567890abcdef",
			authDomain: "https://auth.api-direct.io",
			wantErr:    false,
		},
		{
			name:       "empty region",
			region:     "",
			userPoolID: "us-east-1_abcdef123",
			clientID:   "1234567890abcdef",
			authDomain: "https://auth.api-direct.io",
			wantErr:    false, // AWS SDK will use default region
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewCognitoAuth(tt.region, tt.userPoolID, tt.clientID, tt.authDomain)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, auth)
				assert.Equal(t, tt.userPoolID, auth.userPoolID)
				assert.Equal(t, tt.clientID, auth.clientID)
				assert.Equal(t, tt.authDomain, auth.authDomain)
				assert.Equal(t, 8080, auth.callbackPort)
			}
		})
	}
}

func TestBuildAuthURL(t *testing.T) {
	auth := &CognitoAuth{
		clientID:     "test-client-id",
		authDomain:   "https://auth.example.com",
		callbackPort: 8080,
	}

	codeChallenge := "test-challenge-123"
	authURL := auth.buildAuthURL(codeChallenge)

	// Parse the URL
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	// Verify base URL
	assert.Equal(t, "https", parsedURL.Scheme)
	assert.Equal(t, "auth.example.com", parsedURL.Host)
	assert.Equal(t, "/oauth2/authorize", parsedURL.Path)

	// Verify query parameters
	params := parsedURL.Query()
	assert.Equal(t, "code", params.Get("response_type"))
	assert.Equal(t, "test-client-id", params.Get("client_id"))
	assert.Equal(t, "http://localhost:8080/callback", params.Get("redirect_uri"))
	assert.Equal(t, "openid email profile", params.Get("scope"))
	assert.Equal(t, codeChallenge, params.Get("code_challenge"))
	assert.Equal(t, "S256", params.Get("code_challenge_method"))
}

func TestExchangeCodeForTokens(t *testing.T) {
	auth := &CognitoAuth{
		clientID:     "test-client-id",
		authDomain:   "https://auth.example.com",
		callbackPort: 8080,
	}

	tests := []struct {
		name        string
		code        string
		verifier    string
		mockServer  func(*testing.T) *httptest.Server
		wantErr     bool
		errContains string
		validate    func(*testing.T, *AuthResult)
	}{
		{
			name:     "successful token exchange",
			code:     "valid-auth-code",
			verifier: "test-verifier",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request
					assert.Equal(t, "/oauth2/token", r.URL.Path)
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

					// Parse form data
					err := r.ParseForm()
					require.NoError(t, err)
					assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
					assert.Equal(t, "test-client-id", r.Form.Get("client_id"))
					assert.Equal(t, "valid-auth-code", r.Form.Get("code"))
					assert.Equal(t, "test-verifier", r.Form.Get("code_verifier"))

					// Return successful response
					response := map[string]interface{}{
						"access_token":  "test-access-token",
						"id_token":      "test-id-token",
						"refresh_token": "test-refresh-token",
						"expires_in":    3600,
						"token_type":    "Bearer",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}))
			},
			wantErr: false,
			validate: func(t *testing.T, result *AuthResult) {
				assert.Equal(t, "test-access-token", result.AccessToken)
				assert.Equal(t, "test-id-token", result.IDToken)
				assert.Equal(t, "test-refresh-token", result.RefreshToken)
				assert.Equal(t, 3600, result.ExpiresIn)
			},
		},
		{
			name:     "invalid authorization code",
			code:     "invalid-code",
			verifier: "test-verifier",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid authorization code"}`))
				}))
			},
			wantErr:     true,
			errContains: "token exchange failed with status: 400",
		},
		{
			name:     "server error",
			code:     "valid-code",
			verifier: "test-verifier",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantErr:     true,
			errContains: "token exchange failed with status: 500",
		},
		{
			name:     "invalid JSON response",
			code:     "valid-code",
			verifier: "test-verifier",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{invalid json`))
				}))
			},
			wantErr:     true,
			errContains: "failed to decode token response",
		},
		{
			name:     "missing tokens in response",
			code:     "valid-code",
			verifier: "test-verifier",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Return partial response
					response := map[string]interface{}{
						"token_type": "Bearer",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}))
			},
			wantErr: false,
			validate: func(t *testing.T, result *AuthResult) {
				// Should handle missing fields gracefully
				assert.Empty(t, result.AccessToken)
				assert.Empty(t, result.IDToken)
				assert.Empty(t, result.RefreshToken)
				assert.Equal(t, 0, result.ExpiresIn)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.mockServer(t)
			defer server.Close()

			// Override auth domain with test server URL
			auth.authDomain = server.URL

			result, err := auth.exchangeCodeForTokens(tt.code, tt.verifier)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestGenerateCodeVerifier(t *testing.T) {
	// Test multiple generations to ensure randomness
	verifiers := make(map[string]bool)
	for i := 0; i < 10; i++ {
		verifier := generateCodeVerifier()
		
		// Should be base64 URL encoded
		_, err := base64.RawURLEncoding.DecodeString(verifier)
		assert.NoError(t, err)
		
		// Should be unique
		assert.False(t, verifiers[verifier], "Generated duplicate verifier")
		verifiers[verifier] = true
		
		// Should have appropriate length (32 bytes = ~43 chars in base64)
		assert.GreaterOrEqual(t, len(verifier), 40)
		assert.LessOrEqual(t, len(verifier), 50)
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	tests := []struct {
		name     string
		verifier string
		validate func(*testing.T, string)
	}{
		{
			name:     "valid verifier",
			verifier: "test-verifier-123",
			validate: func(t *testing.T, challenge string) {
				// Should be base64 URL encoded
				decoded, err := base64.RawURLEncoding.DecodeString(challenge)
				assert.NoError(t, err)
				
				// Should be 32 bytes (SHA256 output)
				assert.Len(t, decoded, 32)
				
				// Should be deterministic for same input
				challenge2 := generateCodeChallenge("test-verifier-123")
				assert.Equal(t, challenge, challenge2)
			},
		},
		{
			name:     "empty verifier",
			verifier: "",
			validate: func(t *testing.T, challenge string) {
				// Should still produce valid output
				_, err := base64.RawURLEncoding.DecodeString(challenge)
				assert.NoError(t, err)
			},
		},
		{
			name:     "long verifier",
			verifier: strings.Repeat("a", 1000),
			validate: func(t *testing.T, challenge string) {
				// Should handle long input
				decoded, err := base64.RawURLEncoding.DecodeString(challenge)
				assert.NoError(t, err)
				assert.Len(t, decoded, 32)
			},
		},
		{
			name:     "special characters",
			verifier: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			validate: func(t *testing.T, challenge string) {
				// Should handle special characters
				decoded, err := base64.RawURLEncoding.DecodeString(challenge)
				assert.NoError(t, err)
				assert.Len(t, decoded, 32)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := generateCodeChallenge(tt.verifier)
			assert.NotEmpty(t, challenge)
			if tt.validate != nil {
				tt.validate(t, challenge)
			}
		})
	}
}

func TestAuthResult(t *testing.T) {
	// Test AuthResult structure
	result := &AuthResult{
		AccessToken:  "access-123",
		IDToken:      "id-456",
		RefreshToken: "refresh-789",
		ExpiresIn:    3600,
	}

	assert.Equal(t, "access-123", result.AccessToken)
	assert.Equal(t, "id-456", result.IDToken)
	assert.Equal(t, "refresh-789", result.RefreshToken)
	assert.Equal(t, 3600, result.ExpiresIn)
}

func TestUserInfo(t *testing.T) {
	// Test UserInfo structure
	info := &UserInfo{
		Username: "testuser",
		Email:    "test@example.com",
		Sub:      "12345-67890",
	}

	assert.Equal(t, "testuser", info.Username)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "12345-67890", info.Sub)
}

func TestCognitoAuthStructure(t *testing.T) {
	// Test that CognitoAuth has all required fields
	auth := &CognitoAuth{
		client:       nil,
		userPoolID:   "us-east-1_test",
		clientID:     "client123",
		authDomain:   "https://auth.test.com",
		callbackPort: 8080,
	}

	assert.Equal(t, "us-east-1_test", auth.userPoolID)
	assert.Equal(t, "client123", auth.clientID)
	assert.Equal(t, "https://auth.test.com", auth.authDomain)
	assert.Equal(t, 8080, auth.callbackPort)
}

func TestCallbackServerSetup(t *testing.T) {

	// Test various callback scenarios
	tests := []struct {
		name        string
		queryParams string
		wantError   bool
		wantHTML    string
	}{
		{
			name:        "successful callback with code",
			queryParams: "?code=auth-code-123&state=test-state",
			wantError:   false,
			wantHTML:    "Authentication Successful",
		},
		{
			name:        "callback with error",
			queryParams: "?error=access_denied&error_description=User+denied+access",
			wantError:   true,
			wantHTML:    "Authentication Failed",
		},
		{
			name:        "callback without code",
			queryParams: "?state=test-state",
			wantError:   true,
			wantHTML:    "No authorization code received",
		},
		{
			name:        "empty callback",
			queryParams: "",
			wantError:   true,
			wantHTML:    "No authorization code received",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/callback"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			// Create handler function that would be used in LoginWithBrowser
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				code := r.URL.Query().Get("code")
				if code == "" {
					fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>No authorization code received.</p></body></html>`)
					return
				}
				fmt.Fprintf(w, `<html><body><h1>Authentication Successful!</h1><p>You can close this window.</p></body></html>`)
			})

			// Serve the request
			handler.ServeHTTP(rec, req)

			// Check response
			assert.Contains(t, rec.Body.String(), tt.wantHTML)
		})
	}
}

func TestPKCEFlow(t *testing.T) {
	// Test PKCE (Proof Key for Code Exchange) implementation
	t.Run("verifier and challenge relationship", func(t *testing.T) {
		verifier := generateCodeVerifier()
		challenge := generateCodeChallenge(verifier)

		// Verify that different verifiers produce different challenges
		verifier2 := generateCodeVerifier()
		challenge2 := generateCodeChallenge(verifier2)
		
		assert.NotEqual(t, verifier, verifier2)
		assert.NotEqual(t, challenge, challenge2)

		// Verify that same verifier produces same challenge
		challengeAgain := generateCodeChallenge(verifier)
		assert.Equal(t, challenge, challengeAgain)
	})

	t.Run("PKCE security properties", func(t *testing.T) {
		verifier := generateCodeVerifier()
		challenge := generateCodeChallenge(verifier)

		// Challenge should not reveal the verifier
		assert.NotContains(t, challenge, verifier)
		
		// Challenge should be URL-safe
		assert.NotContains(t, challenge, "+")
		assert.NotContains(t, challenge, "/")
		assert.NotContains(t, challenge, "=")
	})
}

func TestURLConstruction(t *testing.T) {
	auth := &CognitoAuth{
		clientID:     "my-client-123",
		authDomain:   "https://auth.api-direct.io",
		callbackPort: 3000,
	}

	tests := []struct {
		name          string
		challenge     string
		validateURL   func(*testing.T, string)
	}{
		{
			name:      "standard challenge",
			challenge: "abc123_-",
			validateURL: func(t *testing.T, authURL string) {
				assert.Contains(t, authURL, "code_challenge=abc123_-")
				assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback")
			},
		},
		{
			name:      "challenge with special characters",
			challenge: "test+/=challenge",
			validateURL: func(t *testing.T, authURL string) {
				// Should be URL encoded
				assert.Contains(t, authURL, "code_challenge=test%2B%2F%3Dchallenge")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authURL := auth.buildAuthURL(tt.challenge)
			if tt.validateURL != nil {
				tt.validateURL(t, authURL)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateCodeVerifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateCodeVerifier()
	}
}

func BenchmarkGenerateCodeChallenge(b *testing.B) {
	verifier := "test-verifier-for-benchmark"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateCodeChallenge(verifier)
	}
}

func BenchmarkBuildAuthURL(b *testing.B) {
	auth := &CognitoAuth{
		clientID:     "test-client-id",
		authDomain:   "https://auth.example.com",
		callbackPort: 8080,
	}
	challenge := "test-challenge"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.buildAuthURL(challenge)
	}
}

func BenchmarkExchangeCodeForTokens(b *testing.B) {
	// Create a mock server that responds quickly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"token","id_token":"id","refresh_token":"refresh","expires_in":3600}`))
	}))
	defer server.Close()

	auth := &CognitoAuth{
		clientID:     "test-client-id",
		authDomain:   server.URL,
		callbackPort: 8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = auth.exchangeCodeForTokens("test-code", "test-verifier")
	}
}