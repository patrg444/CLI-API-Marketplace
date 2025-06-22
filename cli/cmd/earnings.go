package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	earningsPeriod    string
	earningsFormat    string
	earningsAPI       string
	earningsDetailed  bool
	earningsGroupBy   string
)

// earningsCmd represents the earnings command group
var earningsCmd = &cobra.Command{
	Use:   "earnings",
	Short: "Manage and track your API earnings",
	Long: `View earnings, track revenue, and request payouts for your published APIs.

This command helps you:
- View earnings summaries and detailed breakdowns
- Track revenue by API and time period
- Request payouts when balance is available
- Export earnings data for accounting`,
}

// Earnings subcommands
var earningsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "View earnings summary",
	Long: `View a summary of your earnings including available balance,
pending payouts, and revenue trends.

Examples:
  apidirect earnings summary                   # Current month summary
  apidirect earnings summary --period 30d      # Last 30 days
  apidirect earnings summary --period 2024-Q1  # Q1 2024`,
	RunE: runEarningsSummary,
}

var earningsDetailsCmd = &cobra.Command{
	Use:   "details [api-name]",
	Short: "View detailed earnings breakdown",
	Long: `View detailed earnings breakdown by API, subscription plan,
and time period.

Examples:
  apidirect earnings details                   # All APIs
  apidirect earnings details my-api            # Specific API
  apidirect earnings details --group-by daily  # Daily breakdown
  apidirect earnings details --format csv      # Export as CSV`,
	RunE: runEarningsDetails,
}

var earningsPayoutCmd = &cobra.Command{
	Use:   "payout",
	Short: "Request a payout",
	Long: `Request a payout of your available earnings balance.

Examples:
  apidirect earnings payout                    # Request full balance
  apidirect earnings payout --amount 500       # Request specific amount`,
	RunE: runEarningsPayout,
}

var earningsHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "View payout history",
	Long: `View your payout history including completed and pending payouts.

Examples:
  apidirect earnings history                   # Recent payouts
  apidirect earnings history --period 2024     # All 2024 payouts
  apidirect earnings history --format json     # Export as JSON`,
	RunE: runEarningsHistory,
}

var earningsSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up or update payout method",
	Long: `Set up or update your payout method for receiving earnings.

This will guide you through connecting your Stripe account
for receiving payouts.`,
	RunE: runEarningsSetup,
}

func init() {
	rootCmd.AddCommand(earningsCmd)
	
	// Add subcommands
	earningsCmd.AddCommand(earningsSummaryCmd)
	earningsCmd.AddCommand(earningsDetailsCmd)
	earningsCmd.AddCommand(earningsPayoutCmd)
	earningsCmd.AddCommand(earningsHistoryCmd)
	earningsCmd.AddCommand(earningsSetupCmd)
	
	// Common flags
	for _, cmd := range []*cobra.Command{
		earningsSummaryCmd, earningsDetailsCmd, earningsHistoryCmd,
	} {
		cmd.Flags().StringVarP(&earningsPeriod, "period", "p", "", "Time period (e.g., 7d, 30d, 2024-01, 2024-Q1)")
		cmd.Flags().StringVarP(&earningsFormat, "format", "f", "table", "Output format (table, json, csv)")
	}
	
	// Details-specific flags
	earningsDetailsCmd.Flags().StringVarP(&earningsAPI, "api", "a", "", "Filter by specific API")
	earningsDetailsCmd.Flags().BoolVarP(&earningsDetailed, "detailed", "d", false, "Show transaction-level details")
	earningsDetailsCmd.Flags().StringVarP(&earningsGroupBy, "group-by", "g", "api", "Group results by (api, daily, weekly, monthly)")
	
	// Payout-specific flags
	earningsPayoutCmd.Flags().Float64P("amount", "a", 0, "Payout amount (0 for full balance)")
}

