package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/api-direct/services/payout/store"
	"github.com/yourusername/api-direct/services/payout/stripe"
)

// PayoutWorker handles payout processing
type PayoutWorker struct {
	stripeClient  *stripe.Client
	payoutStore   *store.PayoutStore
	earningsStore *store.EarningsStore
}

// NewPayoutWorker creates a new payout worker
func NewPayoutWorker(stripeClient *stripe.Client, payoutStore *store.PayoutStore, earningsStore *store.EarningsStore) *PayoutWorker {
	return &PayoutWorker{
		stripeClient:  stripeClient,
		payoutStore:   payoutStore,
		earningsStore: earningsStore,
	}
}

// Start begins the payout worker process
func (w *PayoutWorker) Start(ctx context.Context) {
	log.Println("Starting payout worker...")

	// Run monthly payout processing
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Check immediately on start
	w.processPayouts()

	for {
		select {
		case <-ctx.Done():
			log.Println("Payout worker shutting down...")
			return
		case <-ticker.C:
			// Check if it's the first day of the month
			now := time.Now()
			if now.Day() == 1 {
				log.Println("Processing monthly payouts...")
				w.processPayouts()
			}
		}
	}
}

// processPayouts processes payouts for all eligible creators
func (w *PayoutWorker) processPayouts() {
	// Get creators with pending payouts above threshold
	threshold := float64(w.stripeClient.GetPayoutSchedule().MinimumPayoutAmount) / 100
	creators, err := w.earningsStore.GetCreatorsWithPendingPayouts(threshold)
	if err != nil {
		log.Printf("Failed to get creators with pending payouts: %v", err)
		return
	}

	log.Printf("Found %d creators eligible for payout", len(creators))

	// Calculate period
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	// Process each creator
	for _, creator := range creators {
		if err := w.processCreatorPayout(creator.CreatorID, creator.TotalPending, periodStart, periodEnd); err != nil {
			log.Printf("Failed to process payout for creator %s: %v", creator.CreatorID, err)
			continue
		}
	}
}

// processCreatorPayout processes a single creator's payout
func (w *PayoutWorker) processCreatorPayout(creatorID string, amount float64, periodStart, periodEnd time.Time) error {
	// Create payout record
	payout := &store.Payout{
		CreatorID:   creatorID,
		Amount:      amount,
		Currency:    "USD",
		PlatformFee: amount * 0.20,
		NetAmount:   amount * 0.80,
		PeriodStart: periodStart.Format("2006-01-02"),
		PeriodEnd:   periodEnd.Format("2006-01-02"),
		Status:      "pending",
	}

	createdPayout, err := w.payoutStore.CreatePayout(payout)
	if err != nil {
		return fmt.Errorf("failed to create payout record: %w", err)
	}

	// Get earnings breakdown for line items
	earnings, err := w.earningsStore.GetCreatorEarnings(creatorID)
	if err != nil {
		return fmt.Errorf("failed to get earnings breakdown: %w", err)
	}

	// Create line items
	var lineItems []store.PayoutLineItem
	for _, e := range earnings {
		if e.PendingPayout > 0 {
			lineItems = append(lineItems, store.PayoutLineItem{
				PayoutID:     createdPayout.ID,
				APIID:        e.APIID,
				GrossRevenue: e.PendingPayout,
				PlatformFee:  e.PendingPayout * 0.20,
				NetRevenue:   e.PendingPayout * 0.80,
			})
		}
	}

	if err := w.payoutStore.CreatePayoutLineItems(createdPayout.ID, lineItems); err != nil {
		return fmt.Errorf("failed to create payout line items: %w", err)
	}

	// TODO: Create Stripe transfer
	// This would be implemented once we have the actual Stripe Connect account IDs

	// Update payout status
	if err := w.payoutStore.UpdatePayoutStatus(createdPayout.ID, "processing", nil); err != nil {
		return fmt.Errorf("failed to update payout status: %w", err)
	}

	// Deduct from pending payouts
	if err := w.earningsStore.DeductPaidAmount(creatorID, amount); err != nil {
		return fmt.Errorf("failed to deduct paid amount: %w", err)
	}

	log.Printf("Successfully processed payout for creator %s: $%.2f", creatorID, amount)
	return nil
}

// EarningsWorker handles earnings calculation
type EarningsWorker struct {
	earningsStore *store.EarningsStore
}

// NewEarningsWorker creates a new earnings worker
func NewEarningsWorker(earningsStore *store.EarningsStore) *EarningsWorker {
	return &EarningsWorker{
		earningsStore: earningsStore,
	}
}

