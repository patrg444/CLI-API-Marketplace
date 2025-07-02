package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	subscriptionStatus   string
	subscriptionFormat   string
	subscriptionDetailed bool
)

// subscriptionsCmd represents the subscriptions command group
var subscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage your API subscriptions",
	Long: `View and manage your subscriptions to APIs in the marketplace.

This command helps you:
- View active and past subscriptions
- Check usage and billing information
- Cancel or modify subscriptions
- View API keys and endpoints`,
}

// Subscriptions subcommands
var subscriptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your subscriptions",
	Long: `List all your API subscriptions with their current status,
usage, and billing information.

Examples:
  apidirect subscriptions list                 # All subscriptions
  apidirect subscriptions list --status active # Only active
  apidirect subscriptions list --format json   # JSON output`,
	RunE: runSubscriptionsList,
}

var subscriptionsShowCmd = &cobra.Command{
	Use:   "show [subscription-id]",
	Short: "Show subscription details",
	Long: `Show detailed information about a specific subscription including
usage statistics, billing details, and API endpoints.

Examples:
  apidirect subscriptions show sub_123abc      # Show specific subscription
  apidirect subscriptions show sub_123abc -d   # Include usage details`,
	Args: cobra.ExactArgs(1),
	RunE: runSubscriptionsShow,
}

var subscriptionsCancelCmd = &cobra.Command{
	Use:   "cancel [subscription-id]",
	Short: "Cancel a subscription",
	Long: `Cancel an API subscription. The subscription will remain active
until the end of the current billing period.

Examples:
  apidirect subscriptions cancel sub_123abc    # Cancel subscription`,
	Args: cobra.ExactArgs(1),
	RunE: runSubscriptionsCancel,
}

var subscriptionsUsageCmd = &cobra.Command{
	Use:   "usage [subscription-id]",
	Short: "View subscription usage",
	Long: `View detailed usage statistics for a subscription including
API calls, data transfer, and rate limit information.

Examples:
  apidirect subscriptions usage sub_123abc     # Current period usage
  apidirect subscriptions usage sub_123abc -d  # Detailed breakdown`,
	Args: cobra.ExactArgs(1),
	RunE: runSubscriptionsUsage,
}

var subscriptionsKeysCmd = &cobra.Command{
	Use:   "keys [subscription-id]",
	Short: "Manage API keys",
	Long: `View and regenerate API keys for a subscription.

Examples:
  apidirect subscriptions keys sub_123abc      # View API keys`,
	Args: cobra.ExactArgs(1),
	RunE: runSubscriptionsKeys,
}

func init() {
	rootCmd.AddCommand(subscriptionsCmd)
	
	// Add subcommands
	subscriptionsCmd.AddCommand(subscriptionsListCmd)
	subscriptionsCmd.AddCommand(subscriptionsShowCmd)
	subscriptionsCmd.AddCommand(subscriptionsCancelCmd)
	subscriptionsCmd.AddCommand(subscriptionsUsageCmd)
	subscriptionsCmd.AddCommand(subscriptionsKeysCmd)
	
	// List flags
	subscriptionsListCmd.Flags().StringVarP(&subscriptionStatus, "status", "s", "", "Filter by status (active, cancelled, expired)")
	subscriptionsListCmd.Flags().StringVarP(&subscriptionFormat, "format", "f", "table", "Output format (table, json)")
	
	// Show flags
	subscriptionsShowCmd.Flags().BoolVarP(&subscriptionDetailed, "detailed", "d", false, "Show detailed information")
	subscriptionsShowCmd.Flags().StringVarP(&subscriptionFormat, "format", "f", "table", "Output format (table, json)")
	
	// Usage flags
	subscriptionsUsageCmd.Flags().BoolVarP(&subscriptionDetailed, "detailed", "d", false, "Show detailed breakdown")
	subscriptionsUsageCmd.Flags().StringVarP(&subscriptionFormat, "format", "f", "table", "Output format (table, json)")
	
	// Keys flags
	subscriptionsKeysCmd.Flags().StringVarP(&subscriptionFormat, "format", "f", "table", "Output format (table, json)")
}