func runEarningsSummary(cmd *cobra.Command, args []string) error {
	// Parse period
	start, end, err := parsePeriod(earningsPeriod)
	if err != nil {
		return fmt.Errorf("invalid period: %w", err)
	}
	
	// Call API
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/earnings/summary?start=%s&end=%s",
		cfg.APIEndpoint, start.Format("2006-01-02"), end.Format("2006-01-02"))
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var summary struct {
		Period struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"period"`
		TotalEarnings     float64 `json:"total_earnings"`
		AvailableBalance  float64 `json:"available_balance"`
		PendingPayouts    float64 `json:"pending_payouts"`
		LifetimeEarnings  float64 `json:"lifetime_earnings"`
		TotalPayouts      float64 `json:"total_payouts"`
		NextPayoutDate    string  `json:"next_payout_date"`
		PayoutMethod      string  `json:"payout_method"`
		TopAPIs []struct {
			APIName  string  `json:"api_name"`
			APIID    string  `json:"api_id"`
			Earnings float64 `json:"earnings"`
		} `json:"top_apis"`
		RevenueByMonth []struct {
			Month    string  `json:"month"`
			Earnings float64 `json:"earnings"`
		} `json:"revenue_by_month"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return err
	}
	
	// Output based on format
	switch earningsFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(summary)
		
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		
		// Write summary data
		w.Write([]string{"Metric", "Value"})
		w.Write([]string{"Period Start", summary.Period.Start})
		w.Write([]string{"Period End", summary.Period.End})
		w.Write([]string{"Total Earnings", fmt.Sprintf("%.2f", summary.TotalEarnings)})
		w.Write([]string{"Available Balance", fmt.Sprintf("%.2f", summary.AvailableBalance)})
		w.Write([]string{"Pending Payouts", fmt.Sprintf("%.2f", summary.PendingPayouts)})
		w.Write([]string{"Lifetime Earnings", fmt.Sprintf("%.2f", summary.LifetimeEarnings)})
		
		return nil
		
	default:
		// Table format
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("💰 Earnings Summary\n")
		fmt.Printf("Period: %s to %s\n\n", summary.Period.Start, summary.Period.End)
		
		// Main metrics
		fmt.Printf("📊 Current Period:\n")
		fmt.Printf("   Total Earnings: %s\n", color.GreenString("$%.2f", summary.TotalEarnings))
		fmt.Printf("   Available Balance: %s\n", color.GreenString("$%.2f", summary.AvailableBalance))
		if summary.PendingPayouts > 0 {
			fmt.Printf("   Pending Payouts: %s\n", color.YellowString("$%.2f", summary.PendingPayouts))
		}
		
		fmt.Printf("\n💎 Lifetime:\n")
		fmt.Printf("   Total Earnings: %s\n", color.GreenString("$%.2f", summary.LifetimeEarnings))
		fmt.Printf("   Total Payouts: %s\n", color.BlueString("$%.2f", summary.TotalPayouts))
		
		// Payout info
		fmt.Printf("\n💳 Payout Information:\n")
		if summary.PayoutMethod != "" {
			fmt.Printf("   Method: %s\n", summary.PayoutMethod)
			if summary.NextPayoutDate != "" {
				fmt.Printf("   Next Payout: %s\n", summary.NextPayoutDate)
			}
		} else {
			fmt.Printf("   %s\n", color.YellowString("⚠️  No payout method configured"))
			fmt.Printf("   Run 'apidirect earnings setup' to configure\n")
		}
		
		// Top APIs
		if len(summary.TopAPIs) > 0 {
			fmt.Printf("\n🏆 Top Earning APIs:\n")
			for i, api := range summary.TopAPIs {
				if i >= 5 {
					break
				}
				fmt.Printf("   %d. %s: %s\n", i+1, api.APIName, color.GreenString("$%.2f", api.Earnings))
			}
		}
		
		// Revenue trend
		if len(summary.RevenueByMonth) > 0 {
			fmt.Printf("\n📈 Recent Months:\n")
			for _, month := range summary.RevenueByMonth {
				fmt.Printf("   %s: %s\n", month.Month, color.GreenString("$%.2f", month.Earnings))
			}
		}
		
		return nil
	}
}

