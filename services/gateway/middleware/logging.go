package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UsageLogRequest struct {
	SubscriptionID     string `json:"subscription_id"`
	APIKeyID           string `json:"api_key_id"`
	Timestamp          string `json:"timestamp"`
	Endpoint           string `json:"endpoint"`
	Method             string `json:"method"`
	StatusCode         int    `json:"status_code"`
	ResponseTimeMs     int64  `json:"response_time_ms"`
	RequestSizeBytes   int64  `json:"request_size_bytes"`
	ResponseSizeBytes  int64  `json:"response_size_bytes"`
}

// Custom response writer to capture response size and status code
type responseWriter struct {
	gin.ResponseWriter
	size       int64
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size += int64(n)
	return n, err
}

// LogRequest middleware sends usage data to the Metering Service
func LogRequest(meteringServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Get request size
		requestSize := c.Request.ContentLength
		if requestSize == -1 {
			requestSize = 0
		}

		// Wrap response writer to capture response details
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = rw

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(startTime).Milliseconds()

		// Get data from context
		subscriptionID, _ := c.Get("subscription_id")
		apiKeyID, _ := c.Get("api_key_id")

		subscriptionIDStr, _ := subscriptionID.(string)
		apiKeyIDStr, _ := apiKeyID.(string)

		// Only log if we have valid subscription and API key IDs
		if subscriptionIDStr != "" && apiKeyIDStr != "" {
			// Prepare usage log
			usageLog := UsageLogRequest{
				SubscriptionID:    subscriptionIDStr,
				APIKeyID:          apiKeyIDStr,
				Timestamp:         time.Now().UTC().Format(time.RFC3339),
				Endpoint:          c.Request.URL.Path,
				Method:            c.Request.Method,
				StatusCode:        rw.statusCode,
				ResponseTimeMs:    responseTime,
				RequestSizeBytes:  requestSize,
				ResponseSizeBytes: rw.size,
			}

			// Send to metering service asynchronously
			go func() {
				jsonData, err := json.Marshal(usageLog)
				if err != nil {
					fmt.Printf("Failed to marshal usage log: %v\n", err)
					return
				}

				resp, err := http.Post(
					fmt.Sprintf("%s/api/v1/usage", meteringServiceURL),
					"application/json",
					bytes.NewBuffer(jsonData),
				)
				if err != nil {
					fmt.Printf("Failed to send usage log: %v\n", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
					body, _ := io.ReadAll(resp.Body)
					fmt.Printf("Metering service returned error: %d - %s\n", resp.StatusCode, string(body))
				}
			}()
		}
	}
}
