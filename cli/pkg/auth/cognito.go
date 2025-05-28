package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/pkg/browser"
)

// CognitoAuth handles authentication with AWS Cognito
type CognitoAuth struct {
	client       *cognitoidentityprovider.Client
	userPoolID   string
	clientID     string
	authDomain   string
	callbackPort int
}

// AuthResult contains the authentication tokens
type AuthResult struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int
}

// UserInfo contains basic user information
type UserInfo struct {
	Username string
	Email    string
	Sub      string
}

// NewCognitoAuth creates a new Cognito authentication handler
func NewCognitoAuth(region, userPoolID, clientID, authDomain string) (*CognitoAuth, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	return &CognitoAuth{
		client:       client,
		userPoolID:   userPoolID,
		clientID:     clientID,
		authDomain:   authDomain,
		callbackPort: 8080,
	}, nil
}

// LoginWithBrowser performs OAuth2 authentication flow using the browser
func (c *CognitoAuth) LoginWithBrowser() (*AuthResult, error) {
	// Generate PKCE challenge
	verifier := generateCodeVerifier()
	challenge := generateCodeChallenge(verifier)

	// Start local HTTP server for callback
	resultChan := make(chan *AuthResult)
	errorChan := make(chan error)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", c.callbackPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code == "" {
				errorChan <- fmt.Errorf("no authorization code received")
				fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>No authorization code received.</p><script>window.close();</script></body></html>`)
				return
			}

			// Exchange code for tokens
			tokens, err := c.exchangeCodeForTokens(code, verifier)
			if err != nil {
				errorChan <- err
				fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>%s</p><script>window.close();</script></body></html>`, err.Error())
				return
			}

			resultChan <- tokens
			fmt.Fprintf(w, `<html><body><h1>Authentication Successful!</h1><p>You can close this window and return to the terminal.</p><script>window.close();</script></body></html>`)
		}),
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	// Build authorization URL
	authURL := c.buildAuthURL(challenge)

	// Open browser
	if err := browser.OpenURL(authURL); err != nil {
		server.Close()
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for callback
	select {
	case result := <-resultChan:
		server.Close()
		return result, nil
	case err := <-errorChan:
		server.Close()
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("authentication timeout")
	}
}

// RefreshTokens uses a refresh token to get new access and ID tokens
func (c *CognitoAuth) RefreshTokens(refreshToken string) (*AuthResult, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(c.clientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	}

	resp, err := c.client.InitiateAuth(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	return &AuthResult{
		AccessToken:  aws.ToString(resp.AuthenticationResult.AccessToken),
		IDToken:      aws.ToString(resp.AuthenticationResult.IdToken),
		RefreshToken: refreshToken, // Refresh token doesn't change
		ExpiresIn:    int(aws.ToInt32(resp.AuthenticationResult.ExpiresIn)),
	}, nil
}

// GetUserInfo retrieves user information from the access token
func (c *CognitoAuth) GetUserInfo(accessToken string) (*UserInfo, error) {
	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(accessToken),
	}

	resp, err := c.client.GetUser(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	info := &UserInfo{}
	for _, attr := range resp.UserAttributes {
		switch aws.ToString(attr.Name) {
		case "email":
			info.Email = aws.ToString(attr.Value)
		case "sub":
			info.Sub = aws.ToString(attr.Value)
		}
	}
	info.Username = aws.ToString(resp.Username)

	return info, nil
}

// SignOut signs out the user from all devices
func (c *CognitoAuth) SignOut(accessToken string) error {
	input := &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	}

	_, err := c.client.GlobalSignOut(context.Background(), input)
	if err != nil {
		// Ignore error if token is already invalid
		if strings.Contains(err.Error(), "Access Token has been revoked") {
			return nil
		}
		return fmt.Errorf("failed to sign out: %w", err)
	}

	return nil
}

// Helper functions

func (c *CognitoAuth) buildAuthURL(codeChallenge string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/callback", c.callbackPort))
	params.Set("scope", "openid email profile")
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s/oauth2/authorize?%s", c.authDomain, params.Encode())
}

func (c *CognitoAuth) exchangeCodeForTokens(code, verifier string) (*AuthResult, error) {
	tokenURL := fmt.Sprintf("%s/oauth2/token", c.authDomain)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("code", code)
	data.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/callback", c.callbackPort))
	data.Set("code_verifier", verifier)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &AuthResult{
		AccessToken:  result.AccessToken,
		IDToken:      result.IDToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
