package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/api-direct/cli/pkg/auth"
	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	publishDescription string
	publishCategory    string
	publishTags        []string
)

var publishCmd = &cobra.Command{
	Use:   "publish [api-name-or-id]",
	Short: "Publish an API to the marketplace",
	Long:  `Publish a deployed API to the marketplace, making it discoverable by consumers.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiIdentifier := args[0]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("not authenticated. Please run 'apidirect auth login' first")
		}

		// Prepare publish request
		publishData := map[string]interface{}{
			"is_published": true,
		}

		if publishDescription != "" {
			publishData["description"] = publishDescription
		}
		if publishCategory != "" {
			publishData["category"] = publishCategory
		}
		if len(publishTags) > 0 {
			publishData["tags"] = publishTags
		}

		// Make API request to publish
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/marketplace/apis/%s/publish", cfg.APIEndpoint, apiIdentifier)

		body, _ := json.Marshal(publishData)
		resp, err := auth.MakeAuthenticatedRequest("PUT", url, token, body)
		if err != nil {
			return fmt.Errorf("publishing API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Printf("✓ API '%s' has been published to the marketplace\n", apiIdentifier)
			
			// Parse and display the marketplace URL
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				if marketplaceURL, ok := result["marketplace_url"].(string); ok {
					fmt.Printf("  View in marketplace: %s\n", marketplaceURL)
				}
			}
		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			return fmt.Errorf("failed to publish API - %s", errorResp["error"])
		}
		return nil
	},
}

var unpublishCmd = &cobra.Command{
	Use:   "unpublish [api-name-or-id]",
	Short: "Remove an API from the marketplace",
	Long:  `Unpublish an API from the marketplace, making it private and no longer discoverable.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiIdentifier := args[0]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("not authenticated. Please run 'apidirect auth login' first")
		}

		// Make API request to unpublish
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/marketplace/apis/%s/publish", cfg.APIEndpoint, apiIdentifier)

		publishData := map[string]interface{}{
			"is_published": false,
		}
		body, _ := json.Marshal(publishData)

		resp, err := auth.MakeAuthenticatedRequest("PUT", url, token, body)
		if err != nil {
			return fmt.Errorf("unpublishing API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Printf("✓ API '%s' has been removed from the marketplace\n", apiIdentifier)
		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			return fmt.Errorf("failed to unpublish API - %s", errorResp["error"])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(unpublishCmd)

	// Add flags for publish command
	publishCmd.Flags().StringVar(&publishDescription, "description", "", "API description for marketplace listing")
	publishCmd.Flags().StringVar(&publishCategory, "category", "", "API category (e.g., AI/ML, Data, Finance)")
	publishCmd.Flags().StringSliceVar(&publishTags, "tags", []string{}, "Comma-separated tags for the API")
}