// Start begins the earnings worker process
func (w *EarningsWorker) Start(ctx context.Context) {
	log.Println("Starting earnings worker...")

	// Run earnings calculation every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Calculate immediately on start
	w.calculateEarnings()

	for {
		select {
		case <-ctx.Done():
			log.Println("Earnings worker shutting down...")
			return
		case <-ticker.C:
			w.calculateEarnings()
		}
	}
}

// calculateEarnings calculates earnings from billing data
func (w *EarningsWorker) calculateEarnings() {
	now := time.Now()
	
	// Calculate for the current month
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now

	log.Printf("Calculating earnings from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if err := w.earningsStore.CalculateEarningsFromBilling(startDate, endDate); err != nil {
		log.Printf("Failed to calculate earnings: %v", err)
		return
	}

	// Update platform revenue
	// This would aggregate data from various sources
	if err := w.updatePlatformRevenue(now); err != nil {
		log.Printf("Failed to update platform revenue: %v", err)
		return
	}

	// Reset monthly earnings if it's the first day of the month
	if now.Day() == 1 && now.Hour() == 0 {
		log.Println("Resetting monthly earnings...")
		if err := w.earningsStore.ResetMonthlyEarnings(); err != nil {
			log.Printf("Failed to reset monthly earnings: %v", err)
		}
	}
}

// updatePlatformRevenue updates platform-wide revenue metrics
func (w *EarningsWorker) updatePlatformRevenue(date time.Time) error {
	// This would typically aggregate data from various tables
	// For now, using placeholder values
	
	// Get total pending payouts as a proxy for revenue
	totalPending, err := w.earningsStore.GetTotalPendingPayouts()
	if err != nil {
		return fmt.Errorf("failed to get total pending payouts: %w", err)
	}

	// Update platform revenue
	return w.earningsStore.UpdatePlatformRevenue(
		date,
		totalPending,      // gross revenue
		totalPending*0.20, // platform fees
		totalPending*0.80, // creator payouts
		0,                 // transactions (would be calculated)
		0,                 // API calls (would be calculated)
		0,                 // active subscriptions (would be calculated)
	)
}

// MonthlyReportWorker generates monthly reports
type MonthlyReportWorker struct {
	payoutStore   *store.PayoutStore
	earningsStore *store.EarningsStore
}

// NewMonthlyReportWorker creates a new monthly report worker
func NewMonthlyReportWorker(payoutStore *store.PayoutStore, earningsStore *store.EarningsStore) *MonthlyReportWorker {
	return &MonthlyReportWorker{
		payoutStore:   payoutStore,
		earningsStore: earningsStore,
	}
}

// Start begins the monthly report worker process
func (w *MonthlyReportWorker) Start(ctx context.Context) {
	log.Println("Starting monthly report worker...")

	// Run on the first day of each month
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Monthly report worker shutting down...")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Day() == 1 {
				w.generateMonthlyReport(now.AddDate(0, -1, 0))
			}
		}
	}
}

// generateMonthlyReport generates a report for the given month
func (w *MonthlyReportWorker) generateMonthlyReport(month time.Time) {
	log.Printf("Generating monthly report for %s", month.Format("2006-01"))

	// Get platform revenue for the month
	revenue, err := w.earningsStore.GetPlatformRevenue(month)
	if err != nil {
		log.Printf("Failed to get platform revenue: %v", err)
		return
	}

	if revenue == nil {
		log.Printf("No revenue data for %s", month.Format("2006-01"))
		return
	}

	// Get payout statistics
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	payoutStats, err := w.payoutStore.GetPayoutStats(startOfMonth, endOfMonth)
	if err != nil {
		log.Printf("Failed to get payout stats: %v", err)
		return
	}

	// Log report summary
	log.Printf("=== Monthly Report for %s ===", month.Format("2006-01"))
	log.Printf("Gross Revenue: $%.2f", revenue.TotalGrossRevenue)
	log.Printf("Platform Fees: $%.2f (%.1f%%)", revenue.TotalPlatformFees, (revenue.TotalPlatformFees/revenue.TotalGrossRevenue)*100)
	log.Printf("Creator Payouts: $%.2f", revenue.TotalCreatorPayouts)
	log.Printf("Total Transactions: %d", revenue.TotalTransactions)
	log.Printf("Total API Calls: %d", revenue.TotalAPICalls)
	log.Printf("Active Subscriptions: %d", revenue.ActiveSubscriptions)
	log.Printf("Payouts Processed: %v", payoutStats["total_payouts"])
	log.Printf("============================")

	// TODO: Send email report to admins
	// TODO: Generate PDF report
	// TODO: Store report in S3
}
