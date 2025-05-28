package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/apidirect/metering/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the metering service
type Handler struct {
	usageStore       *store.UsageStore
	aggregationStore *store.AggregationStore
}

// NewHandler creates a new handler
func NewHandler(usageStore *store.UsageStore, aggregationStore *store.AggregationStore) *Handler {
	return &Handler{
		usageStore:       usageStore,
		aggregationStore: aggregationStore,
	}
}

// RecordUsage handles usage recording from the API Gateway
func (h *Handler) RecordUsage(c *gin.Context) {
	var req struct {
		SubscriptionID    string `json:"subscription_id" binding:"required"`
		APIKeyID          string `json:"api_key_id" binding:"required"`
		Timestamp         string `json:"timestamp" binding:"required"`
		Endpoint          string `json:"endpoint" binding:"required"`
		Method            string `json:"method" binding:"required"`
		StatusCode        int    `json:"status_code" binding:"required"`
		ResponseTimeMs    int64  `json:"response_time_ms"`
		RequestSizeBytes  int64  `json:"request_size_bytes"`
		ResponseSizeBytes int64  `json:"response_size_bytes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp format"})
		return
	}

	// Create usage record
	record := &store.UsageRecord{
		ID:                uuid.New(),
		SubscriptionID:    req.SubscriptionID,
		APIKeyID:          req.APIKeyID,
		Timestamp:         timestamp,
		Endpoint:          req.Endpoint,
		Method:            req.Method,
		StatusCode:        req.StatusCode,
		ResponseTimeMs:    req.ResponseTimeMs,
		RequestSizeBytes:  req.RequestSizeBytes,
		ResponseSizeBytes: req.ResponseSizeBytes,
	}

	// Store the record
	if err := h.usageStore.RecordUsage(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record usage"})
		return
	}

	// Update real-time counters asynchronously
	go func() {
		ctx := context.Background()
		successful := req.StatusCode < 400
		h.aggregationStore.IncrementUsageCounter(ctx, req.SubscriptionID, successful)
	}()

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// GetUsageSummary returns usage summary for billing purposes
func (h *Handler) GetUsageSummary(c *gin.Context) {
	// Get query parameters
	subscriptionID := c.Query("subscription_id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription_id is required"})
		return
	}

	// Parse date range
	start, end, err := parseDateRange(startStr, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get usage summary
	summary, err := h.aggregationStore.GetUsageSummary(subscriptionID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage summary"})
		return
	}

	// Add real-time data from Redis if within current period
	if isCurrentPeriod(start, end) {
		ctx := c.Request.Context()
		realtimeData, _ := h.aggregationStore.GetRealtimeUsage(ctx, subscriptionID, "monthly")
		if realtimeData != nil {
			// Merge real-time data with aggregated data
			summary = mergeRealtimeData(summary, realtimeData)
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetConsumerUsage returns usage details for a consumer
func (h *Handler) GetConsumerUsage(c *gin.Context) {
	consumerID := c.Param("id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	// Validate consumer access
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")
	
	// Only allow consumers to view their own usage, or platform admins
	if userType != "admin" && userID != consumerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse date range
	start, end, err := parseDateRange(startStr, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get consumer usage summaries
	summaries, err := h.aggregationStore.GetConsumerUsageSummary(consumerID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get consumer usage"})
		return
	}

	// Calculate totals across all subscriptions
	var totalCalls, successfulCalls, failedCalls int64
	for _, summary := range summaries {
		totalCalls += summary.TotalCalls
		successfulCalls += summary.SuccessfulCalls
		failedCalls += summary.FailedCalls
	}

	response := gin.H{
		"consumer_id":      consumerID,
		"period_start":     start,
		"period_end":       end,
		"total_calls":      totalCalls,
		"successful_calls": successfulCalls,
		"failed_calls":     failedCalls,
		"subscriptions":    summaries,
	}

	c.JSON(http.StatusOK, response)
}

// GetAPIUsage returns usage analytics for an API (for creators)
func (h *Handler) GetAPIUsage(c *gin.Context) {
	apiID := c.Param("id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	// TODO: Validate that the user owns this API
	userType, _ := c.Get("user_type")
	if userType != "creator" && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse date range
	start, end, err := parseDateRange(startStr, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API usage summary
	summary, err := h.aggregationStore.GetAPIUsageSummary(apiID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API usage"})
		return
	}

	// Get detailed usage records for analytics
	records, err := h.usageStore.GetUsageByAPI(apiID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage details"})
		return
	}

	// Calculate additional analytics
	analytics := calculateAnalytics(records)

	response := gin.H{
		"api_id":       apiID,
		"period_start": start,
		"period_end":   end,
		"summary":      summary,
		"analytics":    analytics,
	}

	c.JSON(http.StatusOK, response)
}

// GetSubscriptionUsage returns usage for a specific subscription
func (h *Handler) GetSubscriptionUsage(c *gin.Context) {
	subscriptionID := c.Param("id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	// Parse date range
	start, end, err := parseDateRange(startStr, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get usage records
	records, err := h.usageStore.GetUsageBySubscription(subscriptionID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription usage"})
		return
	}

	// Get summary
	summary, err := h.aggregationStore.GetUsageSummary(subscriptionID, start, end)
	if err != nil {
		// If no summary, calculate from records
		summary = calculateSummaryFromRecords(subscriptionID, records, start, end)
	}

	response := gin.H{
		"subscription_id": subscriptionID,
		"period_start":    start,
		"period_end":      end,
		"summary":         summary,
		"recent_calls":    records[:min(10, len(records))], // Last 10 calls
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startStr == "" {
		// Default to current month
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return start, end, err
		}
	}

	if endStr == "" {
		// Default to end of current day
		end = time.Now()
	} else {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return start, end, err
		}
		// Set to end of day
		end = end.Add(24*time.Hour - 1*time.Second)
	}

	return start, end, nil
}

func isCurrentPeriod(start, end time.Time) bool {
	now := time.Now()
	return now.After(start) && now.Before(end)
}

func mergeRealtimeData(summary *store.UsageSummary, realtimeData map[string]string) *store.UsageSummary {
	// Merge real-time Redis data with aggregated data
	// This is a simplified implementation
	return summary
}

func calculateAnalytics(records []*store.UsageRecord) gin.H {
	if len(records) == 0 {
		return gin.H{
			"avg_response_time": 0,
			"error_rate":        0,
			"peak_hour":         nil,
		}
	}

	var totalResponseTime int64
	var errorCount int
	hourCounts := make(map[int]int)

	for _, record := range records {
		totalResponseTime += record.ResponseTimeMs
		if record.StatusCode >= 400 {
			errorCount++
		}
		hour := record.Timestamp.Hour()
		hourCounts[hour]++
	}

	// Find peak hour
	var peakHour int
	var peakCount int
	for hour, count := range hourCounts {
		if count > peakCount {
			peakHour = hour
			peakCount = count
		}
	}

	return gin.H{
		"avg_response_time": totalResponseTime / int64(len(records)),
		"error_rate":        float64(errorCount) / float64(len(records)),
		"peak_hour":         peakHour,
		"peak_hour_calls":   peakCount,
	}
}

func calculateSummaryFromRecords(subscriptionID string, records []*store.UsageRecord, start, end time.Time) *store.UsageSummary {
	summary := &store.UsageSummary{
		SubscriptionID: subscriptionID,
		PeriodStart:    start,
		PeriodEnd:      end,
		TotalCalls:     int64(len(records)),
		EndpointUsage:  make(map[string]int64),
	}

	for _, record := range records {
		if record.StatusCode < 400 {
			summary.SuccessfulCalls++
		} else {
			summary.FailedCalls++
		}
		summary.TotalResponseTime += record.ResponseTimeMs
		summary.TotalRequestSize += record.RequestSizeBytes
		summary.TotalResponseSize += record.ResponseSizeBytes
		summary.EndpointUsage[record.Endpoint]++
	}

	return summary
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