func runSubscriptionsList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// Build URL with filters
	url := fmt.Sprintf("%s/api/v1/subscriptions", cfg.APIEndpoint)
	if subscriptionStatus != "" {
		url += fmt.Sprintf("?status=%s", subscriptionStatus)
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var subscriptions []struct {
		ID            string    `json:"id"`
		APIName       string    `json:"api_name"`
		APIID         string    `json:"api_id"`
		PlanName      string    `json:"plan_name"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
		CurrentPeriod struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		} `json:"current_period"`
		Usage struct {
			Calls     int64 `json:"calls"`
			Limit     int64 `json:"limit"`
			Remaining int64 `json:"remaining"`
		} `json:"usage"`
		Billing struct {
			Amount   float64 `json:"amount"`
			Interval string  `json:"interval"`
			NextDate string  `json:"next_billing_date"`
		} `json:"billing"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return err
	}
	
	// Output based on format
	switch subscriptionFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(subscriptions)
		
	default:
		// Table format
		fmt.Fprintln(cmd.OutOrStdout())
		if len(subscriptions) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No subscriptions found")
			return nil
		}
		
		// Group by status
		active := []struct {
			ID            string    `json:"id"`
			APIName       string    `json:"api_name"`
			APIID         string    `json:"api_id"`
			PlanName      string    `json:"plan_name"`
			Status        string    `json:"status"`
			CreatedAt     time.Time `json:"created_at"`
			CurrentPeriod struct {
				Start time.Time `json:"start"`
				End   time.Time `json:"end"`
			} `json:"current_period"`
			Usage struct {
				Calls     int64 `json:"calls"`
				Limit     int64 `json:"limit"`
				Remaining int64 `json:"remaining"`
			} `json:"usage"`
			Billing struct {
				Amount   float64 `json:"amount"`
				Interval string  `json:"interval"`
				NextDate string  `json:"next_billing_date"`
			} `json:"billing"`
		}{}
		inactive := []struct {
			ID            string    `json:"id"`
			APIName       string    `json:"api_name"`
			APIID         string    `json:"api_id"`
			PlanName      string    `json:"plan_name"`
			Status        string    `json:"status"`
			CreatedAt     time.Time `json:"created_at"`
			CurrentPeriod struct {
				Start time.Time `json:"start"`
				End   time.Time `json:"end"`
			} `json:"current_period"`
			Usage struct {
				Calls     int64 `json:"calls"`
				Limit     int64 `json:"limit"`
				Remaining int64 `json:"remaining"`
			} `json:"usage"`
			Billing struct {
				Amount   float64 `json:"amount"`
				Interval string  `json:"interval"`
				NextDate string  `json:"next_billing_date"`
			} `json:"billing"`
		}{}
		
		for _, sub := range subscriptions {
			if sub.Status == "active" {
				active = append(active, sub)
			} else {
				inactive = append(inactive, sub)
			}
		}
		
		// Show active subscriptions
		if len(active) > 0 {
			color.New(color.FgGreen, color.Bold).Fprintf(cmd.OutOrStdout(), "✅ Active Subscriptions (%d)\n\n", len(active))
			
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "API\tPLAN\tUSAGE\tCOST\tNEXT BILLING\n")
			
			for _, sub := range active {
				usage := fmt.Sprintf("%d/%d", sub.Usage.Calls, sub.Usage.Limit)
				if sub.Usage.Limit == 0 {
					usage = fmt.Sprintf("%d calls", sub.Usage.Calls)
				}
				
				cost := fmt.Sprintf("$%.2f/%s", sub.Billing.Amount, sub.Billing.Interval)
				
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					sub.APIName,
					sub.PlanName,
					usage,
					cost,
					sub.Billing.NextDate,
				)
			}
			w.Flush()
			fmt.Fprintln(cmd.OutOrStdout())
		}
		
		// Show inactive subscriptions
		if len(inactive) > 0 {
			color.New(color.FgYellow, color.Bold).Fprintf(cmd.OutOrStdout(), "⏸️  Inactive Subscriptions (%d)\n\n", len(inactive))
			
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "API\tPLAN\tSTATUS\tCREATED\n")
			
			for _, sub := range inactive {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					sub.APIName,
					sub.PlanName,
					sub.Status,
					sub.CreatedAt.Format("2006-01-02"),
				)
			}
			w.Flush()
		}
		
		// Summary
		fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d subscriptions (%d active)\n", len(subscriptions), len(active))
		
		return nil
	}
}

