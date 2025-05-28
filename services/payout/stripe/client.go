package stripe

import (
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
	"github.com/stripe/stripe-go/v74/loginlink"
	"github.com/stripe/stripe-go/v74/payout"
	"github.com/stripe/stripe-go/v74/transfer"
	"github.com/stripe/stripe-go/v74/webhook"
)

type Client struct {
	secretKey      string
	webhookSecret  string
	platformFeeRate float64
}

// NewClient creates a new Stripe client for Connect operations
func NewClient(secretKey, webhookSecret string) *Client {
	stripe.Key = secretKey
	return &Client{
		secretKey:      secretKey,
		webhookSecret:  webhookSecret,
		platformFeeRate: 0.20, // 20% platform fee
	}
}

// CreateConnectedAccount creates a new Stripe Connect account for a creator
func (c *Client) CreateConnectedAccount(email, creatorID string) (*stripe.Account, error) {
	params := &stripe.AccountParams{
		Type: stripe.String(string(stripe.AccountTypeExpress)),
		Country: stripe.String("US"),
		Email: stripe.String(email),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		BusinessType: stripe.String("individual"),
		Metadata: map[string]string{
			"creator_id": creatorID,
		},
	}

	return account.New(params)
}

// CreateAccountLink creates an onboarding link for Stripe Connect
func (c *Client) CreateAccountLink(accountID, returnURL, refreshURL string) (*stripe.AccountLink, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	return accountlink.New(params)
}

// CreateLoginLink creates a dashboard link for connected accounts
func (c *Client) CreateLoginLink(accountID string) (*stripe.LoginLink, error) {
	return loginlink.New(&stripe.LoginLinkParams{
		Account: stripe.String(accountID),
	})
}

// GetAccount retrieves account details
func (c *Client) GetAccount(accountID string) (*stripe.Account, error) {
	return account.GetByID(accountID, nil)
}

// CreateTransfer transfers funds to a connected account
func (c *Client) CreateTransfer(accountID string, amount int64, currency, description string, metadata map[string]string) (*stripe.Transfer, error) {
	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(currency),
		Destination: stripe.String(accountID),
		Description: stripe.String(description),
		Metadata:    metadata,
	}

	return transfer.New(params)
}

// CreatePayout creates a payout for a connected account
func (c *Client) CreatePayout(accountID string, amount int64, currency string, metadata map[string]string) (*stripe.Payout, error) {
	params := &stripe.PayoutParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Metadata: metadata,
	}
	params.SetStripeAccount(accountID)

	return payout.New(params)
}

// CalculatePlatformFee calculates the platform's commission
func (c *Client) CalculatePlatformFee(grossAmount int64) int64 {
	return int64(float64(grossAmount) * c.platformFeeRate)
}

// CalculateNetAmount calculates the creator's earnings after platform fee
func (c *Client) CalculateNetAmount(grossAmount int64) int64 {
	return grossAmount - c.CalculatePlatformFee(grossAmount)
}

// ValidateWebhookSignature validates a Stripe webhook signature
func (c *Client) ValidateWebhookSignature(payload []byte, header string) (*stripe.Event, error) {
	return webhook.ConstructEvent(payload, header, c.webhookSecret)
}

// AccountStatus represents the onboarding status of a connected account
type AccountStatus struct {
	AccountID        string
	DetailsSubmitted bool
	ChargesEnabled   bool
	PayoutsEnabled   bool
	Requirements     []string
}

// GetAccountStatus returns the current status of a connected account
func (c *Client) GetAccountStatus(accountID string) (*AccountStatus, error) {
	acct, err := c.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	var requirements []string
	if acct.Requirements != nil && acct.Requirements.CurrentlyDue != nil {
		requirements = acct.Requirements.CurrentlyDue
	}

	return &AccountStatus{
		AccountID:        acct.ID,
		DetailsSubmitted: acct.DetailsSubmitted,
		ChargesEnabled:   acct.ChargesEnabled,
		PayoutsEnabled:   acct.PayoutsEnabled,
		Requirements:     requirements,
	}, nil
}

