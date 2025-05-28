package store

import (
	"database/sql"
	"fmt"
	"time"
)

// PayoutStore handles payout operations
type PayoutStore struct {
	db *sql.DB
}

// NewPayoutStore creates a new payout store
func NewPayoutStore(db *sql.DB) *PayoutStore {
	return &PayoutStore{db: db}
}

// CreatePayout creates a new payout record
func (s *PayoutStore) CreatePayout(payout *Payout) (*Payout, error) {
	query := `
		INSERT INTO payouts (creator_id, amount, currency, platform_fee, net_amount, 
							period_start, period_end, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		payout.CreatorID,
		payout.Amount,
		payout.Currency,
		payout.PlatformFee,
		payout.NetAmount,
		payout.PeriodStart,
		payout.PeriodEnd,
		payout.Status,
	).Scan(&payout.ID, &payout.CreatedAt, &payout.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create payout: %w", err)
	}

	return payout, nil
}

// GetPayout retrieves a payout by ID
func (s *PayoutStore) GetPayout(payoutID string) (*Payout, error) {
	query := `
		SELECT id, creator_id, stripe_payout_id, amount, currency, platform_fee, net_amount,
			   period_start, period_end, status, arrival_date, stripe_failure_code,
			   stripe_failure_message, created_at, paid_at, updated_at
		FROM payouts
		WHERE id = $1
	`

	var payout Payout
	err := s.db.QueryRow(query, payoutID).Scan(
		&payout.ID,
		&payout.CreatorID,
		&payout.StripePayoutID,
		&payout.Amount,
		&payout.Currency,
		&payout.PlatformFee,
		&payout.NetAmount,
		&payout.PeriodStart,
		&payout.PeriodEnd,
		&payout.Status,
		&payout.ArrivalDate,
		&payout.StripeFailureCode,
		&payout.StripeFailureMessage,
		&payout.CreatedAt,
		&payout.PaidAt,
		&payout.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payout: %w", err)
	}

	return &payout, nil
}

// GetPayoutsByCreator retrieves all payouts for a creator
func (s *PayoutStore) GetPayoutsByCreator(creatorID string, limit, offset int) ([]Payout, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM payouts WHERE creator_id = $1`
	var total int
	err := s.db.QueryRow(countQuery, creatorID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payouts: %w", err)
	}

	// Get payouts
	query := `
		SELECT id, creator_id, stripe_payout_id, amount, currency, platform_fee, net_amount,
			   period_start, period_end, status, arrival_date, stripe_failure_code,
			   stripe_failure_message, created_at, paid_at, updated_at
		FROM payouts
		WHERE creator_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, creatorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payouts: %w", err)
	}
	defer rows.Close()

	var payouts []Payout
	for rows.Next() {
		var payout Payout
		err := rows.Scan(
			&payout.ID,
			&payout.CreatorID,
			&payout.StripePayoutID,
			&payout.Amount,
			&payout.Currency,
			&payout.PlatformFee,
			&payout.NetAmount,
			&payout.PeriodStart,
			&payout.PeriodEnd,
			&payout.Status,
			&payout.ArrivalDate,
			&payout.StripeFailureCode,
			&payout.StripeFailureMessage,
			&payout.CreatedAt,
			&payout.PaidAt,
			&payout.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payout: %w", err)
		}
		payouts = append(payouts, payout)
	}

	return payouts, total, nil
}

// UpdatePayoutStatus updates the status of a payout
func (s *PayoutStore) UpdatePayoutStatus(payoutID, status string, stripePayoutID *string) error {
	query := `
		UPDATE payouts
		SET status = $2,
			stripe_payout_id = COALESCE($3, stripe_payout_id),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := s.db.Exec(query, payoutID, status, stripePayoutID)
	if err != nil {
		return fmt.Errorf("failed to update payout status: %w", err)
	}

	return nil
}

// MarkPayoutPaid marks a payout as paid
func (s *PayoutStore) MarkPayoutPaid(stripePayoutID string, arrivalDate time.Time) error {
	query := `
		UPDATE payouts
		SET status = 'paid',
			arrival_date = $2,
			paid_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE stripe_payout_id = $1
	`

	_, err := s.db.Exec(query, stripePayoutID, arrivalDate)
	if err != nil {
		return fmt.Errorf("failed to mark payout as paid: %w", err)
	}

	return nil
}

// MarkPayoutFailed marks a payout as failed
func (s *PayoutStore) MarkPayoutFailed(stripePayoutID, failureCode, failureMessage string) error {
	query := `
		UPDATE payouts
		SET status = 'failed',
			stripe_failure_code = $2,
			stripe_failure_message = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE stripe_payout_id = $1
	`

	_, err := s.db.Exec(query, stripePayoutID, failureCode, failureMessage)
	if err != nil {
		return fmt.Errorf("failed to mark payout as failed: %w", err)
	}

	return nil
}

