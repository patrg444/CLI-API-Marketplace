package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	searchCategory   string
	searchTags       []string
	searchSort       string
	searchPriceRange string
	searchFormat     string
	searchLimit      int
	searchOffset     int
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the API marketplace",
	Long: `Search for APIs in the marketplace using keywords, categories, and filters.

Examples:
  apidirect search weather                     # Search for weather APIs
  apidirect search --category data             # Browse data category
  apidirect search payment --tags stripe,paypal # Search with tags
  apidirect search --sort popular              # Sort by popularity
  apidirect search "machine learning" --limit 10`,
	RunE: runSearch,
}

// Browse subcommand
var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse API categories",
	Long: `Browse APIs by category in the marketplace.

Examples:
  apidirect browse                             # List all categories
  apidirect browse --category finance          # Browse finance APIs`,
	RunE: runBrowse,
}

// Trending subcommand
var trendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "View trending APIs",
	Long: `View trending and popular APIs in the marketplace.

Examples:
  apidirect trending                           # Top trending APIs
  apidirect trending --category data           # Trending in category
  apidirect trending --limit 20                # Show more results`,
	RunE: runTrending,
}

// Featured subcommand
var featuredCmd = &cobra.Command{
	Use:   "featured",
	Short: "View featured APIs",
	Long: `View hand-picked featured APIs from the marketplace.

Examples:
  apidirect featured                           # All featured APIs
  apidirect featured --format json             # Export as JSON`,
	RunE: runFeatured,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(trendingCmd)
	rootCmd.AddCommand(featuredCmd)
	
	// Search flags
	searchCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Filter by category")
	searchCmd.Flags().StringSliceVarP(&searchTags, "tags", "t", []string{}, "Filter by tags (comma-separated)")
	searchCmd.Flags().StringVarP(&searchSort, "sort", "s", "relevance", "Sort order (relevance, popular, newest, rating, price)")
	searchCmd.Flags().StringVarP(&searchPriceRange, "price", "p", "", "Price range (free, 0-10, 10-50, 50+)")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json, grid)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "Number of results")
	searchCmd.Flags().IntVar(&searchOffset, "offset", 0, "Offset for pagination")
	
	// Browse flags
	browseCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Browse specific category")
	browseCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json, grid)")
	
	// Trending flags
	trendingCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Filter by category")
	trendingCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Number of results")
	trendingCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json, grid)")
	
	// Featured flags
	featuredCmd.Flags().StringVarP(&searchFormat, "format", "f", "grid", "Output format (table, json, grid)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = strings.Join(args, " ")
	}
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	// Build search URL
	params := url.Values{}
	if query != "" {
		params.Add("q", query)
	}
	if searchCategory != "" {
		params.Add("category", searchCategory)
	}
	if len(searchTags) > 0 {
		params.Add("tags", strings.Join(searchTags, ","))
	}
	params.Add("sort", searchSort)
	if searchPriceRange != "" {
		params.Add("price_range", searchPriceRange)
	}
	params.Add("limit", fmt.Sprintf("%d", searchLimit))
	params.Add("offset", fmt.Sprintf("%d", searchOffset))
	
	searchURL := fmt.Sprintf("%s/api/v1/marketplace/search?%s", cfg.APIEndpoint, params.Encode())
	
	resp, err := makeAuthenticatedRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var results struct {
		Query   string `json:"query"`
		Total   int    `json:"total"`
		Offset  int    `json:"offset"`
		Limit   int    `json:"limit"`
		Results []struct {
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Category      string   `json:"category"`
			Tags          []string `json:"tags"`
			Creator       string   `json:"creator"`
			Rating        float64  `json:"rating"`
			ReviewCount   int      `json:"review_count"`
			Subscribers   int      `json:"subscriber_count"`
			PricingModel  string   `json:"pricing_model"`
			StartingPrice float64  `json:"starting_price"`
			Currency      string   `json:"currency"`
			Featured      bool     `json:"featured"`
			Verified      bool     `json:"verified"`
			Endpoints     int      `json:"endpoint_count"`
			LastUpdated   string   `json:"last_updated"`
		} `json:"results"`
		Facets struct {
			Categories map[string]int `json:"categories"`
			Tags       map[string]int `json:"tags"`
			PriceRanges map[string]int `json:"price_ranges"`
		} `json:"facets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return err
	}
	
	// Output based on format
	switch searchFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
		
	case "grid":
		return displaySearchGrid(results.Results, results.Total, query)
		
	default:
		return displaySearchTable(results.Results, results.Total, query, results.Facets)
	}
}

func runBrowse(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	if searchCategory == "" {
		// List categories
		url := fmt.Sprintf("%s/api/v1/marketplace/categories", cfg.APIEndpoint)
		resp, err := makeAuthenticatedRequest("GET", url, nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return handleErrorResponse(resp)
		}
		
		var categories []struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			APICount    int    `json:"api_count"`
			Icon        string `json:"icon"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
			return err
		}
		
		// Display categories
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("📚 API Categories\n\n")
		
		for _, cat := range categories {
			fmt.Printf("%s %s\n", cat.Icon, color.New(color.Bold).Sprint(cat.Name))
			fmt.Printf("   %s\n", cat.Description)
			fmt.Printf("   %s\n\n", color.New(color.FgHiBlack).Sprintf("%d APIs", cat.APICount))
		}
		
		fmt.Println("Browse a category: apidirect browse --category <name>")
		
		return nil
	}
	
	// Browse specific category
	return runSearch(cmd, []string{})
}

