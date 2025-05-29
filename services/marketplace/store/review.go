package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReviewStore handles review-related database operations
type ReviewStore struct {
	db *sql.DB
}

// NewReviewStore creates a new review store
func NewReviewStore(db *sql.DB) *ReviewStore {
	return &ReviewStore{db: db}
}

// Review represents an API review
type Review struct {
	ID                 string     `json:"id"`
	APIID              string     `json:"api_id"`
	ConsumerID         string     `json:"consumer_id"`
	Rating             int        `json:"rating"`
	Title              string     `json:"title"`
	Comment            string     `json:"comment"`
	HelpfulVotes       int        `json:"helpful_votes"`
	TotalVotes         int        `json:"total_votes"`
	CreatorResponse    *string    `json:"creator_response"`
	ResponseDate       *time.Time `json:"response_date"`
	IsVerifiedPurchase bool       `json:"is_verified_purchase"`
	CreatedAt          time.Time  `json:"created_at"`
	
	// Joined data
	ConsumerName string `json:"consumer_name,omitempty"`
	UserVote     *bool  `json:"user_vote,omitempty"` // true=helpful, false=not helpful, nil=no vote
}

// ReviewStats represents aggregated review statistics
type ReviewStats struct {
	AverageRating      float32 `json:"average_rating"`
	TotalReviews       int     `json:"total_reviews"`
	RatingDistribution map[int]int `json:"rating_distribution"`
}

// SubmitReviewRequest represents a review submission
type SubmitReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Title   string `json:"title" binding:"max=255"`
	Comment string `json:"comment" binding:"required,max=5000"`
}

// ListReviewsParams represents parameters for listing reviews
type ListReviewsParams struct {
	APIID        string
	Sort         string // most_recent, most_helpful, highest_rating, lowest_rating
	VerifiedOnly bool
	Page         int
	Limit        int
}

// GetAPIReviews retrieves reviews for an API
func (s *ReviewStore) GetAPIReviews(params ListReviewsParams, currentUserID string) ([]*Review, int, error) {
	// Base query
	query := `
		SELECT 
			r.id, r.api_id, r.consumer_id, r.rating, r.title, r.comment,
			r.helpful_votes, r.total_votes, r.creator_response, r.response_date,
			r.is_verified_purchase, r.created_at,
			c.company_name as consumer_name,
			v.is_helpful as user_vote
		FROM api_reviews r
		JOIN consumers c ON r.consumer_id = c.id
		LEFT JOIN review_votes v ON r.id = v.review_id AND v.consumer_id = $1
		WHERE r.api_id = $2
	`
	
	countQuery := `
		SELECT COUNT(*) FROM api_reviews 
		WHERE api_id = $1
	`
	
	args := []interface{}{params.APIID}
	countArgs := []interface{}{params.APIID}
	argCount := 2

	// Add verified filter
	if params.VerifiedOnly {
		query += " AND r.is_verified_purchase = true"
		countQuery += " AND is_verified_purchase = true"
	}

	// Get total count
	var totalCount int
	err := s.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting reviews: %w", err)
	}

	// Add ordering
	switch params.Sort {
	case "most_helpful":
		query += " ORDER BY r.helpful_votes DESC, r.created_at DESC"
	case "highest_rating":
		query += " ORDER BY r.rating DESC, r.created_at DESC"
	case "lowest_rating":
		query += " ORDER BY r.rating ASC, r.created_at DESC"
	case "most_recent":
		fallthrough
	default:
		query += " ORDER BY r.created_at DESC"
	}
	
	// Add pagination
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append([]interface{}{currentUserID}, args...)
	args = append(args, params.Limit)
	
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, (params.Page-1)*params.Limit)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying reviews: %w", err)
	}
	defer rows.Close()

	reviews := []*Review{}
	for rows.Next() {
		review := &Review{}
		var userVote sql.NullBool
		
		err := rows.Scan(
			&review.ID, &review.APIID, &review.ConsumerID, &review.Rating,
			&review.Title, &review.Comment, &review.HelpfulVotes, &review.TotalVotes,
			&review.CreatorResponse, &review.ResponseDate, &review.IsVerifiedPurchase,
			&review.CreatedAt, &review.ConsumerName, &userVote,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning review: %w", err)
		}
		
		if userVote.Valid {
			review.UserVote = &userVote.Bool
		}
		
		reviews = append(reviews, review)
	}

	return reviews, totalCount, nil
}

// GetReviewStats retrieves aggregated review statistics for an API
func (s *ReviewStore) GetReviewStats(apiID string) (*ReviewStats, error) {
	query := `
		SELECT 
			COALESCE(average_rating, 0) as average_rating,
			COALESCE(total_reviews, 0) as total_reviews,
			COALESCE(five_star_count, 0) as five_star_count,
			COALESCE(four_star_count, 0) as four_star_count,
			COALESCE(three_star_count, 0) as three_star_count,
			COALESCE(two_star_count, 0) as two_star_count,
			COALESCE(one_star_count, 0) as one_star_count
		FROM api_rating_stats
		WHERE api_id = $1
	`

	stats := &ReviewStats{
		RatingDistribution: make(map[int]int),
	}

	var fiveStar, fourStar, threeStar, twoStar, oneStar int

	err := s.db.QueryRow(query, apiID).Scan(
		&stats.AverageRating,
		&stats.TotalReviews,
		&fiveStar, &fourStar, &threeStar, &twoStar, &oneStar,
	)
	if err == sql.ErrNoRows {
		// No reviews yet
		return stats, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying review stats: %w", err)
	}

	stats.RatingDistribution[5] = fiveStar
	stats.RatingDistribution[4] = fourStar
	stats.RatingDistribution[3] = threeStar
	stats.RatingDistribution[2] = twoStar
	stats.RatingDistribution[1] = oneStar

	return stats, nil
}

