package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/yourusername/api-direct/services/payout/handlers"
	"github.com/yourusername/api-direct/services/payout/middleware"
	"github.com/yourusername/api-direct/services/payout/store"
	"github.com/yourusername/api-direct/services/payout/stripe"
	"github.com/yourusername/api-direct/services/payout/workers"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database connection
	db, err := store.InitDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Stripe client
	stripeClient := stripe.NewClient(
		os.Getenv("STRIPE_SECRET_KEY"),
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)

	// Initialize stores
	payoutStore := store.NewPayoutStore(db)
	accountStore := store.NewAccountStore(db)
	earningsStore := store.NewEarningsStore(db)

	// Initialize handlers
	h := handlers.NewHandlers(stripeClient, payoutStore, accountStore, earningsStore)

	// Initialize workers
	payoutWorker := workers.NewPayoutWorker(stripeClient, payoutStore, earningsStore)
	earningsWorker := workers.NewEarningsWorker(earningsStore)

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go payoutWorker.Start(ctx)
	go earningsWorker.Start(ctx)

	// Setup routes
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Stripe Connect account management
	api.HandleFunc("/accounts/onboard", h.StartOnboarding).Methods("POST")
	api.HandleFunc("/accounts/onboard/callback", h.OnboardingCallback).Methods("GET")
	api.HandleFunc("/accounts/status", h.GetAccountStatus).Methods("GET")
	api.HandleFunc("/accounts/dashboard", h.GetDashboardLink).Methods("GET")

	// Earnings and payouts
	api.HandleFunc("/earnings", h.GetEarnings).Methods("GET")
	api.HandleFunc("/earnings/{apiId}", h.GetAPIEarnings).Methods("GET")
	api.HandleFunc("/payouts", h.ListPayouts).Methods("GET")
	api.HandleFunc("/payouts/{payoutId}", h.GetPayoutDetails).Methods("GET")
	api.HandleFunc("/payouts/upcoming", h.GetUpcomingPayout).Methods("GET")

	// Platform analytics (admin only)
	api.HandleFunc("/platform/revenue", middleware.AdminOnly(h.GetPlatformRevenue)).Methods("GET")
	api.HandleFunc("/platform/analytics", middleware.AdminOnly(h.GetPlatformAnalytics)).Methods("GET")

	// Webhooks
	r.HandleFunc("/webhooks/stripe", h.HandleStripeWebhook).Methods("POST")

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	// Start server
	srv := &http.Server{
		Addr:         ":8086",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("Payout Service starting on port 8086...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
