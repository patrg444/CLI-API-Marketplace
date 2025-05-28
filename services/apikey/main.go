package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/api-direct/services/apikey/handlers"
	"github.com/api-direct/services/apikey/middleware"
	"github.com/api-direct/services/apikey/store"
)

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize store
	apiKeyStore := store.NewPostgresStore(db)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "apikey",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// API Key endpoints
		keys := api.Group("/keys")
		{
			// Generate new API key (requires auth)
			keys.POST("", middleware.AuthRequired(), handlers.GenerateAPIKey(apiKeyStore))
			
			// Validate API key (no auth required - used by gateway)
			keys.POST("/validate", handlers.ValidateAPIKey(apiKeyStore))
			
			// Get API key details (requires auth)
			keys.GET("/:keyId", middleware.AuthRequired(), handlers.GetAPIKey(apiKeyStore))
			
			// List API keys for a consumer (requires auth)
			keys.GET("", middleware.AuthRequired(), handlers.ListAPIKeys(apiKeyStore))
			
			// Revoke API key (requires auth)
			keys.DELETE("/:keyId", middleware.AuthRequired(), handlers.RevokeAPIKey(apiKeyStore))
			
			// Update API key name (requires auth)
			keys.PUT("/:keyId", middleware.AuthRequired(), handlers.UpdateAPIKey(apiKeyStore))
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API Key Management service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
