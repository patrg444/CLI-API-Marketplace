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
	"github.com/api-direct/services/deployment/handlers"
	"github.com/api-direct/services/deployment/k8s"
	"github.com/api-direct/services/deployment/middleware"
)

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "deployment",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthRequired())
	{
		// Deployment endpoints
		api.POST("/deploy/:apiId", handlers.DeployAPI(k8sClient))
		api.GET("/status/:apiId", handlers.GetDeploymentStatus(k8sClient))
		api.DELETE("/deploy/:apiId", handlers.UndeployAPI(k8sClient))
		api.PUT("/scale/:apiId", handlers.ScaleDeployment(k8sClient))
		
		// Environment management
		api.GET("/env/:apiId", handlers.GetEnvironmentVars(k8sClient))
		api.PUT("/env/:apiId", handlers.UpdateEnvironmentVars(k8sClient))
		
		// Logs and metrics
		api.GET("/logs/:apiId", handlers.StreamLogs(k8sClient))
		api.GET("/metrics/:apiId", handlers.GetMetrics(k8sClient))
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
		log.Printf("Deployment service starting on port %s", port)
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

	log.Println("Server exited")
}
