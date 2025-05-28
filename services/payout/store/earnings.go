package store

import (
	"database/sql"
	"fmt"
	"time"
)

// EarningsStore handles creator earnings operations
type EarningsStore struct {
	db *sql.DB
}

// NewEarningsStore creates a new earnings store
func NewEarningsStore(db *sql.DB) *EarningsStore {
	return &EarningsStore{db: db}
}

// GetCreatorEarnings retrieves all earnings for a creator
func (s *EarningsStore) GetCreatorEarnings(creatorID string) ([]CreatorEarnings, error) {
	query := `
		SELECT ce.id, ce.creator_id, ce.api_id, a.name as api_name,
			   ce.current_month_gross, ce.current_month_net,
			   ce.lifetime_gross, ce.lifetime_net, ce.lifetime_payouts,
			   ce.pending_payout, ce.last_updated
		FROM creator_earnings ce
		JOIN apis a ON ce.api_id = a.id
		WHERE ce.creator_id = $1
		ORDER BY ce.current_month_gross DESC
	`

	rows, err := s.db.Query(query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator earnings: %w", err)
	}
	defer rows.Close()

	var earnings []CreatorEarnings
	for rows.Next() {
		var e CreatorEarnings
		err := rows.Scan(
			&e.ID,
			&e.CreatorID,
			&e.APIID,
			&e.APIName,
			&e.CurrentMonthGross,
			&e.CurrentMonthNet,
			&e.LifetimeGross,
			&e.LifetimeNet,
			&e.LifetimePayouts,
			&e.PendingPayout,
			&e.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan earnings: %w", err)
		}
		earnings = append(earnings, e)
	}

	return earnings, nil
}

// GetAPIEarnings retrieves earnings for a specific API
func (s *EarningsStore) GetAPIEarnings(creatorID, apiID string) (*CreatorEarnings, error) {
	query := `
		SELECT ce.id, ce.creator_id, ce.api_id, a.name as api_name,
			   ce.current_month_gross, ce.current_month_net,
			   ce.lifetime_gross, ce.lifetime_net, ce.lifetime_payouts,
			   ce.pending_payout, ce.last_updated
		FROM creator_earnings ce
		JOIN apis a ON ce.api_id = a.id
		WHERE ce.creator_id = $1 AND ce.api_id = $2
	`

	var e CreatorEarnings
	err := s.db.QueryRow(query, creatorID, apiID).Scan(
		&e.ID,
		&e.CreatorID,
		&e.APIID,
		&e.APIName,
		&e.CurrentMonthGross,
		&e.CurrentMonthNet,
		&e.LifetimeGross,
		&e.LifetimeNet,
		&e.LifetimePayouts,
		&e.PendingPayout,
		&e.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API earnings: %w", err)
	}

	return &e, nil
}

// UpdateEarnings updates earnings for a creator's API
func (s *EarningsStore) UpdateEarnings(creatorID, apiID string, monthlyGross, monthlyNet float64) error {
	query := `
		INSERT INTO creator_earnings (creator_id, api_id, current_month_gross, current_month_net,
									  lifetime_gross, lifetime_net, pending_payout)
		VALUES ($1, $2, $3, $4, $3, $4, $4)
		ON CONFLICT (creator_id, api_id) DO UPDATE
		SET current_month_gross = creator_earnings.current_month_gross + EXCLUDED.current_month_gross,
			current_month_net = creator_earnings.current_month_net + EXCLUDED.current_month_net,
			lifetime_gross = creator_earnings.lifetime_gross + EXCLUDED.current_month_gross,
			lifetime_net = creator_earnings.lifetime_net + EXCLUDED.current_month_net,
			pending_payout = creator_earnings.pending_payout + EXCLUDED.current_month_net,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, creatorID, apiID, monthlyGross, monthlyNet)
	if err != nil {
		return fmt.Errorf("failed to update earnings: %w", err)
	}

	return nil
}

// ResetMonthlyEarnings resets current month earnings at the start of a new month
func (s *EarningsStore) ResetMonthlyEarnings() error {
	query := `
		UPDATE creator_earnings
		SET current_month_gross = 0,
			current_month_net = 0,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset monthly earnings: %w", err)
	}

	return nil
}

