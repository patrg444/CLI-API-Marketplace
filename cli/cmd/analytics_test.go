package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock HTTP client for testing
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

func TestAnalyticsUsageCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful usage analytics for all APIs",
			args: []string{"usage"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/usage": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_calls": 15000,
						"unique_consumers": 250,
						"apis": []map[string]interface{}{
							{
								"name": "weather-api",
								"calls": 10000,
								"consumers": 150,
								"error_rate": 0.02,
							},
							{
								"name": "payment-api",
								"calls": 5000,
								"consumers": 100,
								"error_rate": 0.01,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Usage Analytics",
				"Period: 2024-01-01 to 2024-01-31",
				"Total Calls: 15,000",
				"Unique Consumers: 250",
				"weather-api",
				"10,000",
				"payment-api",
				"5,000",
			},
			expectError: false,
		},
		{
			name: "usage analytics for specific API",
			args: []string{"usage", "weather-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/usage": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_calls": 10000,
						"unique_consumers": 150,
						"endpoints": []map[string]interface{}{
							{
								"path": "/weather/{city}",
								"method": "GET",
								"calls": 8000,
								"avg_latency": 125,
							},
							{
								"path": "/forecast/{city}",
								"method": "GET",
								"calls": 2000,
								"avg_latency": 200,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"weather-api",
				"/weather/{city}",
				"8,000",
				"/forecast/{city}",
				"2,000",
			},
			expectError: false,
		},
		{
			name: "error handling for failed API call",
			args: []string{"usage"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/usage": {
					statusCode: 500,
					body:       map[string]string{"error": "Internal server error"},
				},
			},
			expectedOutput: []string{},
			expectError:    true,
		},
		{
			name: "usage analytics with custom period",
			args: []string{"usage", "--period", "7d"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/usage": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-24",
							"end":   "2024-01-31",
						},
						"total_calls": 3500,
						"unique_consumers": 85,
					},
				},
			},
			expectedOutput: []string{
				"Period: 2024-01-24 to 2024-01-31",
				"3,500",
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
			
			// Reset flags
			analyticsPeriod = ""
			analyticsGroupBy = ""
			analyticsFormat = "table"
			
			// Parse flags if any
			analyticsUsageCmd.ParseFlags(tt.args)
			
			// Execute command
			err := runAnalyticsUsage(cmd, tt.args)
			
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

func TestAnalyticsRevenueCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful revenue analytics",
			args: []string{"revenue"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/revenue": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_revenue": 5000.00,
						"subscription_revenue": 4000.00,
						"usage_revenue": 1000.00,
						"new_subscribers": 25,
						"churned_subscribers": 3,
						"apis": []map[string]interface{}{
							{
								"name": "weather-api",
								"revenue": 3000.00,
								"subscribers": 100,
							},
							{
								"name": "payment-api",
								"revenue": 2000.00,
								"subscribers": 50,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Revenue Analytics",
				"Total Revenue: $5,000.00",
				"Subscription Revenue: $4,000.00",
				"Usage Revenue: $1,000.00",
				"New Subscribers: 25",
				"weather-api",
				"$3,000.00",
			},
			expectError: false,
		},
		{
			name: "revenue analytics with breakdown",
			args: []string{"revenue", "--breakdown"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/revenue": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_revenue": 5000.00,
						"daily_breakdown": []map[string]interface{}{
							{
								"date": "2024-01-01",
								"revenue": 150.00,
								"new_subscriptions": 2,
							},
							{
								"date": "2024-01-02",
								"revenue": 175.00,
								"new_subscriptions": 3,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"2024-01-01",
				"$150.00",
				"2024-01-02",
				"$175.00",
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
			
			// Reset flags
			analyticsPeriod = ""
			analyticsBreakdown = false
			analyticsFormat = "table"
			
			// Execute command
			err := runAnalyticsRevenue(cmd, tt.args)
			
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

func TestAnalyticsConsumersCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful consumer analytics",
			args: []string{"consumers"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/consumers": {
					statusCode: 200,
					body: map[string]interface{}{
						"total_consumers": 250,
						"active_consumers": 200,
						"new_consumers": 25,
						"top_consumers": []map[string]interface{}{
							{
								"name": "Acme Corp",
								"id": "consumer_123",
								"total_calls": 5000,
								"total_spent": 500.00,
								"apis_used": 3,
							},
							{
								"name": "Tech Solutions",
								"id": "consumer_456",
								"total_calls": 3000,
								"total_spent": 300.00,
								"apis_used": 2,
							},
						},
						"geographic_distribution": map[string]int{
							"US": 150,
							"EU": 75,
							"Asia": 25,
						},
					},
				},
			},
			expectedOutput: []string{
				"Consumer Analytics",
				"Total Consumers: 250",
				"Active Consumers: 200",
				"Acme Corp",
				"5,000",
				"$500.00",
				"Geographic Distribution",
				"US",
				"150",
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
			
			// Execute command
			err := runAnalyticsConsumers(cmd, tt.args)
			
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

func TestAnalyticsPerformanceCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
		flagSetup      func()
	}{
		{
			name: "successful performance analytics",
			args: []string{"performance"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/performance": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01T00:00:00Z",
							"end":   "2024-01-01T23:59:59Z",
						},
						"average_latency": 125,
						"p99_latency": 500,
						"p95_latency": 300,
						"error_rate": 0.02,
						"uptime": 99.95,
						"total_requests": 10000,
						"status_codes": map[string]int{
							"200": 9500,
							"400": 300,
							"500": 200,
						},
					},
				},
			},
			expectedOutput: []string{
				"Performance Analytics",
				"Average Latency: 125ms",
				"P99 Latency: 500ms",
				"Error Rate: 2.00%",
				"Uptime: 99.95%",
			},
			expectError: false,
		},
		{
			name: "performance analytics for specific API",
			args: []string{"performance", "weather-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/performance": {
					statusCode: 200,
					body: map[string]interface{}{
						"api_name": "weather-api",
						"average_latency": 100,
						"endpoints": []map[string]interface{}{
							{
								"path": "/weather/{city}",
								"method": "GET",
								"avg_latency": 95,
								"error_rate": 0.01,
								"calls": 8000,
							},
							{
								"path": "/forecast/{city}",
								"method": "GET",
								"avg_latency": 150,
								"error_rate": 0.03,
								"calls": 2000,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"weather-api",
				"/weather/{city}",
				"95ms",
				"/forecast/{city}",
				"150ms",
			},
			expectError: false,
		},
		{
			name: "performance analytics JSON format",
			args: []string{"performance", "--format", "json"},
			flagSetup: func() {
				analyticsFormat = "json"
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/analytics/performance": {
					statusCode: 200,
					body: map[string]interface{}{
						"average_latency": 125,
						"error_rate": 0.02,
					},
				},
			},
			expectedOutput: []string{
				`"average_latency"`,
				`"error_rate"`,
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
			
			// Reset flags
			analyticsPeriod = ""
			analyticsFormat = "table"
			
			// Setup flags if needed
			if tt.flagSetup != nil {
				tt.flagSetup()
			}
			
			// Execute command
			err := runAnalyticsPerformance(cmd, tt.args)
			
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
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{1000000000, "1.0B"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    int64 // milliseconds
		expected string
	}{
		{50, "50ms"},
		{1000, "1.0s"},
		{1500, "1.5s"},
		{60000, "1.0m"},
		{90000, "1.5m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test
func TestAnalyticsCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test would run against a real test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/analytics/usage":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"period": map[string]string{
					"start": "2024-01-01",
					"end":   "2024-01-31",
				},
				"total_calls": 1000,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Configure to use test server
	testConfig := &config.Config{
		APIEndpoint: server.URL,
	}
	
	// Note: In a real implementation, we would properly mock the config
	// For now, this is a placeholder

	// Run command
	cmd := &cobra.Command{}
	err := runAnalyticsUsage(cmd, []string{})
	
	assert.NoError(t, err)
}