package metering

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/redis/go-redis/v9"
)

// Mock Redis client
type MockRedisClient struct {
	mock.Mock
	data map[string]interface{}
	mu   sync.Mutex
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]interface{}),
	}
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	val, exists := m.data[key]
	if !exists {
		m.data[key] = int64(1)
		return redis.NewIntResult(1, nil)
	}
	
	intVal := val.(int64) + 1
	m.data[key] = intVal
	return redis.NewIntResult(intVal, nil)
}

func (m *MockRedisClient) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	val, exists := m.data[key]
	if !exists {
		m.data[key] = value
		return redis.NewIntResult(value, nil)
	}
	
	intVal := val.(int64) + value
	m.data[key] = intVal
	return redis.NewIntResult(intVal, nil)
}

// Mock database
type MockDB struct {
	mock.Mock
}

func (m *MockDB) SaveUsageRecord(ctx context.Context, record *UsageRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockDB) GetUsageByPeriod(ctx context.Context, apiKey string, period string) (*UsageRecord, error) {
	args := m.Called(ctx, apiKey, period)
	return args.Get(0).(*UsageRecord), args.Error(1)
}

func TestRecordAPICall(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("records single API call", func(t *testing.T) {
		event := &APICallEvent{
			APIKey:     "key_test123",
			Endpoint:   "/api/data",
			Method:     "GET",
			StatusCode: 200,
			Duration:   150 * time.Millisecond,
			Timestamp:  time.Now(),
		}

		err := service.RecordAPICall(ctx, event)
		assert.NoError(t, err)

		// Verify Redis increments
		key := service.getUsageKey(event.APIKey, event.Timestamp)
		val := mockRedis.data[key]
		assert.Equal(t, int64(1), val)
	})

	t.Run("handles concurrent API calls", func(t *testing.T) {
		apiKey := "key_concurrent"
		numCalls := 100
		var wg sync.WaitGroup

		for i := 0; i < numCalls; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				event := &APICallEvent{
					APIKey:     apiKey,
					Endpoint:   "/api/data",
					Method:     "GET",
					StatusCode: 200,
					Duration:   100 * time.Millisecond,
					Timestamp:  time.Now(),
				}
				service.RecordAPICall(ctx, event)
			}()
		}

		wg.Wait()

		// Verify all calls were recorded
		key := service.getUsageKey(apiKey, time.Now())
		val := mockRedis.data[key]
		assert.Equal(t, int64(numCalls), val)
	})

	t.Run("tracks endpoint-specific usage", func(t *testing.T) {
		apiKey := "key_endpoints"
		endpoints := []string{"/api/users", "/api/posts", "/api/comments"}
		
		for _, endpoint := range endpoints {
			for i := 0; i < 10; i++ {
				event := &APICallEvent{
					APIKey:     apiKey,
					Endpoint:   endpoint,
					Method:     "GET",
					StatusCode: 200,
					Duration:   100 * time.Millisecond,
					Timestamp:  time.Now(),
				}
				service.RecordAPICall(ctx, event)
			}
		}

		// Verify endpoint metrics
		for _, endpoint := range endpoints {
			key := service.getEndpointKey(apiKey, endpoint, time.Now())
			val := mockRedis.data[key]
			assert.Equal(t, int64(10), val)
		}
	})

	t.Run("tracks error rates", func(t *testing.T) {
		apiKey := "key_errors"
		
		// Record successful calls
		for i := 0; i < 90; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		// Record error calls
		for i := 0; i < 10; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 500,
				Duration:   50 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		// Calculate error rate
		errorRate := service.GetErrorRate(ctx, apiKey, time.Now())
		assert.InDelta(t, 0.1, errorRate, 0.01) // 10% error rate
	})
}