// DeductPaidAmount deducts paid amount from pending payout
func (s *EarningsStore) DeductPaidAmount(creatorID string, amount float64) error {
	query := `
		UPDATE creator_earnings
		SET pending_payout = GREATEST(0, pending_payout - $2),
			lifetime_payouts = lifetime_payouts + $2,
			last_updated = CURRENT_TIMESTAMP
		WHERE creator_id = $1
	`

	_, err := s.db.Exec(query, creatorID, amount)
	if err != nil {
		return fmt.Errorf("failed to deduct paid amount: %w", err)
	}

	return nil
}

// GetTotalPendingPayouts gets total pending payouts across all creators
func (s *EarningsStore) GetTotalPendingPayouts() (float64, error) {
	query := `SELECT COALESCE(SUM(pending_payout), 0) FROM creator_earnings`

	var total float64
	err := s.db.QueryRow(query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total pending payouts: %w", err)
	}

	return total, nil
}

// GetCreatorsWithPendingPayouts retrieves creators with pending payouts above threshold
func (s *EarningsStore) GetCreatorsWithPendingPayouts(threshold float64) ([]struct {
	CreatorID     string
	TotalPending  float64
}, error) {
	query := `
		SELECT creator_id, SUM(pending_payout) as total_pending
		FROM creator_earnings
		GROUP BY creator_id
		HAVING SUM(pending_payout) >= $1
		ORDER BY total_pending DESC
	`

	rows, err := s.db.Query(query, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get creators with pending payouts: %w", err)
	}
	defer rows.Close()

	var results []struct {
		CreatorID     string
		TotalPending  float64
	}

	for rows.Next() {
		var r struct {
			CreatorID     string
			TotalPending  float64
		}
		err := rows.Scan(&r.CreatorID, &r.TotalPending)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

// UpdatePlatformRevenue updates platform-wide revenue metrics
func (s *EarningsStore) UpdatePlatformRevenue(month time.Time, grossRevenue, platformFees, creatorPayouts float64, transactions, apiCalls int64, subscriptions int) error {
	query := `
		INSERT INTO platform_revenue (month, total_gross_revenue, total_platform_fees,
									  total_creator_payouts, total_transactions,
									  total_api_calls, active_subscriptions)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (month) DO UPDATE
		SET total_gross_revenue = platform_revenue.total_gross_revenue + EXCLUDED.total_gross_revenue,
			total_platform_fees = platform_revenue.total_platform_fees + EXCLUDED.total_platform_fees,
			total_creator_payouts = platform_revenue.total_creator_payouts + EXCLUDED.total_creator_payouts,
			total_transactions = platform_revenue.total_transactions + EXCLUDED.total_transactions,
			total_api_calls = platform_revenue.total_api_calls + EXCLUDED.total_api_calls,
			active_subscriptions = EXCLUDED.active_subscriptions,
			updated_at = CURRENT_TIMESTAMP
	`

	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	_, err := s.db.Exec(query, monthStart, grossRevenue, platformFees, creatorPayouts, transactions, apiCalls, subscriptions)
	if err != nil {
		return fmt.Errorf("failed to update platform revenue: %w", err)
	}

	return nil
}

// GetPlatformRevenue retrieves platform revenue for a specific month
func (s *EarningsStore) GetPlatformRevenue(month time.Time) (*PlatformRevenue, error) {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	
	query := `
		SELECT id, month, total_gross_revenue, total_platform_fees,
			   total_creator_payouts, total_transactions, total_api_calls,
			   active_subscriptions, created_at, updated_at
		FROM platform_revenue
		WHERE month = $1
	`

	var p PlatformRevenue
	err := s.db.QueryRow(query, monthStart).Scan(
		&p.ID,
		&p.Month,
		&p.TotalGrossRevenue,
		&p.TotalPlatformFees,
		&p.TotalCreatorPayouts,
		&p.TotalTransactions,
		&p.TotalAPICalls,
		&p.ActiveSubscriptions,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get platform revenue: %w", err)
	}

	return &p, nil
}

// GetPlatformRevenueHistory retrieves platform revenue history
func (s *EarningsStore) GetPlatformRevenueHistory(months int) ([]PlatformRevenue, error) {
	query := `
		SELECT id, month, total_gross_revenue, total_platform_fees,
			   total_creator_payouts, total_transactions, total_api_calls,
			   active_subscriptions, created_at, updated_at
		FROM platform_revenue
		ORDER BY month DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform revenue history: %w", err)
	}
	defer rows.Close()

	var revenues []PlatformRevenue
	for rows.Next() {
		var p PlatformRevenue
		err := rows.Scan(
			&p.ID,
			&p.Month,
			&p.TotalGrossRevenue,
			&p.TotalPlatformFees,
			&p.TotalCreatorPayouts,
			&p.TotalTransactions,
			&p.TotalAPICalls,
			&p.ActiveSubscriptions,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan platform revenue: %w", err)
		}
		revenues = append(revenues, p)
	}

	return revenues, nil
}

// CalculateEarningsFromBilling calculates earnings from billing data
func (s *EarningsStore) CalculateEarningsFromBilling(startDate, endDate time.Time) error {
	// This would typically join with billing tables to calculate earnings
	// For now, this is a placeholder that would be called by a worker
	query := `
		WITH billing_summary AS (
			SELECT 
				a.creator_id,
				a.id as api_id,
				SUM(i.amount) as gross_revenue,
				SUM(i.amount * 0.20) as platform_fee,
				SUM(i.amount * 0.80) as net_revenue
			FROM invoices i
			JOIN subscriptions s ON i.subscription_id = s.id
			JOIN pricing_plans pp ON s.pricing_plan_id = pp.id
			JOIN apis a ON pp.api_id = a.id
			WHERE i.status = 'paid'
			  AND i.period_start >= $1
			  AND i.period_end <= $2
			GROUP BY a.creator_id, a.id
		)
		INSERT INTO creator_earnings (creator_id, api_id, current_month_gross, current_month_net,
									  lifetime_gross, lifetime_net, pending_payout)
		SELECT 
			creator_id,
			api_id,
			gross_revenue,
			net_revenue,
			gross_revenue,
			net_revenue,
			net_revenue
		FROM billing_summary
		ON CONFLICT (creator_id, api_id) DO UPDATE
		SET current_month_gross = creator_earnings.current_month_gross + EXCLUDED.current_month_gross,
			current_month_net = creator_earnings.current_month_net + EXCLUDED.current_month_net,
			lifetime_gross = creator_earnings.lifetime_gross + EXCLUDED.current_month_gross,
			lifetime_net = creator_earnings.lifetime_net + EXCLUDED.current_month_net,
			pending_payout = creator_earnings.pending_payout + EXCLUDED.current_month_net,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to calculate earnings from billing: %w", err)
	}

	return nil
}

// GetEarningsSummary returns a summary of earnings for a creator
func (s *EarningsStore) GetEarningsSummary(creatorID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(DISTINCT api_id) as total_apis,
			COALESCE(SUM(current_month_gross), 0) as current_month_gross,
			COALESCE(SUM(current_month_net), 0) as current_month_net,
			COALESCE(SUM(lifetime_gross), 0) as lifetime_gross,
			COALESCE(SUM(lifetime_net), 0) as lifetime_net,
			COALESCE(SUM(lifetime_payouts), 0) as lifetime_payouts,
			COALESCE(SUM(pending_payout), 0) as pending_payout
		FROM creator_earnings
		WHERE creator_id = $1
	`

	var summary struct {
		TotalAPIs          int
		CurrentMonthGross  float64
		CurrentMonthNet    float64
		LifetimeGross      float64
		LifetimeNet        float64
		LifetimePayouts    float64
		PendingPayout      float64
	}

	err := s.db.QueryRow(query, creatorID).Scan(
		&summary.TotalAPIs,
		&summary.CurrentMonthGross,
		&summary.CurrentMonthNet,
		&summary.LifetimeGross,
		&summary.LifetimeNet,
		&summary.LifetimePayouts,
		&summary.PendingPayout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get earnings summary: %w", err)
	}

	return map[string]interface{}{
		"total_apis":           summary.TotalAPIs,
		"current_month_gross":  summary.CurrentMonthGross,
		"current_month_net":    summary.CurrentMonthNet,
		"lifetime_gross":       summary.LifetimeGross,
		"lifetime_net":         summary.LifetimeNet,
		"lifetime_payouts":     summary.LifetimePayouts,
		"pending_payout":       summary.PendingPayout,
		"lifetime_platform_fee": summary.LifetimeGross - summary.LifetimeNet,
	}, nil
}
