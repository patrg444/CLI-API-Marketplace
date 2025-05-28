package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PostgresStore handles database operations for API keys
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// APIKey represents an API key in the database
type APIKey struct {
	ID         string    `json:"id"`
	KeyPrefix  string    `json:"key_prefix"`
	ConsumerID string    `json:"consumer_id"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// APIKeyValidation contains validation response data
type APIKeyValidation struct {
	Valid          bool      `json:"valid"`
	ConsumerID     string    `json:"consumer_id"`
	SubscriptionID string    `json:"subscription_id"`
	APIKeyID       string    `json:"api_key_id"`
	APIID          string    `json:"api_id"`
	RateLimits     RateLimits `json:"rate_limits"`
}

// RateLimits contains rate limit information
type RateLimits struct {
	PerMinute int `json:"per_minute"`
	PerDay    int `json:"per_day"`
	PerMonth  int `json:"per_month"`
}

// GenerateAPIKey creates a new API key
func (s *PostgresStore) GenerateAPIKey(consumerID, name string) (string, *APIKey, error) {
	// Generate a random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	
	// Create the full API key
	fullKey := "sk_" + hex.EncodeToString(keyBytes)
	
	// Create key prefix for display (first 8 chars after prefix)
	keyPrefix := fullKey[:11] + "..."
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])
	
	// Insert into database
	apiKey := &APIKey{
		ID:         uuid.New().String(),
		KeyPrefix:  keyPrefix,
		ConsumerID: consumerID,
		Name:       name,
		IsActive:   true,
		CreatedAt:  time.Now().UTC(),
	}
	
	query := `
		INSERT INTO api_keys (id, key_hash, key_prefix, consumer_id, name, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	
	err := s.db.QueryRow(
		query,
		apiKey.ID,
		keyHash,
		apiKey.KeyPrefix,
		apiKey.ConsumerID,
		apiKey.Name,
		apiKey.IsActive,
		apiKey.CreatedAt,
	).Scan(&apiKey.CreatedAt)
	
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert API key: %w", err)
	}
	
	return fullKey, apiKey, nil
}

// ValidateAPIKey validates an API key and returns subscription information
func (s *PostgresStore) ValidateAPIKey(apiKey, path string) (*APIKeyValidation, error) {
	// Hash the provided key
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])
	
	// Parse the path to get creator and API name
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path format")
	}
	creator := parts[0]
	apiName := parts[1]
	
	// Query for the API key and subscription information
	query := `
		SELECT 
			ak.id,
			ak.consumer_id,
			s.id as subscription_id,
			s.api_id,
			pp.rate_limit_per_minute,
			pp.rate_limit_per_day,
			pp.rate_limit_per_month
		FROM api_keys ak
		JOIN subscriptions s ON s.api_key_id = ak.id
		JOIN apis a ON a.id = s.api_id
		JOIN users u ON u.id = a.user_id
		JOIN api_pricing_plans pp ON pp.id = s.pricing_plan_id
		WHERE ak.key_hash = $1
			AND ak.is_active = true
			AND s.status = 'active'
			AND a.name = $2
			AND u.username = $3
			AND a.is_published = true
	`
	
	var validation APIKeyValidation
	err := s.db.QueryRow(query, keyHash, apiName, creator).Scan(
		&validation.APIKeyID,
		&validation.ConsumerID,
		&validation.SubscriptionID,
		&validation.APIID,
		&validation.RateLimits.PerMinute,
		&validation.RateLimits.PerDay,
		&validation.RateLimits.PerMonth,
	)
	
	if err == sql.ErrNoRows {
		return &APIKeyValidation{Valid: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	
	// Update last used timestamp
	go s.updateLastUsed(validation.APIKeyID)
	
	validation.Valid = true
	return &validation, nil
}

// GetAPIKey retrieves an API key by ID
func (s *PostgresStore) GetAPIKey(keyID, consumerID string) (*APIKey, error) {
	query := `
		SELECT id, key_prefix, consumer_id, name, is_active, created_at, last_used_at
		FROM api_keys
		WHERE id = $1 AND consumer_id = $2
	`
	
	var apiKey APIKey
	err := s.db.QueryRow(query, keyID, consumerID).Scan(
		&apiKey.ID,
		&apiKey.KeyPrefix,
		&apiKey.ConsumerID,
		&apiKey.Name,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.LastUsedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	
	return &apiKey, nil
}

// ListAPIKeys lists all API keys for a consumer
func (s *PostgresStore) ListAPIKeys(consumerID string) ([]*APIKey, error) {
	query := `
		SELECT id, key_prefix, consumer_id, name, is_active, created_at, last_used_at
		FROM api_keys
		WHERE consumer_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := s.db.Query(query, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()
	
	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		err := rows.Scan(
			&key.ID,
			&key.KeyPrefix,
			&key.ConsumerID,
			&key.Name,
			&key.IsActive,
			&key.CreatedAt,
			&key.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, &key)
	}
	
	return keys, nil
}

// RevokeAPIKey revokes an API key
func (s *PostgresStore) RevokeAPIKey(keyID, consumerID string) error {
	query := `
		UPDATE api_keys
		SET is_active = false
		WHERE id = $1 AND consumer_id = $2
	`
	
	result, err := s.db.Exec(query, keyID, consumerID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	
	return nil
}

// UpdateAPIKey updates an API key's name
func (s *PostgresStore) UpdateAPIKey(keyID, consumerID, name string) error {
	query := `
		UPDATE api_keys
		SET name = $3
		WHERE id = $1 AND consumer_id = $2
	`
	
	result, err := s.db.Exec(query, keyID, consumerID, name)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	
	return nil
}

// updateLastUsed updates the last used timestamp for an API key
func (s *PostgresStore) updateLastUsed(keyID string) {
	query := `
		UPDATE api_keys
		SET last_used_at = $2
		WHERE id = $1
	`
	
	_, _ = s.db.Exec(query, keyID, time.Now().UTC())
}

// EnsureConsumer ensures a consumer record exists
func (s *PostgresStore) EnsureConsumer(cognitoUserID, email string) (string, error) {
	// Try to get existing consumer
	var consumerID string
	err := s.db.QueryRow(
		"SELECT id FROM consumers WHERE cognito_user_id = $1",
		cognitoUserID,
	).Scan(&consumerID)
	
	if err == nil {
		return consumerID, nil
	}
	
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check consumer: %w", err)
	}
	
	// Create new consumer
	consumerID = uuid.New().String()
	_, err = s.db.Exec(
		`INSERT INTO consumers (id, cognito_user_id, email, created_at)
		 VALUES ($1, $2, $3, $4)`,
		consumerID,
		cognitoUserID,
		email,
		time.Now().UTC(),
	)
	
	if err != nil {
		return "", fmt.Errorf("failed to create consumer: %w", err)
	}
	
	return consumerID, nil
}
