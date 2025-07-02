package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIDirectError(t *testing.T) {
	tests := []struct {
		name     string
		error    *APIDirectError
		expected string
	}{
		{
			name: "basic error",
			error: &APIDirectError{
				Code:    "TEST_ERROR",
				Message: "This is a test error",
			},
			expected: "[TEST_ERROR] This is a test error",
		},
		{
			name: "error with details",
			error: &APIDirectError{
				Code:    "DETAILED_ERROR",
				Message: "Error with details",
				Details: map[string]interface{}{
					"field":  "username",
					"reason": "already exists",
				},
			},
			expected: "[DETAILED_ERROR] Error with details",
		},
		{
			name: "error with all fields",
			error: &APIDirectError{
				Code:        "FULL_ERROR",
				Message:     "Complete error",
				Details:     map[string]interface{}{"key": "value"},
				Suggestion:  "Try this",
				RetryAfter:  60,
				DocsURL:     "https://docs.example.com",
				Recoverable: true,
			},
			expected: "[FULL_ERROR] Complete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.Error())
		})
	}
}

func TestAPIDirectErrorToJSON(t *testing.T) {
	tests := []struct {
		name  string
		error *APIDirectError
		check func(*testing.T, string)
	}{
		{
			name: "minimal error",
			error: &APIDirectError{
				Code:    "JSON_ERROR",
				Message: "JSON test",
			},
			check: func(t *testing.T, jsonStr string) {
				var parsed APIDirectError
				err := json.Unmarshal([]byte(jsonStr), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "JSON_ERROR", parsed.Code)
				assert.Equal(t, "JSON test", parsed.Message)
				assert.False(t, parsed.Recoverable)
			},
		},
		{
			name: "error with details",
			error: &APIDirectError{
				Code:    "DETAILED",
				Message: "With details",
				Details: map[string]interface{}{
					"count": 42,
					"name":  "test",
				},
				Recoverable: true,
			},
			check: func(t *testing.T, jsonStr string) {
				var parsed APIDirectError
				err := json.Unmarshal([]byte(jsonStr), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "DETAILED", parsed.Code)
				assert.Equal(t, float64(42), parsed.Details["count"]) // JSON numbers are float64
				assert.Equal(t, "test", parsed.Details["name"])
				assert.True(t, parsed.Recoverable)
			},
		},
		{
			name: "error with all fields",
			error: &APIDirectError{
				Code:        "COMPLETE",
				Message:     "All fields",
				Details:     map[string]interface{}{"foo": "bar"},
				Suggestion:  "Do this",
				RetryAfter:  30,
				DocsURL:     "https://docs.test.com",
				Recoverable: false,
			},
			check: func(t *testing.T, jsonStr string) {
				var parsed APIDirectError
				err := json.Unmarshal([]byte(jsonStr), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "COMPLETE", parsed.Code)
				assert.Equal(t, "All fields", parsed.Message)
				assert.Equal(t, "Do this", parsed.Suggestion)
				assert.Equal(t, 30, parsed.RetryAfter)
				assert.Equal(t, "https://docs.test.com", parsed.DocsURL)
				assert.False(t, parsed.Recoverable)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := tt.error.ToJSON()
			assert.NotEmpty(t, jsonStr)
			
			// Verify it's valid JSON
			var data map[string]interface{}
			err := json.Unmarshal([]byte(jsonStr), &data)
			require.NoError(t, err)
			
			// Run specific checks
			if tt.check != nil {
				tt.check(t, jsonStr)
			}
		})
	}
}

func TestNewAuthError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"simple message", "User not authenticated"},
		{"detailed message", "Token expired. Please login again"},
		{"empty message", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthError(tt.message)
			
			assert.Equal(t, ErrorNotAuthenticated, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, "Run 'apidirect login' to authenticate", err.Suggestion)
			assert.Equal(t, "https://docs.api-direct.io/authentication", err.DocsURL)
			assert.True(t, err.Recoverable)
			assert.Nil(t, err.Details)
			assert.Equal(t, 0, err.RetryAfter)
		})
	}
}

