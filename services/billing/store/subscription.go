package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Subscription represents a subscription in the system
type Subscription struct {
	ID                   string    `json:"id"`
	ConsumerID           string    `json:"consumer_id"`
	APIID                string    `json:"api_id"`
	PricingPlanID        string    `json:"pricing_plan_id"`
	APIKeyID             string    `json:"api_key_id"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	Status               string    `json:"status"`
	StartedAt            time.Time `json:"started_at"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
}

// SubscriptionWithDetails includes additional information
type SubscriptionWithDetails struct {
	Subscription
	APIName       string  `json:"api_name"`
	PlanName      string  `json:"plan_name"`
	PlanType      string  `json:"plan_type"`
	MonthlyPrice  float64 `json:"monthly_price,omitempty"`
	PricePerCall  float64 `json:"price_per_call,omitempty"`
	CallLimit     *int    `json:"call_limit,omitempty"`
}

// SubscriptionStore handles subscription data operations
type SubscriptionStore struct {
	db *sql.DB
}

// NewSubscriptionStore creates a new subscription store
func NewSubscriptionStore(db *sql.DB) *SubscriptionStore {
	return &SubscriptionStore{db: db}
}

// Create creates a new subscription
func (s *SubscriptionStore) Create(subscription *Subscription) error {
	query := `
		INSERT INTO subscriptions (
			consumer_id, api_id, pricing_plan_id, api_key_id, 
			stripe_subscription_id, status, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, started_at
	`
	
	err := s.db.QueryRow(
		query,
		subscription.ConsumerID,
		subscription.APIID,
		subscription.PricingPlanID,
		subscription.APIKeyID,
		subscription.StripeSubscriptionID,
		subscription.Status,
		subscription.ExpiresAt,
	).Scan(&subscription.ID, &subscription.StartedAt)
	
	return err
}