// SubmitReview creates a new review
func (s *ReviewStore) SubmitReview(apiID, consumerID string, req SubmitReviewRequest) (*Review, error) {
	// Check if user already reviewed this API
	var existingCount int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM api_reviews WHERE api_id = $1 AND consumer_id = $2",
		apiID, consumerID,
	).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("error checking existing review: %w", err)
	}
	if existingCount > 0 {
		return nil, fmt.Errorf("you have already reviewed this API")
	}

	// Check if this is a verified purchase
	isVerified, err := s.verifyPurchase(consumerID, apiID)
	if err != nil {
		return nil, fmt.Errorf("error verifying purchase: %w", err)
	}

	// Insert review
	review := &Review{
		ID:                 uuid.New().String(),
		APIID:              apiID,
		ConsumerID:         consumerID,
		Rating:             req.Rating,
		Title:              req.Title,
		Comment:            req.Comment,
		IsVerifiedPurchase: isVerified,
		CreatedAt:          time.Now(),
	}

	query := `
		INSERT INTO api_reviews (
			id, api_id, consumer_id, rating, title, comment, is_verified_purchase
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at
	`

	err = s.db.QueryRow(
		query,
		review.ID, review.APIID, review.ConsumerID, review.Rating,
		review.Title, review.Comment, review.IsVerifiedPurchase,
	).Scan(&review.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting review: %w", err)
	}

	return review, nil
}

// VoteOnReview records a helpful/not helpful vote
func (s *ReviewStore) VoteOnReview(reviewID, consumerID string, isHelpful bool) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if user already voted
	var existingVote sql.NullBool
	err = tx.QueryRow(
		"SELECT is_helpful FROM review_votes WHERE review_id = $1 AND consumer_id = $2",
		reviewID, consumerID,
	).Scan(&existingVote)
	
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing vote: %w", err)
	}

	if err == sql.ErrNoRows {
		// No existing vote, insert new one
		_, err = tx.Exec(
			"INSERT INTO review_votes (review_id, consumer_id, is_helpful) VALUES ($1, $2, $3)",
			reviewID, consumerID, isHelpful,
		)
		if err != nil {
			return fmt.Errorf("error inserting vote: %w", err)
		}

		// Update review vote counts
		if isHelpful {
			_, err = tx.Exec(
				"UPDATE api_reviews SET helpful_votes = helpful_votes + 1, total_votes = total_votes + 1 WHERE id = $1",
				reviewID,
			)
		} else {
			_, err = tx.Exec(
				"UPDATE api_reviews SET total_votes = total_votes + 1 WHERE id = $1",
				reviewID,
			)
		}
	} else if existingVote.Bool != isHelpful {
		// User is changing their vote
		_, err = tx.Exec(
			"UPDATE review_votes SET is_helpful = $1 WHERE review_id = $2 AND consumer_id = $3",
			isHelpful, reviewID, consumerID,
		)
		if err != nil {
			return fmt.Errorf("error updating vote: %w", err)
		}

		// Update review vote counts
		if isHelpful {
			_, err = tx.Exec(
				"UPDATE api_reviews SET helpful_votes = helpful_votes + 1 WHERE id = $1",
				reviewID,
			)
		} else {
			_, err = tx.Exec(
				"UPDATE api_reviews SET helpful_votes = helpful_votes - 1 WHERE id = $1",
				reviewID,
			)
		}
	}
	// If vote is the same, do nothing

	if err != nil {
		return fmt.Errorf("error updating vote counts: %w", err)
	}

	return tx.Commit()
}

// RespondToReview adds a creator response to a review
func (s *ReviewStore) RespondToReview(reviewID, creatorID, response string) error {
	// Verify the creator owns the API
	var apiCreatorID string
	err := s.db.QueryRow(`
		SELECT a.user_id 
		FROM api_reviews r
		JOIN apis a ON r.api_id = a.id
		WHERE r.id = $1
	`, reviewID).Scan(&apiCreatorID)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("review not found")
	}
	if err != nil {
		return fmt.Errorf("error verifying creator: %w", err)
	}
	
	if apiCreatorID != creatorID {
		return fmt.Errorf("unauthorized: you are not the creator of this API")
	}

	// Update review with response
	now := time.Now()
	_, err = s.db.Exec(
		"UPDATE api_reviews SET creator_response = $1, response_date = $2 WHERE id = $3",
		response, now, reviewID,
	)
	if err != nil {
		return fmt.Errorf("error updating review: %w", err)
	}

	return nil
}

// verifyPurchase checks if a consumer has an active subscription to an API
func (s *ReviewStore) verifyPurchase(consumerID, apiID string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM subscriptions 
		WHERE consumer_id = $1 
		AND api_id = $2 
		AND status = 'active'
		AND started_at < NOW()
	`, consumerID, apiID).Scan(&count)
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
