package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/api-platform/billing-service/store"
	"github.com/api-platform/billing-service/stripe"
	"github.com/redis/go-redis/v9"
)

// BillingWorker handles background billing tasks
type BillingWorker struct {
	billingStore      *store.BillingStore
	subscriptionStore *store.SubscriptionStore
	invoiceStore      *store.InvoiceStore
	stripeClient      *stripe.Client
	redis             *redis.Client
	meteringServiceURL string
}

// NewBillingWorker creates a new billing worker
func NewBillingWorker(
	billingStore *store.BillingStore,
	subscriptionStore *store.SubscriptionStore,
	invoiceStore *store.InvoiceStore,
	stripeClient *stripe.Client,
	redisClient *redis.Client,
) *BillingWorker {
	return &BillingWorker{
		billingStore:      billingStore,
		subscriptionStore: subscriptionStore,
		invoiceStore:      invoiceStore,
		stripeClient:      stripeClient,
		redis:             redisClient,
		meteringServiceURL: "http://metering-service:8080", // Configure this
	}
}

// StartUsageAggregator starts the usage aggregation worker
func (w *BillingWorker) StartUsageAggregator(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	log.Println("Usage aggregator worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Usage aggregator worker stopped")
			return
		case <-ticker.C:
			if err := w.aggregateUsage(ctx); err != nil {
				log.Printf("Error aggregating usage: %v", err)
			}
		}
	}
}

// StartInvoiceGenerator starts the invoice generation worker
func (w *BillingWorker) StartInvoiceGenerator(ctx context.Context) {
	// Run daily at midnight UTC
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		duration := next.Sub(now)

		select {
		case <-ctx.Done():
			log.Println("Invoice generator worker stopped")
			return
		case <-time.After(duration):
			if err := w.generateInvoices(ctx); err != nil {
				log.Printf("Error generating invoices: %v", err)
			}
		}
	}
}

// StartSubscriptionSyncWorker syncs subscription statuses with Stripe
func (w *BillingWorker) StartSubscriptionSyncWorker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute) // Run every 15 minutes
	defer ticker.Stop()

	log.Println("Subscription sync worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Subscription sync worker stopped")
			return
		case <-ticker.C:
			if err := w.syncSubscriptions(ctx); err != nil {
				log.Printf("Error syncing subscriptions: %v", err)
			}
		}
	}
}

// aggregateUsage fetches usage data from metering service and updates Stripe
func (w *BillingWorker) aggregateUsage(ctx context.Context) error {
	// Get all active metered subscriptions
	// For this phase, we'll focus on subscriptions with pay-per-use pricing
	
	// TODO: Implement actual usage aggregation
	// This would:
	// 1. Query active pay-per-use subscriptions
	// 2. For each subscription, fetch usage from metering service
	// 3. Report usage to Stripe using RecordUsage
	// 4. Store aggregated usage locally for analytics

	log.Println("Running usage aggregation...")

	// Mock implementation for now
	// In production, this would fetch from metering service
	
	return nil
}

// generateInvoices generates invoices for due subscriptions
func (w *BillingWorker) generateInvoices(ctx context.Context) error {
	log.Println("Running invoice generation...")

	// For subscription-based plans, Stripe handles invoice generation automatically
	// This worker would handle custom invoice logic if needed

	// Check for any manual invoice requirements
	// For example, custom enterprise invoices or adjustments

	return nil
}

// syncSubscriptions syncs local subscription status with Stripe
func (w *BillingWorker) syncSubscriptions(ctx context.Context) error {
	log.Println("Running subscription sync...")

	// Get all active subscriptions
	// This is a safety mechanism to ensure our local state matches Stripe

	// In a real implementation, this would:
	// 1. List all active subscriptions from database
	// 2. For each subscription with a Stripe ID, fetch status from Stripe
	// 3. Update local status if different
	// 4. Handle any expired subscriptions

	// Check for expired subscriptions
	expiredSubs, err := w.subscriptionStore.GetExpiredSubscriptions()
	if err != nil {
		return fmt.Errorf("error fetching expired subscriptions: %v", err)
	}

	for _, sub := range expiredSubs {
		log.Printf("Processing expired subscription: %s", sub.ID)
		
		// Update status to expired
		if err := w.subscriptionStore.UpdateStatus(sub.ID, "expired"); err != nil {
			log.Printf("Error updating subscription status: %v", err)
			continue
		}

		// TODO: Call API key service to deactivate keys
		if err := w.deactivateAPIKey(sub.APIKeyID); err != nil {
			log.Printf("Error deactivating API key: %v", err)
		}
	}

	return nil
}

// fetchUsageFromMetering fetches usage data from the metering service
func (w *BillingWorker) fetchUsageFromMetering(subscriptionID string, start, end time.Time) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/usage/subscription/%s?start=%s&end=%s",
		w.meteringServiceURL,
		subscriptionID,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("metering service returned status %d", resp.StatusCode)
	}

	var usage struct {
		TotalCalls int64 `json:"total_calls"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return 0, err
	}

	return usage.TotalCalls, nil
}

// deactivateAPIKey calls the API key service to deactivate a key
func (w *BillingWorker) deactivateAPIKey(keyID string) error {
	// TODO: Implement actual API call to key service
	log.Printf("Would deactivate API key: %s", keyID)
	return nil
}

// UsageReporter handles metered usage reporting to Stripe
type UsageReporter struct {
	stripeClient      *stripe.Client
	subscriptionStore *store.SubscriptionStore
	redis             *redis.Client
}

// NewUsageReporter creates a new usage reporter
func NewUsageReporter(
	stripeClient *stripe.Client,
	subscriptionStore *store.SubscriptionStore,
	redisClient *redis.Client,
) *UsageReporter {
	return &UsageReporter{
		stripeClient:      stripeClient,
		subscriptionStore: subscriptionStore,
		redis:             redisClient,
	}
}

// ReportUsage reports usage to Stripe for a subscription
func (r *UsageReporter) ReportUsage(subscriptionID string, quantity int64) error {
	// Get subscription details
	subscription, err := r.subscriptionStore.GetByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %v", err)
	}

	if subscription.StripeSubscriptionID == "" {
		return fmt.Errorf("subscription has no Stripe ID")
	}

	// Get the Stripe subscription to find the subscription item
	stripeSub, err := r.stripeClient.GetSubscription(subscription.StripeSubscriptionID)
	if err != nil {
		return fmt.Errorf("error getting Stripe subscription: %v", err)
	}

	if len(stripeSub.Items.Data) == 0 {
		return fmt.Errorf("subscription has no items")
	}

	// Report usage for the first item (assuming single-item subscriptions)
	subscriptionItemID := stripeSub.Items.Data[0].ID
	timestamp := time.Now().Unix()

	_, err = r.stripeClient.RecordUsage(subscriptionItemID, quantity, timestamp)
	if err != nil {
		return fmt.Errorf("error recording usage: %v", err)
	}

	// Cache the reported usage in Redis for deduplication
	key := fmt.Sprintf("usage_reported:%s:%d", subscriptionID, timestamp)
	r.redis.Set(context.Background(), key, quantity, 24*time.Hour)

	return nil
}

// BatchReportUsage reports usage for multiple subscriptions in batch
func (r *UsageReporter) BatchReportUsage(usageReports map[string]int64) error {
	for subscriptionID, quantity := range usageReports {
		if err := r.ReportUsage(subscriptionID, quantity); err != nil {
			log.Printf("Error reporting usage for subscription %s: %v", subscriptionID, err)
			// Continue with other subscriptions
		}
	}
	return nil
}