func runEarningsDetails(cmd *cobra.Command, args []string) error {
	apiName := ""
	if len(args) > 0 {
		apiName = args[0]
	}
	
	// Parse period
	start, end, err := parsePeriod(earningsPeriod)
	if err != nil {
		return fmt.Errorf("invalid period: %w", err)
	}
	
	// Build URL
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/earnings/details?start=%s&end=%s&group_by=%s",
		cfg.APIEndpoint, start.Format("2006-01-02"), end.Format("2006-01-02"), earningsGroupBy)
	
	if apiName != "" || earningsAPI != "" {
		api := apiName
		if api == "" {
			api = earningsAPI
		}
		url += fmt.Sprintf("&api=%s", api)
	}
	
	if earningsDetailed {
		url += "&detailed=true"
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var details struct {
		Period struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"period"`
		TotalEarnings float64 `json:"total_earnings"`
		Breakdown []struct {
			Group       string  `json:"group"`        // API name or date
			Earnings    float64 `json:"earnings"`
			Subscribers int     `json:"subscribers"`
			Usage       int64   `json:"usage"`        // API calls
			Plans []struct {
				PlanName    string  `json:"plan_name"`
				Earnings    float64 `json:"earnings"`
				Subscribers int     `json:"subscribers"`
			} `json:"plans,omitempty"`
			Transactions []struct {
				Date        string  `json:"date"`
				Type        string  `json:"type"`
				Description string  `json:"description"`
				Amount      float64 `json:"amount"`
				Fee         float64 `json:"fee"`
				Net         float64 `json:"net"`
			} `json:"transactions,omitempty"`
		} `json:"breakdown"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return err
	}
	
	// Output based on format
	switch earningsFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(details)
		
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		
		// Write headers based on grouping
		if earningsDetailed {
			w.Write([]string{"Group", "Date", "Type", "Description", "Amount", "Fee", "Net"})
			for _, group := range details.Breakdown {
				for _, tx := range group.Transactions {
					w.Write([]string{
						group.Group,
						tx.Date,
						tx.Type,
						tx.Description,
						fmt.Sprintf("%.2f", tx.Amount),
						fmt.Sprintf("%.2f", tx.Fee),
						fmt.Sprintf("%.2f", tx.Net),
					})
				}
			}
		} else {
			w.Write([]string{"Group", "Earnings", "Subscribers", "API Calls"})
			for _, group := range details.Breakdown {
				w.Write([]string{
					group.Group,
					fmt.Sprintf("%.2f", group.Earnings),
					strconv.Itoa(group.Subscribers),
					strconv.FormatInt(group.Usage, 10),
				})
			}
		}
		
		return nil
		
	default:
		// Table format
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("💰 Earnings Details\n")
		fmt.Printf("Period: %s to %s\n", details.Period.Start, details.Period.End)
		fmt.Printf("Total Earnings: %s\n\n", color.GreenString("$%.2f", details.TotalEarnings))
		
		if earningsDetailed && len(details.Breakdown) > 0 && len(details.Breakdown[0].Transactions) > 0 {
			// Transaction view
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "DATE\tTYPE\tDESCRIPTION\tAMOUNT\tFEE\tNET\n")
			
			for _, group := range details.Breakdown {
				if len(group.Transactions) > 0 {
					fmt.Fprintf(w, "\n%s\n", color.New(color.Bold).Sprint(group.Group))
					for _, tx := range group.Transactions {
						fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t$%.2f\t%s\n",
							tx.Date,
							tx.Type,
							tx.Description,
							tx.Amount,
							tx.Fee,
							color.GreenString("$%.2f", tx.Net),
						)
					}
				}
			}
			w.Flush()
		} else {
			// Summary view
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			
			switch earningsGroupBy {
			case "daily", "weekly", "monthly":
				fmt.Fprintf(w, "PERIOD\tEARNINGS\tSUBSCRIBERS\tAPI CALLS\n")
			default:
				fmt.Fprintf(w, "API\tEARNINGS\tSUBSCRIBERS\tAPI CALLS\n")
			}
			
			for _, group := range details.Breakdown {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
					group.Group,
					color.GreenString("$%.2f", group.Earnings),
					group.Subscribers,
					formatNumber(group.Usage),
				)
				
				// Show plan breakdown if available
				if len(group.Plans) > 0 {
					for _, plan := range group.Plans {
						fmt.Fprintf(w, "  └─ %s\t%s\t%d\t-\n",
							plan.PlanName,
							color.GreenString("$%.2f", plan.Earnings),
							plan.Subscribers,
						)
					}
				}
			}
			w.Flush()
		}
		
		return nil
	}
}

