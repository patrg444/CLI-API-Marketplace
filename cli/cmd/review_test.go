package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestReviewSubmitCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		userInput      string
		expectedOutput []string
		expectError    bool
		errorMessage   string
	}{
		{
			name: "submit review with all flags",
			args: []string{"weather-api"},
			flags: map[string]string{
				"rating":  "5",
				"title":   "Excellent API",
				"message": "Great performance and documentation",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews": {
					statusCode: 201,
					body: map[string]interface{}{
						"review_id":  "rev_123",
						"status":     "published",
						"created_at": time.Now(),
					},
				},
			},
			expectedOutput: []string{
				"Review submitted successfully",
				"rev_123",
			},
			expectError: false,
		},
		{
			name: "submit review interactive mode",
			args: []string{"payment-api"},
			flags: map[string]string{
				"rating": "4",
			},
			userInput: "\nGood API\nWorks well but could use better error messages\n\n",
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews": {
					statusCode: 201,
					body: map[string]interface{}{
						"review_id": "rev_456",
						"status":    "published",
					},
				},
			},
			expectedOutput: []string{
				"Submit Review for payment-api",
				"Rating: ★★★★☆",
				"Review submitted successfully",
			},
			expectError: false,
		},
		{
			name: "invalid rating",
			args: []string{"weather-api"},
			flags: map[string]string{
				"rating": "6",
			},
			expectedOutput: []string{},
			expectError:    true,
			errorMessage:   "rating must be between 1 and 5",
		},
		{
			name: "missing review message in interactive mode",
			args: []string{"weather-api"},
			flags: map[string]string{
				"rating": "3",
			},
			userInput:      "\n\n\n", // Empty title, then immediate double newline for empty message
			expectedOutput: []string{},
			expectError:    true,
			errorMessage:   "review message is required",
		},
		{
			name: "API error during submission",
			args: []string{"weather-api"},
			flags: map[string]string{
				"rating":  "5",
				"message": "Great API",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews": {
					statusCode: 400,
					body:       map[string]string{"error": "You must be subscribed to review this API"},
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
			if tt.userInput != "" {
				cmd.SetIn(strings.NewReader(tt.userInput))
			}

			// Reset and set flags
			reviewRating = 0
			cmd.Flags().IntVarP(&reviewRating, "rating", "r", 0, "")
			cmd.Flags().StringP("title", "t", "", "")
			cmd.Flags().StringP("message", "m", "", "")

			if rating, ok := tt.flags["rating"]; ok {
				reviewRating, _ = strconv.Atoi(rating)
			}
			if title, ok := tt.flags["title"]; ok {
				cmd.Flags().Set("title", title)
			}
			if message, ok := tt.flags["message"]; ok {
				cmd.Flags().Set("message", message)
			}

			// Execute command
			err := runReviewSubmit(cmd, tt.args)

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

func TestReviewListCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "list reviews for API",
			args: []string{"weather-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/weather-api": {
					statusCode: 200,
					body: map[string]interface{}{
						"api": map[string]interface{}{
							"name":           "weather-api",
							"average_rating": 4.5,
							"total_reviews":  125,
							"rating_counts": map[string]int{
								"5": 75,
								"4": 35,
								"3": 10,
								"2": 3,
								"1": 2,
							},
						},
						"reviews": []map[string]interface{}{
							{
								"id":              "rev_123",
								"rating":          5,
								"title":           "Excellent API",
								"message":         "Best weather API I've used. Fast and accurate.",
								"author_name":     "John Doe",
								"author_id":       "user_123",
								"created_at":      time.Now().Add(-7 * 24 * time.Hour),
								"verified_purchase": true,
								"helpful_count":   15,
								"not_helpful_count": 2,
								"creator_response": map[string]interface{}{
									"message":    "Thank you for the feedback!",
									"created_at": time.Now().Add(-5 * 24 * time.Hour),
								},
							},
							{
								"id":              "rev_456",
								"rating":          4,
								"title":           "Good but room for improvement",
								"message":         "Works well, but documentation could be better.",
								"author_name":     "Jane Smith",
								"created_at":      time.Now().Add(-14 * 24 * time.Hour),
								"verified_purchase": true,
								"helpful_count":   8,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Reviews for weather-api",
				"Average Rating: ★★★★½ 4.5 (125 reviews)",
				"5★",
				"75 (60%)",
				"Excellent API",
				"John Doe ✓ Verified",
				"Best weather API",
				"Creator Response:",
				"Thank you for the feedback",
				"15 helpful",
			},
			expectError: false,
		},
		{
			name: "filter reviews by rating",
			args: []string{"weather-api"},
			flags: map[string]string{
				"filter": "5",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/weather-api?sort=helpful&limit=20&filter=5": {
					statusCode: 200,
					body: map[string]interface{}{
						"api": map[string]interface{}{
							"name":           "weather-api",
							"average_rating": 5.0,
							"total_reviews":  75,
						},
						"reviews": []map[string]interface{}{
							{
								"id":      "rev_123",
								"rating":  5,
								"title":   "Perfect!",
								"message": "Couldn't ask for more.",
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"★★★★★",
				"Perfect!",
			},
			expectError: false,
		},
		{
			name: "no reviews yet",
			args: []string{"new-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/new-api": {
					statusCode: 200,
					body: map[string]interface{}{
						"api": map[string]interface{}{
							"name":          "new-api",
							"total_reviews": 0,
						},
						"reviews": []map[string]interface{}{},
					},
				},
			},
			expectedOutput: []string{
				"No reviews yet",
			},
			expectError: false,
		},
		{
			name: "JSON format output",
			args: []string{"weather-api"},
			flags: map[string]string{
				"format": "json",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/weather-api": {
					statusCode: 200,
					body: map[string]interface{}{
						"api": map[string]interface{}{
							"name": "weather-api",
						},
						"reviews": []map[string]interface{}{},
					},
				},
			},
			expectedOutput: []string{
				`"api"`,
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
			
			// Add limit flag with default value
			cmd.Flags().IntP("limit", "l", 20, "Number of reviews to show")

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
			reviewFilter = ""
			reviewSort = "helpful"
			reviewFormat = "table"

			if filter, ok := tt.flags["filter"]; ok {
				reviewFilter = filter
			}
			if format, ok := tt.flags["format"]; ok {
				reviewFormat = format
			}

			// Execute command
			err := runReviewList(cmd, tt.args)

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

func TestReviewMyCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "list my reviews",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/my": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"id":           "rev_123",
							"api_name":     "weather-api",
							"api_id":       "api_123",
							"rating":       5,
							"title":        "Excellent API",
							"message":      "Best weather API I've used.",
							"created_at":   time.Now().Add(-7 * 24 * time.Hour),
							"status":       "published",
							"helpful_count": 15,
							"creator_response": map[string]interface{}{
								"message":    "Thanks for your review!",
								"created_at": time.Now().Add(-5 * 24 * time.Hour),
							},
						},
						{
							"id":           "rev_456",
							"api_name":     "payment-api",
							"rating":       4,
							"title":        "Good but could be better",
							"message":      "Works well overall.",
							"created_at":   time.Now().Add(-30 * 24 * time.Hour),
							"status":       "published",
							"helpful_count": 5,
						},
					},
				},
			},
			expectedOutput: []string{
				"My Reviews (2)",
				"weather-api",
				"★★★★★",
				"Excellent API",
				"published",
				"15",
				"Creator Responses:",
				"Thanks for your review!",
			},
			expectError: false,
		},
		{
			name: "no reviews submitted",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/my": {
					statusCode: 200,
					body:       []map[string]interface{}{},
				},
			},
			expectedOutput: []string{
				"You haven't submitted any reviews yet",
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

			// Reset format to table
			reviewFormat = "table"

			// Execute command
			err := runReviewMy(cmd, tt.args)

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

func TestReviewResponseCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful response",
			args: []string{"rev_123"},
			flags: map[string]string{
				"message": "Thank you for your feedback! We're glad you enjoyed our API.",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews/rev_123/respond": {
					statusCode: 201,
					body: map[string]interface{}{
						"status":     "published",
						"created_at": time.Now(),
					},
				},
			},
			expectedOutput: []string{
				"Response posted successfully",
			},
			expectError: false,
		},
		{
			name: "unauthorized response",
			args: []string{"rev_456"},
			flags: map[string]string{
				"message": "Thanks!",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews/rev_456/respond": {
					statusCode: 403,
					body:       map[string]string{"error": "You can only respond to reviews on your own APIs"},
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

			// Set flags
			cmd.Flags().StringP("message", "m", "", "")
			if message, ok := tt.flags["message"]; ok {
				cmd.Flags().Set("message", message)
			}

			// Execute command
			err := runReviewResponse(cmd, tt.args)

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

func TestReviewReportCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "successful report",
			args: []string{"rev_789"},
			flags: map[string]string{
				"reason": "Spam content promoting unrelated services",
			},
			mockResponses: map[string]mockResponse{
				"POST /api/v1/reviews/rev_789/report": {
					statusCode: 201,
					body: map[string]interface{}{
						"report_id": "report_123",
						"status":    "pending_review",
					},
				},
			},
			expectedOutput: []string{
				"Review reported successfully",
				"report_123",
				"24-48 hours",
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

			// Set flags
			cmd.Flags().StringP("reason", "r", "", "")
			if reason, ok := tt.flags["reason"]; ok {
				cmd.Flags().Set("reason", reason)
			}

			// Execute command
			err := runReviewReport(cmd, tt.args)

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

func TestReviewStatsCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "stats for specific API",
			args: []string{"weather-api"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/stats/weather-api": {
					statusCode: 200,
					body: map[string]interface{}{
						"api_name":       "weather-api",
						"average_rating": 4.5,
						"total_reviews":  125,
						"rating_counts": map[string]int{
							"5": 75,
							"4": 35,
							"3": 10,
							"2": 3,
							"1": 2,
						},
						"trends": map[string]interface{}{
							"last_30_days": map[string]interface{}{
								"average_rating": 4.7,
								"review_count":   25,
							},
							"last_90_days": map[string]interface{}{
								"average_rating": 4.6,
								"review_count":   60,
							},
						},
						"response_metrics": map[string]interface{}{
							"total_responses":  80,
							"response_rate":    0.64,
							"avg_response_time": "2.5 days",
						},
						"keywords": []map[string]interface{}{
							{"word": "accurate", "count": 45, "sentiment": "positive"},
							{"word": "fast", "count": 38, "sentiment": "positive"},
							{"word": "documentation", "count": 22, "sentiment": "neutral"},
							{"word": "expensive", "count": 8, "sentiment": "negative"},
						},
						"recent_reviews": []map[string]interface{}{
							{
								"rating":     5,
								"title":      "Excellent!",
								"created_at": time.Now().Add(-2 * 24 * time.Hour),
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Review Statistics: weather-api",
				"Overall Rating: 4.5 ★★★★½",
				"5★",
				"75 (60%)",
				"Last 30 days: 4.7★ (25 reviews)",
				"Response rate: 64% (80/125)",
				"accurate (45)",
				"Excellent!",
			},
			expectError: false,
		},
		{
			name: "stats for all APIs",
			args: []string{},
			flags: map[string]string{
				"all": "true",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/reviews/stats": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"api_name":       "weather-api",
							"api_id":         "api_123",
							"average_rating": 4.5,
							"total_reviews":  125,
							"recent_trend":   "up",
							"response_rate":  0.64,
						},
						{
							"api_name":       "payment-api",
							"api_id":         "api_456",
							"average_rating": 4.2,
							"total_reviews":  87,
							"recent_trend":   "stable",
							"response_rate":  0.80,
						},
					},
				},
			},
			expectedOutput: []string{
				"Review Statistics for Your APIs",
				"weather-api",
				"4.5 ★★★★½",
				"125",
				"↑",
				"64%",
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

			// Reset format to table
			reviewFormat = "table"

			// Set flags
			cmd.Flags().Bool("all", false, "")
			if all, ok := tt.flags["all"]; ok {
				cmd.Flags().Set("all", all)
			}

			// Execute command
			err := runReviewStats(cmd, tt.args)

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
func TestGetStarRating(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{5.0, "★★★★★"},
		{4.5, "★★★★½"},
		{4.0, "★★★★☆"},
		{3.5, "★★★½☆"},
		{3.0, "★★★☆☆"},
		{2.5, "★★½☆☆"},
		{2.0, "★★☆☆☆"},
		{1.5, "★½☆☆☆"},
		{1.0, "★☆☆☆☆"},
		{0.5, "½☆☆☆☆"},
		{0.0, "☆☆☆☆☆"},
		{4.7, "★★★★½"}, // Should round
		{4.3, "★★★★☆"}, // Should round
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.1f", tt.input), func(t *testing.T) {
			result := getStarRating(tt.input)
			// Remove color codes for comparison
			result = strings.ReplaceAll(result, "\x1b[33m", "")
			result = strings.ReplaceAll(result, "\x1b[0m", "")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 5, "He..."},
		{"Hello World", 11, "Hello World"},
		{"Hello World", 8, "Hello..."},
		{"", 5, ""},
		{"Hi", 2, "Hi"},
		{"Hi", 1, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}