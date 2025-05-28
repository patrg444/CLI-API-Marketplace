package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisRateLimiter implements sliding window rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// CheckLimit checks if a request is allowed under the rate limit
// Returns: allowed (bool), remaining (int), resetTime (time.Time), error
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Add(-window).UnixNano()
	
	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()
	
	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	
	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to execute pipeline: %w", err)
	}
	
	// Get current count
	count := countCmd.Val()
	
	// Check if under limit
	if count >= int64(limit) {
		// Get the oldest entry to calculate reset time
		oldestEntries, err := r.client.ZRange(ctx, key, 0, 0).Result()
		if err != nil || len(oldestEntries) == 0 {
			return false, 0, now.Add(window), nil
		}
		
		// Get the score (timestamp) of the oldest entry
		scores, err := r.client.ZMScore(ctx, key, oldestEntries[0]).Result()
		if err != nil || len(scores) == 0 {
			return false, 0, now.Add(window), nil
		}
		
		oldestTime := time.Unix(0, int64(scores[0]))
		resetTime := oldestTime.Add(window)
		
		return false, 0, resetTime, nil
	}
	
	// Add current request
	err = r.client.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d-%d", now.UnixNano(), now.UnixNano()%1000),
	}).Err()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to add request: %w", err)
	}
	
	// Set expiration on the key
	err = r.client.Expire(ctx, key, window).Err()
	if err != nil {
		// Non-critical error, log but don't fail the request
		fmt.Printf("Failed to set expiration on key %s: %v\n", key, err)
	}
	
	remaining := limit - int(count) - 1
	resetTime := now.Add(window)
	
	return true, remaining, resetTime, nil
}

// Reset clears the rate limit for a specific key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
