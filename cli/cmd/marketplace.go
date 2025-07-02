package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/api-direct/cli/pkg/auth"
	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Manage marketplace settings for your APIs",
	Long:  `View and manage marketplace settings, status, and analytics for your APIs.`,
}

var marketplaceInfoCmd = &cobra.Command{
	Use:   "info [api-name-or-id]",
	Short: "Get marketplace information for an API",
	Long:  `Display detailed marketplace information for an API including publish status, pricing plans, and basic analytics.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiIdentifier := args[0]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("not authenticated. Please run 'apidirect auth login' first")
		}

		// Make API request to get marketplace info
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/apis/%s/marketplace", cfg.APIEndpoint, apiIdentifier)

		resp, err := auth.MakeAuthenticatedRequest("GET", url, token, nil)
		if err != nil {
			return fmt.Errorf("getting marketplace info: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var marketplaceData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&marketplaceData); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			// Display marketplace information
			fmt.Printf("\nMarketplace Information for '%s'\n", apiIdentifier)
			fmt.Println(strings.Repeat("=", 50))

			// Publish status
			isPublished, _ := marketplaceData["is_published"].(bool)
			if isPublished {
				fmt.Println("Status: 🟢 Published")
				if publishedAt, ok := marketplaceData["published_at"].(string); ok {
					fmt.Printf("Published: %s\n", publishedAt)
				}
			} else {
				fmt.Println("Status: 🔴 Not Published")
			}

			// Marketplace URL
			if marketplaceURL, ok := marketplaceData["marketplace_url"].(string); ok {
				fmt.Printf("Marketplace URL: %s\n", marketplaceURL)
			}

			// Category and tags
			if category, ok := marketplaceData["category"].(string); ok && category != "" {
				fmt.Printf("Category: %s\n", category)
			}
			if tags, ok := marketplaceData["tags"].([]interface{}); ok && len(tags) > 0 {
				tagStrings := make([]string, len(tags))
				for i, tag := range tags {
					tagStrings[i] = fmt.Sprintf("%v", tag)
				}
				fmt.Printf("Tags: %s\n", strings.Join(tagStrings, ", "))
			}

			// Description
			if description, ok := marketplaceData["description"].(string); ok && description != "" {
				fmt.Printf("\nDescription:\n%s\n", description)
			}

			// Pricing plans summary
			if plans, ok := marketplaceData["pricing_plans"].([]interface{}); ok && len(plans) > 0 {
				fmt.Printf("\nPricing Plans: %d configured\n", len(plans))
				for _, plan := range plans {
					if p, ok := plan.(map[string]interface{}); ok {
						planType := p["type"].(string)
						planName := p["name"].(string)
						
						switch planType {
						case "free":
							fmt.Printf("  • %s (Free)\n", planName)
						case "subscription":
							fmt.Printf("  • %s ($%.2f/month)\n", planName, p["monthly_price"])
						case "pay_per_use":
							fmt.Printf("  • %s ($%.4f/call)\n", planName, p["price_per_call"])
						}
					}
				}
			} else {
				fmt.Println("\nPricing Plans: None configured")
			}

			// Analytics summary
			if analytics, ok := marketplaceData["analytics"].(map[string]interface{}); ok {
				fmt.Println("\nAnalytics:")
				if totalSubs, ok := analytics["total_subscriptions"].(float64); ok {
					fmt.Printf("  • Active Subscriptions: %.0f\n", totalSubs)
				}
				if totalCalls, ok := analytics["total_calls"].(float64); ok {
					fmt.Printf("  • Total API Calls: %.0f\n", totalCalls)
				}
				if avgRating, ok := analytics["average_rating"].(float64); ok && avgRating > 0 {
					fmt.Printf("  • Average Rating: %.1f/5.0\n", avgRating)
				}
				if totalReviews, ok := analytics["total_reviews"].(float64); ok && totalReviews > 0 {
					fmt.Printf("  • Total Reviews: %.0f\n", totalReviews)
				}
			}

			// Documentation status
			if docs, ok := marketplaceData["documentation"].(map[string]interface{}); ok {
				fmt.Println("\nDocumentation:")
				if hasOpenAPI, ok := docs["has_openapi"].(bool); ok && hasOpenAPI {
					fmt.Println("  • ✓ OpenAPI specification uploaded")
				} else {
					fmt.Println("  • ✗ No OpenAPI specification")
				}
				if hasMarkdown, ok := docs["has_markdown"].(bool); ok && hasMarkdown {
					fmt.Println("  • ✓ Additional documentation provided")
				}
			}

			fmt.Println()

		} else if resp.StatusCode == 404 {
			return fmt.Errorf("API '%s' not found or you don't have permission to view it", apiIdentifier)
		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			return fmt.Errorf("failed to get marketplace info - %s", errorResp["error"])
		}
		return nil
	},
}

var marketplaceStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get marketplace statistics for all your APIs",
	Long:  `Display aggregated marketplace statistics across all your published APIs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("not authenticated. Please run 'apidirect auth login' first")
		}

		// Make API request to get marketplace stats
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/marketplace/stats", cfg.APIEndpoint)

		resp, err := auth.MakeAuthenticatedRequest("GET", url, token, nil)
		if err != nil {
			return fmt.Errorf("getting marketplace stats: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var stats map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			// Display statistics
			fmt.Println("\nMarketplace Statistics")
			fmt.Println(strings.Repeat("=", 50))

			if publishedAPIs, ok := stats["published_apis"].(float64); ok {
				fmt.Printf("Published APIs: %.0f\n", publishedAPIs)
			}

			if totalSubscriptions, ok := stats["total_subscriptions"].(float64); ok {
				fmt.Printf("Total Active Subscriptions: %.0f\n", totalSubscriptions)
			}

			if monthlyRevenue, ok := stats["monthly_revenue"].(float64); ok {
				fmt.Printf("Estimated Monthly Revenue: $%.2f\n", monthlyRevenue)
			}

			if totalAPICalls, ok := stats["total_api_calls"].(float64); ok {
				fmt.Printf("Total API Calls (This Month): %.0f\n", totalAPICalls)
			}

			// Top APIs
			if topAPIs, ok := stats["top_apis"].([]interface{}); ok && len(topAPIs) > 0 {
				fmt.Println("\nTop Performing APIs:")
				for i, api := range topAPIs {
					if a, ok := api.(map[string]interface{}); ok {
						fmt.Printf("%d. %s - %.0f subscriptions\n", 
							i+1, 
							a["name"], 
							a["subscriptions"])
					}
				}
			}

			fmt.Println()

		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			return fmt.Errorf("failed to get marketplace stats - %s", errorResp["error"])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(marketplaceCmd)
	marketplaceCmd.AddCommand(marketplaceInfoCmd)
	marketplaceCmd.AddCommand(marketplaceStatsCmd)
}
