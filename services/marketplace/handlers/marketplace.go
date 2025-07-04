package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"marketplace/elasticsearch"
	"marketplace/indexer"
	"marketplace/store"
)

// MarketplaceHandler handles marketplace-related requests
type MarketplaceHandler struct {
	apiStore      *store.APIStore
	searchService *elasticsearch.SearchService
	indexer       *indexer.APIIndexer
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(apiStore *store.APIStore, searchService *elasticsearch.SearchService, indexer *indexer.APIIndexer) *MarketplaceHandler {
	return &MarketplaceHandler{
		apiStore:      apiStore,
		searchService: searchService,
		indexer:       indexer,
	}
}

// ListAPIs handles GET /api/v1/marketplace/apis
func (h *MarketplaceHandler) ListAPIs(c *gin.Context) {
	// Parse query parameters
	params := store.ListParams{
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Page:     1,
		Limit:    12,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 && limit <= 100 {
		params.Limit = limit
	}

	// Get APIs from database
	apis, total, err := h.apiStore.ListAPIs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apis":  apis,
		"total": total,
		"page":  params.Page,
		"limit": params.Limit,
	})
}

// GetAPI handles GET /api/v1/marketplace/apis/:id
func (h *MarketplaceHandler) GetAPI(c *gin.Context) {
	apiID := c.Param("id")

	api, err := h.apiStore.GetAPI(apiID)
	if err != nil {
		if err.Error() == "API not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, api)
}

// GetAPIDocumentation handles GET /api/v1/marketplace/apis/:id/documentation
func (h *MarketplaceHandler) GetAPIDocumentation(c *gin.Context) {
	apiID := c.Param("id")

	doc, err := h.apiStore.GetDocumentation(apiID)
	if err != nil {
		if err.Error() == "documentation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Documentation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// SearchAPIs handles POST /api/v1/marketplace/search
func (h *MarketplaceHandler) SearchAPIs(c *gin.Context) {
	var req elasticsearch.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Perform search
	results, err := h.searchService.Search(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSearchSuggestions handles GET /api/v1/marketplace/search/suggestions
func (h *MarketplaceHandler) GetSearchSuggestions(c *gin.Context) {
	prefix := c.Query("q")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	suggestions, err := h.searchService.GetSuggestions(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
	})
}

// ReindexAll handles POST /api/v1/marketplace/admin/reindex
func (h *MarketplaceHandler) ReindexAll(c *gin.Context) {
	// Start reindexing in background
	go func() {
		if err := h.indexer.ReindexAll(); err != nil {
			// Log error
			// In production, this would be logged to a monitoring service
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Reindexing started",
	})
}

// IndexAPI handles POST /api/v1/marketplace/admin/index/:id
func (h *MarketplaceHandler) IndexAPI(c *gin.Context) {
	apiID := c.Param("id")

	// Index API
	if err := h.indexer.IndexAPI(apiID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API indexed successfully",
	})
}

// PublishAPI handles PUT /api/v1/marketplace/apis/:id/publish
func (h *MarketplaceHandler) PublishAPI(c *gin.Context) {
	apiID := c.Param("id")

	var req struct {
		IsPublished bool   `json:"is_published"`
		Description string `json:"description,omitempty"`
		Category    string `json:"category,omitempty"`
		Tags        []string `json:"tags,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user owns this API
	isOwner, err := h.apiStore.CheckAPIOwnership(userID.(string), apiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify API ownership"})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to publish this API"})
		return
	}

	// Update API marketplace status
	err = h.apiStore.UpdateMarketplaceStatus(apiID, req.IsPublished, req.Description, req.Category, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Index/remove from Elasticsearch
	if req.IsPublished {
		if err := h.indexer.IndexAPI(apiID); err != nil {
			// Log error but don't fail the request
			// The API is published but search indexing failed
		}
	} else {
		if err := h.indexer.RemoveAPI(apiID); err != nil {
			// Log error but don't fail the request
		}
	}

	// Get updated API info
	api, err := h.apiStore.GetAPI(apiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated API info"})
		return
	}

	response := gin.H{
		"message": "API marketplace status updated successfully",
		"api_id":  apiID,
		"is_published": req.IsPublished,
	}

	if req.IsPublished {
		response["marketplace_url"] = fmt.Sprintf("https://marketplace.apidirect.com/api/%s", apiID)
	}

	c.JSON(http.StatusOK, response)
}
