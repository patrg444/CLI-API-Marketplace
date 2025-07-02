package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColorMethod(t *testing.T) {
	tests := []struct {
		method   string
		contains string
	}{
		{"GET", "GET"},     // Would be green in actual output
		{"POST", "POST"},   // Would be blue
		{"PUT", "PUT"},     // Would be yellow
		{"DELETE", "DELETE"}, // Would be red
		{"PATCH", "PATCH"},  // Would be magenta
		{"OPTIONS", "OPTIONS"}, // Default, no color
		{"HEAD", "HEAD"},    // Default, no color
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := colorMethod(tt.method)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestInfoCommand(t *testing.T) {
	// Create mock API response
	mockAPIInfo := map[string]interface{}{
		"id":           "test-api",
		"name":         "test-api",
		"display_name": "Test API",
		"description":  "A test API for testing",
		"long_description": "This is a longer description of the test API",
		"category":     "Testing",
		"tags":         []string{"test", "demo"},
		"version":      "1.0.0",
		"status":       "active",
		"created_at":   time.Now().Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
		"creator": map[string]interface{}{
			"id":            "user123",
			"name":          "Test User",
			"username":      "testuser",
			"verified":      true,
			"joined_date":   "2023-01-01",
			"total_apis":    5,
			"average_rating": 4.5,
		},
		"metrics": map[string]interface{}{
			"subscriber_count":     1000,
			"monthly_calls_avg":    1000000,
			"average_rating":       4.5,
			"total_reviews":        50,
			"response_time_avg_ms": 125,
			"uptime_percentage":    99.95,
			"last_downtime":        nil,
		},
		"pricing": map[string]interface{}{
			"model": "tiered",
			"plans": []map[string]interface{}{
				{
					"id":          "free",
					"name":        "Free",
					"description": "Get started for free",
					"price":       0,
					"currency":    "USD",
					"interval":    "month",
					"features":    []string{"100 API calls/month", "Basic support"},
					"limits": map[string]interface{}{
						"requests_per_month": 100,
					},
					"popular": false,
				},
				{
					"id":          "pro",
					"name":        "Pro",
					"description": "For growing applications",
					"price":       29.99,
					"currency":    "USD",
					"interval":    "month",
					"features":    []string{"10,000 API calls/month", "Priority support", "Advanced features"},
					"limits": map[string]interface{}{
						"requests_per_month": 10000,
						"requests_per_second": 10,
					},
					"popular": true,
				},
			},
			"custom_pricing": true,
			"contact_sales":  "sales@example.com",
		},
		"technical": map[string]interface{}{
			"base_url":        "https://api.example.com/v1",
			"authentication":  []string{"API Key", "OAuth 2.0"},
			"formats":         []string{"JSON", "XML"},
			"sdks":           []string{"Python", "JavaScript", "Go"},
			"openapi_spec_url": "https://api.example.com/openapi.json",
			"postman_collection_url": "https://api.example.com/postman.json",
			"documentation_url": "https://docs.example.com",
			"support_email":    "support@example.com",
			"sla":             "99.9% uptime guarantee",
		},
	}

	tests := []struct {
		name           string
		args           []string
		format         string
		detailed       bool
		mockServer     func(t *testing.T) *httptest.Server
		setupAuth      func(t *testing.T, tempDir string)
		wantErr        bool
		expectedOutput string
		errContains    string
	}{
		{
			name:   "basic info table format",
			args:   []string{"test-api"},
			format: "table",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/api/v1/marketplace/apis/test-api", r.URL.Path)
					assert.Equal(t, "GET", r.Method)
					
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockAPIInfo)
				}))
			},
			setupAuth: func(t *testing.T, tempDir string) {
				// No auth needed for info command
			},
			wantErr:        false,
			expectedOutput: "Test API",
		},
		{
			name:     "detailed info with endpoints",
			args:     []string{"test-api"},
			format:   "table",
			detailed: true,
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/api/v1/marketplace/apis/test-api", r.URL.Path)
					assert.Equal(t, "detailed=true", r.URL.RawQuery)
					
					// Add endpoints to response
					infoWithEndpoints := copyMap(mockAPIInfo)
					infoWithEndpoints["endpoints"] = []map[string]interface{}{
						{
							"path":         "/users",
							"method":       "GET",
							"description":  "List all users",
							"category":     "Users",
							"auth_required": true,
							"rate_limit":   "100/hour",
						},
						{
							"path":         "/users",
							"method":       "POST",
							"description":  "Create a new user",
							"category":     "Users",
							"auth_required": true,
							"rate_limit":   "10/hour",
						},
					}
					
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(infoWithEndpoints)
				}))
			},
			setupAuth: func(t *testing.T, tempDir string) {},
			wantErr:   false,
			expectedOutput: "Endpoints (2)",
		},
		{
			name:   "JSON format output",
			args:   []string{"test-api"},
			format: "json",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockAPIInfo)
				}))
			},
			setupAuth: func(t *testing.T, tempDir string) {},
			wantErr:   false,
			expectedOutput: `"display_name": "Test API"`,
		},
		{
			name:   "API not found",
			args:   []string{"non-existent-api"},
			format: "table",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "API not found",
					})
				}))
			},
			setupAuth:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "API not found",
		},
		{
			name:   "server error",
			args:   []string{"test-api"},
			format: "table",
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			setupAuth:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "status 500",
		},
		{
			name:        "no arguments",
			args:        []string{},
			format:      "table",
			mockServer:  func(t *testing.T) *httptest.Server { return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})) },
			setupAuth:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			tempDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", oldHome)

			// Setup auth if provided
			if tt.setupAuth != nil {
				tt.setupAuth(t, tempDir)
			}

			// Setup mock server
			server := tt.mockServer(t)
			defer server.Close()

			// Create config with mock endpoint
			configDir := filepath.Join(tempDir, ".apidirect")
			os.MkdirAll(configDir, 0755)
			
			config := map[string]interface{}{
				"api": map[string]interface{}{
					"base_url": server.URL,
				},
			}
			configData, _ := json.Marshal(config)
			os.WriteFile(filepath.Join(configDir, "config.json"), configData, 0644)

			// Set command flags
			infoFormat = tt.format
			infoDetailed = tt.detailed

			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs(append([]string{"info"}, tt.args...))
			err := rootCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			output := buf.String()

			// Verify output
			if tt.expectedOutput != "" {
				assert.Contains(t, output, tt.expectedOutput)
			}
			if tt.errContains != "" {
				assert.Contains(t, output, tt.errContains)
			}

			// Reset flags
			infoFormat = "table"
			infoDetailed = false
		})
	}
}

