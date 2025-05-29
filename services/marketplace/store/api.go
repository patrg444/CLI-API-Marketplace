package store

import (
    "database/sql"
)

type APIStore struct {
    db *sql.DB
}

func NewAPIStore(db *sql.DB) *APIStore {
    return &APIStore{db: db}
}

func (s *APIStore) CheckAPIOwnership(userID, apiID string) (bool, error) {
    // Mock implementation - always return true for now
    return true, nil
}

type API struct {
    ID          string
    Name        string
    Description string
    CreatorID   string
}
