package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// InitDB initializes the database connection
func InitDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// CreatorPaymentAccount represents a creator's Stripe Connect account
type CreatorPaymentAccount struct {
	ID                     string
	CreatorID              string
	StripeAccountID        string
	AccountStatus          string
	DetailsSubmitted       bool
	ChargesEnabled         bool
	PayoutsEnabled         bool
	DefaultCurrency        string
	Country                string
	BusinessType           string
	OnboardingCompletedAt  *string
	CreatedAt              string
	UpdatedAt              string
}

// Payout represents a payout record
type Payout struct {
	ID                   string
	CreatorID            string
	StripePayoutID       *string
	Amount               float64
	Currency             string
	PlatformFee          float64
	NetAmount            float64
	PeriodStart          string
	PeriodEnd            string
	Status               string
	ArrivalDate          *string
	StripeFailureCode    *string
	StripeFailureMessage *string
	CreatedAt            string
	PaidAt               *string
	UpdatedAt            string
}

// PayoutLineItem represents a detailed breakdown of a payout
type PayoutLineItem struct {
	ID                 string
	PayoutID           string
	APIID              string
	GrossRevenue       float64
	PlatformFee        float64
	NetRevenue         float64
	TotalSubscriptions int
	TotalAPICalls      int64
	CreatedAt          string
}

// CreatorEarnings represents real-time earnings for a creator
type CreatorEarnings struct {
	ID                 string
	CreatorID          string
	APIID              string
	APIName            string
	CurrentMonthGross  float64
	CurrentMonthNet    float64
	LifetimeGross      float64
	LifetimeNet        float64
	LifetimePayouts    float64
	PendingPayout      float64
	LastUpdated        string
}

// PlatformRevenue represents platform-wide revenue metrics
type PlatformRevenue struct {
	ID                   string
	Month                string
	TotalGrossRevenue    float64
	TotalPlatformFees    float64
	TotalCreatorPayouts  float64
	TotalTransactions    int64
	TotalAPICalls        int64
	ActiveSubscriptions  int
	CreatedAt            string
	UpdatedAt            string
}

// PayoutSummary represents aggregated payout information
type PayoutSummary struct {
	CreatorID         string
	TotalPending      float64
	TotalPaid         float64
	LastPayoutDate    *string
	NextPayoutAmount  float64
	NextPayoutDate    string
}

// APIEarnings represents earnings for a specific API
type APIEarnings struct {
	APIID              string
	APIName            string
	CurrentMonthGross  float64
	CurrentMonthNet    float64
	LifetimeGross      float64
	LifetimeNet        float64
	TotalSubscriptions int
	TotalAPICalls      int64
}
