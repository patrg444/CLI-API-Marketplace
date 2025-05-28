package store

import (
	"database/sql"
	"time"
)

// Consumer represents a consumer in the system
type Consumer struct {
	ID               string    `json:"id"`
	CognitoUserID    string    `json:"cognito_user_id"`
	Email            string    `json:"email"`
	StripeCustomerID string    `json:"stripe_customer_id,omitempty"`
	CompanyName      string    `json:"company_name,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ConsumerStore handles consumer data operations
type ConsumerStore struct {
	db *sql.DB
}

// NewConsumerStore creates a new consumer store
func NewConsumerStore(db *sql.DB) *ConsumerStore {
	return &ConsumerStore{db: db}
}

// Create creates a new consumer
func (s *ConsumerStore) Create(consumer *Consumer) error {
	query := `
		INSERT INTO consumers (cognito_user_id, email, stripe_customer_id, company_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	
	err := s.db.QueryRow(
		query,
		consumer.CognitoUserID,
		consumer.Email,
		consumer.StripeCustomerID,
		consumer.CompanyName,
	).Scan(&consumer.ID, &consumer.CreatedAt)
	
	return err
}

// GetByID retrieves a consumer by ID
func (s *ConsumerStore) GetByID(id string) (*Consumer, error) {
	query := `
		SELECT id, cognito_user_id, email, stripe_customer_id, company_name, created_at
		FROM consumers
		WHERE id = $1
	`
	
	consumer := &Consumer{}
	err := s.db.QueryRow(query, id).Scan(
		&consumer.ID,
		&consumer.CognitoUserID,
		&consumer.Email,
		&consumer.StripeCustomerID,
		&consumer.CompanyName,
		&consumer.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return consumer, err
}

// GetByCognitoID retrieves a consumer by Cognito user ID
func (s *ConsumerStore) GetByCognitoID(cognitoID string) (*Consumer, error) {
	query := `
		SELECT id, cognito_user_id, email, stripe_customer_id, company_name, created_at
		FROM consumers
		WHERE cognito_user_id = $1
	`
	
	consumer := &Consumer{}
	err := s.db.QueryRow(query, cognitoID).Scan(
		&consumer.ID,
		&consumer.CognitoUserID,
		&consumer.Email,
		&consumer.StripeCustomerID,
		&consumer.CompanyName,
		&consumer.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return consumer, err
}

// GetByStripeCustomerID retrieves a consumer by Stripe customer ID
func (s *ConsumerStore) GetByStripeCustomerID(stripeCustomerID string) (*Consumer, error) {
	query := `
		SELECT id, cognito_user_id, email, stripe_customer_id, company_name, created_at
		FROM consumers
		WHERE stripe_customer_id = $1
	`
	
	consumer := &Consumer{}
	err := s.db.QueryRow(query, stripeCustomerID).Scan(
		&consumer.ID,
		&consumer.CognitoUserID,
		&consumer.Email,
		&consumer.StripeCustomerID,
		&consumer.CompanyName,
		&consumer.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return consumer, err
}

// Update updates a consumer
func (s *ConsumerStore) Update(consumer *Consumer) error {
	query := `
		UPDATE consumers
		SET email = $2, stripe_customer_id = $3, company_name = $4
		WHERE id = $1
	`
	
	_, err := s.db.Exec(
		query,
		consumer.ID,
		consumer.Email,
		consumer.StripeCustomerID,
		consumer.CompanyName,
	)
	
	return err
}

// UpdateStripeCustomerID updates the Stripe customer ID for a consumer
func (s *ConsumerStore) UpdateStripeCustomerID(consumerID, stripeCustomerID string) error {
	query := `
		UPDATE consumers
		SET stripe_customer_id = $2
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, consumerID, stripeCustomerID)
	return err
}

// List lists consumers with pagination
func (s *ConsumerStore) List(limit, offset int) ([]*Consumer, error) {
	query := `
		SELECT id, cognito_user_id, email, stripe_customer_id, company_name, created_at
		FROM consumers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var consumers []*Consumer
	for rows.Next() {
		consumer := &Consumer{}
		err := rows.Scan(
			&consumer.ID,
			&consumer.CognitoUserID,
			&consumer.Email,
			&consumer.StripeCustomerID,
			&consumer.CompanyName,
			&consumer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, consumer)
	}
	
	return consumers, rows.Err()
}

// Delete deletes a consumer
func (s *ConsumerStore) Delete(id string) error {
	query := `DELETE FROM consumers WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