func TestUsageAggregation(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("aggregates hourly usage", func(t *testing.T) {
		apiKey := "key_hourly"
		now := time.Now()
		
		// Simulate calls throughout an hour
		for i := 0; i < 60; i++ {
			timestamp := now.Add(time.Duration(i) * time.Minute)
			for j := 0; j < 10; j++ {
				event := &APICallEvent{
					APIKey:     apiKey,
					Endpoint:   "/api/data",
					Method:     "GET",
					StatusCode: 200,
					Duration:   100 * time.Millisecond,
					Timestamp:  timestamp,
				}
				service.RecordAPICall(ctx, event)
			}
		}

		// Aggregate hourly data
		hourlyUsage := service.AggregateHourlyUsage(ctx, apiKey, now)
		assert.Equal(t, 600, hourlyUsage.TotalCalls) // 60 minutes * 10 calls
		assert.Equal(t, apiKey, hourlyUsage.APIKey)
	})

	t.Run("aggregates daily usage with quota tracking", func(t *testing.T) {
		apiKey := "key_daily"
		subscription := &Subscription{
			APIKey:    apiKey,
			DailyQuota: 1000,
			Plan:      "starter",
		}
		
		// Record 800 calls
		for i := 0; i < 800; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		dailyUsage := service.GetDailyUsage(ctx, apiKey, time.Now())
		quotaStatus := service.CheckQuota(ctx, subscription, dailyUsage)
		
		assert.Equal(t, 800, dailyUsage.TotalCalls)
		assert.True(t, quotaStatus.WithinQuota)
		assert.Equal(t, 200, quotaStatus.Remaining)
		assert.Equal(t, 80.0, quotaStatus.PercentUsed)
	})

	t.Run("handles quota exceeded", func(t *testing.T) {
		apiKey := "key_exceeded"
		subscription := &Subscription{
			APIKey:     apiKey,
			DailyQuota: 100,
			Plan:       "free",
		}
		
		// Record 150 calls (exceeding quota)
		for i := 0; i < 150; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		dailyUsage := service.GetDailyUsage(ctx, apiKey, time.Now())
		quotaStatus := service.CheckQuota(ctx, subscription, dailyUsage)
		
		assert.False(t, quotaStatus.WithinQuota)
		assert.Equal(t, -50, quotaStatus.Remaining)
		assert.Equal(t, 150.0, quotaStatus.PercentUsed)
		assert.True(t, quotaStatus.ShouldBlock)
	})
}

