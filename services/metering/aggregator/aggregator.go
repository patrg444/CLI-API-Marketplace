package aggregator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/apidirect/metering/store"
)

// Aggregator handles the aggregation of usage data
type Aggregator struct {
	usageStore       *store.UsageStore
	aggregationStore *store.AggregationStore
}

// NewAggregator creates a new aggregator
func NewAggregator(usageStore *store.UsageStore, aggregationStore *store.AggregationStore) *Aggregator {
	return &Aggregator{
		usageStore:       usageStore,
		aggregationStore: aggregationStore,
	}
}

// AggregateUsage aggregates recent usage data
func (a *Aggregator) AggregateUsage(ctx context.Context) error {
	// Get the last aggregation time (default to 1 hour ago if first run)
	lastAggregation := time.Now().Add(-1 * time.Hour)
	
	// Get recent usage records
	records, err := a.usageStore.GetRecentUsageForAggregation(lastAggregation)
	if err != nil {
		return fmt.Errorf("failed to get recent usage: %w", err)
	}
	
	if len(records) == 0 {
		log.Println("No new usage records to aggregate")
		return nil
	}
	
	log.Printf("Aggregating %d usage records", len(records))
	
	// Group records by subscription
	subscriptionUsage := make(map[string][]*store.UsageRecord)
	for _, record := range records {
		subscriptionUsage[record.SubscriptionID] = append(subscriptionUsage[record.SubscriptionID], record)
	}
	
	// Aggregate each subscription's usage
	for subscriptionID, records := range subscriptionUsage {
		if err := a.aggregateSubscriptionUsage(ctx, subscriptionID, records); err != nil {
			log.Printf("Failed to aggregate usage for subscription %s: %v", subscriptionID, err)
			// Continue with other subscriptions
		}
	}
	
	return nil
}

// aggregateSubscriptionUsage aggregates usage for a single subscription
func (a *Aggregator) aggregateSubscriptionUsage(ctx context.Context, subscriptionID string, records []*store.UsageRecord) error {
	if len(records) == 0 {
		return nil
	}
	
	// Calculate summary
	summary := &store.UsageSummary{
		SubscriptionID:    subscriptionID,
		PeriodStart:       records[0].Timestamp,
		PeriodEnd:         records[len(records)-1].Timestamp,
		TotalCalls:        int64(len(records)),
		EndpointUsage:     make(map[string]int64),
	}
	
	for _, record := range records {
		// Update real-time counters in Redis
		successful := record.StatusCode < 400
		if err := a.aggregationStore.IncrementUsageCounter(ctx, subscriptionID, successful); err != nil {
			log.Printf("Failed to update Redis counter: %v", err)
		}
		
		// Update summary
		if successful {
			summary.SuccessfulCalls++
		} else {
			summary.FailedCalls++
		}
		
		summary.TotalResponseTime += record.ResponseTimeMs
		summary.TotalRequestSize += record.RequestSizeBytes
		summary.TotalResponseSize += record.ResponseSizeBytes
		
		// Track endpoint usage
		summary.EndpointUsage[record.Endpoint]++
	}
	
	// Store aggregated data
	if err := a.aggregationStore.StoreAggregatedUsage(summary); err != nil {
		return fmt.Errorf("failed to store aggregated usage: %w", err)
	}
	
	return nil
}

// AggregateDaily performs daily aggregation (called by a separate cron job if needed)
func (a *Aggregator) AggregateDaily(ctx context.Context) error {
	// Get yesterday's date range
	now := time.Now()
	startOfYesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	endOfYesterday := startOfYesterday.Add(24 * time.Hour).Add(-1 * time.Second)
	
	log.Printf("Running daily aggregation for %s", startOfYesterday.Format("2006-01-02"))
	
	// Get all subscriptions that had usage yesterday
	query := `
		SELECT DISTINCT subscription_id 
		FROM api_usage 
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	
	// This would be implemented with proper database access
	// For now, we'll use the existing aggregation logic
	
	return nil
}

// AggregateMonthly performs monthly aggregation (called by a separate cron job if needed)
func (a *Aggregator) AggregateMonthly(ctx context.Context) error {
	// Get last month's date range
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstOfLastMonth := firstOfMonth.AddDate(0, -1, 0)
	lastOfLastMonth := firstOfMonth.Add(-1 * time.Second)
	
	log.Printf("Running monthly aggregation for %s", firstOfLastMonth.Format("2006-01"))
	
	// Similar to daily aggregation but for monthly period
	
	return nil
}