func TestNewProjectValidationError(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		details map[string]interface{}
	}{
		{
			name:    "missing main file",
			code:    ErrorMissingMainFile,
			message: "main.py not found",
			details: map[string]interface{}{
				"searched": []string{"main.py", "app.py"},
			},
		},
		{
			name:    "invalid config",
			code:    ErrorInvalidProjectConfig,
			message: "Invalid manifest.yaml",
			details: map[string]interface{}{
				"line":   10,
				"column": 5,
				"error":  "unexpected value",
			},
		},
		{
			name:    "no details",
			code:    ErrorProjectNotFound,
			message: "Project not found",
			details: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProjectValidationError(tt.code, tt.message, tt.details)
			
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.details, err.Details)
			assert.Equal(t, "https://docs.api-direct.io/project-setup", err.DocsURL)
			assert.True(t, err.Recoverable)
			assert.Empty(t, err.Suggestion)
			assert.Equal(t, 0, err.RetryAfter)
		})
	}
}

func TestNewHostedDeploymentError(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		retryAfter  int
		recoverable bool
	}{
		{
			name:        "build failed",
			code:        ErrorContainerBuildFailed,
			message:     "Docker build failed",
			retryAfter:  0,
			recoverable: true,
		},
		{
			name:        "deployment timeout",
			code:        ErrorDeploymentTimeout,
			message:     "Deployment timed out",
			retryAfter:  60,
			recoverable: true,
		},
		{
			name:        "quota exceeded",
			code:        ErrorQuotaExceeded,
			message:     "API limit reached",
			retryAfter:  0,
			recoverable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHostedDeploymentError(tt.code, tt.message, tt.retryAfter)
			
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.retryAfter, err.RetryAfter)
			assert.Equal(t, "https://docs.api-direct.io/hosted-deployment", err.DocsURL)
			assert.Equal(t, tt.recoverable, err.Recoverable)
			assert.Nil(t, err.Details)
			assert.Empty(t, err.Suggestion)
		})
	}
}

func TestNewBYOAError(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		details     map[string]interface{}
		recoverable bool
	}{
		{
			name:    "AWS credentials invalid",
			code:    ErrorAWSCredentials,
			message: "Invalid AWS credentials",
			details: map[string]interface{}{
				"region": "us-west-2",
				"error":  "UnauthorizedException",
			},
			recoverable: true,
		},
		{
			name:    "IAM permission denied",
			code:    ErrorIAMPermissionDenied,
			message: "Insufficient IAM permissions",
			details: map[string]interface{}{
				"required_permissions": []string{"ec2:CreateInstances", "iam:CreateRole"},
			},
			recoverable: false,
		},
		{
			name:        "terraform failed",
			code:        ErrorTerraformFailed,
			message:     "Terraform apply failed",
			details:     nil,
			recoverable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBYOAError(tt.code, tt.message, tt.details)
			
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.details, err.Details)
			assert.Equal(t, "https://docs.api-direct.io/byoa-setup", err.DocsURL)
			assert.Equal(t, tt.recoverable, err.Recoverable)
			assert.Empty(t, err.Suggestion)
			assert.Equal(t, 0, err.RetryAfter)
		})
	}
}

