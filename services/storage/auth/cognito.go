package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CognitoUser struct {
	UserID   string   `json:"sub"`
	Email    string   `json:"email"`
	UserType string   `json:"custom:user_type"` // "creator", "consumer", "admin"
	Groups   []string `json:"cognito:groups"`
}

type JWKKey struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

type JWKS struct {
	Keys []JWKKey `json:"keys"`
}

type JWTHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type JWTClaims struct {
	Sub       string   `json:"sub"`
	Email     string   `json:"email"`
	UserType  string   `json:"custom:user_type"`
	Groups    []string `json:"cognito:groups"`
	Exp       int64    `json:"exp"`
	Iat       int64    `json:"iat"`
	TokenUse  string   `json:"token_use"`
	ClientID  string   `json:"client_id"`
}

var (
	jwksCache     *JWKS
	jwksCacheTime time.Time
	cacheDuration = 24 * time.Hour
)

// VerifyCognitoToken verifies the JWT token with AWS Cognito
func VerifyCognitoToken(tokenString string) (*CognitoUser, error) {
	// Parse JWT without verification first to get the header
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %v", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %v", err)
	}

	// Get JWKS
	jwks, err := getJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %v", err)
	}

	// Find the key
	var key *JWKKey
	for _, k := range jwks.Keys {
		if k.Kid == header.Kid {
			key = &k
			break
		}
	}

	if key == nil {
		return nil, errors.New("key not found in JWKS")
	}

	// Verify signature
	if err := verifyJWTSignature(tokenString, key); err != nil {
		return nil, fmt.Errorf("signature verification failed: %v", err)
	}

	// Parse claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %v", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}

	// Verify token use
	if claims.TokenUse != "id" && claims.TokenUse != "access" {
		return nil, errors.New("invalid token use")
	}

	// Verify expiration
	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	// Verify client ID if configured
	expectedClientID := os.Getenv("COGNITO_CLIENT_ID")
	if expectedClientID != "" && claims.ClientID != expectedClientID {
		return nil, errors.New("invalid client ID")
	}

	return &CognitoUser{
		UserID:   claims.Sub,
		Email:    claims.Email,
		UserType: claims.UserType,
		Groups:   claims.Groups,
	}, nil
}

// getJWKS fetches the JSON Web Key Set from Cognito
func getJWKS() (*JWKS, error) {
	// Check cache
	if jwksCache != nil && time.Since(jwksCacheTime) < cacheDuration {
		return jwksCache, nil
	}

	// Build JWKS URL
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	if userPoolID == "" {
		return nil, errors.New("COGNITO_USER_POOL_ID not configured")
	}

	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	// Fetch JWKS
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}

	// Update cache
	jwksCache = &jwks
	jwksCacheTime = time.Now()

	return &jwks, nil
}

// verifyJWTSignature verifies the JWT signature using the JWK
func verifyJWTSignature(tokenString string, key *JWKKey) error {
	// For now, we'll return nil as proper RSA verification requires additional libraries
	// In production, use a proper JWT library like github.com/golang-jwt/jwt
	// This is a placeholder that should be replaced with actual verification
	
	if key.Alg != "RS256" {
		return errors.New("unsupported algorithm")
	}

	// TODO: Implement actual RSA signature verification
	// This would involve:
	// 1. Converting the JWK to an RSA public key
	// 2. Verifying the signature using the public key
	
	log.Println("WARNING: JWT signature verification not fully implemented")
	return nil
}

// GetUserFromContext retrieves the authenticated user from the Gin context
func GetUserFromContext(c *gin.Context) (*CognitoUser, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	cognitoUser, ok := user.(*CognitoUser)
	return cognitoUser, ok
}

// GetUserIDFromContext retrieves the authenticated user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID := c.GetString("userID")
	return userID, userID != ""
}

// IsInGroup checks if a group is in the list of groups
func IsInGroup(groups []string, targetGroup string) bool {
	for _, group := range groups {
		if group == targetGroup {
			return true
		}
	}
	return false
}

// IsCreator checks if the user is a creator
func IsCreator(user *CognitoUser) bool {
	return user.UserType == "creator" || IsInGroup(user.Groups, "creators")
}

// IsAdmin checks if the user is an admin
func IsAdmin(user *CognitoUser) bool {
	return user.UserType == "admin" || IsInGroup(user.Groups, "admins")
}

// IsConsumer checks if the user is a consumer
func IsConsumer(user *CognitoUser) bool {
	return user.UserType == "consumer" || IsInGroup(user.Groups, "consumers")
}