func runSubscriptionsShow(cmd *cobra.Command, args []string) error {
	subscriptionID := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/subscriptions/%s", cfg.APIEndpoint, subscriptionID)
	if subscriptionDetailed {
		url += "?detailed=true"
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var subscription struct {
		ID          string    `json:"id"`
		APIName     string    `json:"api_name"`
		APIID       string    `json:"api_id"`
		APIEndpoint string    `json:"api_endpoint"`
		PlanName    string    `json:"plan_name"`
		PlanType    string    `json:"plan_type"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		
		CurrentPeriod struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		} `json:"current_period"`
		
		Usage struct {
			Calls           int64   `json:"calls"`
			Limit           int64   `json:"limit"`
			Remaining       int64   `json:"remaining"`
			DataTransferGB  float64 `json:"data_transfer_gb"`
			AverageLatency  int64   `json:"average_latency_ms"`
			ErrorRate       float64 `json:"error_rate"`
		} `json:"usage"`
		
		Billing struct {
			Amount         float64 `json:"amount"`
			Currency       string  `json:"currency"`
			Interval       string  `json:"interval"`
			NextBillingDate string `json:"next_billing_date"`
			PaymentMethod  string  `json:"payment_method"`
			LastInvoice    string  `json:"last_invoice_id"`
		} `json:"billing"`
		
		Features []string `json:"features"`
		
		APIKey struct {
			Key       string    `json:"key"`
			CreatedAt time.Time `json:"created_at"`
			LastUsed  time.Time `json:"last_used"`
		} `json:"api_key"`
		
		// Detailed fields
		UsageHistory []struct {
			Date  string `json:"date"`
			Calls int64  `json:"calls"`
		} `json:"usage_history,omitempty"`
		
		Invoices []struct {
			ID     string  `json:"id"`
			Date   string  `json:"date"`
			Amount float64 `json:"amount"`
			Status string  `json:"status"`
			URL    string  `json:"url"`
		} `json:"invoices,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&subscription); err != nil {
		return err
	}
	
	// Output based on format
	switch subscriptionFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(subscription)
		
	default:
		// Table format
		fmt.Fprintln(cmd.OutOrStdout())
		color.New(color.FgCyan, color.Bold).Fprintf(cmd.OutOrStdout(), "📋 Subscription Details\n\n")
		
		// Basic info
		fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", subscription.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "API: %s\n", color.CyanString(subscription.APIName))
		fmt.Fprintf(cmd.OutOrStdout(), "Plan: %s (%s)\n", subscription.PlanName, subscription.PlanType)
		
		// Status
		statusColor := color.FgGreen
		if subscription.Status != "active" {
			statusColor = color.FgYellow
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", color.New(statusColor).Sprint(subscription.Status))
		fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", subscription.CreatedAt.Format("2006-01-02"))
		
		// API Access
		fmt.Fprintf(cmd.OutOrStdout(), "\n🔌 API Access\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Endpoint: %s\n", color.BlueString(subscription.APIEndpoint))
		
		// Safely display API key
		apiKey := subscription.APIKey.Key
		if len(apiKey) > 12 {
			fmt.Fprintf(cmd.OutOrStdout(), "API Key: %s...%s\n", apiKey[:10], apiKey[len(apiKey)-4:])
		} else {
			// For short keys, just show them partially
			fmt.Fprintf(cmd.OutOrStdout(), "API Key: %s...\n", apiKey[:min(4, len(apiKey))])
		}
		
		if !subscription.APIKey.LastUsed.IsZero() {
			fmt.Fprintf(cmd.OutOrStdout(), "Last Used: %s\n", subscription.APIKey.LastUsed.Format("2006-01-02 15:04:05"))
		}
		
		// Usage
		fmt.Fprintf(cmd.OutOrStdout(), "\n📊 Usage (Current Period: %s to %s)\n", 
			subscription.CurrentPeriod.Start.Format("Jan 02"),
			subscription.CurrentPeriod.End.Format("Jan 02"))
		
		if subscription.Usage.Limit > 0 {
			usagePercent := float64(subscription.Usage.Calls) / float64(subscription.Usage.Limit) * 100
			fmt.Fprintf(cmd.OutOrStdout(), "API Calls: %d / %d (%.1f%%)\n", 
				subscription.Usage.Calls, subscription.Usage.Limit, usagePercent)
			fmt.Fprintf(cmd.OutOrStdout(), "Remaining: %d calls\n", subscription.Usage.Remaining)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "API Calls: %d (unlimited)\n", subscription.Usage.Calls)
		}
		
		if subscription.Usage.DataTransferGB > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Data Transfer: %.2f GB\n", subscription.Usage.DataTransferGB)
		}
		if subscription.Usage.AverageLatency > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Avg Latency: %d ms\n", subscription.Usage.AverageLatency)
		}
		if subscription.Usage.ErrorRate > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Error Rate: %.2f%%\n", subscription.Usage.ErrorRate*100)
		}
		
		// Billing
		fmt.Fprintf(cmd.OutOrStdout(), "\n💳 Billing\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Cost: %s%.2f / %s\n", 
			getCurrencySymbol(subscription.Billing.Currency),
			subscription.Billing.Amount,
			subscription.Billing.Interval)
		fmt.Fprintf(cmd.OutOrStdout(), "Next Billing: %s\n", subscription.Billing.NextBillingDate)
		fmt.Fprintf(cmd.OutOrStdout(), "Payment Method: %s\n", subscription.Billing.PaymentMethod)
		
		// Features
		if len(subscription.Features) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n✨ Features\n")
			for _, feature := range subscription.Features {
				fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", feature)
			}
		}
		
		// Detailed information
		if subscriptionDetailed {
			// Usage history
			if len(subscription.UsageHistory) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n📈 Usage History (Last 7 Days)\n")
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "DATE\tCALLS\n")
				for _, day := range subscription.UsageHistory {
					fmt.Fprintf(w, "%s\t%d\n", day.Date, day.Calls)
				}
				w.Flush()
			}
			
			// Recent invoices
			if len(subscription.Invoices) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n📄 Recent Invoices\n")
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "DATE\tAMOUNT\tSTATUS\n")
				for _, invoice := range subscription.Invoices {
					fmt.Fprintf(w, "%s\t$%.2f\t%s\n", 
						invoice.Date, invoice.Amount, invoice.Status)
				}
				w.Flush()
			}
		}
		
		// Actions hint
		fmt.Fprintf(cmd.OutOrStdout(), "\n💡 Actions:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • View usage details: apidirect subscriptions usage %s\n", subscription.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  • Regenerate API key: apidirect subscriptions keys %s --regenerate\n", subscription.ID)
		if subscription.Status == "active" {
			fmt.Fprintf(cmd.OutOrStdout(), "  • Cancel subscription: apidirect subscriptions cancel %s\n", subscription.ID)
		}
		
		return nil
	}
}