func TestNewQuotaError(t *testing.T) {
	tests := []struct {
		name         string
		currentUsage int
		limit        int
		plan         string
	}{
		{
			name:         "free plan limit",
			currentUsage: 3,
			limit:        3,
			plan:         "free",
		},
		{
			name:         "pro plan limit",
			currentUsage: 10,
			limit:        10,
			plan:         "pro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewQuotaError(tt.currentUsage, tt.limit, tt.plan)
			
			assert.Equal(t, ErrorQuotaExceeded, err.Code)
			assert.Equal(t, fmt.Sprintf("You have reached your plan limit of %d APIs", tt.limit), err.Message)
			assert.Equal(t, tt.currentUsage, err.Details["current_usage"])
			assert.Equal(t, tt.limit, err.Details["limit"])
			assert.Equal(t, tt.plan, err.Details["current_plan"])
			assert.Equal(t, "Upgrade your plan at https://console.api-direct.io/billing", err.Suggestion)
			assert.Equal(t, "https://docs.api-direct.io/pricing", err.DocsURL)
			assert.False(t, err.Recoverable)
			assert.Equal(t, 0, err.RetryAfter)
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Verify all error constants are unique
	constants := []string{
		ErrorNotAuthenticated,
		ErrorInvalidCredentials,
		ErrorTokenExpired,
		ErrorProjectNotFound,
		ErrorInvalidProjectConfig,
		ErrorMissingMainFile,
		ErrorUnsupportedRuntime,
		ErrorInvalidEndpoints,
		ErrorContainerBuildFailed,
		ErrorImagePushFailed,
		ErrorDeploymentFailed,
		ErrorDeploymentTimeout,
		ErrorQuotaExceeded,
		ErrorInsufficientPlan,
		ErrorAWSCredentials,
		ErrorIAMPermissionDenied,
		ErrorTerraformFailed,
		ErrorAWSResourceLimit,
		ErrorServiceUnavailable,
		ErrorNetworkTimeout,
		ErrorRateLimited,
		ErrorCodePackagingFailed,
		ErrorFileUploadFailed,
		ErrorInvalidCode,
	}

	seen := make(map[string]bool)
	for _, constant := range constants {
		assert.NotEmpty(t, constant, "Error constant should not be empty")
		assert.False(t, seen[constant], "Duplicate error constant: %s", constant)
		seen[constant] = true
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		response ErrorResponse
	}{
		{
			name: "success false with error",
			response: ErrorResponse{
				Success: false,
				Error: &APIDirectError{
					Code:    "TEST_ERROR",
					Message: "Test error message",
				},
			},
		},
		{
			name: "success false with detailed error",
			response: ErrorResponse{
				Success: false,
				Error: &APIDirectError{
					Code:        "DETAILED_ERROR",
					Message:     "Detailed error",
					Details:     map[string]interface{}{"key": "value"},
					Suggestion:  "Try this",
					RetryAfter:  30,
					DocsURL:     "https://docs.test.com",
					Recoverable: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.response)
			require.NoError(t, err)
			
			// Unmarshal back
			var parsed ErrorResponse
			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			
			assert.Equal(t, tt.response.Success, parsed.Success)
			assert.Equal(t, tt.response.Error.Code, parsed.Error.Code)
			assert.Equal(t, tt.response.Error.Message, parsed.Error.Message)
		})
	}
}

func TestOutputError(t *testing.T) {
	tests := []struct {
		name       string
		error      *APIDirectError
		jsonFormat bool
		checkJSON  func(*testing.T, string)
		checkText  func(*testing.T, string)
	}{
		{
			name: "JSON format minimal",
			error: &APIDirectError{
				Code:    "TEST_ERROR",
				Message: "Test message",
			},
			jsonFormat: true,
			checkJSON: func(t *testing.T, output string) {
				var response ErrorResponse
				err := json.Unmarshal([]byte(output), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, "TEST_ERROR", response.Error.Code)
				assert.Equal(t, "Test message", response.Error.Message)
			},
		},
		{
			name: "JSON format complete",
			error: &APIDirectError{
				Code:        "COMPLETE_ERROR",
				Message:     "Complete error",
				Details:     map[string]interface{}{"detail": "value"},
				Suggestion:  "Do this",
				RetryAfter:  60,
				DocsURL:     "https://docs.test.com",
				Recoverable: true,
			},
			jsonFormat: true,
			checkJSON: func(t *testing.T, output string) {
				var response ErrorResponse
				err := json.Unmarshal([]byte(output), &response)
				require.NoError(t, err)
				assert.Equal(t, "COMPLETE_ERROR", response.Error.Code)
				assert.Equal(t, "Do this", response.Error.Suggestion)
				assert.Equal(t, 60, response.Error.RetryAfter)
			},
		},
		{
			name: "Human format minimal",
			error: &APIDirectError{
				Code:    "TEST_ERROR",
				Message: "Test message",
			},
			jsonFormat: false,
			checkText: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌ Error: Test message")
				assert.Contains(t, output, "🔍 Error Code: TEST_ERROR")
				assert.NotContains(t, output, "💡 Suggestion:")
				assert.NotContains(t, output, "📖 Documentation:")
				assert.NotContains(t, output, "⏰ Retry after:")
			},
		},
		{
			name: "Human format complete",
			error: &APIDirectError{
				Code:        "COMPLETE_ERROR",
				Message:     "Complete error",
				Suggestion:  "Do this",
				RetryAfter:  60,
				DocsURL:     "https://docs.test.com",
				Recoverable: true,
			},
			jsonFormat: false,
			checkText: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌ Error: Complete error")
				assert.Contains(t, output, "🔍 Error Code: COMPLETE_ERROR")
				assert.Contains(t, output, "💡 Suggestion: Do this")
				assert.Contains(t, output, "📖 Documentation: https://docs.test.com")
				assert.Contains(t, output, "⏰ Retry after: 60 seconds")
			},
		},
		{
			name: "Human format no code",
			error: &APIDirectError{
				Code:    "",
				Message: "Error without code",
			},
			jsonFormat: false,
			checkText: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌ Error: Error without code")
				assert.NotContains(t, output, "🔍 Error Code:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			output := captureOutput(func() {
				OutputError(tt.error, tt.jsonFormat)
			})
			
			if tt.jsonFormat && tt.checkJSON != nil {
				tt.checkJSON(t, output)
			}
			if !tt.jsonFormat && tt.checkText != nil {
				tt.checkText(t, output)
			}
		})
	}
}

