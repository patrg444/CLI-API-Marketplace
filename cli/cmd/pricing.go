package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/api-direct/cli/pkg/auth"
	"github.com/api-direct/cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	pricingPlanFile string
)

var pricingCmd = &cobra.Command{
	Use:   "pricing",
	Short: "Manage API pricing plans",
	Long:  `Manage pricing plans for your APIs in the marketplace.`,
}

var setPricingCmd = &cobra.Command{
	Use:   "set [api-name-or-id]",
	Short: "Set pricing plans for an API",
	Long: `Set pricing plans for an API using a JSON configuration file.
	
Example pricing configuration file:
{
  "plans": [
    {
      "name": "Free Tier",
      "type": "free",
      "call_limit": 1000,
      "rate_limit_per_minute": 10,
      "rate_limit_per_day": 1000
    },
    {
      "name": "Basic",
      "type": "subscription",
      "monthly_price": 29.99,
      "call_limit": 100000,
      "rate_limit_per_minute": 60,
      "rate_limit_per_day": 50000
    },
    {
      "name": "Pay As You Go",
      "type": "pay_per_use",
      "price_per_call": 0.001,
      "rate_limit_per_minute": 100,
      "rate_limit_per_day": 100000
    }
  ]
}`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiIdentifier := args[0]

		if pricingPlanFile == "" {
			fmt.Println("Error: Please specify a pricing plan file with --plan-file")
			os.Exit(1)
		}

		// Read pricing plan file
		planData, err := ioutil.ReadFile(pricingPlanFile)
		if err != nil {
			fmt.Printf("Error reading pricing plan file: %v\n", err)
			os.Exit(1)
		}

		// Validate JSON
		var pricingConfig map[string]interface{}
		if err := json.Unmarshal(planData, &pricingConfig); err != nil {
			fmt.Printf("Error: Invalid JSON in pricing plan file: %v\n", err)
			os.Exit(1)
		}

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			fmt.Printf("Error: Not authenticated. Please run 'apidirect auth login' first.\n")
			os.Exit(1)
		}

		// Make API request to set pricing
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/apis/%s/pricing", cfg.APIEndpoint, apiIdentifier)

		resp, err := auth.MakeAuthenticatedRequest("PUT", url, token, planData)
		if err != nil {
			fmt.Printf("Error setting pricing plans: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Printf("✓ Pricing plans updated successfully for API '%s'\n", apiIdentifier)
			
			// Display the plans
			if plans, ok := pricingConfig["plans"].([]interface{}); ok {
				fmt.Println("\nConfigured pricing plans:")
				for _, plan := range plans {
					if p, ok := plan.(map[string]interface{}); ok {
						fmt.Printf("  - %s (%s)\n", p["name"], p["type"])
					}
				}
			}
		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			fmt.Printf("Error: Failed to set pricing plans - %s\n", errorResp["error"])
			os.Exit(1)
		}
	},
}

var getPricingCmd = &cobra.Command{
	Use:   "get [api-name-or-id]",
	Short: "Get current pricing plans for an API",
	Long:  `Retrieve and display the current pricing plans configured for an API.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiIdentifier := args[0]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			fmt.Printf("Error: Not authenticated. Please run 'apidirect auth login' first.\n")
			os.Exit(1)
		}

		// Make API request to get pricing
		cfg := config.Get()
		url := fmt.Sprintf("%s/api/v1/apis/%s/pricing", cfg.APIEndpoint, apiIdentifier)

		resp, err := auth.MakeAuthenticatedRequest("GET", url, token, nil)
		if err != nil {
			fmt.Printf("Error getting pricing plans: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var pricingData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&pricingData); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				os.Exit(1)
			}

			// Pretty print the pricing plans
			fmt.Printf("Pricing plans for API '%s':\n\n", apiIdentifier)
			
			if plans, ok := pricingData["plans"].([]interface{}); ok {
				for i, plan := range plans {
					if p, ok := plan.(map[string]interface{}); ok {
						fmt.Printf("Plan %d: %s\n", i+1, p["name"])
						fmt.Printf("  Type: %s\n", p["type"])
						
						switch p["type"] {
						case "free":
							fmt.Println("  Price: Free")
						case "subscription":
							fmt.Printf("  Price: $%.2f/month\n", p["monthly_price"])
						case "pay_per_use":
							fmt.Printf("  Price: $%.4f/call\n", p["price_per_call"])
						}
						
						if callLimit, ok := p["call_limit"].(float64); ok && callLimit > 0 {
							fmt.Printf("  Call Limit: %.0f/month\n", callLimit)
						} else {
							fmt.Println("  Call Limit: Unlimited")
						}
						
						fmt.Printf("  Rate Limits:\n")
						fmt.Printf("    - Per minute: %.0f\n", p["rate_limit_per_minute"])
						fmt.Printf("    - Per day: %.0f\n", p["rate_limit_per_day"])
						fmt.Println()
					}
				}
			} else {
				fmt.Println("No pricing plans configured.")
			}
		} else {
			var errorResp map[string]string
			json.NewDecoder(resp.Body).Decode(&errorResp)
			fmt.Printf("Error: Failed to get pricing plans - %s\n", errorResp["error"])
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pricingCmd)
	pricingCmd.AddCommand(setPricingCmd)
	pricingCmd.AddCommand(getPricingCmd)

	// Add flags
	setPricingCmd.Flags().StringVar(&pricingPlanFile, "plan-file", "", "Path to pricing plan JSON file")
	setPricingCmd.MarkFlagRequired("plan-file")
}
