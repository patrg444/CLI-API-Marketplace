package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// UsageRecord represents a single API usage record
type UsageRecord struct {
	ID                uuid.UUID `json:"id"`
	SubscriptionID    string    `json:"subscription_id"`
	APIKeyID          string    `json:"api_key_id"`
	Timestamp         time.Time `json:"timestamp"`
	Endpoint          string    `json:"endpoint"`
	Method            string    `json:"method"`
	StatusCode        int       `json:"status_code"`
	ResponseTimeMs    int64     `json:"response_time_ms"`
	RequestSizeBytes  int64     `json:"request_size_bytes"`
	ResponseSizeBytes int64     `json:"response_size_bytes"`
}

// UsageStore handles database operations for usage records
type UsageStore struct {
	db *sql.DB
}

// NewUsageStore creates a new usage store
func NewUsageStore(db *sql.DB) *UsageStore {
	return &UsageStore{db: db}
}

// RecordUsage stores a new usage record
func (s *UsageStore) RecordUsage(record *UsageRecord) error {
	query := `
		INSERT INTO api_usage (
			id, subscription_id, api_key_id, timestamp, endpoint, method,
			status_code, response_time_ms, request_size_bytes, response_size_bytes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}

	_, err := s.db.Exec(
		query,
		record.ID,
		record.SubscriptionID,
		record.APIKeyID,
		record.Timestamp,
		record.Endpoint,
		record.Method,
		record.StatusCode,
		record.ResponseTimeMs,
		record.RequestSizeBytes,
		record.ResponseSizeBytes,
	)

	return err
}

// GetUsageBySubscription retrieves usage records for a subscription within a time range
func (s *UsageStore) GetUsageBySubscription(subscriptionID string, start, end time.Time) ([]*UsageRecord, error) {
	query := `
		SELECT id, subscription_id, api_key_id, timestamp, endpoint, method,
			   status_code, response_time_ms, request_size_bytes, response_size_bytes
		FROM api_usage
		WHERE subscription_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
	`

	rows, err := s.db.Query(query, subscriptionID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SubscriptionID,
			&record.APIKeyID,
			&record.Timestamp,
			&record.Endpoint,
			&record.Method,
			&record.StatusCode,
			&record.ResponseTimeMs,
			&record.RequestSizeBytes,
			&record.ResponseSizeBytes,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// GetUsageByConsumer retrieves all usage records for a consumer within a time range
func (s *UsageStore) GetUsageByConsumer(consumerID string, start, end time.Time) ([]*UsageRecord, error) {
	query := `
		SELECT u.id, u.subscription_id, u.api_key_id, u.timestamp, u.endpoint, u.method,
			   u.status_code, u.response_time_ms, u.request_size_bytes, u.response_size_bytes
		FROM api_usage u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE s.consumer_id = $1 AND u.timestamp >= $2 AND u.timestamp <= $3
		ORDER BY u.timestamp DESC
	`

	rows, err := s.db.Query(query, consumerID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SubscriptionID,
			&record.APIKeyID,
			&record.Timestamp,
			&record.Endpoint,
			&record.Method,
			&record.StatusCode,
			&record.ResponseTimeMs,
			&record.RequestSizeBytes,
			&record.ResponseSizeBytes,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// GetUsageByAPI retrieves all usage records for an API within a time range
func (s *UsageStore) GetUsageByAPI(apiID string, start, end time.Time) ([]*UsageRecord, error) {
	query := `
		SELECT u.id, u.subscription_id, u.api_key_id, u.timestamp, u.endpoint, u.method,
			   u.status_code, u.response_time_ms, u.request_size_bytes, u.response_size_bytes
		FROM api_usage u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE s.api_id = $1 AND u.timestamp >= $2 AND u.timestamp <= $3
		ORDER BY u.timestamp DESC
	`

	rows, err := s.db.Query(query, apiID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SubscriptionID,
			&record.APIKeyID,
			&record.Timestamp,
			&record.Endpoint,
			&record.Method,
			&record.StatusCode,
			&record.ResponseTimeMs,
			&record.RequestSizeBytes,
			&record.ResponseSizeBytes,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// GetRecentUsageForAggregation retrieves recent usage records that haven't been aggregated
func (s *UsageStore) GetRecentUsageForAggregation(since time.Time) ([]*UsageRecord, error) {
	query := `
		SELECT id, subscription_id, api_key_id, timestamp, endpoint, method,
			   status_code, response_time_ms, request_size_bytes, response_size_bytes
		FROM api_usage
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SubscriptionID,
			&record.APIKeyID,
			&record.Timestamp,
			&record.Endpoint,
			&record.Method,
			&record.StatusCode,
			&record.ResponseTimeMs,
			&record.RequestSizeBytes,
			&record.ResponseSizeBytes,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}
