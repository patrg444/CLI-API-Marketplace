package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// UsageSummary represents aggregated usage data
type UsageSummary struct {
	SubscriptionID    string    `json:"subscription_id"`
	ConsumerID        string    `json:"consumer_id"`
	APIID             string    `json:"api_id"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	TotalCalls        int64     `json:"total_calls"`
	SuccessfulCalls   int64     `json:"successful_calls"`
	FailedCalls       int64     `json:"failed_calls"`
	TotalResponseTime int64     `json:"total_response_time_ms"`
	TotalRequestSize  int64     `json:"total_request_size_bytes"`
	TotalResponseSize int64     `json:"total_response_size_bytes"`
	EndpointUsage     map[string]int64 `json:"endpoint_usage"`
}

// AggregationStore handles aggregated usage data
type AggregationStore struct {
	db    *sql.DB
	redis *redis.Client
}

// NewAggregationStore creates a new aggregation store
func NewAggregationStore(db *sql.DB, redis *redis.Client) *AggregationStore {
	return &AggregationStore{
		db:    db,
		redis: redis,
	}
}

// IncrementUsageCounter increments real-time usage counters in Redis
func (s *AggregationStore) IncrementUsageCounter(ctx context.Context, subscriptionID string, successful bool) error {
	// Create keys for different time windows
	now := time.Now()
	dayKey := fmt.Sprintf("usage:%s:daily:%s", subscriptionID, now.Format("2006-01-02"))
	monthKey := fmt.Sprintf("usage:%s:monthly:%s", subscriptionID, now.Format("2006-01"))
	
	// Increment total calls
	pipe := s.redis.Pipeline()
	pipe.HIncrBy(ctx, dayKey, "total_calls", 1)
	pipe.HIncrBy(ctx, monthKey, "total_calls", 1)
	
	// Increment success/failure counters
	if successful {
		pipe.HIncrBy(ctx, dayKey, "successful_calls", 1)
		pipe.HIncrBy(ctx, monthKey, "successful_calls", 1)
	} else {
		pipe.HIncrBy(ctx, dayKey, "failed_calls", 1)
		pipe.HIncrBy(ctx, monthKey, "failed_calls", 1)
	}
	
	// Set expiration (keep data for 35 days)
	pipe.Expire(ctx, dayKey, 35*24*time.Hour)
	pipe.Expire(ctx, monthKey, 35*24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	return err
}

// GetRealtimeUsage gets current usage from Redis
func (s *AggregationStore) GetRealtimeUsage(ctx context.Context, subscriptionID string, period string) (map[string]string, error) {
	var key string
	now := time.Now()
	
	switch period {
	case "daily":
		key = fmt.Sprintf("usage:%s:daily:%s", subscriptionID, now.Format("2006-01-02"))
	case "monthly":
		key = fmt.Sprintf("usage:%s:monthly:%s", subscriptionID, now.Format("2006-01"))
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}
	
	return s.redis.HGetAll(ctx, key).Result()
}

// StoreAggregatedUsage stores aggregated usage data in PostgreSQL
func (s *AggregationStore) StoreAggregatedUsage(summary *UsageSummary) error {
	// Create or update aggregation table (this could be a separate table in production)
	// For now, we'll store in a JSON column for flexibility
	
	// This would typically be a separate aggregation table
	// For this implementation, we'll use a temporary approach
	query := `
		INSERT INTO platform_config (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE
		SET value = $2, updated_at = CURRENT_TIMESTAMP
	`
	
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	
	key := fmt.Sprintf("usage_summary:%s:%s", summary.SubscriptionID, summary.PeriodStart.Format("2006-01-02"))
	_, err = s.db.Exec(query, key, summaryJSON)
	
	return err
}

// GetUsageSummary retrieves usage summary for a subscription
func (s *AggregationStore) GetUsageSummary(subscriptionID string, start, end time.Time) (*UsageSummary, error) {
	// First, try to get from aggregated data
	query := `
		SELECT value
		FROM platform_config
		WHERE key LIKE $1
		ORDER BY updated_at DESC
		LIMIT 1
	`
	
	keyPattern := fmt.Sprintf("usage_summary:%s:%%", subscriptionID)
	
	var summaryJSON []byte
	err := s.db.QueryRow(query, keyPattern).Scan(&summaryJSON)
	if err == sql.ErrNoRows {
		// No aggregated data, calculate from raw usage
		return s.calculateUsageSummary(subscriptionID, start, end)
	}
	if err != nil {
		return nil, err
	}
	
	var summary UsageSummary
	if err := json.Unmarshal(summaryJSON, &summary); err != nil {
		return nil, err
	}
	
	return &summary, nil
}

// calculateUsageSummary calculates usage summary from raw usage records
func (s *AggregationStore) calculateUsageSummary(subscriptionID string, start, end time.Time) (*UsageSummary, error) {
	query := `
		SELECT 
			s.subscription_id,
			s.consumer_id,
			s.api_id,
			COUNT(*) as total_calls,
			SUM(CASE WHEN u.status_code < 400 THEN 1 ELSE 0 END) as successful_calls,
			SUM(CASE WHEN u.status_code >= 400 THEN 1 ELSE 0 END) as failed_calls,
			COALESCE(SUM(u.response_time_ms), 0) as total_response_time,
			COALESCE(SUM(u.request_size_bytes), 0) as total_request_size,
			COALESCE(SUM(u.response_size_bytes), 0) as total_response_size
		FROM api_usage u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.subscription_id = $1 
			AND u.timestamp >= $2 
			AND u.timestamp <= $3
		GROUP BY s.subscription_id, s.consumer_id, s.api_id
	`
	
	var summary UsageSummary
	err := s.db.QueryRow(query, subscriptionID, start, end).Scan(
		&summary.SubscriptionID,
		&summary.ConsumerID,
		&summary.APIID,
		&summary.TotalCalls,
		&summary.SuccessfulCalls,
		&summary.FailedCalls,
		&summary.TotalResponseTime,
		&summary.TotalRequestSize,
		&summary.TotalResponseSize,
	)
	if err != nil {
		return nil, err
	}
	
	summary.PeriodStart = start
	summary.PeriodEnd = end
	
	// Get endpoint breakdown
	endpointQuery := `
		SELECT endpoint, COUNT(*) as count
		FROM api_usage
		WHERE subscription_id = $1 
			AND timestamp >= $2 
			AND timestamp <= $3
		GROUP BY endpoint
	`
	
	rows, err := s.db.Query(endpointQuery, subscriptionID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	summary.EndpointUsage = make(map[string]int64)
	for rows.Next() {
		var endpoint string
		var count int64
		if err := rows.Scan(&endpoint, &count); err != nil {
			return nil, err
		}
		summary.EndpointUsage[endpoint] = count
	}
	
	return &summary, rows.Err()
}

// GetConsumerUsageSummary gets usage summary for all of a consumer's subscriptions
func (s *AggregationStore) GetConsumerUsageSummary(consumerID string, start, end time.Time) ([]*UsageSummary, error) {
	query := `
		SELECT DISTINCT s.id
		FROM subscriptions s
		WHERE s.consumer_id = $1 AND s.status = 'active'
	`
	
	rows, err := s.db.Query(query, consumerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []*UsageSummary
	for rows.Next() {
		var subscriptionID string
		if err := rows.Scan(&subscriptionID); err != nil {
			return nil, err
		}
		
		summary, err := s.GetUsageSummary(subscriptionID, start, end)
		if err != nil {
			continue // Skip if no usage
		}
		summaries = append(summaries, summary)
	}
	
	return summaries, rows.Err()
}

// GetAPIUsageSummary gets usage summary for an API across all subscriptions
func (s *AggregationStore) GetAPIUsageSummary(apiID string, start, end time.Time) (*UsageSummary, error) {
	query := `
		SELECT 
			$1 as api_id,
			COUNT(*) as total_calls,
			SUM(CASE WHEN u.status_code < 400 THEN 1 ELSE 0 END) as successful_calls,
			SUM(CASE WHEN u.status_code >= 400 THEN 1 ELSE 0 END) as failed_calls,
			COALESCE(SUM(u.response_time_ms), 0) as total_response_time,
			COALESCE(SUM(u.request_size_bytes), 0) as total_request_size,
			COALESCE(SUM(u.response_size_bytes), 0) as total_response_size
		FROM api_usage u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE s.api_id = $1 
			AND u.timestamp >= $2 
			AND u.timestamp <= $3
	`
	
	var summary UsageSummary
	err := s.db.QueryRow(query, apiID, start, end).Scan(
		&summary.APIID,
		&summary.TotalCalls,
		&summary.SuccessfulCalls,
		&summary.FailedCalls,
		&summary.TotalResponseTime,
		&summary.TotalRequestSize,
		&summary.TotalResponseSize,
	)
	if err != nil {
		return nil, err
	}
	
	summary.PeriodStart = start
	summary.PeriodEnd = end
	
	return &summary, nil
}
