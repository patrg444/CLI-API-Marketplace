package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionsListCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "list all subscriptions",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"id":        "sub_123",
							"api_name":  "weather-api",
							"api_id":    "api_123",
							"plan_name": "Pro",
							"status":    "active",
							"created_at": time.Now().Add(-30 * 24 * time.Hour),
							"current_period": map[string]interface{}{
								"start": time.Now().Add(-10 * 24 * time.Hour),
								"end":   time.Now().Add(20 * 24 * time.Hour),
							},
							"usage": map[string]interface{}{
								"calls":     8000,
								"limit":     10000,
								"remaining": 2000,
							},
							"billing": map[string]interface{}{
								"amount":             49.99,
								"interval":           "monthly",
								"next_billing_date": "2024-02-01",
							},
						},
						{
							"id":        "sub_456",
							"api_name":  "payment-api",
							"api_id":    "api_456",
							"plan_name": "Basic",
							"status":    "cancelled",
							"created_at": time.Now().Add(-60 * 24 * time.Hour),
						},
					},
				},
			},
			expectedOutput: []string{
				"Active Subscriptions (1)",
				"weather-api",
				"Pro",
				"8000/10000",
				"$49.99/monthly",
				"Inactive Subscriptions (1)",
				"payment-api",
				"cancelled",
			},
			expectError: false,
		},
		{
			name: "filter by status",
			args: []string{},
			flags: map[string]string{
				"status": "active",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions?status=active": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"id":        "sub_123",
							"api_name":  "weather-api",
							"status":    "active",
						},
					},
				},
			},
			expectedOutput: []string{
				"Active Subscriptions (1)",
				"weather-api",
			},
			expectError: false,
		},
		{
			name: "no subscriptions",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions": {
					statusCode: 200,
					body:       []map[string]interface{}{},
				},
			},
			expectedOutput: []string{
				"No subscriptions found",
			},
			expectError: false,
		},
		{
			name: "JSON format output",
			args: []string{},
			flags: map[string]string{
				"format": "json",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"id":       "sub_123",
							"api_name": "weather-api",
						},
					},
				},
			},
			expectedOutput: []string{
				`"id"`,
				`"sub_123"`,
				`"api_name"`,
				`"weather-api"`,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldClient := httpClient
			httpClient = &mockHTTPClient{responses: tt.mockResponses}
			defer func() { httpClient = oldClient }()

			// Capture output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Setup test environment
			cleanup := setupTestAuth(t)
			defer cleanup()

			// Create config with mock server URL
			tempDir := os.Getenv("HOME") // setupTestAuth sets this
			configDir := filepath.Join(tempDir, ".apidirect")
			os.MkdirAll(configDir, 0755)
			
			config := map[string]interface{}{
				"api": map[string]interface{}{
					"base_url": "http://test-server",
				},
			}
			configData, _ := json.Marshal(config)
			os.WriteFile(filepath.Join(configDir, "config.json"), configData, 0644)

			// Reset and set flags
			subscriptionStatus = ""
			subscriptionFormat = "table"

			if status, ok := tt.flags["status"]; ok {
				subscriptionStatus = status
			}
			if format, ok := tt.flags["format"]; ok {
				subscriptionFormat = format
			}

			// Execute command
			err := runSubscriptionsList(cmd, tt.args)

			// Check error
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestSubscriptionsShowCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "show subscription details",
			args: []string{"sub_123"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123": {
					statusCode: 200,
					body: map[string]interface{}{
						"id":           "sub_123",
						"api_name":     "weather-api",
						"api_id":       "api_123",
						"api_endpoint": "https://api.weather.com/v1",
						"plan_name":    "Pro",
						"plan_type":    "subscription",
						"status":       "active",
						"created_at":   time.Now().Add(-30 * 24 * time.Hour),
						"current_period": map[string]interface{}{
							"start": time.Now().Add(-10 * 24 * time.Hour),
							"end":   time.Now().Add(20 * 24 * time.Hour),
						},
						"usage": map[string]interface{}{
							"calls":              8000,
							"limit":              10000,
							"remaining":          2000,
							"data_transfer_gb":   1.5,
							"average_latency_ms": 125,
							"error_rate":         0.02,
						},
						"billing": map[string]interface{}{
							"amount":             49.99,
							"currency":           "USD",
							"interval":           "monthly",
							"next_billing_date":  "2024-02-01",
							"payment_method":     "Visa ending in 4242",
						},
						"features": []string{
							"10,000 API calls/month",
							"Priority support",
							"Advanced analytics",
						},
						"api_key": map[string]interface{}{
							"key":        "sk_test_abcdef123456",
							"created_at": time.Now().Add(-30 * 24 * time.Hour),
							"last_used":  time.Now().Add(-1 * time.Hour),
						},
					},
				},
			},
			expectedOutput: []string{
				"Subscription Details",
				"sub_123",
				"weather-api",
				"Pro (subscription)",
				"Status: active",
				"https://api.weather.com/v1",
				"sk_test_ab...3456",
				"8000 / 10000",
				"$49.99 / monthly",
				"10,000 API calls/month",
				"Priority support",
			},
			expectError: false,
		},
		{
			name: "show detailed subscription with history",
			args: []string{"sub_123"},
			flags: map[string]string{
				"detailed": "true",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123?detailed=true": {
					statusCode: 200,
					body: map[string]interface{}{
						"id":         "sub_123",
						"api_name":   "weather-api",
						"status":     "active",
						"usage_history": []map[string]interface{}{
							{"date": "2024-01-25", "calls": 350},
							{"date": "2024-01-24", "calls": 420},
							{"date": "2024-01-23", "calls": 380},
						},
						"invoices": []map[string]interface{}{
							{
								"id":     "inv_123",
								"date":   "2024-01-01",
								"amount": 49.99,
								"status": "paid",
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Usage History",
				"2024-01-25",
				"350",
				"Recent Invoices",
				"2024-01-01",
				"$49.99",
				"paid",
			},
			expectError: false,
		},
		{
			name: "subscription not found",
			args: []string{"sub_invalid"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_invalid": {
					statusCode: 404,
					body:       map[string]string{"error": "Subscription not found"},
				},
			},
			expectedOutput: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldClient := httpClient
			httpClient = &mockHTTPClient{responses: tt.mockResponses}
			defer func() { httpClient = oldClient }()

			// Capture output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Reset and set flags
			subscriptionDetailed = false
			subscriptionFormat = "table"

			if detailed, ok := tt.flags["detailed"]; ok {
				subscriptionDetailed = detailed == "true"
			}

			// Setup test environment
			cleanup := setupTestAuth(t)
			defer cleanup()

			// Create config with mock server URL
			tempDir := os.Getenv("HOME") // setupTestAuth sets this
			configDir := filepath.Join(tempDir, ".apidirect")
			os.MkdirAll(configDir, 0755)
			
			config := map[string]interface{}{
				"api": map[string]interface{}{
					"base_url": "http://test-server",
				},
			}
			configData, _ := json.Marshal(config)
			os.WriteFile(filepath.Join(configDir, "config.json"), configData, 0644)

			// Execute command
			err := runSubscriptionsShow(cmd, tt.args)

			// Check error
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestSubscriptionsCancelCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		userInput      string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful cancellation",
			args: []string{"sub_123"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123": {
					statusCode: 200,
					body: map[string]interface{}{
						"api_name":  "weather-api",
						"plan_name": "Pro",
						"billing": map[string]interface{}{
							"next_billing_date": "2024-02-01",
						},
					},
				},
				"POST /api/v1/subscriptions/sub_123/cancel": {
					statusCode: 200,
					body: map[string]interface{}{
						"status":        "cancelled",
						"cancelled_at":  "2024-01-25",
						"active_until":  "2024-02-01",
						"refund_amount": 0,
					},
				},
			},
			userInput: "y\n", // Confirm cancellation
			expectedOutput: []string{
				"Cancel Subscription",
				"weather-api",
				"Pro",
				"remain active until: 2024-02-01",
				"Subscription cancelled successfully",
				"Active until: 2024-02-01",
			},
			expectError: false,
		},
		{
			name: "cancellation with refund",
			args: []string{"sub_456"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_456": {
					statusCode: 200,
					body: map[string]interface{}{
						"api_name":  "payment-api",
						"plan_name": "Enterprise",
					},
				},
				"POST /api/v1/subscriptions/sub_456/cancel": {
					statusCode: 200,
					body: map[string]interface{}{
						"status":        "cancelled",
						"refund_amount": 125.50,
					},
				},
			},
			userInput: "y\n",
			expectedOutput: []string{
				"Refund amount: $125.50",
			},
			expectError: false,
		},
		{
			name: "user cancels cancellation",
			args: []string{"sub_123"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123": {
					statusCode: 200,
					body:       map[string]interface{}{},
				},
			},
			userInput: "n\n", // Do not confirm
			expectedOutput: []string{
				"Cancellation aborted",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldClient := httpClient
			httpClient = &mockHTTPClient{responses: tt.mockResponses}
			defer func() { httpClient = oldClient }()

			// Mock user input
			oldStdin := stdin
			stdin = strings.NewReader(tt.userInput)
			defer func() { stdin = oldStdin }()

			// Capture output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := runSubscriptionsCancel(cmd, tt.args)

			// Check error
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestSubscriptionsUsageCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "basic usage report",
			args: []string{"sub_123"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123/usage": {
					statusCode: 200,
					body: map[string]interface{}{
						"subscription_id": "sub_123",
						"api_name":        "weather-api",
						"period": map[string]interface{}{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"summary": map[string]interface{}{
							"total_calls":       8000,
							"successful_calls":  7800,
							"failed_calls":      200,
							"call_limit":        10000,
							"calls_remaining":   2000,
							"data_transfer_gb":  1.5,
							"average_latency_ms": 125,
							"uptime_percentage": 99.95,
						},
						"by_endpoint": []map[string]interface{}{
							{
								"endpoint":      "/weather/{city}",
								"method":        "GET",
								"calls":         6000,
								"errors":        100,
								"avg_latency_ms": 120,
							},
							{
								"endpoint":      "/forecast/{city}",
								"method":        "GET",
								"calls":         2000,
								"errors":        100,
								"avg_latency_ms": 140,
							},
						},
						"rate_limits": map[string]interface{}{
							"requests_per_second": 10,
							"requests_per_minute": 300,
							"requests_per_hour":   10000,
							"current_usage": map[string]interface{}{
								"second": 2,
								"minute": 50,
								"hour":   800,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Usage Report: weather-api",
				"Total Calls: 8000 / 10000",
				"Success Rate: 97.5%",
				"Average Latency: 125 ms",
				"Rate Limits",
				"10/sec, 300/min, 10000/hour",
				"/weather/{city}",
				"6000",
			},
			expectError: false,
		},
		{
			name: "detailed usage with daily breakdown",
			args: []string{"sub_123"},
			flags: map[string]string{
				"detailed": "true",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123/usage?detailed=true": {
					statusCode: 200,
					body: map[string]interface{}{
						"subscription_id": "sub_123",
						"api_name":        "weather-api",
						"by_day": []map[string]interface{}{
							{"date": "2024-01-25", "calls": 350, "errors": 5},
							{"date": "2024-01-24", "calls": 420, "errors": 8},
						},
						"error_breakdown": []map[string]interface{}{
							{"status_code": 400, "count": 150, "percentage": 75},
							{"status_code": 500, "count": 50, "percentage": 25},
						},
					},
				},
			},
			expectedOutput: []string{
				"Daily Usage",
				"2024-01-25",
				"350",
				"Error Breakdown",
				"400",
				"75.0%",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldClient := httpClient
			httpClient = &mockHTTPClient{responses: tt.mockResponses}
			defer func() { httpClient = oldClient }()

			// Capture output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Reset and set flags
			subscriptionDetailed = false
			subscriptionFormat = "table"

			if detailed, ok := tt.flags["detailed"]; ok {
				subscriptionDetailed = detailed == "true"
			}

			// Execute command
			err := runSubscriptionsUsage(cmd, tt.args)

			// Check error
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestSubscriptionsKeysCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		userInput      string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "view API keys",
			args: []string{"sub_123"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/subscriptions/sub_123/keys": {
					statusCode: 200,
					body: map[string]interface{}{
						"api_name":     "weather-api",
						"api_endpoint": "https://api.weather.com/v1",
						"keys": []map[string]interface{}{
							{
								"key":         "sk_test_abcdef123456789",
								"name":        "Production Key",
								"created_at":  time.Now().Add(-30 * 24 * time.Hour),
								"last_used":   time.Now().Add(-1 * time.Hour),
								"calls_today": 350,
								"status":      "active",
							},
						},
						"documentation_url": "https://api.weather.com/docs",
						"examples": map[string]interface{}{
							"curl":   "curl -H 'X-API-Key: YOUR_KEY' https://api.weather.com/v1/weather/london",
							"python": "requests.get('https://api.weather.com/v1/weather/london', headers={'X-API-Key': 'YOUR_KEY'})",
							"nodejs": "axios.get('https://api.weather.com/v1/weather/london', {headers: {'X-API-Key': 'YOUR_KEY'}})",
						},
					},
				},
			},
			expectedOutput: []string{
				"API Keys: weather-api",
				"https://api.weather.com/v1",
				"sk_test_abcd...6789",
				"Production Key",
				"active",
				"350 calls today",
				"Quick Start Examples",
				"curl",
			},
			expectError: false,
		},
		{
			name: "regenerate API key",
			args: []string{"sub_123"},
			flags: map[string]string{
				"regenerate": "true",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/subscriptions/sub_123/keys/regenerate": {
					statusCode: 200,
					body: map[string]interface{}{
						"key":        "sk_test_newkey987654321",
						"created_at": time.Now(),
					},
				},
			},
			userInput: "y\n", // Confirm regeneration
			expectedOutput: []string{
				"Regenerate API Key",
				"invalidate your current API key",
				"API key regenerated successfully",
				"sk_test_newkey987654321",
				"Save this key securely",
			},
			expectError: false,
		},
		{
			name: "user cancels regeneration",
			args: []string{"sub_123"},
			flags: map[string]string{
				"regenerate": "true",
			},
			userInput: "n\n",
			expectedOutput: []string{
				"Regeneration cancelled",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldClient := httpClient
			httpClient = &mockHTTPClient{responses: tt.mockResponses}
			defer func() { httpClient = oldClient }()

			// Mock user input
			if tt.userInput != "" {
				oldStdin := stdin
				stdin = strings.NewReader(tt.userInput)
				defer func() { stdin = oldStdin }()
			}

			// Capture output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set flags
			cmd.Flags().Bool("regenerate", false, "")
			if regenerate, ok := tt.flags["regenerate"]; ok {
				cmd.Flags().Set("regenerate", regenerate)
			}

			// Execute command
			err := runSubscriptionsKeys(cmd, tt.args)

			// Check error
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// Test helper functions
func TestGetCurrencySymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"USD", "$"},
		{"EUR", "€"},
		{"GBP", "£"},
		{"JPY", "¥"},
		{"usd", "$"}, // Test case insensitive
		{"XYZ", "XYZ "}, // Unknown currency
		{"", " "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getCurrencySymbol(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}