func runSubscriptionsCancel(cmd *cobra.Command, args []string) error {
	subscriptionID := args[0]
	
	// Get subscription details first
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/subscriptions/%s", cfg.APIEndpoint, subscriptionID)
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	var subscription struct {
		APIName string `json:"api_name"`
		PlanName string `json:"plan_name"`
		Billing struct {
			NextBillingDate string `json:"next_billing_date"`
		} `json:"billing"`
	}
	
	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&subscription)
	}
	resp.Body.Close()
	
	// Confirm cancellation
	fmt.Fprintf(cmd.OutOrStdout(), "\n⚠️  Cancel Subscription\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "API: %s\n", subscription.APIName)
	fmt.Fprintf(cmd.OutOrStdout(), "Plan: %s\n", subscription.PlanName)
	fmt.Fprintf(cmd.OutOrStdout(), "\nThe subscription will remain active until: %s\n", subscription.Billing.NextBillingDate)
	
	if !confirmAction("\nAre you sure you want to cancel this subscription?") {
		fmt.Fprintln(cmd.OutOrStdout(), "Cancellation aborted")
		return nil
	}
	
	// Cancel subscription
	cancelURL := fmt.Sprintf("%s/api/v1/subscriptions/%s/cancel", cfg.APIEndpoint, subscriptionID)
	resp, err = makeAuthenticatedRequest("POST", cancelURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		Status           string `json:"status"`
		CancelledAt      string `json:"cancelled_at"`
		ActiveUntil      string `json:"active_until"`
		RefundAmount     float64 `json:"refund_amount,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Fprintln(cmd.OutOrStdout())
	color.Green("✅ Subscription cancelled successfully")
	fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", result.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "Active until: %s\n", result.ActiveUntil)
	if result.RefundAmount > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Refund amount: $%.2f\n", result.RefundAmount)
	}
	
	return nil
}

func runSubscriptionsUsage(cmd *cobra.Command, args []string) error {
	subscriptionID := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/subscriptions/%s/usage", cfg.APIEndpoint, subscriptionID)
	if subscriptionDetailed {
		url += "?detailed=true"
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var usage struct {
		SubscriptionID string `json:"subscription_id"`
		APIName        string `json:"api_name"`
		Period struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"period"`
		
		Summary struct {
			TotalCalls      int64   `json:"total_calls"`
			SuccessfulCalls int64   `json:"successful_calls"`
			FailedCalls     int64   `json:"failed_calls"`
			CallLimit       int64   `json:"call_limit"`
			CallsRemaining  int64   `json:"calls_remaining"`
			DataTransferGB  float64 `json:"data_transfer_gb"`
			AverageLatency  int64   `json:"average_latency_ms"`
			Uptime          float64 `json:"uptime_percentage"`
		} `json:"summary"`
		
		ByEndpoint []struct {
			Endpoint string `json:"endpoint"`
			Method   string `json:"method"`
			Calls    int64  `json:"calls"`
			Errors   int64  `json:"errors"`
			AvgLatency int64 `json:"avg_latency_ms"`
		} `json:"by_endpoint"`
		
		ByDay []struct {
			Date  string `json:"date"`
			Calls int64  `json:"calls"`
			Errors int64 `json:"errors"`
		} `json:"by_day,omitempty"`
		
		ErrorBreakdown []struct {
			StatusCode int   `json:"status_code"`
			Count      int64 `json:"count"`
			Percentage float64 `json:"percentage"`
		} `json:"error_breakdown,omitempty"`
		
		RateLimits struct {
			RequestsPerSecond int `json:"requests_per_second"`
			RequestsPerMinute int `json:"requests_per_minute"`
			RequestsPerHour   int `json:"requests_per_hour"`
			CurrentUsage struct {
				Second int `json:"second"`
				Minute int `json:"minute"`
				Hour   int `json:"hour"`
			} `json:"current_usage"`
		} `json:"rate_limits"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return err
	}
	
	// Output based on format
	switch subscriptionFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(usage)
		
	default:
		// Table format
		fmt.Fprintln(cmd.OutOrStdout())
		color.New(color.FgCyan, color.Bold).Fprintf(cmd.OutOrStdout(), "📊 Usage Report: %s\n", usage.APIName)
		fmt.Fprintf(cmd.OutOrStdout(), "Period: %s to %s\n\n", usage.Period.Start, usage.Period.End)
		
		// Summary
		fmt.Fprintf(cmd.OutOrStdout(), "📈 Summary\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Total Calls: %d", usage.Summary.TotalCalls)
		if usage.Summary.CallLimit > 0 {
			usagePercent := float64(usage.Summary.TotalCalls) / float64(usage.Summary.CallLimit) * 100
			fmt.Fprintf(cmd.OutOrStdout(), " / %d (%.1f%%)\n", usage.Summary.CallLimit, usagePercent)
			fmt.Fprintf(cmd.OutOrStdout(), "Remaining: %d calls\n", usage.Summary.CallsRemaining)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), " (unlimited)\n")
		}
		
		successRate := float64(usage.Summary.SuccessfulCalls) / float64(usage.Summary.TotalCalls) * 100
		fmt.Fprintf(cmd.OutOrStdout(), "Success Rate: %.1f%% (%d successful, %d failed)\n",
			successRate, usage.Summary.SuccessfulCalls, usage.Summary.FailedCalls)
		
		if usage.Summary.DataTransferGB > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Data Transfer: %.2f GB\n", usage.Summary.DataTransferGB)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Average Latency: %d ms\n", usage.Summary.AverageLatency)
		fmt.Fprintf(cmd.OutOrStdout(), "Uptime: %.2f%%\n", usage.Summary.Uptime)
		
		// Rate limits
		fmt.Fprintf(cmd.OutOrStdout(), "\n⚡ Rate Limits\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Allowed: %d/sec, %d/min, %d/hour\n",
			usage.RateLimits.RequestsPerSecond,
			usage.RateLimits.RequestsPerMinute,
			usage.RateLimits.RequestsPerHour)
		fmt.Fprintf(cmd.OutOrStdout(), "Current: %d/sec, %d/min, %d/hour\n",
			usage.RateLimits.CurrentUsage.Second,
			usage.RateLimits.CurrentUsage.Minute,
			usage.RateLimits.CurrentUsage.Hour)
		
		// Endpoint breakdown
		if len(usage.ByEndpoint) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n🔗 Top Endpoints\n")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ENDPOINT\tMETHOD\tCALLS\tERRORS\tAVG LATENCY\n")
			
			for i, ep := range usage.ByEndpoint {
				if i >= 10 {
					break
				}
				errorRate := float64(ep.Errors) / float64(ep.Calls) * 100
				fmt.Fprintf(w, "%s\t%s\t%d\t%d (%.1f%%)\t%d ms\n",
					ep.Endpoint,
					ep.Method,
					ep.Calls,
					ep.Errors,
					errorRate,
					ep.AvgLatency,
				)
			}
			w.Flush()
		}
		
		// Detailed views
		if subscriptionDetailed {
			// Daily breakdown
			if len(usage.ByDay) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n📅 Daily Usage\n")
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "DATE\tCALLS\tERRORS\n")
				for _, day := range usage.ByDay {
					fmt.Fprintf(w, "%s\t%d\t%d\n", day.Date, day.Calls, day.Errors)
				}
				w.Flush()
			}
			
			// Error breakdown
			if len(usage.ErrorBreakdown) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n❌ Error Breakdown\n")
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "STATUS CODE\tCOUNT\tPERCENTAGE\n")
				for _, err := range usage.ErrorBreakdown {
					fmt.Fprintf(w, "%d\t%d\t%.1f%%\n", 
						err.StatusCode, err.Count, err.Percentage)
				}
				w.Flush()
			}
		}
		
		return nil
	}
}

func runSubscriptionsKeys(cmd *cobra.Command, args []string) error {
	subscriptionID := args[0]
	regenerate, _ := cmd.Flags().GetBool("regenerate")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	if regenerate {
		// Confirm regeneration
		fmt.Fprintf(cmd.OutOrStdout(), "\n⚠️  Regenerate API Key\n\n")
		fmt.Fprintln(cmd.OutOrStdout(), "This will invalidate your current API key.")
		fmt.Fprintln(cmd.OutOrStdout(), "You'll need to update it in all your applications.")
		
		if !confirmAction("\nContinue with regeneration?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Regeneration cancelled")
			return nil
		}
		
		// Regenerate key
		url := fmt.Sprintf("%s/api/v1/subscriptions/%s/keys/regenerate", cfg.APIEndpoint, subscriptionID)
		resp, err := makeAuthenticatedRequest("POST", url, nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return handleErrorResponse(resp)
		}
		
		var result struct {
			Key       string    `json:"key"`
			CreatedAt time.Time `json:"created_at"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}
		
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("✅ API key regenerated successfully"))
		fmt.Fprintf(cmd.OutOrStdout(), "\nNew API Key: %s\n", color.YellowString(result.Key))
		fmt.Fprintln(cmd.OutOrStdout(), "\n⚠️  Save this key securely - it won't be shown again!")
		
		return nil
	}
	
	// View keys
	url := fmt.Sprintf("%s/api/v1/subscriptions/%s/keys", cfg.APIEndpoint, subscriptionID)
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var keys struct {
		APIName     string `json:"api_name"`
		APIEndpoint string `json:"api_endpoint"`
		Keys []struct {
			Key         string    `json:"key"`
			Name        string    `json:"name"`
			CreatedAt   time.Time `json:"created_at"`
			LastUsed    time.Time `json:"last_used"`
			CallsToday  int64     `json:"calls_today"`
			Status      string    `json:"status"`
		} `json:"keys"`
		Documentation string `json:"documentation_url"`
		Examples      struct {
			Curl   string `json:"curl"`
			Python string `json:"python"`
			NodeJS string `json:"nodejs"`
		} `json:"examples"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return err
	}
	
	// Output based on format
	switch subscriptionFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(keys)
		
	default:
		// Table format
		fmt.Fprintln(cmd.OutOrStdout())
		color.New(color.FgCyan, color.Bold).Fprintf(cmd.OutOrStdout(), "🔑 API Keys: %s\n\n", keys.APIName)
		
		fmt.Fprintf(cmd.OutOrStdout(), "Endpoint: %s\n", color.BlueString(keys.APIEndpoint))
		if keys.Documentation != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Documentation: %s\n", color.BlueString(keys.Documentation))
		}
		fmt.Fprintln(cmd.OutOrStdout())
		
		// Keys list
		for _, key := range keys.Keys {
			statusColor := color.FgGreen
			if key.Status != "active" {
				statusColor = color.FgRed
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "Key: %s...%s\n", key.Key[:12], key.Key[len(key.Key)-4:])
			if key.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", key.Name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", color.New(statusColor).Sprint(key.Status))
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", key.CreatedAt.Format("2006-01-02"))
			if !key.LastUsed.IsZero() {
				fmt.Fprintf(cmd.OutOrStdout(), "Last Used: %s (%d calls today)\n", 
					key.LastUsed.Format("2006-01-02 15:04:05"), key.CallsToday)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
		
		// Usage examples
		fmt.Fprintf(cmd.OutOrStdout(), "📚 Quick Start Examples\n\n")
		
		if keys.Examples.Curl != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "cURL:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", color.New(color.FgHiBlack).Sprint(keys.Examples.Curl))
		}
		
		if keys.Examples.Python != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Python:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", color.New(color.FgHiBlack).Sprint(keys.Examples.Python))
		}
		
		if keys.Examples.NodeJS != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Node.js:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", color.New(color.FgHiBlack).Sprint(keys.Examples.NodeJS))
		}
		
		// Actions
		fmt.Fprintf(cmd.OutOrStdout(), "💡 To regenerate your API key:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   apidirect subscriptions keys %s --regenerate\n", subscriptionID)
		
		return nil
	}
}

// Add regenerate flag to keys command
func init() {
	subscriptionsKeysCmd.Flags().Bool("regenerate", false, "Regenerate API key")
}

// Helper function
func getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"JPY": "¥",
		"INR": "₹",
		"KRW": "₩",
		"CNY": "¥",
		"AUD": "$",
		"CAD": "$",
		"CHF": "CHF ",
		"SEK": "SEK ",
		"NOK": "NOK ",
		"DKK": "DKK ",
		"PLN": "PLN ",
		"CZK": "CZK ",
		"HUF": "HUF ",
		"RON": "RON ",
		"BGN": "BGN ",
		"HRK": "HRK ",
		"RUB": "₽",
		"TRY": "₺",
		"BRL": "R$",
		"ZAR": "R",
		"MXN": "$",
		"IDR": "Rp",
		"MYR": "RM",
		"PHP": "₱",
		"SGD": "$",
		"THB": "฿",
		"VND": "₫",
	}
	if symbol, ok := symbols[strings.ToUpper(currency)]; ok {
		return symbol
	}
	return currency + " "
}