func runTrending(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	params := url.Values{}
	params.Add("period", "week")
	if searchCategory != "" {
		params.Add("category", searchCategory)
	}
	params.Add("limit", fmt.Sprintf("%d", searchLimit))
	
	trendingURL := fmt.Sprintf("%s/api/v1/marketplace/trending?%s", cfg.APIEndpoint, params.Encode())
	
	resp, err := makeAuthenticatedRequest("GET", trendingURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var trending struct {
		Period string `json:"period"`
		APIs []struct {
			Rank          int      `json:"rank"`
			RankChange    int      `json:"rank_change"`
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Category      string   `json:"category"`
			Tags          []string `json:"tags"`
			Creator       string   `json:"creator"`
			Rating        float64  `json:"rating"`
			ReviewCount   int      `json:"review_count"`
			Subscribers   int      `json:"subscriber_count"`
			GrowthRate    float64  `json:"growth_rate"`
			PricingModel  string   `json:"pricing_model"`
			StartingPrice float64  `json:"starting_price"`
		} `json:"apis"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&trending); err != nil {
		return err
	}
	
	// Output based on format
	switch searchFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(trending)
		
	default:
		// Display trending
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Printf("🔥 Trending APIs")
		if searchCategory != "" {
			fmt.Printf(" in %s", searchCategory)
		}
		fmt.Printf("\n\n")
		
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "RANK\tAPI\tCATEGORY\tRATING\tGROWTH\tPRICE\n")
		
		for _, api := range trending.APIs {
			// Rank change indicator
			rankIndicator := ""
			if api.RankChange > 0 {
				rankIndicator = color.GreenString("↑%d", api.RankChange)
			} else if api.RankChange < 0 {
				rankIndicator = color.RedString("↓%d", -api.RankChange)
			} else {
				rankIndicator = "→"
			}
			
			// Price display
			price := "Free"
			if api.StartingPrice > 0 {
				price = fmt.Sprintf("From $%.0f", api.StartingPrice)
			}
			
			fmt.Fprintf(w, "#%d %s\t%s\t%s\t%.1f★ (%d)\t+%.0f%%\t%s\n",
				api.Rank,
				rankIndicator,
				api.Name,
				api.Category,
				api.Rating,
				api.ReviewCount,
				api.GrowthRate*100,
				price,
			)
		}
		w.Flush()
		
		return nil
	}
}

func runFeatured(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/v1/marketplace/featured", cfg.APIEndpoint)
	
	resp, err := makeAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}
	
	var featured struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		APIs []struct {
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Category      string   `json:"category"`
			Tags          []string `json:"tags"`
			Creator       string   `json:"creator"`
			Rating        float64  `json:"rating"`
			ReviewCount   int      `json:"review_count"`
			FeaturedText  string   `json:"featured_text"`
			Badge         string   `json:"badge"`
			PricingModel  string   `json:"pricing_model"`
			StartingPrice float64  `json:"starting_price"`
		} `json:"apis"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&featured); err != nil {
		return err
	}
	
	// Output based on format
	switch searchFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(featured)
		
	case "table":
		return displayFeaturedTable(featured.APIs, featured.Title, featured.Description)
		
	default:
		return displayFeaturedGrid(featured.APIs, featured.Title, featured.Description)
	}
}

