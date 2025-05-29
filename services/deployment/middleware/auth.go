package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/deployment/auth"
)

// AuthRequired middleware requires a valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from Bearer scheme
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Verify token with Cognito using shared auth package
		user, err := auth.VerifyCognitoToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user", user)
		c.Set("userID", user.UserID)
		c.Set("userType", user.UserType)
		c.Next()
	}
}

// CreatorOnly middleware ensures the user is a creator
func CreatorOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := auth.GetUserFromContext(c)
		if !exists || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user is a creator
		if !auth.IsCreator(user) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Creator access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly middleware ensures the user has admin privileges
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := auth.GetUserFromContext(c)
		if !exists || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user is an admin
		if !auth.IsAdmin(user) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
