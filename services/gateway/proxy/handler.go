package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler manages proxying requests to creator functions
type Handler struct {
	meteringServiceURL string
	httpClient         *http.Client
}

// NewHandler creates a new proxy handler
func NewHandler(meteringServiceURL string) *Handler {
	return &Handler{
		meteringServiceURL: meteringServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest forwards a request to the appropriate creator function
func (h *Handler) ProxyRequest(c *gin.Context, targetURL string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid target URL",
			"code":  "INVALID_TARGET",
		})
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the director to preserve the original path
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Preserve the original path after the API name
		path := c.Param("path")
		if path != "" {
			req.URL.Path = path
		}
		
		// Add consumer information to headers for the creator function
		if consumerID, exists := c.Get("consumer_id"); exists {
			req.Header.Set("X-Consumer-ID", consumerID.(string))
		}
		if subscriptionID, exists := c.Get("subscription_id"); exists {
			req.Header.Set("X-Subscription-ID", subscriptionID.(string))
		}
		
		// Remove sensitive headers
		req.Header.Del("X-API-Key")
		req.Header.Del("Authorization")
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Failed to reach API endpoint",
			"code":  "GATEWAY_ERROR",
			"details": err.Error(),
		})
	}

	// Serve the request
	proxy.ServeHTTP(c.Writer, c.Request)
}

// GetCreatorFunctionURL determines the URL for a creator's function
func (h *Handler) GetCreatorFunctionURL(creator, apiName string) (string, error) {
	// In production, this would query the deployment service or database
	// to get the actual URL of the deployed function
	// For now, we'll use a simple pattern
	
	// Format: http://<api-name>-<creator>.api-direct.svc.cluster.local:8080
	// This assumes functions are deployed as Kubernetes services
	functionURL := fmt.Sprintf("http://%s-%s.api-direct.svc.cluster.local:8080", 
		strings.ToLower(apiName), 
		strings.ToLower(creator))
	
	return functionURL, nil
}

// ValidateContentType checks if the content type is acceptable
func (h *Handler) ValidateContentType(contentType string) bool {
	acceptableTypes := []string{
		"application/json",
		"application/xml",
		"text/plain",
		"text/html",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}
	
	// Extract the main type without parameters
	mainType := strings.Split(contentType, ";")[0]
	mainType = strings.TrimSpace(strings.ToLower(mainType))
	
	for _, acceptable := range acceptableTypes {
		if mainType == acceptable {
			return true
		}
	}
	
	return false
}

// StreamResponse streams the response from the creator function to the client
func (h *Handler) StreamResponse(c *gin.Context, resp *http.Response) {
	defer resp.Body.Close()
	
	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	
	// Set status code
	c.Status(resp.StatusCode)
	
	// Stream body
	_, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		// Log the error but don't try to write a response
		// as headers have already been sent
		fmt.Printf("Error streaming response: %v\n", err)
	}
}
