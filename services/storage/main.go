package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/storage/handlers"
	"github.com/api-direct/services/storage/middleware"
	"github.com/api-direct/services/storage/s3client"
)

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize S3 client
	s3Client, err := s3client.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "storage",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthRequired())
	{
		// Code upload endpoints
		api.POST("/upload/:apiId", handlers.UploadCode(s3Client))
		api.GET("/download/:apiId/:version", handlers.DownloadCode(s3Client))
		api.GET("/versions/:apiId", handlers.ListVersions(s3Client))
		api.DELETE("/code/:apiId/:version", handlers.DeleteVersion(s3Client))
		
		// Metadata endpoints
		api.GET("/metadata/:apiId/:version", handlers.GetMetadata(s3Client))
		api.PUT("/metadata/:apiId/:version", handlers.UpdateMetadata(s3Client))
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
		log.Printf("Storage service starting on port %s", port)
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
