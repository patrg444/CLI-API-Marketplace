package store

import (
	"database/sql"
	"fmt"
)

// APIStore handles API-related database operations
type APIStore struct {
	db *sql.DB
}

// NewAPIStore creates a new API store
func NewAPIStore(db *sql.DB) *APIStore {
	return &APIStore{db: db}
}

// CheckAPIOwnership verifies if a user owns an API
func (s *APIStore) CheckAPIOwnership(userID, apiID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM apis 
		WHERE id = $1 AND creator_id = $2 AND deleted_at IS NULL
	`
	
	err := s.db.QueryRow(query, apiID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check API ownership: %w", err)
	}
	
	return count > 0, nil
}

// GetAPICreatorID returns the creator ID for an API
func (s *APIStore) GetAPICreatorID(apiID string) (string, error) {
	var creatorID string
	query := `
		SELECT creator_id 
		FROM apis 
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	err := s.db.QueryRow(query, apiID).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("API not found")
		}
		return "", fmt.Errorf("failed to get API creator: %w", err)
	}
	
	return creatorID, nil
}

// GetAPIsByCreator returns all APIs created by a user
func (s *APIStore) GetAPIsByCreator(creatorID string) ([]string, error) {
	query := `
		SELECT id 
		FROM apis 
		WHERE creator_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	rows, err := s.db.Query(query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get APIs by creator: %w", err)
	}
	defer rows.Close()
	
	var apiIDs []string
	for rows.Next() {
		var apiID string
		if err := rows.Scan(&apiID); err != nil {
			return nil, fmt.Errorf("failed to scan API ID: %w", err)
		}
		apiIDs = append(apiIDs, apiID)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return apiIDs, nil
}

// IsAPIPublished checks if an API is published
func (s *APIStore) IsAPIPublished(apiID string) (bool, error) {
	var published bool
	query := `
		SELECT is_published 
		FROM apis 
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	err := s.db.QueryRow(query, apiID).Scan(&published)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("API not found")
		}
		return false, fmt.Errorf("failed to check API published status: %w", err)
	}
	
	return published, nil
}
