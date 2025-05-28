package store

import (
	"database/sql"
)

// BillingStore aggregates all stores
type BillingStore struct {
	db           *sql.DB
	Consumer     *ConsumerStore
	Subscription *SubscriptionStore
	Invoice      *InvoiceStore
	PricingPlan  *PricingPlanStore
}

// NewBillingStore creates a new billing store
func NewBillingStore(db *sql.DB) *BillingStore {
	return &BillingStore{
		db:           db,
		Consumer:     NewConsumerStore(db),
		Subscription: NewSubscriptionStore(db),
		Invoice:      NewInvoiceStore(db),
		PricingPlan:  NewPricingPlanStore(db),
	}
}

// BeginTx starts a database transaction
func (s *BillingStore) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// PricingPlan represents an API pricing plan
type PricingPlan struct {
	ID                 string                 `json:"id"`
	APIID              string                 `json:"api_id"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	PricePerCall       *float64               `json:"price_per_call,omitempty"`
	MonthlyPrice       *float64               `json:"monthly_price,omitempty"`
	CallLimit          *int                   `json:"call_limit,omitempty"`
	RateLimitPerMinute *int                   `json:"rate_limit_per_minute,omitempty"`
	RateLimitPerDay    *int                   `json:"rate_limit_per_day,omitempty"`
	RateLimitPerMonth  *int                   `json:"rate_limit_per_month,omitempty"`
	Features           map[string]interface{} `json:"features,omitempty"`
	IsActive           bool                   `json:"is_active"`
	StripePriceID      string                 `json:"stripe_price_id,omitempty"`
}

// PricingPlanStore handles pricing plan operations
type PricingPlanStore struct {
	db *sql.DB
}

// NewPricingPlanStore creates a new pricing plan store
func NewPricingPlanStore(db *sql.DB) *PricingPlanStore {
	return &PricingPlanStore{db: db}
}

// GetByID retrieves a pricing plan by ID
func (s *PricingPlanStore) GetByID(id string) (*PricingPlan, error) {
	query := `
		SELECT 
			id, api_id, name, type, price_per_call, monthly_price,
			call_limit, rate_limit_per_minute, rate_limit_per_day,
			rate_limit_per_month, features, is_active
		FROM api_pricing_plans
		WHERE id = $1
	`
	
	plan := &PricingPlan{}
	var features sql.NullString
	
	err := s.db.QueryRow(query, id).Scan(
		&plan.ID,
		&plan.APIID,
		&plan.Name,
		&plan.Type,
		&plan.PricePerCall,
		&plan.MonthlyPrice,
		&plan.CallLimit,
		&plan.RateLimitPerMinute,
		&plan.RateLimitPerDay,
		&plan.RateLimitPerMonth,
		&features,
		&plan.IsActive,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	// Parse features JSON if present
	if features.Valid {
		// In production, you would unmarshal the JSON here
		// plan.Features = parseJSON(features.String)
	}
	
	return plan, nil
}

// ListByAPI lists all pricing plans for an API
func (s *PricingPlanStore) ListByAPI(apiID string) ([]*PricingPlan, error) {
	query := `
		SELECT 
			id, api_id, name, type, price_per_call, monthly_price,
			call_limit, rate_limit_per_minute, rate_limit_per_day,
			rate_limit_per_month, features, is_active
		FROM api_pricing_plans
		WHERE api_id = $1 AND is_active = true
		ORDER BY monthly_price ASC NULLS FIRST
	`
	
	rows, err := s.db.Query(query, apiID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var plans []*PricingPlan
	for rows.Next() {
		plan := &PricingPlan{}
		var features sql.NullString
		
		err := rows.Scan(
			&plan.ID,
			&plan.APIID,
			&plan.Name,
			&plan.Type,
			&plan.PricePerCall,
			&plan.MonthlyPrice,
			&plan.CallLimit,
			&plan.RateLimitPerMinute,
			&plan.RateLimitPerDay,
			&plan.RateLimitPerMonth,
			&features,
			&plan.IsActive,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse features JSON if present
		if features.Valid {
			// In production, you would unmarshal the JSON here
			// plan.Features = parseJSON(features.String)
		}
		
		plans = append(plans, plan)
	}
	
	return plans, rows.Err()
}

// UpdateStripePriceID updates the Stripe price ID for a plan
func (s *PricingPlanStore) UpdateStripePriceID(planID, stripePriceID string) error {
	// Note: This column would need to be added to the schema
	query := `
		UPDATE api_pricing_plans
		SET stripe_price_id = $2
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, planID, stripePriceID)
	return err
}

// GetPricingPlanWithAPI gets a pricing plan with API details
func (s *PricingPlanStore) GetPricingPlanWithAPI(planID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			p.id, p.api_id, p.name, p.type, p.price_per_call, p.monthly_price,
			p.call_limit, p.rate_limit_per_minute, p.rate_limit_per_day,
			p.rate_limit_per_month, p.features, p.is_active,
			a.name as api_name, a.user_id as creator_id
		FROM api_pricing_plans p
		JOIN apis a ON p.api_id = a.id
		WHERE p.id = $1
	`
	
	var plan PricingPlan
	var apiName string
	var creatorID string
	var features sql.NullString
	
	err := s.db.QueryRow(query, planID).Scan(
		&plan.ID,
		&plan.APIID,
		&plan.Name,
		&plan.Type,
		&plan.PricePerCall,
		&plan.MonthlyPrice,
		&plan.CallLimit,
		&plan.RateLimitPerMinute,
		&plan.RateLimitPerDay,
		&plan.RateLimitPerMonth,
		&features,
		&plan.IsActive,
		&apiName,
		&creatorID,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	result := map[string]interface{}{
		"plan":       plan,
		"api_name":   apiName,
		"creator_id": creatorID,
	}
	
	return result, nil
}
