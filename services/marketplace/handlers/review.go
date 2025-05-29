package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"marketplace/auth"
	"marketplace/middleware"
	"marketplace/store"
)

// ReviewHandler handles review-related requests
type ReviewHandler struct {
	reviewStore   *store.ReviewStore
	apiStore      *store.APIStore
	consumerStore *store.ConsumerStore
	db            *sql.DB
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewStore *store.ReviewStore, apiStore *store.APIStore, db *sql.DB) *ReviewHandler {
	return &ReviewHandler{
		reviewStore:   reviewStore,
		apiStore:      apiStore,
		consumerStore: store.NewConsumerStore(db),
		db:            db,
	}
}

// GetAPIReviews handles GET /api/v1/marketplace/apis/:id/reviews
func (h *ReviewHandler) GetAPIReviews(c *gin.Context) {
	apiID := c.Param("id")

	// Parse query parameters
	params := store.ListReviewsParams{
		APIID:        apiID,
		Sort:         c.Query("sort"),
		VerifiedOnly: c.Query("verified_only") == "true",
		Page:         1,
		Limit:        10,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	// Get current user ID if authenticated
	currentUserID := ""
	// Try to get consumer ID from context (if authenticated)
	user, exists := auth.GetUserFromContext(c)
	if exists && user != nil {
		// Convert user ID to consumer ID
		consumerID, err := h.consumerStore.GetConsumerID(user.UserID)
		if err == nil {
			currentUserID = consumerID
		} else {
			// User might not have a consumer record yet (not subscribed to any APIs)
			currentUserID = ""
		}
	}

	// Get reviews
	reviews, total, err := h.reviewStore.GetAPIReviews(params, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"total":   total,
		"page":    params.Page,
		"limit":   params.Limit,
	})
}

// GetReviewStats handles GET /api/v1/marketplace/apis/:id/reviews/stats
func (h *ReviewHandler) GetReviewStats(c *gin.Context) {
	apiID := c.Param("id")

	stats, err := h.reviewStore.GetReviewStats(apiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SubmitReview handles POST /api/v1/marketplace/apis/:id/reviews
func (h *ReviewHandler) SubmitReview(c *gin.Context) {
	apiID := c.Param("id")

	// Get user from context
	user, exists := auth.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get or create consumer ID from user ID
	consumerID, err := h.consumerStore.GetOrCreateConsumerID(user.UserID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to identify consumer"})
		return
	}

	// Parse request
	var req store.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Submit review
	review, err := h.reviewStore.SubmitReview(apiID, consumerID, req)
	if err != nil {
		if err.Error() == "you have already reviewed this API" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger reindexing of the API to update stats
	// This would be done via an event system in production
	// For now, we'll skip this

	c.JSON(http.StatusCreated, review)
}

// VoteOnReview handles POST /api/v1/marketplace/reviews/:id/vote
func (h *ReviewHandler) VoteOnReview(c *gin.Context) {
	reviewID := c.Param("id")

	// Get user from context
	user, exists := auth.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get or create consumer ID from user ID
	consumerID, err := h.consumerStore.GetOrCreateConsumerID(user.UserID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to identify consumer"})
		return
	}

	// Parse request
	var req struct {
		IsHelpful bool `json:"is_helpful"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Vote on review
	err = h.reviewStore.VoteOnReview(reviewID, consumerID, req.IsHelpful)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}

// RespondToReview handles POST /api/v1/marketplace/reviews/:id/response
func (h *ReviewHandler) RespondToReview(c *gin.Context) {
	reviewID := c.Param("id")

	// Get creator ID from context
	creatorID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req struct {
		Response string `json:"response" binding:"required,max=2000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add response
	err := h.reviewStore.RespondToReview(reviewID, creatorID, req.Response)
	if err != nil {
		if err.Error() == "review not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized: you are not the creator of this API" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response added successfully"})
}
