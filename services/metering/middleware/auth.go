package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type CognitoTokenInfo struct {
	Sub        string                 `json:"sub"`
	Email      string                 `json:"email"`
	Attributes map[string]interface{} `json:"custom:attributes"`
}

// AuthRequired middleware validates JWT tokens
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token with Cognito
		tokenInfo, err := validateCognitoToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", tokenInfo.Sub)
		c.Set("email", tokenInfo.Email)
		
		// Extract user type from custom attributes
		if tokenInfo.Attributes != nil {
			if userType, ok := tokenInfo.Attributes["user_type"].(string); ok {
				c.Set("user_type", userType)
			}
		}

		c.Next()
	}
}

// validateCognitoToken validates the JWT token with AWS Cognito
func validateCognitoToken(token string) (*CognitoTokenInfo, error) {
	// In a production environment, this would validate the token with Cognito
	// For now, we'll implement a basic validation that can be enhanced later
	
	cognitoURL := viper.GetString("COGNITO_URL")
	if cognitoURL == "" {
		// For local development, return a mock token info
		return &CognitoTokenInfo{
			Sub:   "test-user-id",
			Email: "test@example.com",
			Attributes: map[string]interface{}{
				"user_type": "creator",
			},
		}, nil
	}

	// Make request to Cognito to validate token
	req, err := http.NewRequestWithContext(context.Background(), "GET", cognitoURL+"/oauth2/userInfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed")
	}

	var tokenInfo CognitoTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}
