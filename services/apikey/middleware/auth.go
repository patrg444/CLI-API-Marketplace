package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"sub"`
	Email    string `json:"email"`
	UserType string `json:"custom:user_type"`
	jwt.RegisteredClaims
}

// AuthRequired middleware validates JWT tokens from Cognito
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH",
			})
			c.Abort()
			return
		}

		// Extract token from Bearer scheme
		tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing algorithm
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// In production, this would fetch the public key from Cognito's JWKS endpoint
			// For now, we'll use a placeholder
			publicKey := os.Getenv("COGNITO_PUBLIC_KEY")
			if publicKey == "" {
				// In development, accept any valid JWT structure
				// In production, this MUST validate against Cognito's public key
				return []byte("development-key"), nil
			}

			return []byte(publicKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("user_type", claims.UserType)

		// For consumer endpoints, we need to ensure the user is a consumer
		if strings.Contains(c.Request.URL.Path, "/consumer") && claims.UserType != "consumer" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Consumer access required",
				"code":  "CONSUMER_ONLY",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ConsumerAuth is a specialized auth middleware that ensures the user is a consumer
func ConsumerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run the standard auth
		AuthRequired()(c)
		
		// If auth failed, return
		if c.IsAborted() {
			return
		}

		// Check user type
		userType, exists := c.Get("user_type")
		if !exists || userType != "consumer" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This endpoint is only accessible to consumers",
				"code":  "CONSUMER_ONLY",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
