package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/gateway/ratelimit"
)

// RateLimit middleware enforces rate limits based on subscription
func RateLimit(limiter *ratelimit.RedisRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get rate limits from context (set by API key validation)
		rateLimitsInterface, exists := c.Get("rate_limits")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limits not found",
				"code":  "RATE_LIMITS_MISSING",
			})
			c.Abort()
			return
		}

		// Type assert the rate limits
		rateLimits, ok := rateLimitsInterface.(struct {
			PerMinute int `json:"per_minute"`
			PerDay    int `json:"per_day"`
			PerMonth  int `json:"per_month"`
		})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid rate limits format",
				"code":  "RATE_LIMITS_INVALID",
			})
			c.Abort()
			return
		}

		// Get API key ID for rate limiting
		apiKeyID, _ := c.Get("api_key_id")
		apiKeyIDStr, ok := apiKeyID.(string)
		if !ok || apiKeyIDStr == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "API key ID not found",
				"code":  "API_KEY_ID_MISSING",
			})
			c.Abort()
			return
		}

		// Check per-minute rate limit
		if rateLimits.PerMinute > 0 {
			key := fmt.Sprintf("rate:minute:%s", apiKeyIDStr)
			allowed, remaining, resetTime, err := limiter.CheckLimit(c.Request.Context(), key, rateLimits.PerMinute, time.Minute)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to check rate limit",
					"code":  "RATE_LIMIT_ERROR",
				})
				c.Abort()
				return
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit-Minute", fmt.Sprintf("%d", rateLimits.PerMinute))
			c.Header("X-RateLimit-Remaining-Minute", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				c.Header("Retry-After", fmt.Sprintf("%d", resetTime.Unix()-time.Now().Unix()))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded (per minute)",
					"code":  "RATE_LIMIT_EXCEEDED",
					"retry_after": resetTime.Unix(),
				})
				c.Abort()
				return
			}
		}

		// Check per-day rate limit
		if rateLimits.PerDay > 0 {
			key := fmt.Sprintf("rate:day:%s", apiKeyIDStr)
			allowed, remaining, resetTime, err := limiter.CheckLimit(c.Request.Context(), key, rateLimits.PerDay, 24*time.Hour)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to check rate limit",
					"code":  "RATE_LIMIT_ERROR",
				})
				c.Abort()
				return
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit-Day", fmt.Sprintf("%d", rateLimits.PerDay))
			c.Header("X-RateLimit-Remaining-Day", fmt.Sprintf("%d", remaining))

			if !allowed {
				c.Header("Retry-After", fmt.Sprintf("%d", resetTime.Unix()-time.Now().Unix()))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded (per day)",
					"code":  "RATE_LIMIT_EXCEEDED",
					"retry_after": resetTime.Unix(),
				})
				c.Abort()
				return
			}
		}

		// Check per-month rate limit
		if rateLimits.PerMonth > 0 {
			key := fmt.Sprintf("rate:month:%s", apiKeyIDStr)
			allowed, remaining, resetTime, err := limiter.CheckLimit(c.Request.Context(), key, rateLimits.PerMonth, 30*24*time.Hour)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to check rate limit",
					"code":  "RATE_LIMIT_ERROR",
				})
				c.Abort()
				return
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit-Month", fmt.Sprintf("%d", rateLimits.PerMonth))
			c.Header("X-RateLimit-Remaining-Month", fmt.Sprintf("%d", remaining))

			if !allowed {
				c.Header("Retry-After", fmt.Sprintf("%d", resetTime.Unix()-time.Now().Unix()))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded (per month)",
					"code":  "RATE_LIMIT_EXCEEDED",
					"retry_after": resetTime.Unix(),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
