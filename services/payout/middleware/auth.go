package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type contextKey string

const (
	UserContextKey    = contextKey("user")
	CreatorContextKey = contextKey("creatorId")
)

type UserClaims struct {
	UserID     string `json:"sub"`
	Email      string `json:"email"`
	UserType   string `json:"custom:user_type"`
	CreatorID  string `json:"custom:creator_id"`
	ConsumerID string `json:"custom:consumer_id"`
	jwt.StandardClaims
}

// AuthMiddleware validates JWT tokens from AWS Cognito
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// In production, validate against Cognito JWKs
			// For now, using a placeholder
			return []byte("your-secret-key"), nil
		})

		if err != nil || !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		// Ensure user is a creator for payout service
		if claims.UserType != "creator" {
			respondWithError(w, http.StatusForbidden, "Access denied: Creator access required")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, CreatorContextKey, claims.CreatorID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnly middleware ensures only platform admins can access certain endpoints
func AdminOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*UserClaims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if user has admin role
		// In production, this would check against a proper role system
		if claims.UserType != "admin" {
			respondWithError(w, http.StatusForbidden, "Admin access required")
			return
		}

		handler(w, r)
	}
}

// GetCreatorID extracts creator ID from request context
func GetCreatorID(r *http.Request) string {
	creatorID, _ := r.Context().Value(CreatorContextKey).(string)
	return creatorID
}

// GetUserClaims extracts user claims from request context
func GetUserClaims(r *http.Request) *UserClaims {
	claims, _ := r.Context().Value(UserContextKey).(*UserClaims)
	return claims
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
