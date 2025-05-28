package store

import (
	"database/sql"
	"fmt"
	"time"
)

// AccountStore handles creator payment account operations
type AccountStore struct {
	db *sql.DB
}

// NewAccountStore creates a new account store
func NewAccountStore(db *sql.DB) *AccountStore {
	return &AccountStore{db: db}
}

// CreateAccount creates a new creator payment account
func (s *AccountStore) CreateAccount(creatorID, stripeAccountID string) (*CreatorPaymentAccount, error) {
	query := `
		INSERT INTO creator_payment_accounts (creator_id, stripe_account_id, account_status)
		VALUES ($1, $2, 'pending')
		RETURNING id, creator_id, stripe_account_id, account_status, details_submitted, 
				  charges_enabled, payouts_enabled, default_currency, country, business_type,
				  onboarding_completed_at, created_at, updated_at
	`

	var account CreatorPaymentAccount
	err := s.db.QueryRow(query, creatorID, stripeAccountID).Scan(
		&account.ID,
		&account.CreatorID,
		&account.StripeAccountID,
		&account.AccountStatus,
		&account.DetailsSubmitted,
		&account.ChargesEnabled,
		&account.PayoutsEnabled,
		&account.DefaultCurrency,
		&account.Country,
		&account.BusinessType,
		&account.OnboardingCompletedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &account, nil
}

// GetAccountByCreatorID retrieves an account by creator ID
func (s *AccountStore) GetAccountByCreatorID(creatorID string) (*CreatorPaymentAccount, error) {
	query := `
		SELECT id, creator_id, stripe_account_id, account_status, details_submitted, 
			   charges_enabled, payouts_enabled, default_currency, country, business_type,
			   onboarding_completed_at, created_at, updated_at
		FROM creator_payment_accounts
		WHERE creator_id = $1
	`

	var account CreatorPaymentAccount
	err := s.db.QueryRow(query, creatorID).Scan(
		&account.ID,
		&account.CreatorID,
		&account.StripeAccountID,
		&account.AccountStatus,
		&account.DetailsSubmitted,
		&account.ChargesEnabled,
		&account.PayoutsEnabled,
		&account.DefaultCurrency,
		&account.Country,
		&account.BusinessType,
		&account.OnboardingCompletedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetAccountByStripeID retrieves an account by Stripe account ID
func (s *AccountStore) GetAccountByStripeID(stripeAccountID string) (*CreatorPaymentAccount, error) {
	query := `
		SELECT id, creator_id, stripe_account_id, account_status, details_submitted, 
			   charges_enabled, payouts_enabled, default_currency, country, business_type,
			   onboarding_completed_at, created_at, updated_at
		FROM creator_payment_accounts
		WHERE stripe_account_id = $1
	`

	var account CreatorPaymentAccount
	err := s.db.QueryRow(query, stripeAccountID).Scan(
		&account.ID,
		&account.CreatorID,
		&account.StripeAccountID,
		&account.AccountStatus,
		&account.DetailsSubmitted,
		&account.ChargesEnabled,
		&account.PayoutsEnabled,
		&account.DefaultCurrency,
		&account.Country,
		&account.BusinessType,
		&account.OnboardingCompletedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by stripe id: %w", err)
	}

	return &account, nil
}

// UpdateAccountStatus updates the account status and capabilities
func (s *AccountStore) UpdateAccountStatus(stripeAccountID string, status string, detailsSubmitted, chargesEnabled, payoutsEnabled bool) error {
	query := `
		UPDATE creator_payment_accounts
		SET account_status = $2,
			details_submitted = $3,
			charges_enabled = $4,
			payouts_enabled = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE stripe_account_id = $1
	`

	_, err := s.db.Exec(query, stripeAccountID, status, detailsSubmitted, chargesEnabled, payoutsEnabled)
	if err != nil {
		return fmt.Errorf("failed to update account status: %w", err)
	}

	return nil
}

// CompleteOnboarding marks the account as onboarded
func (s *AccountStore) CompleteOnboarding(creatorID string) error {
	query := `
		UPDATE creator_payment_accounts
		SET account_status = 'active',
			onboarding_completed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE creator_id = $1 AND charges_enabled = true AND payouts_enabled = true
	`

	result, err := s.db.Exec(query, creatorID)
	if err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not ready for onboarding completion")
	}

	return nil
}

// GetActiveAccounts retrieves all active creator accounts
func (s *AccountStore) GetActiveAccounts() ([]CreatorPaymentAccount, error) {
	query := `
		SELECT id, creator_id, stripe_account_id, account_status, details_submitted, 
			   charges_enabled, payouts_enabled, default_currency, country, business_type,
			   onboarding_completed_at, created_at, updated_at
		FROM creator_payment_accounts
		WHERE account_status = 'active' AND payouts_enabled = true
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active accounts: %w", err)
	}
	defer rows.Close()

	var accounts []CreatorPaymentAccount
	for rows.Next() {
		var account CreatorPaymentAccount
		err := rows.Scan(
			&account.ID,
			&account.CreatorID,
			&account.StripeAccountID,
			&account.AccountStatus,
			&account.DetailsSubmitted,
			&account.ChargesEnabled,
			&account.PayoutsEnabled,
			&account.DefaultCurrency,
			&account.Country,
			&account.BusinessType,
			&account.OnboardingCompletedAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccountStats returns statistics about creator accounts
func (s *AccountStore) GetAccountStats() (map[string]int, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE account_status = 'pending') as pending,
			COUNT(*) FILTER (WHERE account_status = 'onboarding') as onboarding,
			COUNT(*) FILTER (WHERE account_status = 'active') as active,
			COUNT(*) FILTER (WHERE account_status = 'restricted') as restricted,
			COUNT(*) FILTER (WHERE account_status = 'disabled') as disabled,
			COUNT(*) as total
		FROM creator_payment_accounts
	`

	var pending, onboarding, active, restricted, disabled, total int
	err := s.db.QueryRow(query).Scan(&pending, &onboarding, &active, &restricted, &disabled, &total)
	if err != nil {
		return nil, fmt.Errorf("failed to get account stats: %w", err)
	}

	return map[string]int{
		"pending":     pending,
		"onboarding":  onboarding,
		"active":      active,
		"restricted":  restricted,
		"disabled":    disabled,
		"total":       total,
	}, nil
}

// UpdateAccountDetails updates account details after Stripe webhook
func (s *AccountStore) UpdateAccountDetails(stripeAccountID string, country, currency, businessType string) error {
	query := `
		UPDATE creator_payment_accounts
		SET country = COALESCE(NULLIF($2, ''), country),
			default_currency = COALESCE(NULLIF($3, ''), default_currency),
			business_type = COALESCE(NULLIF($4, ''), business_type),
			updated_at = CURRENT_TIMESTAMP
		WHERE stripe_account_id = $1
	`

	_, err := s.db.Exec(query, stripeAccountID, country, currency, businessType)
	if err != nil {
		return fmt.Errorf("failed to update account details: %w", err)
	}

	return nil
}

// DisableAccount disables a creator's payment account
func (s *AccountStore) DisableAccount(creatorID string, reason string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update account status
	query := `
		UPDATE creator_payment_accounts
		SET account_status = 'disabled',
			updated_at = CURRENT_TIMESTAMP
		WHERE creator_id = $1
	`

	_, err = tx.Exec(query, creatorID)
	if err != nil {
		return fmt.Errorf("failed to disable account: %w", err)
	}

	// Log the reason (you might want to add an audit log table)
	// For now, we'll just update the account

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAccountsRequiringAction retrieves accounts that need attention
func (s *AccountStore) GetAccountsRequiringAction() ([]CreatorPaymentAccount, error) {
	query := `
		SELECT id, creator_id, stripe_account_id, account_status, details_submitted, 
			   charges_enabled, payouts_enabled, default_currency, country, business_type,
			   onboarding_completed_at, created_at, updated_at
		FROM creator_payment_accounts
		WHERE account_status IN ('pending', 'onboarding')
		   OR (details_submitted = false AND created_at < NOW() - INTERVAL '7 days')
		   OR (charges_enabled = false AND payouts_enabled = false)
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts requiring action: %w", err)
	}
	defer rows.Close()

	var accounts []CreatorPaymentAccount
	for rows.Next() {
		var account CreatorPaymentAccount
		err := rows.Scan(
			&account.ID,
			&account.CreatorID,
			&account.StripeAccountID,
			&account.AccountStatus,
			&account.DetailsSubmitted,
			&account.ChargesEnabled,
			&account.PayoutsEnabled,
			&account.DefaultCurrency,
			&account.Country,
			&account.BusinessType,
			&account.OnboardingCompletedAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
