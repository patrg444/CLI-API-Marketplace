package cmd

import (
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
	reviewRating  int
	reviewFormat  string
	reviewFilter  string
	reviewSort    string
)

// reviewCmd represents the review command group
var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Manage API reviews and ratings",
	Long: `Submit, view, and manage reviews for APIs in the marketplace.

This command helps you:
- Submit reviews for APIs you've used
- View reviews before subscribing
- Respond to reviews on your APIs (creators)
- Report inappropriate reviews`,
}

// Review subcommands
var reviewSubmitCmd = &cobra.Command{
	Use:   "submit [api-name]",
	Short: "Submit a review for an API",
	Long: `Submit a review and rating for an API you've subscribed to.

Examples:
  apidirect review submit weather-api --rating 5
  apidirect review submit payment-api -r 4 -t "Great API!"`,
	Args: cobra.ExactArgs(1),
	RunE: runReviewSubmit,
}

var reviewListCmd = &cobra.Command{
	Use:   "list [api-name]",
	Short: "List reviews for an API",
	Long: `View reviews and ratings for a specific API.

Examples:
  apidirect review list weather-api               # All reviews
  apidirect review list weather-api --filter 5    # Only 5-star reviews
  apidirect review list weather-api --sort newest # Sort by date`,
	Args: cobra.ExactArgs(1),
	RunE: runReviewList,
}

var reviewMyCmd = &cobra.Command{
	Use:   "my",
	Short: "View your reviews",
	Long: `View all reviews you've submitted and their current status.

Examples:
  apidirect review my                    # All your reviews
  apidirect review my --format json      # Export as JSON`,
	RunE: runReviewMy,
}

var reviewResponseCmd = &cobra.Command{
	Use:   "respond [review-id]",
	Short: "Respond to a review (creators only)",
	Long: `Respond to a review on one of your APIs.

Examples:
  apidirect review respond rev_123abc -m "Thanks for the feedback!"`,
	Args: cobra.ExactArgs(1),
	RunE: runReviewResponse,
}

var reviewReportCmd = &cobra.Command{
	Use:   "report [review-id]",
	Short: "Report an inappropriate review",
	Long: `Report a review that violates community guidelines.

Examples:
  apidirect review report rev_123abc -r "Spam content"`,
	Args: cobra.ExactArgs(1),
	RunE: runReviewReport,
}

var reviewStatsCmd = &cobra.Command{
	Use:   "stats [api-name]",
	Short: "View review statistics for an API",
	Long: `View detailed review statistics and trends for an API.

Examples:
  apidirect review stats weather-api     # Review statistics
  apidirect review stats --all           # Stats for all your APIs (creators)`,
	RunE: runReviewStats,
}

func init() {
	rootCmd.AddCommand(reviewCmd)
	
	// Add subcommands
	reviewCmd.AddCommand(reviewSubmitCmd)
	reviewCmd.AddCommand(reviewListCmd)
	reviewCmd.AddCommand(reviewMyCmd)
	reviewCmd.AddCommand(reviewResponseCmd)
	reviewCmd.AddCommand(reviewReportCmd)
	reviewCmd.AddCommand(reviewStatsCmd)
	
	// Submit flags
	reviewSubmitCmd.Flags().IntVarP(&reviewRating, "rating", "r", 0, "Rating (1-5 stars)")
	reviewSubmitCmd.Flags().StringP("title", "t", "", "Review title")
	reviewSubmitCmd.Flags().StringP("message", "m", "", "Review message")
	reviewSubmitCmd.MarkFlagRequired("rating")
	
	// List flags
	reviewListCmd.Flags().StringVarP(&reviewFilter, "filter", "f", "", "Filter reviews (rating, verified)")
	reviewListCmd.Flags().StringVarP(&reviewSort, "sort", "s", "helpful", "Sort order (helpful, newest, oldest, rating)")
	reviewListCmd.Flags().StringVar(&reviewFormat, "format", "table", "Output format (table, json)")
	reviewListCmd.Flags().IntP("limit", "l", 20, "Number of reviews to show")
	
	// My reviews flags
	reviewMyCmd.Flags().StringVar(&reviewFormat, "format", "table", "Output format (table, json)")
	
	// Response flags
	reviewResponseCmd.Flags().StringP("message", "m", "", "Response message")
	reviewResponseCmd.MarkFlagRequired("message")
	
	// Report flags
	reviewReportCmd.Flags().StringP("reason", "r", "", "Reason for reporting")
	reviewReportCmd.MarkFlagRequired("reason")
	
	// Stats flags
	reviewStatsCmd.Flags().Bool("all", false, "Show stats for all your APIs (creators only)")
	reviewStatsCmd.Flags().StringVar(&reviewFormat, "format", "table", "Output format (table, json)")
}