func TestInfoOutputFormatting(t *testing.T) {
	t.Run("format number", func(t *testing.T) {
		tests := []struct {
			input    int64
			expected string
		}{
			{0, "0"},
			{100, "100"},
			{1000, "1,000"},
			{10000, "10,000"},
			{100000, "100,000"},
			{1000000, "1,000,000"},
			{1234567890, "1,234,567,890"},
		}

		for _, tt := range tests {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("get currency symbol", func(t *testing.T) {
		tests := []struct {
			currency string
			expected string
		}{
			{"USD", "$"},
			{"EUR", "€"},
			{"GBP", "£"},
			{"JPY", "¥"},
			{"INR", "₹"},
			{"KRW", "₩"},
			{"CNY", "¥"},
			{"AUD", "$"},
			{"CAD", "$"},
			{"CHF", "CHF "},
			{"SEK", "SEK "},
			{"NOK", "NOK "},
			{"DKK", "DKK "},
			{"PLN", "PLN "},
			{"CZK", "CZK "},
			{"HUF", "HUF "},
			{"RON", "RON "},
			{"BGN", "BGN "},
			{"HRK", "HRK "},
			{"RUB", "₽"},
			{"TRY", "₺"},
			{"BRL", "R$"},
			{"ZAR", "R"},
			{"MXN", "$"},
			{"IDR", "Rp"},
			{"MYR", "RM"},
			{"PHP", "₱"},
			{"SGD", "$"},
			{"THB", "฿"},
			{"VND", "₫"},
			{"XYZ", "XYZ "}, // Unknown currency
		}

		for _, tt := range tests {
			result := getCurrencySymbol(tt.currency)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("get star rating", func(t *testing.T) {
		tests := []struct {
			rating   float64
			expected string
		}{
			{0, "☆☆☆☆☆"},
			{0.5, "☆☆☆☆☆"},
			{1, "★☆☆☆☆"},
			{1.5, "★☆☆☆☆"},
			{2, "★★☆☆☆"},
			{2.5, "★★☆☆☆"},
			{3, "★★★☆☆"},
			{3.5, "★★★☆☆"},
			{4, "★★★★☆"},
			{4.5, "★★★★☆"},
			{5, "★★★★★"},
		}

		for _, tt := range tests {
			result := getStarRatingSimple(tt.rating)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("truncate string", func(t *testing.T) {
		tests := []struct {
			input    string
			maxLen   int
			expected string
		}{
			{"short", 10, "short"},
			{"exactly ten", 11, "exactly ten"},
			{"this is a very long string that needs truncation", 20, "this is a very lo..."},
			{"", 10, ""},
			{"test", 0, "..."},
			{"test", 3, "..."},
			{"test", 4, "test"},
		}

		for _, tt := range tests {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		}
	})
}

func TestInfoCommandWithReviews(t *testing.T) {
	mockAPIInfo := map[string]interface{}{
		"id":           "test-api",
		"name":         "test-api",
		"display_name": "Test API",
		"description":  "A test API",
		"category":     "Testing",
		"tags":         []string{"test"},
		"version":      "1.0.0",
		"status":       "active",
		"created_at":   time.Now().Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
		"creator": map[string]interface{}{
			"id":       "user123",
			"name":     "Test User",
			"username": "testuser",
		},
		"metrics": map[string]interface{}{
			"subscriber_count":     100,
			"response_time_avg_ms": 100,
			"uptime_percentage":    99.9,
		},
		"pricing": map[string]interface{}{
			"model": "free",
			"plans": []map[string]interface{}{},
		},
		"technical": map[string]interface{}{
			"base_url":       "https://api.example.com",
			"authentication": []string{"API Key"},
			"formats":        []string{"JSON"},
		},
		"recent_reviews": []map[string]interface{}{
			{
				"rating":     5,
				"title":      "Excellent API!",
				"message":    "This API is fantastic. Easy to use, well documented, and reliable.",
				"author_name": "Happy Developer",
				"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"verified_purchase": true,
			},
			{
				"rating":     4,
				"title":      "Good but could be better",
				"message":    "The API works well but the rate limits are a bit restrictive for my use case.",
				"author_name": "Another Dev",
				"created_at": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
				"verified_purchase": false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockAPIInfo)
	}))
	defer server.Close()

	// Set up test environment
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config
	configDir := filepath.Join(tempDir, ".apidirect")
	os.MkdirAll(configDir, 0755)
	
	config := map[string]interface{}{
		"api": map[string]interface{}{
			"base_url": server.URL,
		},
	}
	configData, _ := json.Marshal(config)
	os.WriteFile(filepath.Join(configDir, "config.json"), configData, 0644)

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"info", "test-api"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	// Verify reviews are displayed
	assert.Contains(t, output, "Recent Reviews")
	assert.Contains(t, output, "Excellent API!")
	assert.Contains(t, output, "Happy Developer")
	assert.Contains(t, output, "★★★★★") // 5 stars
	assert.Contains(t, output, "✓") // Verified purchase
}

func TestInfoCommandWithChangelog(t *testing.T) {
	mockAPIInfo := map[string]interface{}{
		"id":           "test-api",
		"name":         "test-api",
		"display_name": "Test API",
		"description":  "A test API",
		"category":     "Testing",
		"tags":         []string{"test"},
		"version":      "2.0.0",
		"status":       "active",
		"created_at":   time.Now().Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
		"creator": map[string]interface{}{
			"id":       "user123",
			"name":     "Test User",
			"username": "testuser",
		},
		"metrics": map[string]interface{}{
			"subscriber_count":     100,
			"response_time_avg_ms": 100,
			"uptime_percentage":    99.9,
		},
		"pricing": map[string]interface{}{
			"model": "free",
			"plans": []map[string]interface{}{},
		},
		"technical": map[string]interface{}{
			"base_url":       "https://api.example.com",
			"authentication": []string{"API Key"},
			"formats":        []string{"JSON"},
		},
		"changelog": []map[string]interface{}{
			{
				"version":        "2.0.0",
				"date":          time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
				"description":   "Major update with new endpoints and improved performance",
				"breaking_change": true,
			},
			{
				"version":        "1.5.0",
				"date":          time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
				"description":   "Added support for batch operations",
				"breaking_change": false,
			},
			{
				"version":        "1.4.0",
				"date":          time.Now().Add(-60 * 24 * time.Hour).Format(time.RFC3339),
				"description":   "Bug fixes and performance improvements",
				"breaking_change": false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockAPIInfo)
	}))
	defer server.Close()

	// Set up test environment
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config
	configDir := filepath.Join(tempDir, ".apidirect")
	os.MkdirAll(configDir, 0755)
	
	config := map[string]interface{}{
		"api": map[string]interface{}{
			"base_url": server.URL,
		},
	}
	configData, _ := json.Marshal(config)
	os.WriteFile(filepath.Join(configDir, "config.json"), configData, 0644)

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"info", "test-api"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	// Verify changelog is displayed
	assert.Contains(t, output, "Recent Changes")
	assert.Contains(t, output, "v2.0.0")
	assert.Contains(t, output, "BREAKING")
	assert.Contains(t, output, "Major update")
}

// Helper function to copy map
func copyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}


