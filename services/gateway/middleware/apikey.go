package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIKeyValidationRequest struct {
	APIKey string `json:"api_key"`
	Path   string `json:"path"`
}

type APIKeyValidationResponse struct {
	Valid          bool   `json:"valid"`
	ConsumerID     string `json:"consumer_id"`
	SubscriptionID string `json:"subscription_id"`
	APIKeyID       string `json:"api_key_id"`
	APIID          string `json:"api_id"`
	RateLimits     struct {
		PerMinute int `json:"per_minute"`
		PerDay    int `json:"per_day"`
		PerMonth  int `json:"per_month"`
	} `json:"rate_limits"`
	Error string `json:"error,omitempty"`
}

// ValidateAPIKey middleware validates API keys by calling the API Key Management Service
func ValidateAPIKey(apiKeyServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Bearer token format
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		// Extract creator and API name from path
		creator := c.Param("creator")
		apiName := c.Param("apiName")
		if creator == "" || apiName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid API path",
				"code":  "INVALID_PATH",
			})
			c.Abort()
			return
		}

		// Validate API key with the API Key Management Service
		validationReq := APIKeyValidationRequest{
			APIKey: apiKey,
			Path:   fmt.Sprintf("%s/%s", creator, apiName),
		}

		jsonData, err := json.Marshal(validationReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "MARSHAL_ERROR",
			})
			c.Abort()
			return
		}

		// Call API Key Management Service
		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/keys/validate", apiKeyServiceURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "API key validation service unavailable",
				"code":  "SERVICE_UNAVAILABLE",
			})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read validation response",
				"code":  "READ_ERROR",
			})
			c.Abort()
			return
		}

		// Parse validation response
		var validationResp APIKeyValidationResponse
		if err := json.Unmarshal(body, &validationResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse validation response",
				"code":  "PARSE_ERROR",
			})
			c.Abort()
			return
		}

		// Check if API key is valid
		if !validationResp.Valid {
			statusCode := http.StatusUnauthorized
			if validationResp.Error == "API_NOT_FOUND" {
				statusCode = http.StatusNotFound
			}
			c.JSON(statusCode, gin.H{
				"error": validationResp.Error,
				"code":  "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		// Store validation data in context for use by other middleware
		c.Set("consumer_id", validationResp.ConsumerID)
		c.Set("subscription_id", validationResp.SubscriptionID)
		c.Set("api_key_id", validationResp.APIKeyID)
		c.Set("api_id", validationResp.APIID)
		c.Set("rate_limits", validationResp.RateLimits)

		c.Next()
	}
}