func runEarningsPayout(cmd *cobra.Command, args []string) error {
	amount, _ := cmd.Flags().GetFloat64("amount")
	
	// Get current balance first
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// Get summary to check balance
	summaryURL := fmt.Sprintf("%s/api/v1/earnings/summary", cfg.APIEndpoint)
	resp, err := makeAuthenticatedRequest("GET", summaryURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var summary struct {
		AvailableBalance float64 `json:"available_balance"`
		PayoutMethod     string  `json:"payout_method"`
		MinimumPayout    float64 `json:"minimum_payout"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return err
	}
	
	// Check if payout method is configured
	if summary.PayoutMethod == "" {
		return fmt.Errorf("no payout method configured. Run 'apidirect earnings setup' first")
	}
	
	// Determine payout amount
	if amount == 0 {
		amount = summary.AvailableBalance
	}
	
	// Validate amount
	if amount > summary.AvailableBalance {
		return fmt.Errorf("requested amount ($%.2f) exceeds available balance ($%.2f)", 
			amount, summary.AvailableBalance)
	}
	
	if amount < summary.MinimumPayout {
		return fmt.Errorf("requested amount ($%.2f) is below minimum payout ($%.2f)", 
			amount, summary.MinimumPayout)
	}
	
	// Confirm payout
	fmt.Printf("\n💰 Payout Request\n")
	fmt.Printf("Amount: %s\n", color.GreenString("$%.2f", amount))
	fmt.Printf("Method: %s\n", summary.PayoutMethod)
	fmt.Printf("Available Balance: $%.2f\n", summary.AvailableBalance)
	fmt.Printf("Remaining Balance: $%.2f\n", summary.AvailableBalance-amount)
	
	if !confirmAction("\nRequest this payout?") {
		fmt.Println("Payout cancelled")
		return nil
	}
	
	// Request payout
	payoutData := struct {
		Amount float64 `json:"amount"`
	}{
		Amount: amount,
	}
	
	data, _ := json.Marshal(payoutData)
	payoutURL := fmt.Sprintf("%s/api/v1/earnings/payout", cfg.APIEndpoint)
	
	resp, err = makeAuthenticatedRequest("POST", payoutURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		PayoutID      string `json:"payout_id"`
		Amount        float64 `json:"amount"`
		Status        string `json:"status"`
		EstimatedDate string `json:"estimated_date"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Println()
	color.Green("✅ Payout requested successfully!")
	fmt.Printf("Payout ID: %s\n", result.PayoutID)
	fmt.Printf("Amount: %s\n", color.GreenString("$%.2f", result.Amount))
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Estimated Arrival: %s\n", result.EstimatedDate)
	
	return nil
}

func runEarningsHistory(cmd *cobra.Command, args []string) error {
	// Parse period
	start, end, err := parsePeriod(earningsPeriod)
	if err != nil {
		return fmt.Errorf("invalid period: %w", err)
	}
	
	// Call API
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/earnings/payouts?start=%s&end=%s",
		cfg.APIEndpoint, start.Format("2006-01-02"), end.Format("2006-01-02"))
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var history struct {
		Payouts []struct {
			ID            string  `json:"id"`
			Date          string  `json:"date"`
			Amount        float64 `json:"amount"`
			Fee           float64 `json:"fee"`
			Net           float64 `json:"net"`
			Status        string  `json:"status"`
			Method        string  `json:"method"`
			ArrivalDate   string  `json:"arrival_date"`
			Description   string  `json:"description"`
		} `json:"payouts"`
		Summary struct {
			TotalPayouts int     `json:"total_payouts"`
			TotalAmount  float64 `json:"total_amount"`
			TotalFees    float64 `json:"total_fees"`
			TotalNet     float64 `json:"total_net"`
		} `json:"summary"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return err
	}
	
	// Output based on format
	switch earningsFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(history)
		
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		
		// Write headers
		w.Write([]string{"ID", "Date", "Amount", "Fee", "Net", "Status", "Method", "Arrival Date", "Description"})
		
		for _, payout := range history.Payouts {
			w.Write([]string{
				payout.ID,
				payout.Date,
				fmt.Sprintf("%.2f", payout.Amount),
				fmt.Sprintf("%.2f", payout.Fee),
				fmt.Sprintf("%.2f", payout.Net),
				payout.Status,
				payout.Method,
				payout.ArrivalDate,
				payout.Description,
			})
		}
		
		return nil
		
	default:
		// Table format
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("💰 Payout History\n")
		fmt.Printf("Period: %s to %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
		
		if len(history.Payouts) == 0 {
			fmt.Println("No payouts found for this period")
			return nil
		}
		
		// Summary
		fmt.Printf("📊 Summary:\n")
		fmt.Printf("   Total Payouts: %d\n", history.Summary.TotalPayouts)
		fmt.Printf("   Total Amount: %s\n", color.GreenString("$%.2f", history.Summary.TotalAmount))
		fmt.Printf("   Total Fees: $%.2f\n", history.Summary.TotalFees)
		fmt.Printf("   Total Net: %s\n\n", color.GreenString("$%.2f", history.Summary.TotalNet))
		
		// Payout list
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "DATE\tAMOUNT\tFEE\tNET\tSTATUS\tMETHOD\n")
		
		for _, payout := range history.Payouts {
			statusColor := color.New(color.FgGreen)
			if payout.Status == "pending" {
				statusColor = color.New(color.FgYellow)
			} else if payout.Status == "failed" {
				statusColor = color.New(color.FgRed)
			}
			
			fmt.Fprintf(w, "%s\t$%.2f\t$%.2f\t%s\t%s\t%s\n",
				payout.Date,
				payout.Amount,
				payout.Fee,
				color.GreenString("$%.2f", payout.Net),
				statusColor.Sprint(payout.Status),
				payout.Method,
			)
		}
		w.Flush()
		
		return nil
	}
}

func runEarningsSetup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("💳 Payout Setup\n\n")
	
	// Check current status
	statusURL := fmt.Sprintf("%s/api/v1/earnings/payout-status", cfg.APIEndpoint)
	resp, err := makeAuthenticatedRequest("GET", statusURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var status struct {
		HasAccount       bool   `json:"has_account"`
		AccountStatus    string `json:"account_status"`
		RequiresAction   bool   `json:"requires_action"`
		OnboardingURL    string `json:"onboarding_url"`
		DashboardURL     string `json:"dashboard_url"`
	}
	
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return err
		}
		
		if status.HasAccount && status.AccountStatus == "active" {
			fmt.Println("✅ Your payout method is already configured and active!")
			fmt.Println()
			fmt.Println("You can manage your payout settings at:")
			fmt.Printf("%s\n", color.BlueString(status.DashboardURL))
			return nil
		}
		
		if status.HasAccount && status.RequiresAction {
			fmt.Println("⚠️  Your payout account requires additional information")
			fmt.Println()
			fmt.Println("Please complete setup at:")
			fmt.Printf("%s\n", color.BlueString(status.OnboardingURL))
			return nil
		}
	}
	
	// Start new setup
	fmt.Println("This will set up your Stripe Connect account for receiving payouts.")
	fmt.Println("You'll be redirected to Stripe to complete the onboarding process.")
	fmt.Println()
	
	if !confirmAction("Continue with payout setup?") {
		fmt.Println("Setup cancelled")
		return nil
	}
	
	// Create onboarding session
	onboardingURL := fmt.Sprintf("%s/api/v1/earnings/setup", cfg.APIEndpoint)
	resp, err = makeAuthenticatedRequest("POST", onboardingURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		OnboardingURL string `json:"onboarding_url"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Println()
	fmt.Println("📋 Next Steps:")
	fmt.Println("1. Open the following URL in your browser:")
	fmt.Printf("   %s\n", color.BlueString(result.OnboardingURL))
	fmt.Println("2. Complete the Stripe onboarding process")
	fmt.Println("3. Return here and run 'apidirect earnings summary' to verify setup")
	
	// Try to open browser
	if err := openBrowser(result.OnboardingURL); err == nil {
		fmt.Println()
		fmt.Println("✅ Browser opened automatically")
	}
	
	return nil
}

// Helper functions
func parsePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now()
	
	if period == "" {
		// Default to current month
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return start, end, nil
	}
	
	// Check for relative periods (7d, 30d, etc)
	if strings.HasSuffix(period, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(period, "d"))
		if err == nil {
			end := now
			start := now.AddDate(0, 0, -days)
			return start, end, nil
		}
	}
	
	// Check for quarters (2024-Q1, etc)
	if strings.Contains(period, "-Q") {
		parts := strings.Split(period, "-Q")
		if len(parts) == 2 {
			year, err := strconv.Atoi(parts[0])
			if err == nil {
				quarter, err := strconv.Atoi(parts[1])
				if err == nil && quarter >= 1 && quarter <= 4 {
					startMonth := (quarter-1)*3 + 1
					start := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
					end := start.AddDate(0, 3, 0).Add(-time.Second)
					return start, end, nil
				}
			}
		}
	}
	
	// Check for year (2024)
	if len(period) == 4 {
		year, err := strconv.Atoi(period)
		if err == nil {
			start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(1, 0, 0).Add(-time.Second)
			return start, end, nil
		}
	}
	
	// Check for year-month (2024-01)
	if len(period) == 7 && period[4] == '-' {
		t, err := time.Parse("2006-01", period)
		if err == nil {
			start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, 0).Add(-time.Second)
			return start, end, nil
		}
	}
	
	return time.Time{}, time.Time{}, fmt.Errorf("unsupported period format: %s", period)
}

// formatNumber moved to utils.go

// openBrowser moved to utils.go