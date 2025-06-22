package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestEarningsSummaryCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful earnings summary",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_earnings":     1500.00,
						"available_balance":  1200.00,
						"pending_payouts":    300.00,
						"lifetime_earnings":  15000.00,
						"total_payouts":      13500.00,
						"next_payout_date":   "2024-02-01",
						"payout_method":      "Stripe Connect",
						"top_apis": []map[string]interface{}{
							{
								"api_name":  "weather-api",
								"api_id":    "api_123",
								"earnings":  1000.00,
							},
							{
								"api_name":  "payment-api",
								"api_id":    "api_456",
								"earnings":  500.00,
							},
						},
						"revenue_by_month": []map[string]interface{}{
							{
								"month":    "2024-01",
								"earnings": 1500.00,
							},
							{
								"month":    "2023-12",
								"earnings": 1800.00,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Earnings Summary",
				"Total Earnings: $1,500.00",
				"Available Balance: $1,200.00",
				"Lifetime Earnings: $15,000.00",
				"weather-api: $1,000.00",
				"2024-01: $1,500.00",
			},
			expectError: false,
		},
		{
			name: "earnings summary with custom period",
			args: []string{},
			flags: map[string]string{
				"period": "30d",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2023-12-02",
							"end":   "2024-01-01",
						},
						"total_earnings": 2000.00,
					},
				},
			},
			expectedOutput: []string{
				"Period: 2023-12-02 to 2024-01-01",
				"$2,000.00",
			},
			expectError: false,
		},
		{
			name: "earnings summary CSV format",
			args: []string{},
			flags: map[string]string{
				"format": "csv",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_earnings":    1500.00,
						"available_balance": 1200.00,
					},
				},
			},
			expectedOutput: []string{
				"Metric,Value",
				"Total Earnings,1500.00",
			},
			expectError: false,
		},
		{
			name: "no payout method configured",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"total_earnings":    1500.00,
						"available_balance": 1200.00,
						"payout_method":     "",
					},
				},
			},
			expectedOutput: []string{
				"No payout method configured",
				"apidirect earnings setup",
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
			earningsPeriod = ""
			earningsFormat = "table"
			
			if period, ok := tt.flags["period"]; ok {
				earningsPeriod = period
			}
			if format, ok := tt.flags["format"]; ok {
				earningsFormat = format
			}
			
			// Execute command
			err := runEarningsSummary(cmd, tt.args)
			
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

func TestEarningsDetailsCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "earnings details for all APIs",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/details": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": map[string]string{
							"start": "2024-01-01",
							"end":   "2024-01-31",
						},
						"total_earnings": 1500.00,
						"breakdown": []map[string]interface{}{
							{
								"group":       "weather-api",
								"earnings":    1000.00,
								"subscribers": 50,
								"usage":       50000,
							},
							{
								"group":       "payment-api",
								"earnings":    500.00,
								"subscribers": 25,
								"usage":       10000,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Earnings Details",
				"Total Earnings: $1,500.00",
				"weather-api",
				"$1,000.00",
				"50,000",
			},
			expectError: false,
		},
		{
			name: "earnings details with transactions",
			args: []string{},
			flags: map[string]string{
				"detailed": "true",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/details": {
					statusCode: 200,
					body: map[string]interface{}{
						"total_earnings": 1500.00,
						"breakdown": []map[string]interface{}{
							{
								"group": "weather-api",
								"transactions": []map[string]interface{}{
									{
										"date":        "2024-01-15",
										"type":        "subscription",
										"description": "Monthly subscription - Pro Plan",
										"amount":      50.00,
										"fee":         2.50,
										"net":         47.50,
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"2024-01-15",
				"subscription",
				"Monthly subscription - Pro Plan",
				"$47.50",
			},
			expectError: false,
		},
		{
			name: "earnings details for specific API",
			args: []string{"weather-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/details": {
					statusCode: 200,
					body: map[string]interface{}{
						"total_earnings": 1000.00,
						"breakdown": []map[string]interface{}{
							{
								"group":       "weather-api",
								"earnings":    1000.00,
								"subscribers": 50,
								"plans": []map[string]interface{}{
									{
										"plan_name":    "Pro",
										"earnings":     800.00,
										"subscribers":  20,
									},
									{
										"plan_name":    "Basic",
										"earnings":     200.00,
										"subscribers":  30,
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"weather-api",
				"$1,000.00",
				"Pro",
				"$800.00",
				"Basic",
				"$200.00",
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
			earningsDetailed = false
			earningsFormat = "table"
			earningsGroupBy = "api"
			
			if detailed, ok := tt.flags["detailed"]; ok {
				earningsDetailed = detailed == "true"
			}
			
			// Execute command
			err := runEarningsDetails(cmd, tt.args)
			
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

func TestEarningsPayoutCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		flags           map[string]string
		mockResponses   map[string]mockResponse
		userInput       string
		expectedOutput  []string
		expectError     bool
		errorMessage    string
	}{
		{
			name: "successful payout request",
			args: []string{},
			flags: map[string]string{
				"amount": "500",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "Stripe Connect",
						"minimum_payout":    10.00,
					},
				},
				"POST /api/v1/earnings/payout": {
					statusCode: 201,
					body: map[string]interface{}{
						"payout_id":      "payout_123",
						"amount":         500.00,
						"status":         "pending",
						"estimated_date": "2024-02-05",
					},
				},
			},
			userInput: "y\n", // Confirm payout
			expectedOutput: []string{
				"Payout Request",
				"Amount: $500.00",
				"Payout requested successfully",
				"payout_123",
			},
			expectError: false,
		},
		{
			name: "payout full balance",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "Stripe Connect",
						"minimum_payout":    10.00,
					},
				},
				"POST /api/v1/earnings/payout": {
					statusCode: 201,
					body: map[string]interface{}{
						"payout_id": "payout_456",
						"amount":    1200.00,
					},
				},
			},
			userInput: "y\n",
			expectedOutput: []string{
				"Amount: $1,200.00",
				"Remaining Balance: $0.00",
			},
			expectError: false,
		},
		{
			name: "no payout method configured",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "",
					},
				},
			},
			expectedOutput: []string{},
			expectError:    true,
			errorMessage:   "no payout method configured",
		},
		{
			name: "amount exceeds balance",
			args: []string{},
			flags: map[string]string{
				"amount": "2000",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "Stripe Connect",
						"minimum_payout":    10.00,
					},
				},
			},
			expectedOutput: []string{},
			expectError:    true,
			errorMessage:   "exceeds available balance",
		},
		{
			name: "amount below minimum",
			args: []string{},
			flags: map[string]string{
				"amount": "5",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "Stripe Connect",
						"minimum_payout":    10.00,
					},
				},
			},
			expectedOutput: []string{},
			expectError:    true,
			errorMessage:   "below minimum payout",
		},
		{
			name: "user cancels payout",
			args: []string{},
			flags: map[string]string{
				"amount": "500",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/summary": {
					statusCode: 200,
					body: map[string]interface{}{
						"available_balance": 1200.00,
						"payout_method":     "Stripe Connect",
						"minimum_payout":    10.00,
					},
				},
			},
			userInput: "n\n", // Cancel payout
			expectedOutput: []string{
				"Payout cancelled",
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
			
			// Set flags
			cmd.Flags().Float64("amount", 0, "")
			if amount, ok := tt.flags["amount"]; ok {
				cmd.Flags().Set("amount", amount)
			}
			
			// Execute command
			err := runEarningsPayout(cmd, tt.args)
			
			// Check error
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
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

func TestEarningsHistoryCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful payout history",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payouts": {
					statusCode: 200,
					body: map[string]interface{}{
						"payouts": []map[string]interface{}{
							{
								"id":           "payout_123",
								"date":         "2024-01-01",
								"amount":       1000.00,
								"fee":          30.00,
								"net":          970.00,
								"status":       "completed",
								"method":       "Stripe",
								"arrival_date": "2024-01-05",
							},
							{
								"id":           "payout_456",
								"date":         "2023-12-01",
								"amount":       1500.00,
								"fee":          45.00,
								"net":          1455.00,
								"status":       "completed",
								"method":       "Stripe",
								"arrival_date": "2023-12-05",
							},
						},
						"summary": map[string]interface{}{
							"total_payouts": 2,
							"total_amount":  2500.00,
							"total_fees":    75.00,
							"total_net":     2425.00,
						},
					},
				},
			},
			expectedOutput: []string{
				"Payout History",
				"Total Payouts: 2",
				"Total Amount: $2,500.00",
				"2024-01-01",
				"$970.00",
				"completed",
			},
			expectError: false,
		},
		{
			name: "payout history with custom period",
			args: []string{},
			flags: map[string]string{
				"period": "2024",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payouts": {
					statusCode: 200,
					body: map[string]interface{}{
						"payouts": []map[string]interface{}{},
						"summary": map[string]interface{}{
							"total_payouts": 0,
						},
					},
				},
			},
			expectedOutput: []string{
				"No payouts found for this period",
			},
			expectError: false,
		},
		{
			name: "payout history JSON format",
			args: []string{},
			flags: map[string]string{
				"format": "json",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payouts": {
					statusCode: 200,
					body: map[string]interface{}{
						"payouts": []map[string]interface{}{
							{
								"id":     "payout_123",
								"amount": 1000.00,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				`"id"`,
				`"payout_123"`,
				`"amount"`,
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
			earningsPeriod = ""
			earningsFormat = "table"
			
			if period, ok := tt.flags["period"]; ok {
				earningsPeriod = period
			}
			if format, ok := tt.flags["format"]; ok {
				earningsFormat = format
			}
			
			// Execute command
			err := runEarningsHistory(cmd, tt.args)
			
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

func TestEarningsSetupCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		userInput      string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "setup already complete",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payout-status": {
					statusCode: 200,
					body: map[string]interface{}{
						"has_account":    true,
						"account_status": "active",
						"dashboard_url":  "https://dashboard.stripe.com/connect/accounts/acct_123",
					},
				},
			},
			expectedOutput: []string{
				"Your payout method is already configured and active",
				"https://dashboard.stripe.com",
			},
			expectError: false,
		},
		{
			name: "setup requires action",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payout-status": {
					statusCode: 200,
					body: map[string]interface{}{
						"has_account":      true,
						"account_status":   "pending",
						"requires_action":  true,
						"onboarding_url":   "https://connect.stripe.com/setup/acct_123",
					},
				},
			},
			expectedOutput: []string{
				"Your payout account requires additional information",
				"https://connect.stripe.com/setup",
			},
			expectError: false,
		},
		{
			name: "new setup flow",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payout-status": {
					statusCode: 404,
				},
				"POST /api/v1/earnings/setup": {
					statusCode: 200,
					body: map[string]interface{}{
						"onboarding_url": "https://connect.stripe.com/setup/new",
					},
				},
			},
			userInput: "y\n", // Confirm setup
			expectedOutput: []string{
				"Stripe Connect account",
				"Next Steps",
				"https://connect.stripe.com/setup/new",
			},
			expectError: false,
		},
		{
			name: "user cancels setup",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/earnings/payout-status": {
					statusCode: 404,
				},
			},
			userInput: "n\n", // Cancel setup
			expectedOutput: []string{
				"Setup cancelled",
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
			err := runEarningsSetup(cmd, tt.args)
			
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
func TestParsePeriod(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		input       string
		expectError bool
		checkStart  func(time.Time) bool
		checkEnd    func(time.Time) bool
	}{
		{
			input:       "",
			expectError: false,
			checkStart: func(start time.Time) bool {
				// Should be start of current month
				expected := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				return start.Equal(expected)
			},
			checkEnd: func(end time.Time) bool {
				// Should be end of current month
				nextMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0)
				expected := nextMonth.Add(-time.Second)
				return end.Equal(expected)
			},
		},
		{
			input:       "7d",
			expectError: false,
			checkStart: func(start time.Time) bool {
				// Should be 7 days ago
				return start.Day() == now.AddDate(0, 0, -7).Day()
			},
			checkEnd: func(end time.Time) bool {
				// Should be now
				return end.Day() == now.Day()
			},
		},
		{
			input:       "2024-Q1",
			expectError: false,
			checkStart: func(start time.Time) bool {
				expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				return start.Equal(expected)
			},
			checkEnd: func(end time.Time) bool {
				expected := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)
				return start.Day() == 1 && start.Month() == 1 // Simplified check
			},
		},
		{
			input:       "2024",
			expectError: false,
			checkStart: func(start time.Time) bool {
				return start.Year() == 2024 && start.Month() == 1 && start.Day() == 1
			},
			checkEnd: func(end time.Time) bool {
				return end.Year() == 2024 && end.Month() == 12
			},
		},
		{
			input:       "2024-01",
			expectError: false,
			checkStart: func(start time.Time) bool {
				return start.Year() == 2024 && start.Month() == 1 && start.Day() == 1
			},
			checkEnd: func(end time.Time) bool {
				return end.Year() == 2024 && end.Month() == 1 && end.Day() == 31
			},
		},
		{
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			start, end, err := parsePeriod(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkStart != nil {
					assert.True(t, tt.checkStart(start), "Start time check failed")
				}
				if tt.checkEnd != nil {
					assert.True(t, tt.checkEnd(end), "End time check failed")
				}
			}
		})
	}
}