package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/gateway/proxy"
)

// ProxyToFunction handles proxying requests to creator functions
func ProxyToFunction(proxyHandler *proxy.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get creator and API name from URL
		creator := c.Param("creator")
		apiName := c.Param("apiName")
		
		if creator == "" || apiName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid API path",
				"code":  "INVALID_PATH",
			})
			return
		}
		
		// Validate content type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !proxyHandler.ValidateContentType(contentType) {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Unsupported content type",
					"code":  "UNSUPPORTED_CONTENT_TYPE",
				})
				return
			}
		}
		
		// Get the target URL for the creator's function
		targetURL, err := proxyHandler.GetCreatorFunctionURL(creator, apiName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to determine target URL",
				"code":  "TARGET_URL_ERROR",
				"details": err.Error(),
			})
			return
		}
		
		// Add debugging information in development
		if gin.Mode() == gin.DebugMode {
			fmt.Printf("Proxying request to: %s\n", targetURL)
		}
		
		// Proxy the request
		proxyHandler.ProxyRequest(c, targetURL)
	}
}