func TestPerformanceMetrics(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("calculates average response time", func(t *testing.T) {
		apiKey := "key_performance"
		durations := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			150 * time.Millisecond,
			250 * time.Millisecond,
			300 * time.Millisecond,
		}
		
		for _, duration := range durations {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   duration,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		metrics := service.GetPerformanceMetrics(ctx, apiKey, time.Now())
		assert.InDelta(t, 200, metrics.AvgResponseTime.Milliseconds(), 10)
		assert.Equal(t, 100, metrics.MinResponseTime.Milliseconds())
		assert.Equal(t, 300, metrics.MaxResponseTime.Milliseconds())
	})

	t.Run("calculates percentiles", func(t *testing.T) {
		apiKey := "key_percentiles"
		
		// Generate 1000 calls with varying response times
		for i := 0; i < 1000; i++ {
			duration := time.Duration(i%200+50) * time.Millisecond
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   duration,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		percentiles := service.GetResponseTimePercentiles(ctx, apiKey, time.Now())
		assert.True(t, percentiles.P50 < percentiles.P95)
		assert.True(t, percentiles.P95 < percentiles.P99)
		assert.True(t, percentiles.P99 <= 250*time.Millisecond)
	})
}

func TestGeographicDistribution(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("tracks usage by geographic region", func(t *testing.T) {
		apiKey := "key_geographic"
		regions := map[string]int{
			"us-east-1": 500,
			"eu-west-1": 300,
			"ap-south-1": 200,
		}
		
		for region, count := range regions {
			for i := 0; i < count; i++ {
				event := &APICallEvent{
					APIKey:     apiKey,
					Endpoint:   "/api/data",
					Method:     "GET",
					StatusCode: 200,
					Duration:   100 * time.Millisecond,
					Timestamp:  time.Now(),
					Region:     region,
					ClientIP:   "1.2.3.4",
				}
				service.RecordAPICall(ctx, event)
			}
		}

		geoStats := service.GetGeographicDistribution(ctx, apiKey, time.Now())
		assert.Equal(t, 500, geoStats.Regions["us-east-1"])
		assert.Equal(t, 300, geoStats.Regions["eu-west-1"])
		assert.Equal(t, 200, geoStats.Regions["ap-south-1"])
		assert.Equal(t, 1000, geoStats.TotalCalls)
	})
}

func TestRealTimeAlerts(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	alertChan := make(chan *Alert, 10)
	
	service := &MeteringService{
		redis:     mockRedis,
		db:        mockDB,
		alertChan: alertChan,
	}

	t.Run("triggers high error rate alert", func(t *testing.T) {
		apiKey := "key_high_errors"
		
		// Generate high error rate
		for i := 0; i < 100; i++ {
			statusCode := 200
			if i < 30 { // 30% error rate
				statusCode = 500
			}
			
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: statusCode,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		// Check for alert
		select {
		case alert := <-alertChan:
			assert.Equal(t, "HIGH_ERROR_RATE", alert.Type)
			assert.Equal(t, apiKey, alert.APIKey)
			assert.True(t, alert.Value >= 0.25) // At least 25% error rate
		case <-time.After(1 * time.Second):
			t.Fatal("Expected alert not received")
		}
	})

	t.Run("triggers quota warning alert", func(t *testing.T) {
		apiKey := "key_quota_warning"
		subscription := &Subscription{
			APIKey:     apiKey,
			DailyQuota: 100,
			Plan:       "starter",
		}
		
		// Use 85% of quota
		for i := 0; i < 85; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		service.CheckQuotaAlert(ctx, subscription)
		
		select {
		case alert := <-alertChan:
			assert.Equal(t, "QUOTA_WARNING", alert.Type)
			assert.Equal(t, apiKey, alert.APIKey)
			assert.Equal(t, 85.0, alert.Value)
		case <-time.After(1 * time.Second):
			t.Fatal("Expected quota warning not received")
		}
	})

	t.Run("triggers spike detection alert", func(t *testing.T) {
		apiKey := "key_spike"
		
		// Normal traffic baseline (10 calls per minute)
		baseTime := time.Now().Add(-1 * time.Hour)
		for i := 0; i < 60; i++ {
			for j := 0; j < 10; j++ {
				event := &APICallEvent{
					APIKey:     apiKey,
					Endpoint:   "/api/data",
					Method:     "GET",
					StatusCode: 200,
					Duration:   100 * time.Millisecond,
					Timestamp:  baseTime.Add(time.Duration(i) * time.Minute),
				}
				service.RecordAPICall(ctx, event)
			}
		}

		// Sudden spike (100 calls in 1 minute)
		for i := 0; i < 100; i++ {
			event := &APICallEvent{
				APIKey:     apiKey,
				Endpoint:   "/api/data",
				Method:     "GET",
				StatusCode: 200,
				Duration:   100 * time.Millisecond,
				Timestamp:  time.Now(),
			}
			service.RecordAPICall(ctx, event)
		}

		service.DetectTrafficSpike(ctx, apiKey)
		
		select {
		case alert := <-alertChan:
			assert.Equal(t, "TRAFFIC_SPIKE", alert.Type)
			assert.Equal(t, apiKey, alert.APIKey)
			assert.True(t, alert.Value > 5.0) // At least 5x normal traffic
		case <-time.After(1 * time.Second):
			t.Fatal("Expected spike alert not received")
		}
	})
}

func TestDataPersistence(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("persists aggregated data to database", func(t *testing.T) {
		apiKey := "key_persist"
		period := "2023-06-15"
		
		expectedRecord := &UsageRecord{
			APIKey:         apiKey,
			Period:         period,
			TotalCalls:     1000,
			SuccessfulCalls: 950,
			FailedCalls:    50,
			TotalDuration:  100 * time.Second,
			UniqueIPs:      150,
		}
		
		mockDB.On("SaveUsageRecord", ctx, mock.MatchedBy(func(record *UsageRecord) bool {
			return record.APIKey == apiKey && record.Period == period
		})).Return(nil)

		err := service.PersistUsageData(ctx, expectedRecord)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("handles persistence failures with retry", func(t *testing.T) {
		apiKey := "key_retry"
		record := &UsageRecord{
			APIKey: apiKey,
			Period: "2023-06-15",
		}
		
		// First call fails, second succeeds
		mockDB.On("SaveUsageRecord", ctx, record).Return(errors.New("database error")).Once()
		mockDB.On("SaveUsageRecord", ctx, record).Return(nil).Once()

		err := service.PersistUsageDataWithRetry(ctx, record, 3)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})
}

func TestConcurrencyAndPerformance(t *testing.T) {
	ctx := context.Background()
	mockRedis := NewMockRedisClient()
	mockDB := new(MockDB)
	
	service := &MeteringService{
		redis: mockRedis,
		db:    mockDB,
	}

	t.Run("handles high throughput", func(t *testing.T) {
		numGoroutines := 100
		callsPerGoroutine := 1000
		apiKey := "key_performance"
		
		start := time.Now()
		var wg sync.WaitGroup
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < callsPerGoroutine; j++ {
					event := &APICallEvent{
						APIKey:     apiKey,
						Endpoint:   "/api/data",
						Method:     "GET",
						StatusCode: 200,
						Duration:   100 * time.Millisecond,
						Timestamp:  time.Now(),
					}
					service.RecordAPICall(ctx, event)
				}
			}()
		}
		
		wg.Wait()
		duration := time.Since(start)
		
		// Verify all calls recorded
		key := service.getUsageKey(apiKey, time.Now())
		totalCalls := mockRedis.data[key].(int64)
		assert.Equal(t, int64(numGoroutines*callsPerGoroutine), totalCalls)
		
		// Performance check - should handle at least 10k calls/second
		callsPerSecond := float64(totalCalls) / duration.Seconds()
		assert.Greater(t, callsPerSecond, 10000.0)
	})
}

// Helper function to get usage key
func (s *MeteringService) getUsageKey(apiKey string, timestamp time.Time) string {
	return fmt.Sprintf("usage:%s:%s", apiKey, timestamp.Format("2006-01-02:15"))
}

// Helper function to get endpoint key
func (s *MeteringService) getEndpointKey(apiKey, endpoint string, timestamp time.Time) string {
	return fmt.Sprintf("endpoint:%s:%s:%s", apiKey, endpoint, timestamp.Format("2006-01-02:15"))
}