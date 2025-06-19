package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"marketplace/elasticsearch"
	"marketplace/handlers"
	"marketplace/indexer"
	"marketplace/middleware"
	"marketplace/store"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewClient(os.Getenv("ELASTICSEARCH_URL"))
	if err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	// Initialize Elasticsearch indices
	if err := elasticsearch.InitializeIndices(esClient); err != nil {
		log.Fatal("Failed to initialize Elasticsearch indices:", err)
	}

	// Initialize stores
	apiStore := store.NewAPIStore(db)
	reviewStore := store.NewReviewStore(db)

	// Initialize indexer
	apiIndexer := indexer.NewAPIIndexer(esClient, apiStore)

	// Initialize search service
	searchService := elasticsearch.NewSearchService(esClient)

	// Initialize handlers
	marketplaceHandler := handlers.NewMarketplaceHandler(apiStore, searchService, apiIndexer)
	reviewHandler := handlers.NewReviewHandler(reviewStore, apiStore)

	// Setup Gin router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "marketplace"})
	})

	// Public routes (no auth required)
	public := r.Group("/api/v1/marketplace")
	{
		// API Discovery
		public.GET("/apis", marketplaceHandler.ListAPIs)
		public.GET("/apis/:id", marketplaceHandler.GetAPI)
		public.GET("/apis/:id/documentation", marketplaceHandler.GetAPIDocumentation)
		
		// Search
		public.POST("/search", marketplaceHandler.SearchAPIs)
		public.GET("/search/suggestions", marketplaceHandler.GetSearchSuggestions)
		
		// Reviews (read-only)
		public.GET("/apis/:id/reviews", reviewHandler.GetAPIReviews)
		public.GET("/apis/:id/reviews/stats", reviewHandler.GetReviewStats)
	}

	// Authenticated routes
	auth := r.Group("/api/v1/marketplace")
	auth.Use(middleware.AuthRequired())
	{
		// API Publishing
		auth.PUT("/apis/:id/publish", marketplaceHandler.PublishAPI)
		
		// Review Management
		auth.POST("/apis/:id/reviews", reviewHandler.SubmitReview)
		auth.POST("/reviews/:id/vote", reviewHandler.VoteOnReview)
		
		// Creator-only routes
		auth.POST("/reviews/:id/response", middleware.CreatorOnly(), reviewHandler.RespondToReview)
		
		// Admin routes for indexing
		auth.POST("/admin/reindex", middleware.AdminOnly(), marketplaceHandler.ReindexAll)
		auth.POST("/admin/index/:id", middleware.AdminOnly(), marketplaceHandler.IndexAPI)
	}

	// Start background workers
	go startBackgroundWorkers(apiIndexer)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Marketplace service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func startBackgroundWorkers(indexer *indexer.APIIndexer) {
	// Periodic reindexing
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Starting periodic reindex...")
		if err := indexer.ReindexAll(); err != nil {
			log.Printf("Reindex failed: %v", err)
		}
	}
}
