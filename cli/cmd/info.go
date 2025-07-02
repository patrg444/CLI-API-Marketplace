package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	infoFormat   string
	infoDetailed bool
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [api-name]",
	Short: "View detailed information about an API",
	Long: `View comprehensive information about an API including its description,
pricing plans, endpoints, authentication methods, and usage statistics.

Examples:
  apidirect info weather-api
  apidirect info payment-gateway --detailed
  apidirect info my-api --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	
	infoCmd.Flags().StringVarP(&infoFormat, "format", "f", "table", "Output format (table, json)")
	infoCmd.Flags().BoolVarP(&infoDetailed, "detailed", "d", false, "Show detailed information including all endpoints")
}

func runInfo(cmd *cobra.Command, args []string) error {
	apiName := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// Build URL
	url := fmt.Sprintf("%s/api/v1/marketplace/apis/%s", cfg.APIEndpoint, apiName)
	if infoDetailed {
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
	
	var apiInfo struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		DisplayName     string    `json:"display_name"`
		Description     string    `json:"description"`
		LongDescription string    `json:"long_description"`
		Category        string    `json:"category"`
		Tags            []string  `json:"tags"`
		Version         string    `json:"version"`
		Status          string    `json:"status"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		
		Creator struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Username     string `json:"username"`
			Verified     bool   `json:"verified"`
			JoinedDate   string `json:"joined_date"`
			TotalAPIs    int    `json:"total_apis"`
			AverageRating float64 `json:"average_rating"`
		} `json:"creator"`
		
		Metrics struct {
			Subscribers      int     `json:"subscriber_count"`
			MonthlyCallsAvg  int64   `json:"monthly_calls_avg"`
			AverageRating    float64 `json:"average_rating"`
			TotalReviews     int     `json:"total_reviews"`
			ResponseTimeAvg  int     `json:"response_time_avg_ms"`
			Uptime           float64 `json:"uptime_percentage"`
			LastDowntime     *string `json:"last_downtime"`
		} `json:"metrics"`
		
		Pricing struct {
			Model string `json:"model"`
			Plans []struct {
				ID           string   `json:"id"`
				Name         string   `json:"name"`
				Description  string   `json:"description"`
				Price        float64  `json:"price"`
				Currency     string   `json:"currency"`
				Interval     string   `json:"interval"`
				Features     []string `json:"features"`
				Limits       struct {
					RequestsPerMonth *int    `json:"requests_per_month,omitempty"`
					RequestsPerSecond *int   `json:"requests_per_second,omitempty"`
					DataTransferGB   *int    `json:"data_transfer_gb,omitempty"`
				} `json:"limits"`
				Popular bool `json:"popular"`
			} `json:"plans"`
			CustomPricing bool   `json:"custom_pricing"`
			ContactSales  string `json:"contact_sales,omitempty"`
		} `json:"pricing"`
		
		Technical struct {
			BaseURL         string   `json:"base_url"`
			Authentication  []string `json:"authentication"`
			Formats         []string `json:"formats"`
			SDKs            []string `json:"sdks"`
			OpenAPISpec     string   `json:"openapi_spec_url"`
			PostmanURL      string   `json:"postman_collection_url"`
			DocumentationURL string  `json:"documentation_url"`
			SupportEmail    string   `json:"support_email"`
			SLA             string   `json:"sla"`
		} `json:"technical"`
		
		Endpoints []struct {
			Path        string   `json:"path"`
			Method      string   `json:"method"`
			Description string   `json:"description"`
			Category    string   `json:"category"`
			AuthRequired bool    `json:"auth_required"`
			RateLimit   string   `json:"rate_limit"`
			Examples    []struct {
				Language string `json:"language"`
				Code     string `json:"code"`
			} `json:"examples,omitempty"`
		} `json:"endpoints,omitempty"`
		
		RecentReviews []struct {
			Rating       int       `json:"rating"`
			Title        string    `json:"title"`
			Message      string    `json:"message"`
			AuthorName   string    `json:"author_name"`
			CreatedAt    time.Time `json:"created_at"`
			Verified     bool      `json:"verified_purchase"`
		} `json:"recent_reviews,omitempty"`
		
		Changelog []struct {
			Version     string    `json:"version"`
			Date        time.Time `json:"date"`
			Description string    `json:"description"`
			Breaking    bool      `json:"breaking_change"`
		} `json:"changelog,omitempty"`
		
		SimilarAPIs []struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Rating      float64 `json:"rating"`
			Subscribers int     `json:"subscribers"`
		} `json:"similar_apis,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&apiInfo); err != nil {
		return err
	}
	
	// Output based on format
	switch infoFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(apiInfo)
		
	default:
		// Table format with rich display
		fmt.Fprintln(cmd.OutOrStdout())
		
		// Header with name and status
		statusColor := color.FgGreen
		if apiInfo.Status != "active" {
			statusColor = color.FgYellow
		}
		
		color.New(color.FgCyan, color.Bold).Fprintf(cmd.OutOrStdout(), "%s", apiInfo.DisplayName)
		fmt.Fprintf(cmd.OutOrStdout(), " ")
		color.New(statusColor).Fprintf(cmd.OutOrStdout(), "[%s]", strings.ToUpper(apiInfo.Status))
		if apiInfo.Creator.Verified {
			color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), " ✓ Verified")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
		
		// Basic info
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", apiInfo.Description)
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
		
		// Category and tags
		fmt.Fprintf(cmd.OutOrStdout(), "📁 Category: %s\n", apiInfo.Category)
		if len(apiInfo.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "🏷️  Tags: %s\n", strings.Join(apiInfo.Tags, ", "))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "🔖 Version: %s\n", apiInfo.Version)
		
		// Creator info
		fmt.Fprintf(cmd.OutOrStdout(), "\n👤 Creator\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   Name: %s (@%s)", apiInfo.Creator.Name, apiInfo.Creator.Username)
		if apiInfo.Creator.Verified {
			fmt.Fprintf(cmd.OutOrStdout(), " ✓")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   Member since: %s\n", apiInfo.Creator.JoinedDate)
		fmt.Fprintf(cmd.OutOrStdout(), "   Total APIs: %d\n", apiInfo.Creator.TotalAPIs)
		if apiInfo.Creator.AverageRating > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Average rating: %.1f★\n", apiInfo.Creator.AverageRating)
		}
		
		// Metrics
		fmt.Fprintf(cmd.OutOrStdout(), "\n📊 Metrics\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   Subscribers: %s\n", formatNumber(int64(apiInfo.Metrics.Subscribers)))
		if apiInfo.Metrics.MonthlyCallsAvg > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Monthly calls (avg): %s\n", formatNumber(apiInfo.Metrics.MonthlyCallsAvg))
		}
		if apiInfo.Metrics.TotalReviews > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Rating: %.1f★ (%d reviews)\n", 
				apiInfo.Metrics.AverageRating, apiInfo.Metrics.TotalReviews)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "   Avg response time: %dms\n", apiInfo.Metrics.ResponseTimeAvg)
		fmt.Fprintf(cmd.OutOrStdout(), "   Uptime: %.2f%%\n", apiInfo.Metrics.Uptime)
		
		// Pricing
		fmt.Fprintf(cmd.OutOrStdout(), "\n💰 Pricing (%s)\n", apiInfo.Pricing.Model)
		if len(apiInfo.Pricing.Plans) > 0 {
			for _, plan := range apiInfo.Pricing.Plans {
				prefix := "   "
				if plan.Popular {
					prefix = "   ⭐ "
				}
				
				if plan.Price == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "%s%s - Free\n", prefix, plan.Name)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s%s - %s%.2f/%s\n", 
						prefix, plan.Name, 
						getCurrencySymbol(plan.Currency), plan.Price, plan.Interval)
				}
				
				if plan.Description != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "      %s\n", plan.Description)
				}
				
				// Show key features
				for i, feature := range plan.Features {
					if i < 3 { // Show first 3 features
						fmt.Fprintf(cmd.OutOrStdout(), "      • %s\n", feature)
					}
				}
				
				// Show limits
				if plan.Limits.RequestsPerMonth != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "      • %s API calls/month\n", 
						formatNumber(int64(*plan.Limits.RequestsPerMonth)))
				}
			}
		}
		
		if apiInfo.Pricing.CustomPricing {
			fmt.Fprintf(cmd.OutOrStdout(), "   📞 Custom pricing available - %s\n", apiInfo.Pricing.ContactSales)
		}
		
		// Technical details
		fmt.Fprintf(cmd.OutOrStdout(), "\n🔧 Technical Details\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   Base URL: %s\n", color.BlueString(apiInfo.Technical.BaseURL))
		fmt.Fprintf(cmd.OutOrStdout(), "   Authentication: %s\n", strings.Join(apiInfo.Technical.Authentication, ", "))
		fmt.Fprintf(cmd.OutOrStdout(), "   Formats: %s\n", strings.Join(apiInfo.Technical.Formats, ", "))
		if len(apiInfo.Technical.SDKs) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   SDKs: %s\n", strings.Join(apiInfo.Technical.SDKs, ", "))
		}
		
		// Documentation links
		fmt.Fprintf(cmd.OutOrStdout(), "\n📚 Resources\n")
		if apiInfo.Technical.DocumentationURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   Documentation: %s\n", color.BlueString(apiInfo.Technical.DocumentationURL))
		}
		if apiInfo.Technical.OpenAPISpec != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   OpenAPI Spec: %s\n", color.BlueString(apiInfo.Technical.OpenAPISpec))
		}
		if apiInfo.Technical.PostmanURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   Postman Collection: %s\n", color.BlueString(apiInfo.Technical.PostmanURL))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "   Support: %s\n", apiInfo.Technical.SupportEmail)
		if apiInfo.Technical.SLA != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   SLA: %s\n", apiInfo.Technical.SLA)
		}
		
		// Endpoints (if detailed)
		if infoDetailed && len(apiInfo.Endpoints) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n🔗 Endpoints (%d)\n", len(apiInfo.Endpoints))
			
			// Group by category
			endpointsByCategory := make(map[string][]struct {
				Path        string   `json:"path"`
				Method      string   `json:"method"`
				Description string   `json:"description"`
				Category    string   `json:"category"`
				AuthRequired bool    `json:"auth_required"`
				RateLimit   string   `json:"rate_limit"`
				Examples    []struct {
					Language string `json:"language"`
					Code     string `json:"code"`
				} `json:"examples,omitempty"`
			})
			
			for _, ep := range apiInfo.Endpoints {
				category := ep.Category
				if category == "" {
					category = "General"
				}
				endpointsByCategory[category] = append(endpointsByCategory[category], ep)
			}
			
			for category, endpoints := range endpointsByCategory {
				fmt.Fprintf(cmd.OutOrStdout(), "\n   %s:\n", category)
				for _, ep := range endpoints {
					auth := ""
					if ep.AuthRequired {
						auth = " 🔒"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "   %s %s%s\n", 
						colorMethod(ep.Method), ep.Path, auth)
					fmt.Fprintf(cmd.OutOrStdout(), "      %s\n", ep.Description)
					if ep.RateLimit != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "      Rate limit: %s\n", ep.RateLimit)
					}
				}
			}
		}
		
		// Recent reviews
		if len(apiInfo.RecentReviews) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n⭐ Recent Reviews\n")
			for i, review := range apiInfo.RecentReviews {
				if i >= 3 { // Show only first 3
					break
				}
				verified := ""
				if review.Verified {
					verified = " ✓"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "   %s %s\n", getStarRatingSimple(float64(review.Rating)), review.Title)
				fmt.Fprintf(cmd.OutOrStdout(), "   \"%s\"\n", truncate(review.Message, 80))
				fmt.Fprintf(cmd.OutOrStdout(), "   — %s%s • %s\n", 
					review.AuthorName, verified, 
					review.CreatedAt.Format("Jan 2, 2006"))
				if i < len(apiInfo.RecentReviews)-1 && i < 2 {
					fmt.Fprintln(cmd.OutOrStdout(), )
				}
			}
		}
		
		// Recent changelog
		if len(apiInfo.Changelog) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n📝 Recent Changes\n")
			for i, change := range apiInfo.Changelog {
				if i >= 3 { // Show only recent 3
					break
				}
				breaking := ""
				if change.Breaking {
					breaking = " ⚠️  BREAKING"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "   v%s (%s)%s\n", 
					change.Version, 
					change.Date.Format("Jan 2, 2006"),
					breaking)
				fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", change.Description)
				if i < len(apiInfo.Changelog)-1 && i < 2 {
					fmt.Fprintln(cmd.OutOrStdout(), )
				}
			}
		}
		
		// Similar APIs
		if len(apiInfo.SimilarAPIs) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n🔍 Similar APIs\n")
			for _, similar := range apiInfo.SimilarAPIs {
				fmt.Fprintf(cmd.OutOrStdout(), "   • %s - %s", similar.Name, truncate(similar.Description, 50))
				if similar.Rating > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), " (%.1f★)", similar.Rating)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\n")
			}
		}
		
		// Actions
		fmt.Fprintf(cmd.OutOrStdout(), "\n💡 Actions:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "   Subscribe: apidirect subscribe %s\n", apiInfo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "   View reviews: apidirect review list %s\n", apiInfo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "   Try it out: %s\n", apiInfo.Technical.DocumentationURL)
		
		return nil
	}
}

// Helper function to get star rating (without half stars for info display)
func getStarRatingSimple(rating float64) string {
	full := int(rating)
	empty := 5 - full
	
	stars := strings.Repeat("★", full)
	stars += strings.Repeat("☆", empty)
	
	return stars
}

// Helper function to color HTTP methods
func colorMethod(method string) string {
	switch method {
	case "GET":
		return color.GreenString(method)
	case "POST":
		return color.BlueString(method)
	case "PUT":
		return color.YellowString(method)
	case "DELETE":
		return color.RedString(method)
	case "PATCH":
		return color.MagentaString(method)
	default:
		return method
	}
}