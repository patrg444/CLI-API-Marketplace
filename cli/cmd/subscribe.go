package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	subscribePlan     string
	subscribeConfirm  bool
	subscribeTrial    bool
)

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe [api-name]",
	Short: "Subscribe to an API",
	Long: `Subscribe to an API from the marketplace. This will create a subscription
and generate API keys for accessing the API.

Examples:
  apidirect subscribe weather-api
  apidirect subscribe weather-api --plan pro
  apidirect subscribe payment-gateway --trial`,
	Args: cobra.ExactArgs(1),
	RunE: runSubscribe,
}

func init() {
	rootCmd.AddCommand(subscribeCmd)
	
	subscribeCmd.Flags().StringVarP(&subscribePlan, "plan", "p", "", "Specific plan to subscribe to")
	subscribeCmd.Flags().BoolVarP(&subscribeConfirm, "yes", "y", false, "Skip confirmation prompt")
	subscribeCmd.Flags().BoolVar(&subscribeTrial, "trial", false, "Start with a free trial if available")
}

func runSubscribe(cmd *cobra.Command, args []string) error {
	apiName := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// First, get API info to show available plans
	infoURL := fmt.Sprintf("%s/api/v1/marketplace/apis/%s", cfg.APIEndpoint, apiName)
	resp, err := makeAuthenticatedRequest("GET", infoURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var apiInfo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Creator     struct {
			Name string `json:"name"`
		} `json:"creator"`
		Pricing struct {
			Model string `json:"model"`
			Plans []struct {
				ID          string   `json:"id"`
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Price       float64  `json:"price"`
				Currency    string   `json:"currency"`
				Interval    string   `json:"interval"`
				Features    []string `json:"features"`
				Limits      struct {
					RequestsPerMonth *int `json:"requests_per_month"`
				} `json:"limits"`
				Popular bool `json:"popular"`
				Trial   *struct {
					Days int `json:"days"`
				} `json:"trial,omitempty"`
			} `json:"plans"`
		} `json:"pricing"`
		Metrics struct {
			Subscribers   int     `json:"subscriber_count"`
			AverageRating float64 `json:"average_rating"`
			TotalReviews  int     `json:"total_reviews"`
		} `json:"metrics"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&apiInfo); err != nil {
		return err
	}
	
	// Display API info
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("📦 %s\n", apiInfo.DisplayName)
	fmt.Printf("%s\n", apiInfo.Description)
	fmt.Printf("by %s", apiInfo.Creator.Name)
	if apiInfo.Metrics.TotalReviews > 0 {
		fmt.Printf(" • %.1f★ (%d reviews)", apiInfo.Metrics.AverageRating, apiInfo.Metrics.TotalReviews)
	}
	fmt.Printf(" • %d subscribers\n", apiInfo.Metrics.Subscribers)
	
	// Check if already subscribed
	checkURL := fmt.Sprintf("%s/api/v1/subscriptions?api_id=%s&status=active", cfg.APIEndpoint, apiInfo.ID)
	checkResp, err := makeAuthenticatedRequest("GET", checkURL, nil)
	if err == nil && checkResp.StatusCode == http.StatusOK {
		var existing []struct {
			ID string `json:"id"`
		}
		if json.NewDecoder(checkResp.Body).Decode(&existing) == nil && len(existing) > 0 {
			checkResp.Body.Close()
			return fmt.Errorf("you already have an active subscription to this API")
		}
		checkResp.Body.Close()
	}
	
	// Display available plans
	fmt.Printf("\n💰 Available Plans (%s pricing):\n\n", apiInfo.Pricing.Model)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "PLAN\tPRICE\tFEATURES\n")
	
	planMap := make(map[string]struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Currency    string   `json:"currency"`
		Interval    string   `json:"interval"`
		Features    []string `json:"features"`
		Limits      struct {
			RequestsPerMonth *int `json:"requests_per_month"`
		} `json:"limits"`
		Popular bool `json:"popular"`
		Trial   *struct {
			Days int `json:"days"`
		} `json:"trial,omitempty"`
	})
	
	for _, plan := range apiInfo.Pricing.Plans {
		planMap[plan.ID] = plan
		
		// Format price
		price := "Free"
		if plan.Price > 0 {
			price = fmt.Sprintf("%s%.2f/%s", getCurrencySymbol(plan.Currency), plan.Price, plan.Interval)
		}
		
		// Mark popular plan
		name := plan.Name
		if plan.Popular {
			name = "⭐ " + name
		}
		
		// Show trial availability
		if plan.Trial != nil && plan.Trial.Days > 0 {
			price += fmt.Sprintf(" (%d-day trial)", plan.Trial.Days)
		}
		
		// Show key features
		features := ""
		if plan.Limits.RequestsPerMonth != nil {
			features = fmt.Sprintf("%s requests/month", formatNumber(int64(*plan.Limits.RequestsPerMonth)))
		} else {
			features = "Unlimited requests"
		}
		if len(plan.Features) > 0 {
			features += " • " + plan.Features[0]
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, price, features)
		
		// Show plan ID for reference
		fmt.Fprintf(w, "  ID: %s\t\t%s\n", plan.ID, plan.Description)
		
		// Show more features
		if len(plan.Features) > 1 {
			for i := 1; i < len(plan.Features) && i < 4; i++ {
				fmt.Fprintf(w, "\t\t• %s\n", plan.Features[i])
			}
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
	
	// Select plan
	selectedPlan := subscribePlan
	if selectedPlan == "" {
		// If trial requested, find a plan with trial
		if subscribeTrial {
			for _, plan := range apiInfo.Pricing.Plans {
				if plan.Trial != nil && plan.Trial.Days > 0 {
					selectedPlan = plan.ID
					fmt.Printf("Selected plan with trial: %s\n", plan.Name)
					break
				}
			}
			if selectedPlan == "" {
				return fmt.Errorf("no plans offer a free trial")
			}
		} else {
			// Interactive plan selection
			fmt.Print("Enter plan ID to subscribe: ")
			fmt.Scanln(&selectedPlan)
		}
	}
	
	// Validate plan exists
	plan, exists := planMap[selectedPlan]
	if !exists {
		return fmt.Errorf("invalid plan ID: %s", selectedPlan)
	}
	
	// Confirm subscription
	if !subscribeConfirm {
		fmt.Printf("\n📋 Subscription Summary:\n")
		fmt.Printf("API: %s\n", apiInfo.DisplayName)
		fmt.Printf("Plan: %s\n", plan.Name)
		
		if plan.Price > 0 {
			fmt.Printf("Price: %s%.2f/%s\n", getCurrencySymbol(plan.Currency), plan.Price, plan.Interval)
			if subscribeTrial && plan.Trial != nil {
				fmt.Printf("Trial: %d days free\n", plan.Trial.Days)
			}
		} else {
			fmt.Printf("Price: Free\n")
		}
		
		if plan.Limits.RequestsPerMonth != nil {
			fmt.Printf("Limit: %s API calls/month\n", formatNumber(int64(*plan.Limits.RequestsPerMonth)))
		}
		
		if !confirmAction("\nProceed with subscription?") {
			fmt.Println("Subscription cancelled")
			return nil
		}
	}
	
	// Create subscription
	subscribeData := struct {
		APIID     string `json:"api_id"`
		PlanID    string `json:"plan_id"`
		StartTrial bool   `json:"start_trial,omitempty"`
	}{
		APIID:     apiInfo.ID,
		PlanID:    plan.ID,
		StartTrial: subscribeTrial && plan.Trial != nil,
	}
	
	data, _ := json.Marshal(subscribeData)
	subscribeURL := fmt.Sprintf("%s/api/v1/subscriptions", cfg.APIEndpoint)
	
	resp, err = makeAuthenticatedRequest("POST", subscribeURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		SubscriptionID string `json:"subscription_id"`
		Status         string `json:"status"`
		APIKey         string `json:"api_key"`
		APIEndpoint    string `json:"api_endpoint"`
		TrialEnds      string `json:"trial_ends,omitempty"`
		NextBilling    string `json:"next_billing_date,omitempty"`
		Documentation  string `json:"documentation_url"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	// Success!
	fmt.Println()
	color.Green("✅ Successfully subscribed to %s!", apiInfo.DisplayName)
	fmt.Printf("\n📋 Subscription Details:\n")
	fmt.Printf("Subscription ID: %s\n", result.SubscriptionID)
	fmt.Printf("Status: %s\n", result.Status)
	
	if result.TrialEnds != "" {
		fmt.Printf("Trial ends: %s\n", result.TrialEnds)
	} else if result.NextBilling != "" {
		fmt.Printf("Next billing: %s\n", result.NextBilling)
	}
	
	fmt.Printf("\n🔑 API Access:\n")
	fmt.Printf("Endpoint: %s\n", color.BlueString(result.APIEndpoint))
	fmt.Printf("API Key: %s\n", color.YellowString(result.APIKey))
	
	fmt.Printf("\n📚 Resources:\n")
	fmt.Printf("Documentation: %s\n", color.BlueString(result.Documentation))
	
	fmt.Printf("\n💡 Quick Start:\n")
	fmt.Printf("curl -H \"X-API-Key: %s\" %s/endpoint\n", result.APIKey, result.APIEndpoint)
	
	fmt.Printf("\n🔧 Manage Subscription:\n")
	fmt.Printf("View details: apidirect subscriptions show %s\n", result.SubscriptionID)
	fmt.Printf("Check usage: apidirect subscriptions usage %s\n", result.SubscriptionID)
	
	return nil
}