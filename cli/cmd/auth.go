package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/api-direct/cli/pkg/auth"
	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to API-Direct",
	Long: `Log in to API-Direct using your browser. This will open your default browser
to complete the authentication process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printInfo("Opening browser for authentication...")
		
		// Load config to get Cognito settings
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get Cognito settings from environment or config
		region := os.Getenv("APIDIRECT_REGION")
		if region == "" {
			region = cfg.API.Region
		}
		
		userPoolID := os.Getenv("APIDIRECT_COGNITO_POOL")
		if userPoolID == "" {
			userPoolID = cfg.API.CognitoPool
		}
		
		clientID := os.Getenv("APIDIRECT_COGNITO_CLIENT")
		if clientID == "" {
			clientID = cfg.API.CognitoClient
		}
		
		authDomain := os.Getenv("APIDIRECT_AUTH_DOMAIN")
		if authDomain == "" {
			// Construct from pool ID and region if not provided
			if userPoolID != "" && region != "" {
				authDomain = fmt.Sprintf("https://api-direct-dev-auth.auth.%s.amazoncognito.com", region)
			}
		}

		// Validate required settings
		if userPoolID == "" || clientID == "" || authDomain == "" {
			return fmt.Errorf("missing required Cognito configuration. Please set APIDIRECT_COGNITO_POOL, APIDIRECT_COGNITO_CLIENT, and APIDIRECT_AUTH_DOMAIN environment variables")
		}

		// Create Cognito auth handler
		cognitoAuth, err := auth.NewCognitoAuth(region, userPoolID, clientID, authDomain)
		if err != nil {
			return fmt.Errorf("failed to initialize authentication: %w", err)
		}

		// Perform login
		result, err := cognitoAuth.LoginWithBrowser()
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Get user info
		userInfo, err := cognitoAuth.GetUserInfo(result.AccessToken)
		if err != nil {
			printWarning("Failed to retrieve user information")
		}

		// Calculate expiration time
		expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

		// Update config with auth info
		authConfig := config.AuthConfig{
			AccessToken:  result.AccessToken,
			IDToken:      result.IDToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    expiresAt,
		}

		if userInfo != nil {
			authConfig.Username = userInfo.Username
			authConfig.Email = userInfo.Email
		}

		if err := config.UpdateAuth(authConfig); err != nil {
			return fmt.Errorf("failed to save authentication: %w", err)
		}

		printSuccess("Successfully logged in!")
		if userInfo != nil && userInfo.Email != "" {
			fmt.Printf("Logged in as: %s\n", userInfo.Email)
		}
		
		return nil
	},
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from API-Direct",
	Long:  `Log out from API-Direct and clear stored credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config to get auth info
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Auth.AccessToken == "" {
			printInfo("You are not logged in")
			return nil
		}

		// Try to sign out from Cognito (ignore errors)
		region := os.Getenv("APIDIRECT_REGION")
		if region == "" {
			region = cfg.API.Region
		}
		
		userPoolID := os.Getenv("APIDIRECT_COGNITO_POOL")
		if userPoolID == "" {
			userPoolID = cfg.API.CognitoPool
		}
		
		clientID := os.Getenv("APIDIRECT_COGNITO_CLIENT")
		if clientID == "" {
			clientID = cfg.API.CognitoClient
		}
		
		authDomain := os.Getenv("APIDIRECT_AUTH_DOMAIN")

		if userPoolID != "" && clientID != "" {
			cognitoAuth, err := auth.NewCognitoAuth(region, userPoolID, clientID, authDomain)
			if err == nil {
				cognitoAuth.SignOut(cfg.Auth.AccessToken)
			}
		}

		// Clear local auth info
		if err := config.ClearAuth(); err != nil {
			return fmt.Errorf("failed to clear authentication: %w", err)
		}

		printSuccess("Successfully logged out")
		return nil
	},
}

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  `Display information about the currently authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if authenticated
		if !config.IsAuthenticated() {
			printError("Not authenticated")
			fmt.Println("Please run 'apidirect login' to authenticate")
			return nil
		}

		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Display user info
		fmt.Println("Logged in as:")
		if cfg.Auth.Email != "" {
			fmt.Printf("  Email:    %s\n", cfg.Auth.Email)
		}
		if cfg.Auth.Username != "" {
			fmt.Printf("  Username: %s\n", cfg.Auth.Username)
		}
		
		// Check token expiration
		if !cfg.Auth.ExpiresAt.IsZero() {
			remaining := time.Until(cfg.Auth.ExpiresAt)
			if remaining > 0 {
				fmt.Printf("  Token expires in: %s\n", remaining.Round(time.Minute))
			} else {
				printWarning("Token has expired. Please run 'apidirect login' to re-authenticate")
			}
		}

		// Try to refresh user info from Cognito
		region := os.Getenv("APIDIRECT_REGION")
		if region == "" {
			region = cfg.API.Region
		}
		
		userPoolID := os.Getenv("APIDIRECT_COGNITO_POOL")
		if userPoolID == "" {
			userPoolID = cfg.API.CognitoPool
		}
		
		clientID := os.Getenv("APIDIRECT_COGNITO_CLIENT")
		if clientID == "" {
			clientID = cfg.API.CognitoClient
		}
		
		authDomain := os.Getenv("APIDIRECT_AUTH_DOMAIN")

		if userPoolID != "" && clientID != "" && cfg.Auth.AccessToken != "" {
			cognitoAuth, err := auth.NewCognitoAuth(region, userPoolID, clientID, authDomain)
			if err == nil {
				userInfo, err := cognitoAuth.GetUserInfo(cfg.Auth.AccessToken)
				if err == nil {
					fmt.Printf("\nAccount details:\n")
					fmt.Printf("  User ID: %s\n", userInfo.Sub)
					if userInfo.Email != cfg.Auth.Email && userInfo.Email != "" {
						// Update cached email if different
						cfg.Auth.Email = userInfo.Email
						config.UpdateAuth(cfg.Auth)
					}
				}
			}
		}

		return nil
	},
}
