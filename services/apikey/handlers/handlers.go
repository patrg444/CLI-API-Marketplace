package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/apikey/store"
)

// GenerateAPIKeyRequest represents the request to generate a new API key
type GenerateAPIKeyRequest struct {
	Name           string `json:"name" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
}

// GenerateAPIKeyResponse represents the response with the new API key
type GenerateAPIKeyResponse struct {
	APIKey    string          `json:"api_key"`
	KeyInfo   *store.APIKey   `json:"key_info"`
}

// ValidateAPIKeyRequest represents the request to validate an API key
type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// UpdateAPIKeyRequest represents the request to update an API key
type UpdateAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// GenerateAPIKey creates a new API key for a consumer
func GenerateAPIKey(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get consumer info from auth middleware
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		userType, _ := c.Get("user_type")
		
		// Ensure user is a consumer
		if userType != "consumer" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only consumers can generate API keys",
				"code":  "CONSUMER_ONLY",
			})
			return
		}
		
		// Parse request
		var req GenerateAPIKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
				"code":  "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}
		
		// Ensure consumer record exists
		consumerID, err := s.EnsureConsumer(userID.(string), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ensure consumer record",
				"code":  "CONSUMER_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Generate the API key
		fullKey, keyInfo, err := s.GenerateAPIKey(consumerID, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate API key",
				"code":  "GENERATION_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Return the full key only once
		c.JSON(http.StatusCreated, GenerateAPIKeyResponse{
			APIKey:  fullKey,
			KeyInfo: keyInfo,
		})
	}
}

// ValidateAPIKey validates an API key and returns subscription information
func ValidateAPIKey(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request
		var req ValidateAPIKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
				"code":  "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}
		
		// Validate the API key
		validation, err := s.ValidateAPIKey(req.APIKey, req.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate API key",
				"code":  "VALIDATION_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Return validation result
		c.JSON(http.StatusOK, validation)
	}
}

// GetAPIKey retrieves details about a specific API key
func GetAPIKey(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get consumer info from auth middleware
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		
		// Get key ID from URL
		keyID := c.Param("keyId")
		if keyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Key ID required",
				"code":  "KEY_ID_REQUIRED",
			})
			return
		}
		
		// Ensure consumer record exists
		consumerID, err := s.EnsureConsumer(userID.(string), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ensure consumer record",
				"code":  "CONSUMER_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Get the API key
		apiKey, err := s.GetAPIKey(keyID, consumerID)
		if err != nil {
			if err.Error() == "API key not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "API key not found",
					"code":  "KEY_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get API key",
				"code":  "GET_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, apiKey)
	}
}

// ListAPIKeys lists all API keys for the authenticated consumer
func ListAPIKeys(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get consumer info from auth middleware
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		
		// Ensure consumer record exists
		consumerID, err := s.EnsureConsumer(userID.(string), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ensure consumer record",
				"code":  "CONSUMER_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// List API keys
		keys, err := s.ListAPIKeys(consumerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list API keys",
				"code":  "LIST_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"keys": keys,
			"count": len(keys),
		})
	}
}

// RevokeAPIKey revokes a specific API key
func RevokeAPIKey(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get consumer info from auth middleware
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		
		// Get key ID from URL
		keyID := c.Param("keyId")
		if keyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Key ID required",
				"code":  "KEY_ID_REQUIRED",
			})
			return
		}
		
		// Ensure consumer record exists
		consumerID, err := s.EnsureConsumer(userID.(string), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ensure consumer record",
				"code":  "CONSUMER_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Revoke the API key
		err = s.RevokeAPIKey(keyID, consumerID)
		if err != nil {
			if err.Error() == "API key not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "API key not found",
					"code":  "KEY_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to revoke API key",
				"code":  "REVOKE_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "API key revoked successfully",
		})
	}
}

// UpdateAPIKey updates an API key's name
func UpdateAPIKey(s *store.PostgresStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get consumer info from auth middleware
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		
		// Get key ID from URL
		keyID := c.Param("keyId")
		if keyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Key ID required",
				"code":  "KEY_ID_REQUIRED",
			})
			return
		}
		
		// Parse request
		var req UpdateAPIKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
				"code":  "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}
		
		// Ensure consumer record exists
		consumerID, err := s.EnsureConsumer(userID.(string), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ensure consumer record",
				"code":  "CONSUMER_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Update the API key
		err = s.UpdateAPIKey(keyID, consumerID, req.Name)
		if err != nil {
			if err.Error() == "API key not found" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "API key not found",
					"code":  "KEY_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update API key",
				"code":  "UPDATE_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "API key updated successfully",
		})
	}
}
