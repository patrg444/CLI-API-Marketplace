package errors

import (
	"encoding/json"
	"fmt"
)

// APIDirectError represents a structured error with code and details
type APIDirectError struct {
	Code        string                 `json:"error_code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Suggestion  string                 `json:"suggestion,omitempty"`
	RetryAfter  int                    `json:"retry_after_seconds,omitempty"`
	DocsURL     string                 `json:"docs_url,omitempty"`
	Recoverable bool                   `json:"recoverable"`
}

func (e *APIDirectError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *APIDirectError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// Error code constants for AI agents to parse
const (
	// Authentication errors
	ErrorNotAuthenticated       = "NOT_AUTHENTICATED"
	ErrorInvalidCredentials     = "INVALID_CREDENTIALS" 
	ErrorTokenExpired          = "TOKEN_EXPIRED"
	
	// Project validation errors
	ErrorProjectNotFound       = "PROJECT_NOT_FOUND"
	ErrorInvalidProjectConfig  = "INVALID_PROJECT_CONFIG"
	ErrorMissingMainFile       = "MISSING_MAIN_FILE"
	ErrorUnsupportedRuntime    = "UNSUPPORTED_RUNTIME"
	ErrorInvalidEndpoints      = "INVALID_ENDPOINTS"
	
	// Hosted deployment errors
	ErrorContainerBuildFailed  = "CONTAINER_BUILD_FAILED"
	ErrorImagePushFailed       = "IMAGE_PUSH_FAILED"
	ErrorDeploymentFailed      = "DEPLOYMENT_FAILED"
	ErrorDeploymentTimeout     = "DEPLOYMENT_TIMEOUT"
	ErrorQuotaExceeded         = "QUOTA_EXCEEDED"
	ErrorInsufficientPlan      = "INSUFFICIENT_PLAN"
	
	// BYOA deployment errors
	ErrorAWSCredentials        = "AWS_CREDENTIALS_INVALID"
	ErrorIAMPermissionDenied   = "IAM_PERMISSION_DENIED"
	ErrorTerraformFailed       = "TERRAFORM_FAILED"
	ErrorAWSResourceLimit      = "AWS_RESOURCE_LIMIT"
	
	// Network and service errors
	ErrorServiceUnavailable    = "SERVICE_UNAVAILABLE"
	ErrorNetworkTimeout        = "NETWORK_TIMEOUT"
	ErrorRateLimited           = "RATE_LIMITED"
	
	// File and code errors
	ErrorCodePackagingFailed   = "CODE_PACKAGING_FAILED"
	ErrorFileUploadFailed      = "FILE_UPLOAD_FAILED"
	ErrorInvalidCode           = "INVALID_CODE"
)

// Common error constructors for AI-friendly errors

func NewAuthError(message string) *APIDirectError {
	return &APIDirectError{
		Code:        ErrorNotAuthenticated,
		Message:     message,
		Suggestion:  "Run 'apidirect login' to authenticate",
		DocsURL:     "https://docs.api-direct.io/authentication",
		Recoverable: true,
	}
}

func NewProjectValidationError(code, message string, details map[string]interface{}) *APIDirectError {
	return &APIDirectError{
		Code:        code,
		Message:     message,
		Details:     details,
		DocsURL:     "https://docs.api-direct.io/project-setup",
		Recoverable: true,
	}
}

func NewHostedDeploymentError(code, message string, retryAfter int) *APIDirectError {
	return &APIDirectError{
		Code:        code,
		Message:     message,
		RetryAfter:  retryAfter,
		DocsURL:     "https://docs.api-direct.io/hosted-deployment",
		Recoverable: code != ErrorQuotaExceeded,
	}
}

func NewBYOAError(code, message string, details map[string]interface{}) *APIDirectError {
	return &APIDirectError{
		Code:        code,
		Message:     message,
		Details:     details,
		DocsURL:     "https://docs.api-direct.io/byoa-setup",
		Recoverable: code != ErrorIAMPermissionDenied,
	}
}

func NewQuotaError(currentUsage, limit int, plan string) *APIDirectError {
	return &APIDirectError{
		Code:    ErrorQuotaExceeded,
		Message: fmt.Sprintf("You have reached your plan limit of %d APIs", limit),
		Details: map[string]interface{}{
			"current_usage": currentUsage,
			"limit":         limit,
			"current_plan":  plan,
		},
		Suggestion:  "Upgrade your plan at https://console.api-direct.io/billing",
		DocsURL:     "https://docs.api-direct.io/pricing",
		Recoverable: false,
	}
}

// Error response for JSON output
type ErrorResponse struct {
	Success bool            `json:"success"`
	Error   *APIDirectError `json:"error"`
}

// Helper function to output errors in appropriate format
func OutputError(err *APIDirectError, jsonFormat bool) {
	if jsonFormat {
		response := ErrorResponse{
			Success: false,
			Error:   err,
		}
		data, _ := json.Marshal(response)
		fmt.Println(string(data))
	} else {
		// Human-readable format
		fmt.Printf("❌ Error: %s\n", err.Message)
		if err.Code != "" {
			fmt.Printf("🔍 Error Code: %s\n", err.Code)
		}
		if err.Suggestion != "" {
			fmt.Printf("💡 Suggestion: %s\n", err.Suggestion)
		}
		if err.DocsURL != "" {
			fmt.Printf("📖 Documentation: %s\n", err.DocsURL)
		}
		if err.RetryAfter > 0 {
			fmt.Printf("⏰ Retry after: %d seconds\n", err.RetryAfter)
		}
	}
}