// Display helper functions
func displaySearchTable(results []struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	Creator       string   `json:"creator"`
	Rating        float64  `json:"rating"`
	ReviewCount   int      `json:"review_count"`
	Subscribers   int      `json:"subscriber_count"`
	PricingModel  string   `json:"pricing_model"`
	StartingPrice float64  `json:"starting_price"`
	Currency      string   `json:"currency"`
	Featured      bool     `json:"featured"`
	Verified      bool     `json:"verified"`
	Endpoints     int      `json:"endpoint_count"`
	LastUpdated   string   `json:"last_updated"`
}, total int, query string, facets struct {
	Categories map[string]int `json:"categories"`
	Tags       map[string]int `json:"tags"`
	PriceRanges map[string]int `json:"price_ranges"`
}) error {
	fmt.Println()
	if query != "" {
		color.New(color.FgCyan, color.Bold).Printf("🔍 Search Results for \"%s\"\n", query)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("🔍 Browse APIs\n")
	}
	fmt.Printf("Found %d APIs\n\n", total)
	
	if len(results) == 0 {
		fmt.Println("No APIs found matching your criteria")
		return nil
	}
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tCATEGORY\tRATING\tPRICE\tSUBSCRIBERS\n")
	
	for _, api := range results {
		// Badges
		badges := ""
		if api.Featured {
			badges += "⭐ "
		}
		if api.Verified {
			badges += "✓ "
		}
		
		// Price
		price := "Free"
		if api.StartingPrice > 0 {
			price = fmt.Sprintf("From $%.0f/%s", api.StartingPrice, getPricingInterval(api.PricingModel))
		}
		
		// Rating
		rating := fmt.Sprintf("%.1f★ (%d)", api.Rating, api.ReviewCount)
		if api.ReviewCount == 0 {
			rating = "No reviews"
		}
		
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%d\n",
			badges,
			api.Name,
			api.Category,
			rating,
			price,
			api.Subscribers,
		)
	}
	w.Flush()
	
	// Show facets if available
	if searchCategory == "" && len(facets.Categories) > 0 {
		fmt.Printf("\n📊 Filter by Category:\n")
		for cat, count := range facets.Categories {
			fmt.Printf("  • %s (%d)\n", cat, count)
		}
	}
	
	// Navigation hint
	fmt.Printf("\n💡 View details: apidirect info <api-name>\n")
	if total > searchLimit {
		fmt.Printf("   Next page: add --offset %d\n", searchOffset+searchLimit)
	}
	
	return nil
}

func displaySearchGrid(results []struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	Creator       string   `json:"creator"`
	Rating        float64  `json:"rating"`
	ReviewCount   int      `json:"review_count"`
	Subscribers   int      `json:"subscriber_count"`
	PricingModel  string   `json:"pricing_model"`
	StartingPrice float64  `json:"starting_price"`
	Currency      string   `json:"currency"`
	Featured      bool     `json:"featured"`
	Verified      bool     `json:"verified"`
	Endpoints     int      `json:"endpoint_count"`
	LastUpdated   string   `json:"last_updated"`
}, total int, query string) error {
	fmt.Println()
	if query != "" {
		color.New(color.FgCyan, color.Bold).Printf("🔍 Search Results for \"%s\"\n", query)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("🔍 Browse APIs\n")
	}
	fmt.Printf("Found %d APIs\n\n", total)
	
	for i, api := range results {
		if i > 0 {
			fmt.Println(strings.Repeat("─", 60))
		}
		
		// Header with badges
		header := api.Name
		if api.Featured {
			header = "⭐ " + header
		}
		if api.Verified {
			header += " ✓"
		}
		color.New(color.Bold).Println(header)
		
		// Description
		fmt.Printf("%s\n", truncate(api.Description, 80))
		
		// Metadata
		fmt.Printf("\n📁 %s", api.Category)
		if len(api.Tags) > 0 {
			fmt.Printf(" • 🏷️  %s", strings.Join(api.Tags[:min(3, len(api.Tags))], ", "))
		}
		fmt.Printf("\n")
		
		// Stats
		if api.ReviewCount > 0 {
			fmt.Printf("⭐ %.1f (%d reviews) • ", api.Rating, api.ReviewCount)
		}
		fmt.Printf("👥 %d subscribers • ", api.Subscribers)
		fmt.Printf("🔗 %d endpoints\n", api.Endpoints)
		
		// Creator and pricing
		fmt.Printf("👤 %s • ", api.Creator)
		if api.StartingPrice > 0 {
			fmt.Printf("💰 From $%.0f/%s\n", api.StartingPrice, getPricingInterval(api.PricingModel))
		} else {
			fmt.Printf("💰 Free\n")
		}
	}
	
	return nil
}

