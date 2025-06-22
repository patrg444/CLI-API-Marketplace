package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSearchCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "search with query",
			args: []string{"weather"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/search": {
					statusCode: 200,
					body: map[string]interface{}{
						"query":  "weather",
						"total":  25,
						"offset": 0,
						"limit":  20,
						"results": []map[string]interface{}{
							{
								"id":             "api_123",
								"name":           "Weather API Pro",
								"description":    "Professional weather data API with global coverage",
								"category":       "Data & Analytics",
								"tags":           []string{"weather", "forecast", "climate"},
								"creator":        "WeatherTech",
								"rating":         4.8,
								"review_count":   156,
								"subscriber_count": 1250,
								"pricing_model":  "subscription_monthly",
								"starting_price": 49.99,
								"currency":       "USD",
								"featured":       true,
								"verified":       true,
								"endpoint_count": 12,
							},
							{
								"id":             "api_456",
								"name":           "Simple Weather",
								"description":    "Easy-to-use weather API for basic needs",
								"category":       "Data & Analytics",
								"tags":           []string{"weather", "simple"},
								"creator":        "DevTools",
								"rating":         4.2,
								"review_count":   45,
								"subscriber_count": 320,
								"pricing_model":  "free",
								"starting_price": 0,
							},
						},
						"facets": map[string]interface{}{
							"categories": map[string]int{
								"Data & Analytics": 20,
								"IoT":              5,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Search Results for \"weather\"",
				"Found 25 APIs",
				"⭐ ✓ Weather API Pro",
				"Data & Analytics",
				"4.8★ (156)",
				"From $49.99/mo",
				"1250",
				"Simple Weather",
				"Free",
			},
			expectError: false,
		},
		{
			name: "search with filters",
			args: []string{},
			flags: map[string]string{
				"category": "Finance",
				"tags":     "payments,stripe",
				"sort":     "popular",
				"price":    "0-10",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/search": {
					statusCode: 200,
					body: map[string]interface{}{
						"total":   5,
						"results": []map[string]interface{}{},
					},
				},
			},
			expectedOutput: []string{
				"Found 5 APIs",
			},
			expectError: false,
		},
		{
			name: "no results found",
			args: []string{"nonexistent"},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/search": {
					statusCode: 200,
					body: map[string]interface{}{
						"query":   "nonexistent",
						"total":   0,
						"results": []map[string]interface{}{},
					},
				},
			},
			expectedOutput: []string{
				"No APIs found matching your criteria",
			},
			expectError: false,
		},
		{
			name: "grid format output",
			args: []string{"payment"},
			flags: map[string]string{
				"format": "grid",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/search": {
					statusCode: 200,
					body: map[string]interface{}{
						"total": 1,
						"results": []map[string]interface{}{
							{
								"name":           "Stripe Connect API",
								"description":    "Full-featured payment processing API",
								"category":       "Finance",
								"tags":           []string{"payments", "stripe", "processing"},
								"creator":        "FinTech Solutions",
								"rating":         4.9,
								"review_count":   523,
								"subscriber_count": 3200,
								"featured":       true,
								"verified":       true,
								"endpoint_count": 45,
								"starting_price": 99.00,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"⭐ Stripe Connect API ✓",
				"Full-featured payment processing",
				"Finance",
				"payments, stripe, processing",
				"4.9 (523 reviews)",
				"3200 subscribers",
				"45 endpoints",
				"From $99.00/mo",
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
			searchCategory = ""
			searchTags = []string{}
			searchSort = "relevance"
			searchPriceRange = ""
			searchFormat = "table"
			searchLimit = 20
			searchOffset = 0

			if category, ok := tt.flags["category"]; ok {
				searchCategory = category
			}
			if tags, ok := tt.flags["tags"]; ok {
				searchTags = strings.Split(tags, ",")
			}
			if sort, ok := tt.flags["sort"]; ok {
				searchSort = sort
			}
			if price, ok := tt.flags["price"]; ok {
				searchPriceRange = price
			}
			if format, ok := tt.flags["format"]; ok {
				searchFormat = format
			}

			// Execute command
			err := runSearch(cmd, tt.args)

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

func TestBrowseCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "list categories",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/categories": {
					statusCode: 200,
					body: []map[string]interface{}{
						{
							"name":        "Data & Analytics",
							"slug":        "data-analytics",
							"description": "APIs for data processing, analytics, and insights",
							"api_count":   245,
							"icon":        "📊",
						},
						{
							"name":        "Finance",
							"slug":        "finance",
							"description": "Payment processing, banking, and financial data APIs",
							"api_count":   187,
							"icon":        "💰",
						},
						{
							"name":        "AI & Machine Learning",
							"slug":        "ai-ml",
							"description": "AI models, NLP, computer vision, and ML services",
							"api_count":   156,
							"icon":        "🤖",
						},
					},
				},
			},
			expectedOutput: []string{
				"API Categories",
				"📊 Data & Analytics",
				"APIs for data processing",
				"245 APIs",
				"💰 Finance",
				"🤖 AI & Machine Learning",
			},
			expectError: false,
		},
		{
			name: "browse specific category",
			args: []string{},
			flags: map[string]string{
				"category": "finance",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/search": {
					statusCode: 200,
					body: map[string]interface{}{
						"total": 187,
						"results": []map[string]interface{}{
							{
								"name":     "Stripe API",
								"category": "Finance",
								"rating":   4.9,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Found 187 APIs",
				"Stripe API",
				"Finance",
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
			searchCategory = ""
			searchFormat = "table"

			if category, ok := tt.flags["category"]; ok {
				searchCategory = category
			}

			// Execute command
			err := runBrowse(cmd, tt.args)

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

func TestTrendingCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "trending APIs",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/trending": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": "week",
						"apis": []map[string]interface{}{
							{
								"rank":           1,
								"rank_change":    3,
								"id":             "api_123",
								"name":           "AI Assistant API",
								"description":    "Advanced AI chat and completion API",
								"category":       "AI & Machine Learning",
								"rating":         4.9,
								"review_count":   234,
								"subscriber_count": 5600,
								"growth_rate":    0.25,
								"pricing_model":  "pay_per_use",
								"starting_price": 0.001,
							},
							{
								"rank":           2,
								"rank_change":    -1,
								"id":             "api_456",
								"name":           "Weather API Pro",
								"category":       "Data & Analytics",
								"rating":         4.7,
								"growth_rate":    0.15,
								"pricing_model":  "subscription_monthly",
								"starting_price": 29.99,
							},
							{
								"rank":           3,
								"rank_change":    0,
								"name":           "SMS Gateway",
								"category":       "Communication",
								"rating":         4.5,
								"growth_rate":    0.10,
								"pricing_model":  "free",
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Trending APIs",
				"#1 ↑3",
				"AI Assistant API",
				"AI & Machine Learning",
				"4.9★ (234)",
				"+25%",
				"From $0.00",
				"#2 ↓1",
				"Weather API Pro",
				"#3 →",
				"Free",
			},
			expectError: false,
		},
		{
			name: "trending in category",
			args: []string{},
			flags: map[string]string{
				"category": "finance",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/trending": {
					statusCode: 200,
					body: map[string]interface{}{
						"period": "week",
						"apis":   []map[string]interface{}{},
					},
				},
			},
			expectedOutput: []string{
				"Trending APIs in finance",
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
			searchCategory = ""
			searchLimit = 10
			searchFormat = "table"

			if category, ok := tt.flags["category"]; ok {
				searchCategory = category
			}

			// Execute command
			err := runTrending(cmd, tt.args)

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

func TestFeaturedCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		mockResponses  map[string]mockResponse
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "featured APIs grid view",
			args: []string{},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/featured": {
					statusCode: 200,
					body: map[string]interface{}{
						"title":       "Featured APIs of the Month",
						"description": "Hand-picked APIs showcasing innovation and reliability",
						"apis": []map[string]interface{}{
							{
								"id":            "api_789",
								"name":          "Vision AI Pro",
								"description":   "State-of-the-art computer vision API with 100+ models",
								"category":      "AI & Machine Learning",
								"tags":          []string{"vision", "ai", "image-recognition"},
								"creator":       "AI Labs",
								"rating":        4.9,
								"review_count":  456,
								"featured_text": "Best-in-class accuracy for object detection",
								"badge":         "Editor's Choice",
								"pricing_model": "subscription_monthly",
								"starting_price": 199.00,
							},
							{
								"id":            "api_101",
								"name":          "Global SMS Gateway",
								"description":   "Reliable SMS delivery to 200+ countries",
								"category":      "Communication",
								"featured_text": "99.9% delivery rate worldwide",
								"badge":         "Most Popular",
								"pricing_model": "pay_per_use",
								"starting_price": 0.01,
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Featured APIs of the Month",
				"Hand-picked APIs",
				"[Editor's Choice]",
				"Vision AI Pro",
				"Best-in-class accuracy",
				"AI & Machine Learning",
				"vision, ai, image-recognition",
				"4.9 (456 reviews)",
				"From $199.00/mo",
				"[Most Popular]",
				"Global SMS Gateway",
			},
			expectError: false,
		},
		{
			name: "featured APIs table view",
			args: []string{},
			flags: map[string]string{
				"format": "table",
			},
			mockResponses: map[string]mockResponse{
				"GET /api/v1/marketplace/featured": {
					statusCode: 200,
					body: map[string]interface{}{
						"title": "Featured APIs",
						"apis": []map[string]interface{}{
							{
								"name":          "Vision AI Pro",
								"category":      "AI & Machine Learning",
								"rating":        4.9,
								"review_count":  456,
								"featured_text": "Best accuracy",
								"badge":         "Editor's Choice",
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Featured APIs",
				"[Editor's Choice] Vision AI Pro",
				"AI & Machine Learning",
				"4.9★ (456)",
				"Best accuracy",
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
			searchFormat = "grid"

			if format, ok := tt.flags["format"]; ok {
				searchFormat = format
			}

			// Execute command
			err := runFeatured(cmd, tt.args)

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
func TestGetPricingInterval(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"subscription_monthly", "mo"},
		{"subscription_yearly", "yr"},
		{"pay_per_use", "use"},
		{"one_time", "once"},
		{"unknown", "mo"},
		{"", "mo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getPricingInterval(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d,%d", tt.a, tt.b), func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}