// PayoutSummary represents a summary of payouts for a period
type PayoutSummary struct {
	TotalGross      int64
	TotalPlatformFee int64
	TotalNet        int64
	PayoutCount     int
	Period          string
}

// CreateBulkPayouts processes payouts for multiple creators
func (c *Client) CreateBulkPayouts(payouts []CreatorPayout) ([]*stripe.Transfer, []error) {
	var transfers []*stripe.Transfer
	var errors []error

	for _, p := range payouts {
		// Calculate amounts
		platformFee := c.CalculatePlatformFee(p.GrossAmount)
		netAmount := p.GrossAmount - platformFee

		// Create transfer to connected account
		transfer, err := c.CreateTransfer(
			p.StripeAccountID,
			netAmount,
			"usd",
			fmt.Sprintf("Payout for %s", p.Period),
			map[string]string{
				"creator_id":    p.CreatorID,
				"period":        p.Period,
				"gross_amount":  fmt.Sprintf("%d", p.GrossAmount),
				"platform_fee":  fmt.Sprintf("%d", platformFee),
			},
		)

		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create transfer for creator %s: %w", p.CreatorID, err))
			continue
		}

		transfers = append(transfers, transfer)
	}

	return transfers, errors
}

// CreatorPayout represents a payout to be processed
type CreatorPayout struct {
	CreatorID       string
	StripeAccountID string
	GrossAmount     int64
	Period          string
}

// WebhookEventType represents different Stripe webhook events we handle
type WebhookEventType string

const (
	AccountUpdated          WebhookEventType = "account.updated"
	PayoutPaid             WebhookEventType = "payout.paid"
	PayoutFailed           WebhookEventType = "payout.failed"
	TransferCreated        WebhookEventType = "transfer.created"
	AccountApplicationAuthorized WebhookEventType = "account.application.authorized"
	AccountApplicationDeauthorized WebhookEventType = "account.application.deauthorized"
)

// HandleWebhookEvent processes a Stripe webhook event
func (c *Client) HandleWebhookEvent(event *stripe.Event) error {
	switch WebhookEventType(event.Type) {
	case AccountUpdated:
		// Handle account updates
		return c.handleAccountUpdated(event)
	case PayoutPaid:
		// Handle successful payouts
		return c.handlePayoutPaid(event)
	case PayoutFailed:
		// Handle failed payouts
		return c.handlePayoutFailed(event)
	case TransferCreated:
		// Handle transfer creation
		return c.handleTransferCreated(event)
	default:
		// Log unhandled event types
		return nil
	}
}

func (c *Client) handleAccountUpdated(event *stripe.Event) error {
	// Implementation for account update handling
	return nil
}

func (c *Client) handlePayoutPaid(event *stripe.Event) error {
	// Implementation for successful payout handling
	return nil
}

func (c *Client) handlePayoutFailed(event *stripe.Event) error {
	// Implementation for failed payout handling
	return nil
}

func (c *Client) handleTransferCreated(event *stripe.Event) error {
	// Implementation for transfer creation handling
	return nil
}

// GetPayoutSchedule returns the default payout schedule for the platform
func (c *Client) GetPayoutSchedule() PayoutSchedule {
	return PayoutSchedule{
		Interval: "monthly",
		DayOfMonth: 1, // First day of each month
		MinimumPayoutAmount: 2500, // $25.00 minimum
	}
}

// PayoutSchedule represents when payouts are processed
type PayoutSchedule struct {
	Interval            string
	DayOfMonth          int
	MinimumPayoutAmount int64
}

// ShouldProcessPayout determines if a payout should be processed based on amount and schedule
func (c *Client) ShouldProcessPayout(amount int64, lastPayoutDate time.Time) bool {
	schedule := c.GetPayoutSchedule()
	
	// Check minimum amount
	if amount < schedule.MinimumPayoutAmount {
		return false
	}

	// Check if it's time for monthly payout
	now := time.Now()
	if schedule.Interval == "monthly" && now.Day() == schedule.DayOfMonth {
		// Check if we haven't already processed this month
		return lastPayoutDate.Month() != now.Month() || lastPayoutDate.Year() != now.Year()
	}

	return false
}