func displayFeaturedTable(apis []struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	Creator       string   `json:"creator"`
	Rating        float64  `json:"rating"`
	ReviewCount   int      `json:"review_count"`
	FeaturedText  string   `json:"featured_text"`
	Badge         string   `json:"badge"`
	PricingModel  string   `json:"pricing_model"`
	StartingPrice float64  `json:"starting_price"`
}, title, description string) error {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("⭐ %s\n", title)
	fmt.Printf("%s\n\n", description)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "API\tCATEGORY\tRATING\tFEATURED FOR\n")
	
	for _, api := range apis {
		badge := ""
		if api.Badge != "" {
			badge = fmt.Sprintf("[%s] ", api.Badge)
		}
		
		rating := fmt.Sprintf("%.1f★ (%d)", api.Rating, api.ReviewCount)
		if api.ReviewCount == 0 {
			rating = "New"
		}
		
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n",
			badge,
			api.Name,
			api.Category,
			rating,
			api.FeaturedText,
		)
	}
	w.Flush()
	
	return nil
}

func displayFeaturedGrid(apis []struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	Creator       string   `json:"creator"`
	Rating        float64  `json:"rating"`
	ReviewCount   int      `json:"review_count"`
	FeaturedText  string   `json:"featured_text"`
	Badge         string   `json:"badge"`
	PricingModel  string   `json:"pricing_model"`
	StartingPrice float64  `json:"starting_price"`
}, title, description string) error {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("⭐ %s\n", title)
	fmt.Printf("%s\n\n", description)
	
	for i, api := range apis {
		if i > 0 {
			fmt.Println()
		}
		
		// Badge and name
		if api.Badge != "" {
			color.New(color.FgYellow, color.Bold).Printf("[%s] ", api.Badge)
		}
		color.New(color.Bold).Println(api.Name)
		
		// Featured text
		color.New(color.FgGreen).Printf("✨ %s\n", api.FeaturedText)
		
		// Description
		fmt.Printf("%s\n", api.Description)
		
		// Details
		fmt.Printf("\n📁 %s", api.Category)
		if len(api.Tags) > 0 {
			fmt.Printf(" • 🏷️  %s", strings.Join(api.Tags, ", "))
		}
		fmt.Printf("\n")
		
		// Rating and price
		if api.ReviewCount > 0 {
			fmt.Printf("⭐ %.1f (%d reviews) • ", api.Rating, api.ReviewCount)
		}
		if api.StartingPrice > 0 {
			fmt.Printf("💰 From $%.0f/%s", api.StartingPrice, getPricingInterval(api.PricingModel))
		} else {
			fmt.Printf("💰 Free")
		}
		fmt.Printf(" • 👤 %s\n", api.Creator)
		
		fmt.Println(strings.Repeat("─", 60))
	}
	
	return nil
}

func getPricingInterval(model string) string {
	switch model {
	case "subscription_monthly":
		return "mo"
	case "subscription_yearly":
		return "yr"
	case "pay_per_use":
		return "use"
	case "one_time":
		return "once"
	default:
		return "mo"
	}
}

// min moved to utils.go