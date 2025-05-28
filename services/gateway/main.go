package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/api-direct/services/gateway/handlers"
	"github.com/api-direct/services/gateway/middleware"
	"github.com/api-direct/services/gateway/proxy"
	"github.com/api-direct/services/gateway/ratelimit"
)

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	apiKeyServiceURL := os.Getenv("API_KEY_SERVICE_URL")
	if apiKeyServiceURL == "" {
		apiKeyServiceURL = "http://localhost:8083"
	}

	meteringServiceURL := os.Getenv("METERING_SERVICE_URL")
	if meteringServiceURL == "" {
		meteringServiceURL = "http://localhost:8084"
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRedisRateLimiter(redisClient)

	// Initialize proxy handler
	proxyHandler := proxy.NewHandler(meteringServiceURL)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "gateway",
			"timestamp": time.Now().UTC(),
		})
	})

	// API Gateway routes - all requests go through API key validation and rate limiting
	api := router.Group("/api")
	api.Use(middleware.ValidateAPIKey(apiKeyServiceURL))
	api.Use(middleware.RateLimit(rateLimiter))
	api.Use(middleware.LogRequest(meteringServiceURL))
	{
		// Proxy all requests to the appropriate creator function
		api.Any("/:creator/:apiName/*path", handlers.ProxyToFunction(proxyHandler))
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
		log.Printf("API Gateway service starting on port %s", port)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server exited")
}