// CreatePayoutLineItems creates line items for a payout
func (s *PayoutStore) CreatePayoutLineItems(payoutID string, items []PayoutLineItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO payout_line_items (payout_id, api_id, gross_revenue, platform_fee, 
									   net_revenue, total_subscriptions, total_api_calls)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(
			payoutID,
			item.APIID,
			item.GrossRevenue,
			item.PlatformFee,
			item.NetRevenue,
			item.TotalSubscriptions,
			item.TotalAPICalls,
		)
		if err != nil {
			return fmt.Errorf("failed to insert line item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPayoutLineItems retrieves line items for a payout
func (s *PayoutStore) GetPayoutLineItems(payoutID string) ([]PayoutLineItem, error) {
	query := `
		SELECT pli.id, pli.payout_id, pli.api_id, pli.gross_revenue, pli.platform_fee,
			   pli.net_revenue, pli.total_subscriptions, pli.total_api_calls, pli.created_at
		FROM payout_line_items pli
		WHERE pli.payout_id = $1
		ORDER BY pli.gross_revenue DESC
	`

	rows, err := s.db.Query(query, payoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout line items: %w", err)
	}
	defer rows.Close()

	var items []PayoutLineItem
	for rows.Next() {
		var item PayoutLineItem
		err := rows.Scan(
			&item.ID,
			&item.PayoutID,
			&item.APIID,
			&item.GrossRevenue,
			&item.PlatformFee,
			&item.NetRevenue,
			&item.TotalSubscriptions,
			&item.TotalAPICalls,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetPendingPayouts retrieves all pending payouts
func (s *PayoutStore) GetPendingPayouts() ([]Payout, error) {
	query := `
		SELECT id, creator_id, stripe_payout_id, amount, currency, platform_fee, net_amount,
			   period_start, period_end, status, arrival_date, stripe_failure_code,
			   stripe_failure_message, created_at, paid_at, updated_at
		FROM payouts
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payouts: %w", err)
	}
	defer rows.Close()

	var payouts []Payout
	for rows.Next() {
		var payout Payout
		err := rows.Scan(
			&payout.ID,
			&payout.CreatorID,
			&payout.StripePayoutID,
			&payout.Amount,
			&payout.Currency,
			&payout.PlatformFee,
			&payout.NetAmount,
			&payout.PeriodStart,
			&payout.PeriodEnd,
			&payout.Status,
			&payout.ArrivalDate,
			&payout.StripeFailureCode,
			&payout.StripeFailureMessage,
			&payout.CreatedAt,
			&payout.PaidAt,
			&payout.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payout: %w", err)
		}
		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// GetPayoutSummary retrieves payout summary for a creator
func (s *PayoutStore) GetPayoutSummary(creatorID string) (*PayoutSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'pending' THEN net_amount ELSE 0 END), 0) as total_pending,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN net_amount ELSE 0 END), 0) as total_paid,
			MAX(CASE WHEN status = 'paid' THEN paid_at ELSE NULL END) as last_payout_date
		FROM payouts
		WHERE creator_id = $1
	`

	var summary PayoutSummary
	summary.CreatorID = creatorID

	err := s.db.QueryRow(query, creatorID).Scan(
		&summary.TotalPending,
		&summary.TotalPaid,
		&summary.LastPayoutDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout summary: %w", err)
	}

	return &summary, nil
}

// GetUpcomingPayoutAmount calculates the upcoming payout amount for a creator
func (s *PayoutStore) GetUpcomingPayoutAmount(creatorID string, periodStart, periodEnd time.Time) (float64, error) {
	// This query would join with billing data to calculate upcoming payout
	// For now, returning from creator_earnings table
	query := `
		SELECT COALESCE(SUM(pending_payout), 0)
		FROM creator_earnings
		WHERE creator_id = $1
	`

	var amount float64
	err := s.db.QueryRow(query, creatorID).Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get upcoming payout amount: %w", err)
	}

	return amount, nil
}

// GetPayoutStats returns statistics about payouts
func (s *PayoutStore) GetPayoutStats(startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_payouts,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_payouts,
			COUNT(*) FILTER (WHERE status = 'processing') as processing_payouts,
			COUNT(*) FILTER (WHERE status = 'paid') as paid_payouts,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_payouts,
			COALESCE(SUM(amount), 0) as total_gross_amount,
			COALESCE(SUM(platform_fee), 0) as total_platform_fees,
			COALESCE(SUM(net_amount), 0) as total_net_amount,
			COUNT(DISTINCT creator_id) as unique_creators
		FROM payouts
		WHERE created_at >= $1 AND created_at <= $2
	`

	var stats struct {
		TotalPayouts       int
		PendingPayouts     int
		ProcessingPayouts  int
		PaidPayouts        int
		FailedPayouts      int
		TotalGrossAmount   float64
		TotalPlatformFees  float64
		TotalNetAmount     float64
		UniqueCreators     int
	}

	err := s.db.QueryRow(query, startDate, endDate).Scan(
		&stats.TotalPayouts,
		&stats.PendingPayouts,
		&stats.ProcessingPayouts,
		&stats.PaidPayouts,
		&stats.FailedPayouts,
		&stats.TotalGrossAmount,
		&stats.TotalPlatformFees,
		&stats.TotalNetAmount,
		&stats.UniqueCreators,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout stats: %w", err)
	}

	return map[string]interface{}{
		"total_payouts":        stats.TotalPayouts,
		"pending_payouts":      stats.PendingPayouts,
		"processing_payouts":   stats.ProcessingPayouts,
		"paid_payouts":         stats.PaidPayouts,
		"failed_payouts":       stats.FailedPayouts,
		"total_gross_amount":   stats.TotalGrossAmount,
		"total_platform_fees":  stats.TotalPlatformFees,
		"total_net_amount":     stats.TotalNetAmount,
		"unique_creators":      stats.UniqueCreators,
		"average_payout":       stats.TotalNetAmount / float64(max(stats.TotalPayouts, 1)),
	}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
