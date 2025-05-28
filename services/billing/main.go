package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/api-platform/billing-service/handlers"
	"github.com/api-platform/billing-service/middleware"
	"github.com/api-platform/billing-service/store"
	"github.com/api-platform/billing-service/stripe"
	"github.com/api-platform/billing-service/webhooks"
	"github.com/api-platform/billing-service/workers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis connection
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}

	redisClient := redis.NewClient(opt)
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("STRIPE_SECRET_KEY is required")
	}

	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if stripeWebhookSecret == "" {
		log.Fatal("STRIPE_WEBHOOK_SECRET is required")
	}

	stripeClient := stripe.NewClient(stripeKey)

	// Initialize stores
	billingStore := store.NewBillingStore(db)
	consumerStore := store.NewConsumerStore(db)
	subscriptionStore := store.NewSubscriptionStore(db)
	invoiceStore := store.NewInvoiceStore(db)

	// Initialize handlers
	billingHandler := handlers.NewBillingHandler(
		billingStore,
		consumerStore,
		subscriptionStore,
		invoiceStore,
		stripeClient,
		redisClient,
	)

	// Initialize webhook handler
	webhookHandler := webhooks.NewStripeWebhookHandler(
		stripeWebhookSecret,
		billingStore,
		consumerStore,
		subscriptionStore,
		invoiceStore,
	)

	// Initialize workers
	billingWorker := workers.NewBillingWorker(
		billingStore,
		subscriptionStore,
		invoiceStore,
		stripeClient,
		redisClient,
	)

	// Start background workers
	go billingWorker.StartUsageAggregator(ctx)
	go billingWorker.StartInvoiceGenerator(ctx)
	go billingWorker.StartSubscriptionSyncWorker(ctx)

	// Setup routes
	r := mux.NewRouter()

	// API routes with authentication
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Consumer routes
	api.HandleFunc("/consumers/register", billingHandler.RegisterConsumer).Methods("POST")
	api.HandleFunc("/consumers/{consumerId}", billingHandler.GetConsumer).Methods("GET")

	// Subscription routes
	api.HandleFunc("/subscriptions", billingHandler.CreateSubscription).Methods("POST")
	api.HandleFunc("/subscriptions", billingHandler.ListSubscriptions).Methods("GET")
	api.HandleFunc("/subscriptions/{subscriptionId}", billingHandler.GetSubscription).Methods("GET")
	api.HandleFunc("/subscriptions/{subscriptionId}/cancel", billingHandler.CancelSubscription).Methods("PUT")
	api.HandleFunc("/subscriptions/{subscriptionId}/upgrade", billingHandler.UpgradeSubscription).Methods("PUT")
	api.HandleFunc("/subscriptions/{subscriptionId}/usage", billingHandler.GetSubscriptionUsage).Methods("GET")

	// Payment method routes
	api.HandleFunc("/payment-methods", billingHandler.AddPaymentMethod).Methods("POST")
	api.HandleFunc("/payment-methods", billingHandler.ListPaymentMethods).Methods("GET")
	api.HandleFunc("/payment-methods/{paymentMethodId}", billingHandler.RemovePaymentMethod).Methods("DELETE")
	api.HandleFunc("/payment-methods/{paymentMethodId}/default", billingHandler.SetDefaultPaymentMethod).Methods("PUT")

	// Invoice routes
	api.HandleFunc("/invoices", billingHandler.ListInvoices).Methods("GET")
	api.HandleFunc("/invoices/{invoiceId}", billingHandler.GetInvoice).Methods("GET")
	api.HandleFunc("/invoices/{invoiceId}/download", billingHandler.DownloadInvoice).Methods("GET")

	// Usage summary routes (for creators)
	api.HandleFunc("/apis/{apiId}/usage", billingHandler.GetAPIUsageSummary).Methods("GET")
	api.HandleFunc("/apis/{apiId}/earnings", billingHandler.GetAPIEarnings).Methods("GET")

	// Webhook route (no auth required)
	r.HandleFunc("/webhooks/stripe", webhookHandler.HandleStripeWebhook).Methods("POST")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Billing service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