func TestErrorRecoverability(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() *APIDirectError
		recoverable bool
	}{
		{
			name:        "auth error is recoverable",
			constructor: func() *APIDirectError { return NewAuthError("test") },
			recoverable: true,
		},
		{
			name:        "project validation error is recoverable",
			constructor: func() *APIDirectError { return NewProjectValidationError("CODE", "msg", nil) },
			recoverable: true,
		},
		{
			name:        "quota error is not recoverable",
			constructor: func() *APIDirectError { return NewQuotaError(5, 5, "free") },
			recoverable: false,
		},
		{
			name: "hosted deployment error (quota) not recoverable",
			constructor: func() *APIDirectError {
				return NewHostedDeploymentError(ErrorQuotaExceeded, "msg", 0)
			},
			recoverable: false,
		},
		{
			name: "hosted deployment error (other) is recoverable",
			constructor: func() *APIDirectError {
				return NewHostedDeploymentError(ErrorDeploymentTimeout, "msg", 60)
			},
			recoverable: true,
		},
		{
			name: "BYOA error (IAM denied) not recoverable",
			constructor: func() *APIDirectError {
				return NewBYOAError(ErrorIAMPermissionDenied, "msg", nil)
			},
			recoverable: false,
		},
		{
			name: "BYOA error (other) is recoverable",
			constructor: func() *APIDirectError {
				return NewBYOAError(ErrorTerraformFailed, "msg", nil)
			},
			recoverable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()
			assert.Equal(t, tt.recoverable, err.Recoverable)
		})
	}
}

// Helper function to capture stdout
func captureOutput(f func()) string {
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	
	// Run the function
	f()
	
	// Restore stdout and close writer
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	var buf strings.Builder
	io.Copy(&buf, r)
	r.Close()
	
	return buf.String()
}

// Benchmark tests
func BenchmarkAPIDirectErrorError(b *testing.B) {
	err := &APIDirectError{
		Code:    "BENCH_ERROR",
		Message: "Benchmark error message",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkAPIDirectErrorToJSON(b *testing.B) {
	err := &APIDirectError{
		Code:        "BENCH_ERROR",
		Message:     "Benchmark error message",
		Details:     map[string]interface{}{"key": "value", "number": 42},
		Suggestion:  "Benchmark suggestion",
		RetryAfter:  30,
		DocsURL:     "https://docs.test.com",
		Recoverable: true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ToJSON()
	}
}