// GetByID retrieves a subscription by ID
func (s *SubscriptionStore) GetByID(id string) (*Subscription, error) {
	query := `
		SELECT 
			id, consumer_id, api_id, pricing_plan_id, api_key_id,
			stripe_subscription_id, status, started_at, cancelled_at, expires_at
		FROM subscriptions
		WHERE id = $1
	`
	
	subscription := &Subscription{}
	err := s.db.QueryRow(query, id).Scan(
		&subscription.ID,
		&subscription.ConsumerID,
		&subscription.APIID,
		&subscription.PricingPlanID,
		&subscription.APIKeyID,
		&subscription.StripeSubscriptionID,
		&subscription.Status,
		&subscription.StartedAt,
		&subscription.CancelledAt,
		&subscription.ExpiresAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return subscription, err
}

// GetByStripeID retrieves a subscription by Stripe subscription ID
func (s *SubscriptionStore) GetByStripeID(stripeID string) (*Subscription, error) {
	query := `
		SELECT 
			id, consumer_id, api_id, pricing_plan_id, api_key_id,
			stripe_subscription_id, status, started_at, cancelled_at, expires_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`
	
	subscription := &Subscription{}
	err := s.db.QueryRow(query, stripeID).Scan(
		&subscription.ID,
		&subscription.ConsumerID,
		&subscription.APIID,
		&subscription.PricingPlanID,
		&subscription.APIKeyID,
		&subscription.StripeSubscriptionID,
		&subscription.Status,
		&subscription.StartedAt,
		&subscription.CancelledAt,
		&subscription.ExpiresAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return subscription, err
}

// ListByConsumer lists all subscriptions for a consumer
func (s *SubscriptionStore) ListByConsumer(consumerID string) ([]*SubscriptionWithDetails, error) {
	query := `
		SELECT 
			s.id, s.consumer_id, s.api_id, s.pricing_plan_id, s.api_key_id,
			s.stripe_subscription_id, s.status, s.started_at, s.cancelled_at, s.expires_at,
			a.name as api_name, p.name as plan_name, p.type as plan_type,
			p.monthly_price, p.price_per_call, p.call_limit
		FROM subscriptions s
		JOIN apis a ON s.api_id = a.id
		JOIN api_pricing_plans p ON s.pricing_plan_id = p.id
		WHERE s.consumer_id = $1
		ORDER BY s.started_at DESC
	`
	
	rows, err := s.db.Query(query, consumerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*SubscriptionWithDetails
	for rows.Next() {
		sub := &SubscriptionWithDetails{}
		err := rows.Scan(
			&sub.ID,
			&sub.ConsumerID,
			&sub.APIID,
			&sub.PricingPlanID,
			&sub.APIKeyID,
			&sub.StripeSubscriptionID,
			&sub.Status,
			&sub.StartedAt,
			&sub.CancelledAt,
			&sub.ExpiresAt,
			&sub.APIName,
			&sub.PlanName,
			&sub.PlanType,
			&sub.MonthlyPrice,
			&sub.PricePerCall,
			&sub.CallLimit,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, rows.Err()
}

// ListByAPI lists all subscriptions for an API
func (s *SubscriptionStore) ListByAPI(apiID string) ([]*Subscription, error) {
	query := `
		SELECT 
			id, consumer_id, api_id, pricing_plan_id, api_key_id,
			stripe_subscription_id, status, started_at, cancelled_at, expires_at
		FROM subscriptions
		WHERE api_id = $1
		ORDER BY started_at DESC
	`
	
	rows, err := s.db.Query(query, apiID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ConsumerID,
			&sub.APIID,
			&sub.PricingPlanID,
			&sub.APIKeyID,
			&sub.StripeSubscriptionID,
			&sub.Status,
			&sub.StartedAt,
			&sub.CancelledAt,
			&sub.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, rows.Err()
}

// Update updates a subscription
func (s *SubscriptionStore) Update(subscription *Subscription) error {
	query := `
		UPDATE subscriptions
		SET 
			pricing_plan_id = $2,
			stripe_subscription_id = $3,
			status = $4,
			cancelled_at = $5,
			expires_at = $6
		WHERE id = $1
	`
	
	_, err := s.db.Exec(
		query,
		subscription.ID,
		subscription.PricingPlanID,
		subscription.StripeSubscriptionID,
		subscription.Status,
		subscription.CancelledAt,
		subscription.ExpiresAt,
	)
	
	return err
}

// UpdateStatus updates the status of a subscription
func (s *SubscriptionStore) UpdateStatus(id, status string) error {
	query := `
		UPDATE subscriptions
		SET status = $2
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, id, status)
	return err
}

// Cancel cancels a subscription
func (s *SubscriptionStore) Cancel(id string, cancelledAt time.Time) error {
	query := `
		UPDATE subscriptions
		SET status = 'cancelled', cancelled_at = $2
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, id, cancelledAt)
	return err
}

// GetActiveCount gets the count of active subscriptions for an API
func (s *SubscriptionStore) GetActiveCount(apiID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE api_id = $1 AND status = 'active'
	`
	
	var count int
	err := s.db.QueryRow(query, apiID).Scan(&count)
	return count, err
}

// GetActiveByPlan gets active subscriptions by pricing plan
func (s *SubscriptionStore) GetActiveByPlan(planID string) ([]*Subscription, error) {
	query := `
		SELECT 
			id, consumer_id, api_id, pricing_plan_id, api_key_id,
			stripe_subscription_id, status, started_at, cancelled_at, expires_at
		FROM subscriptions
		WHERE pricing_plan_id = $1 AND status = 'active'
		ORDER BY started_at DESC
	`
	
	rows, err := s.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ConsumerID,
			&sub.APIID,
			&sub.PricingPlanID,
			&sub.APIKeyID,
			&sub.StripeSubscriptionID,
			&sub.Status,
			&sub.StartedAt,
			&sub.CancelledAt,
			&sub.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, rows.Err()
}

// CheckExistingSubscription checks if a consumer already has an active subscription to an API
func (s *SubscriptionStore) CheckExistingSubscription(consumerID, apiID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE consumer_id = $1 AND api_id = $2 AND status IN ('active', 'trial')
	`
	
	var count int
	err := s.db.QueryRow(query, consumerID, apiID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// GetWithDetails retrieves a subscription with additional details
func (s *SubscriptionStore) GetWithDetails(id string) (*SubscriptionWithDetails, error) {
	query := `
		SELECT 
			s.id, s.consumer_id, s.api_id, s.pricing_plan_id, s.api_key_id,
			s.stripe_subscription_id, s.status, s.started_at, s.cancelled_at, s.expires_at,
			a.name as api_name, p.name as plan_name, p.type as plan_type,
			p.monthly_price, p.price_per_call, p.call_limit
		FROM subscriptions s
		JOIN apis a ON s.api_id = a.id
		JOIN api_pricing_plans p ON s.pricing_plan_id = p.id
		WHERE s.id = $1
	`
	
	sub := &SubscriptionWithDetails{}
	err := s.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.ConsumerID,
		&sub.APIID,
		&sub.PricingPlanID,
		&sub.APIKeyID,
		&sub.StripeSubscriptionID,
		&sub.Status,
		&sub.StartedAt,
		&sub.CancelledAt,
		&sub.ExpiresAt,
		&sub.APIName,
		&sub.PlanName,
		&sub.PlanType,
		&sub.MonthlyPrice,
		&sub.PricePerCall,
		&sub.CallLimit,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return sub, err
}

// GetExpiredSubscriptions gets subscriptions that have expired
func (s *SubscriptionStore) GetExpiredSubscriptions() ([]*Subscription, error) {
	query := `
		SELECT 
			id, consumer_id, api_id, pricing_plan_id, api_key_id,
			stripe_subscription_id, status, started_at, cancelled_at, expires_at
		FROM subscriptions
		WHERE status = 'active' AND expires_at < NOW()
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ConsumerID,
			&sub.APIID,
			&sub.PricingPlanID,
			&sub.APIKeyID,
			&sub.StripeSubscriptionID,
			&sub.Status,
			&sub.StartedAt,
			&sub.CancelledAt,
			&sub.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, rows.Err()
}

// GetSubscriptionStats gets statistics for subscriptions
func (s *SubscriptionStore) GetSubscriptionStats(apiID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'active') as active_count,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_count,
			COUNT(*) FILTER (WHERE status = 'trial') as trial_count,
			COUNT(*) as total_count
		FROM subscriptions
		WHERE api_id = $1
	`
	
	var activeCount, cancelledCount, trialCount, totalCount int
	err := s.db.QueryRow(query, apiID).Scan(
		&activeCount,
		&cancelledCount,
		&trialCount,
		&totalCount,
	)
	
	if err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"active_count":    activeCount,
		"cancelled_count": cancelledCount,
		"trial_count":     trialCount,
		"total_count":     totalCount,
		"churn_rate":      float64(cancelledCount) / float64(totalCount) * 100,
	}
	
	return stats, nil
}

// Delete deletes a subscription (use with caution)
func (s *SubscriptionStore) Delete(id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
