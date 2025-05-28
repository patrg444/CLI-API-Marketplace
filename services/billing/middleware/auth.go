package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	CognitoUserID string `json:"cognito:username"`
	Email         string `json:"email"`
	UserType      string `json:"custom:user_type"`
	jwt.RegisteredClaims
}

// UserContext holds user information from JWT
type UserContext struct {
	CognitoUserID string
	Email         string
	UserType      string
}

// ContextKey type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user context
	UserContextKey ContextKey = "user"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Extract bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// In production, this would verify against Cognito's public keys
			// For now, we'll use a placeholder
			return []byte("your-secret-key"), nil
		})

		if err != nil || !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Create user context
		userContext := &UserContext{
			CognitoUserID: claims.CognitoUserID,
			Email:         claims.Email,
			UserType:      claims.UserType,
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), UserContextKey, userContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserContext retrieves user context from request
func GetUserContext(r *http.Request) (*UserContext, error) {
	ctx := r.Context().Value(UserContextKey)
	if ctx == nil {
		return nil, fmt.Errorf("user context not found")
	}

	userContext, ok := ctx.(*UserContext)
	if !ok {
		return nil, fmt.Errorf("user context has wrong type")
	}

	return userContext, nil
}

// RequireConsumer ensures the user is a consumer
func RequireConsumer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userContext, err := GetUserContext(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if userContext.UserType != "consumer" {
			respondWithError(w, http.StatusForbidden, "Requires consumer account")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequireCreator ensures the user is a creator
func RequireCreator(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userContext, err := GetUserContext(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if userContext.UserType != "creator" {
			respondWithError(w, http.StatusForbidden, "Requires creator account")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Error marshaling JSON"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