func runReviewSubmit(cmd *cobra.Command, args []string) error {
	apiName := args[0]
	title, _ := cmd.Flags().GetString("title")
	message, _ := cmd.Flags().GetString("message")
	
	// Validate rating
	if reviewRating < 1 || reviewRating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	
	// Interactive mode if no message provided
	if message == "" {
		fmt.Printf("\n⭐ Submit Review for %s\n", color.CyanString(apiName))
		fmt.Printf("Rating: %s\n", strings.Repeat("★", reviewRating)+strings.Repeat("☆", 5-reviewRating))
		
		if title == "" {
			fmt.Print("\nTitle (optional): ")
			fmt.Scanln(&title)
		}
		
		fmt.Println("\nWrite your review (press Enter twice to finish):")
		var lines []string
		for {
			var line string
			fmt.Scanln(&line)
			if line == "" && len(lines) > 0 {
				break
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		message = strings.Join(lines, "\n")
		
		if message == "" {
			return fmt.Errorf("review message is required")
		}
	}
	
	// Submit review
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	reviewData := struct {
		APIName string `json:"api_name"`
		Rating  int    `json:"rating"`
		Title   string `json:"title,omitempty"`
		Message string `json:"message"`
	}{
		APIName: apiName,
		Rating:  reviewRating,
		Title:   title,
		Message: message,
	}
	
	data, _ := json.Marshal(reviewData)
	url := fmt.Sprintf("%s/api/v1/reviews", cfg.APIEndpoint)
	
	resp, err := makeAuthenticatedRequest("POST", url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		ReviewID  string    `json:"review_id"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Println()
	color.Green("✅ Review submitted successfully!")
	fmt.Printf("Review ID: %s\n", result.ReviewID)
	fmt.Printf("Status: %s\n", result.Status)
	
	return nil
}

func runReviewList(cmd *cobra.Command, args []string) error {
	apiName := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// Build URL with query parameters
	url := fmt.Sprintf("%s/api/v1/reviews/%s?sort=%s&limit=%d", 
		cfg.APIEndpoint, apiName, reviewSort, limit)
	
	if reviewFilter != "" {
		url += fmt.Sprintf("&filter=%s", reviewFilter)
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var reviewsData struct {
		API struct {
			Name          string  `json:"name"`
			AverageRating float64 `json:"average_rating"`
			TotalReviews  int     `json:"total_reviews"`
			RatingCounts  map[string]int `json:"rating_counts"`
		} `json:"api"`
		Reviews []struct {
			ID           string    `json:"id"`
			Rating       int       `json:"rating"`
			Title        string    `json:"title"`
			Message      string    `json:"message"`
			AuthorName   string    `json:"author_name"`
			AuthorID     string    `json:"author_id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Verified     bool      `json:"verified_purchase"`
			Helpful      int       `json:"helpful_count"`
			NotHelpful   int       `json:"not_helpful_count"`
			Response     *struct {
				Message   string    `json:"message"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"creator_response,omitempty"`
		} `json:"reviews"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&reviewsData); err != nil {
		return err
	}
	
	// Output based on format
	switch reviewFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(reviewsData)
		
	default:
		// Table format
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("⭐ Reviews for %s\n", reviewsData.API.Name)
		
		// API summary
		fmt.Printf("\n📊 Summary\n")
		fmt.Printf("Average Rating: %s %.1f (%d reviews)\n", 
			getStarRating(reviewsData.API.AverageRating),
			reviewsData.API.AverageRating,
			reviewsData.API.TotalReviews)
		
		// Rating distribution
		fmt.Printf("\nRating Distribution:\n")
		for i := 5; i >= 1; i-- {
			count := reviewsData.API.RatingCounts[strconv.Itoa(i)]
			percentage := float64(count) / float64(reviewsData.API.TotalReviews) * 100
			bar := strings.Repeat("█", int(percentage/5))
			fmt.Printf("%d★ %s %d (%.0f%%)\n", i, bar, count, percentage)
		}
		
		// Reviews
		if len(reviewsData.Reviews) > 0 {
			fmt.Printf("\n📝 Reviews\n")
			for _, review := range reviewsData.Reviews {
				fmt.Println(strings.Repeat("-", 60))
				
				// Header
				fmt.Printf("%s ", getStarRating(float64(review.Rating)))
				if review.Title != "" {
					fmt.Printf("%s\n", color.New(color.Bold).Sprint(review.Title))
				} else {
					fmt.Println()
				}
				
				// Author and date
				verifiedBadge := ""
				if review.Verified {
					verifiedBadge = color.GreenString(" ✓ Verified")
				}
				fmt.Printf("by %s%s • %s\n", 
					review.AuthorName, 
					verifiedBadge,
					review.CreatedAt.Format("Jan 2, 2006"))
				
				// Review text
				fmt.Printf("\n%s\n", review.Message)
				
				// Creator response
				if review.Response != nil {
					fmt.Printf("\n  %s Creator Response:\n", color.BlueString("↳"))
					fmt.Printf("  %s\n", review.Response.Message)
					fmt.Printf("  %s\n", color.New(color.FgHiBlack).Sprintf(
						"— %s", review.Response.CreatedAt.Format("Jan 2, 2006")))
				}
				
				// Helpful counts
				if review.Helpful > 0 || review.NotHelpful > 0 {
					fmt.Printf("\n%s %d helpful • %d not helpful\n",
						color.New(color.FgHiBlack).Sprint("👍"),
						review.Helpful,
						review.NotHelpful)
				}
			}
			fmt.Println(strings.Repeat("-", 60))
		} else {
			fmt.Println("\nNo reviews yet")
		}
		
		return nil
	}
}

func runReviewMy(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/reviews/my", cfg.APIEndpoint)
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var myReviews []struct {
		ID        string    `json:"id"`
		APIName   string    `json:"api_name"`
		APIID     string    `json:"api_id"`
		Rating    int       `json:"rating"`
		Title     string    `json:"title"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Status    string    `json:"status"`
		Helpful   int       `json:"helpful_count"`
		Response  *struct {
			Message   string    `json:"message"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"creator_response,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&myReviews); err != nil {
		return err
	}
	
	// Output based on format
	switch reviewFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(myReviews)
		
	default:
		// Table format
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("📝 My Reviews (%d)\n\n", len(myReviews))
		
		if len(myReviews) == 0 {
			fmt.Println("You haven't submitted any reviews yet")
			return nil
		}
		
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "API\tRATING\tTITLE\tDATE\tSTATUS\tHELPFUL\n")
		
		for _, review := range myReviews {
			title := review.Title
			if title == "" {
				title = truncate(review.Message, 30)
			}
			
			statusColor := color.FgGreen
			if review.Status != "published" {
				statusColor = color.FgYellow
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
				review.APIName,
				getStarRating(float64(review.Rating)),
				title,
				review.CreatedAt.Format("Jan 2"),
				color.New(statusColor).Sprint(review.Status),
				review.Helpful,
			)
		}
		w.Flush()
		
		// Show any with responses
		hasResponses := false
		for _, review := range myReviews {
			if review.Response != nil {
				if !hasResponses {
					fmt.Printf("\n💬 Creator Responses:\n")
					hasResponses = true
				}
				fmt.Printf("\n%s (%s):\n", review.APIName, getStarRating(float64(review.Rating)))
				fmt.Printf("Your review: %s\n", truncate(review.Message, 60))
				fmt.Printf("Response: %s\n", review.Response.Message)
			}
		}
		
		return nil
	}
}

func runReviewResponse(cmd *cobra.Command, args []string) error {
	reviewID := args[0]
	message, _ := cmd.Flags().GetString("message")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	responseData := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	
	data, _ := json.Marshal(responseData)
	url := fmt.Sprintf("%s/api/v1/reviews/%s/respond", cfg.APIEndpoint, reviewID)
	
	resp, err := makeAuthenticatedRequest("POST", url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Println()
	color.Green("✅ Response posted successfully!")
	
	return nil
}

func runReviewReport(cmd *cobra.Command, args []string) error {
	reviewID := args[0]
	reason, _ := cmd.Flags().GetString("reason")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	reportData := struct {
		Reason string `json:"reason"`
	}{
		Reason: reason,
	}
	
	data, _ := json.Marshal(reportData)
	url := fmt.Sprintf("%s/api/v1/reviews/%s/report", cfg.APIEndpoint, reviewID)
	
	resp, err := makeAuthenticatedRequest("POST", url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		ReportID string `json:"report_id"`
		Status   string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	fmt.Println()
	color.Green("✅ Review reported successfully")
	fmt.Printf("Report ID: %s\n", result.ReportID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Println("\nOur team will review this report within 24-48 hours.")
	
	return nil
}

func runReviewStats(cmd *cobra.Command, args []string) error {
	showAll, _ := cmd.Flags().GetBool("all")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	var url string
	if showAll {
		url = fmt.Sprintf("%s/api/v1/reviews/stats", cfg.APIEndpoint)
	} else if len(args) > 0 {
		url = fmt.Sprintf("%s/api/v1/reviews/stats/%s", cfg.APIEndpoint, args[0])
	} else {
		return fmt.Errorf("specify an API name or use --all flag")
	}
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	if showAll {
		// Multiple API stats
		var stats []struct {
			APIName       string  `json:"api_name"`
			APIID         string  `json:"api_id"`
			AverageRating float64 `json:"average_rating"`
			TotalReviews  int     `json:"total_reviews"`
			RatingCounts  map[string]int `json:"rating_counts"`
			RecentTrend   string  `json:"recent_trend"` // up, down, stable
			ResponseRate  float64 `json:"response_rate"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return err
		}
		
		switch reviewFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(stats)
			
		default:
			fmt.Println()
			color.New(color.FgCyan, color.Bold).Printf("📊 Review Statistics for Your APIs\n\n")
			
			if len(stats) == 0 {
				fmt.Println("No APIs with reviews found")
				return nil
			}
			
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "API\tAVG RATING\tREVIEWS\tTREND\tRESPONSE RATE\n")
			
			for _, api := range stats {
				trend := api.RecentTrend
				if trend == "up" {
					trend = color.GreenString("↑")
				} else if trend == "down" {
					trend = color.RedString("↓")
				} else {
					trend = "→"
				}
				
				fmt.Fprintf(w, "%s\t%.1f %s\t%d\t%s\t%.0f%%\n",
					api.APIName,
					api.AverageRating,
					getStarRating(api.AverageRating),
					api.TotalReviews,
					trend,
					api.ResponseRate*100,
				)
			}
			w.Flush()
		}
		
	} else {
		// Single API detailed stats
		var stats struct {
			APIName       string  `json:"api_name"`
			AverageRating float64 `json:"average_rating"`
			TotalReviews  int     `json:"total_reviews"`
			RatingCounts  map[string]int `json:"rating_counts"`
			
			Trends struct {
				Last30Days struct {
					AverageRating float64 `json:"average_rating"`
					ReviewCount   int     `json:"review_count"`
				} `json:"last_30_days"`
				Last90Days struct {
					AverageRating float64 `json:"average_rating"`
					ReviewCount   int     `json:"review_count"`
				} `json:"last_90_days"`
			} `json:"trends"`
			
			ResponseMetrics struct {
				TotalResponses   int     `json:"total_responses"`
				ResponseRate     float64 `json:"response_rate"`
				AvgResponseTime  string  `json:"avg_response_time"`
			} `json:"response_metrics"`
			
			Keywords []struct {
				Word  string `json:"word"`
				Count int    `json:"count"`
				Sentiment string `json:"sentiment"`
			} `json:"keywords"`
			
			RecentReviews []struct {
				Rating    int       `json:"rating"`
				Title     string    `json:"title"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"recent_reviews"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return err
		}
		
		switch reviewFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(stats)
			
		default:
			fmt.Println()
			color.New(color.FgCyan, color.Bold).Printf("📊 Review Statistics: %s\n\n", stats.APIName)
			
			// Overall stats
			fmt.Printf("⭐ Overall Rating: %.1f %s (%d reviews)\n",
				stats.AverageRating,
				getStarRating(stats.AverageRating),
				stats.TotalReviews)
			
			// Rating distribution
			fmt.Printf("\n📊 Rating Distribution:\n")
			for i := 5; i >= 1; i-- {
				count := stats.RatingCounts[strconv.Itoa(i)]
				percentage := float64(count) / float64(stats.TotalReviews) * 100
				bar := strings.Repeat("█", int(percentage/5))
				fmt.Printf("%d★ %s %d (%.0f%%)\n", i, bar, count, percentage)
			}
			
			// Trends
			fmt.Printf("\n📈 Trends:\n")
			fmt.Printf("Last 30 days: %.1f★ (%d reviews)\n", 
				stats.Trends.Last30Days.AverageRating,
				stats.Trends.Last30Days.ReviewCount)
			fmt.Printf("Last 90 days: %.1f★ (%d reviews)\n",
				stats.Trends.Last90Days.AverageRating,
				stats.Trends.Last90Days.ReviewCount)
			
			// Response metrics
			fmt.Printf("\n💬 Response Metrics:\n")
			fmt.Printf("Response rate: %.0f%% (%d/%d)\n",
				stats.ResponseMetrics.ResponseRate*100,
				stats.ResponseMetrics.TotalResponses,
				stats.TotalReviews)
			if stats.ResponseMetrics.AvgResponseTime != "" {
				fmt.Printf("Avg response time: %s\n", stats.ResponseMetrics.AvgResponseTime)
			}
			
			// Top keywords
			if len(stats.Keywords) > 0 {
				fmt.Printf("\n🏷️  Top Keywords:\n")
				for i, kw := range stats.Keywords {
					if i >= 5 {
						break
					}
					sentimentColor := color.FgWhite
					if kw.Sentiment == "positive" {
						sentimentColor = color.FgGreen
					} else if kw.Sentiment == "negative" {
						sentimentColor = color.FgRed
					}
					fmt.Printf("  • %s (%d)\n", 
						color.New(sentimentColor).Sprint(kw.Word), kw.Count)
				}
			}
			
			// Recent reviews
			if len(stats.RecentReviews) > 0 {
				fmt.Printf("\n🕐 Recent Reviews:\n")
				for _, review := range stats.RecentReviews {
					fmt.Printf("  %s %s - %s\n",
						getStarRating(float64(review.Rating)),
						review.Title,
						review.CreatedAt.Format("Jan 2"))
				}
			}
		}
	}
	
	return nil
}

// Helper functions
func getStarRating(rating float64) string {
	full := int(rating)
	half := rating - float64(full) >= 0.5
	empty := 5 - full
	if half {
		empty--
	}
	
	stars := strings.Repeat("★", full)
	if half {
		stars += "½"
	}
	stars += strings.Repeat("☆", empty)
	
	return color.YellowString(stars)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}