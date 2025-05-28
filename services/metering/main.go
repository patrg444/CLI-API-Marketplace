package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apidirect/metering/aggregator"
	"github.com/apidirect/metering/handlers"
	"github.com/apidirect/metering/middleware"
	"github.com/apidirect/metering/store"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8084")
	viper.SetDefault("DATABASE_URL", "postgresql://apidirect:localpassword@localhost:5432/apidirect?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("GIN_MODE", "debug")

	// Initialize database
	db, err := sql.Open("postgres", viper.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize Redis
	opt, err := redis.ParseURL(viper.GetString("REDIS_URL"))
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opt)
	
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize stores
	usageStore := store.NewUsageStore(db)
	aggregationStore := store.NewAggregationStore(db, redisClient)

	// Initialize aggregator
	agg := aggregator.NewAggregator(usageStore, aggregationStore)

	// Start cron job for aggregation
	c := cron.New()
	
	// Run aggregation every 5 minutes
	_, err = c.AddFunc("*/5 * * * *", func() {
		log.Println("Running usage aggregation...")
		if err := agg.AggregateUsage(ctx); err != nil {
			log.Printf("Aggregation error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// Initialize Gin router
	gin.SetMode(viper.GetString("GIN_MODE"))
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(gin.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Initialize handlers
	h := handlers.NewHandler(usageStore, aggregationStore)

	// API routes
	api := r.Group("/api/v1")
	{
		// Usage ingestion endpoint (called by Gateway)
		api.POST("/usage", h.RecordUsage)

		// Query endpoints (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			// Get usage summary for billing
			protected.GET("/usage/summary", h.GetUsageSummary)
			
			// Get consumer usage details
			protected.GET("/usage/consumer/:id", h.GetConsumerUsage)
			
			// Get API usage analytics (for creators)
			protected.GET("/usage/api/:id", h.GetAPIUsage)
			
			// Get usage for a specific subscription
			protected.GET("/usage/subscription/:id", h.GetSubscriptionUsage)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", viper.GetString("PORT")),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Metering service started on port %s", viper.GetString("PORT"))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
