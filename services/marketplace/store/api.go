package store

import (
    "database/sql"
    "fmt"
    "strings"
    "time"
)

type APIStore struct {
    db *sql.DB
}

func NewAPIStore(db *sql.DB) *APIStore {
    return &APIStore{db: db}
}

type API struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatorID   string    `json:"creator_id"`
    Category    string    `json:"category"`
    Tags        []string  `json:"tags"`
    IsPublished bool      `json:"is_published"`
    Endpoint    string    `json:"endpoint"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type ListParams struct {
    Category string
    Search   string
    Page     int
    Limit    int
}

func (s *APIStore) CheckAPIOwnership(userID, apiID string) (bool, error) {
    var count int
    query := `SELECT COUNT(*) FROM apis WHERE id = $1 AND creator_id = $2`
    err := s.db.QueryRow(query, apiID, userID).Scan(&count)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (s *APIStore) UpdateMarketplaceStatus(apiID string, isPublished bool, description, category string, tags []string) error {
    query := `
        UPDATE apis 
        SET is_published = $1, description = $2, category = $3, tags = $4, updated_at = NOW()
        WHERE id = $5
    `
    
    tagsStr := ""
    if len(tags) > 0 {
        tagsStr = strings.Join(tags, ",")
    }
    
    _, err := s.db.Exec(query, isPublished, description, category, tagsStr, apiID)
    return err
}

func (s *APIStore) GetAPI(apiID string) (*API, error) {
    var api API
    var tagsStr string
    
    query := `
        SELECT id, name, description, creator_id, category, tags, is_published, endpoint, created_at, updated_at
        FROM apis 
        WHERE id = $1
    `
    
    err := s.db.QueryRow(query, apiID).Scan(
        &api.ID, &api.Name, &api.Description, &api.CreatorID, 
        &api.Category, &tagsStr, &api.IsPublished, &api.Endpoint,
        &api.CreatedAt, &api.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("API not found")
        }
        return nil, err
    }
    
    if tagsStr != "" {
        api.Tags = strings.Split(tagsStr, ",")
    }
    
    return &api, nil
}

func (s *APIStore) ListAPIs(params ListParams) ([]*API, int, error) {
    var apis []*API
    var total int
    
    // Build WHERE clause
    whereClause := "WHERE is_published = true"
    args := []interface{}{}
    argIndex := 1
    
    if params.Category != "" {
        whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
        args = append(args, params.Category)
        argIndex++
    }
    
    if params.Search != "" {
        whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
        searchTerm := "%" + params.Search + "%"
        args = append(args, searchTerm)
        argIndex++
    }
    
    // Get total count
    countQuery := "SELECT COUNT(*) FROM apis " + whereClause
    err := s.db.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    offset := (params.Page - 1) * params.Limit
    query := fmt.Sprintf(`
        SELECT id, name, description, creator_id, category, tags, is_published, endpoint, created_at, updated_at
        FROM apis %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereClause, argIndex, argIndex+1)
    
    args = append(args, params.Limit, offset)
    
    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var api API
        var tagsStr string
        
        err := rows.Scan(
            &api.ID, &api.Name, &api.Description, &api.CreatorID,
            &api.Category, &tagsStr, &api.IsPublished, &api.Endpoint,
            &api.CreatedAt, &api.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        
        if tagsStr != "" {
            api.Tags = strings.Split(tagsStr, ",")
        }
        
        apis = append(apis, &api)
    }
    
    return apis, total, nil
}

func (s *APIStore) GetDocumentation(apiID string) (interface{}, error) {
    // Mock implementation - return placeholder documentation
    return map[string]interface{}{
        "api_id": apiID,
        "openapi": "3.0.0",
        "info": map[string]interface{}{
            "title": "API Documentation",
            "version": "1.0.0",
        },
        "paths": map[string]interface{}{},
    }, nil
}

func (s *APIStore) GetAllPublishedAPIs() ([]*API, error) {
    var apis []*API
    
    query := `
        SELECT id, name, description, creator_id, category, tags, is_published, endpoint, created_at, updated_at
        FROM apis 
        WHERE is_published = true
        ORDER BY created_at DESC
    `
    
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var api API
        var tagsStr string
        
        err := rows.Scan(
            &api.ID, &api.Name, &api.Description, &api.CreatorID,
            &api.Category, &tagsStr, &api.IsPublished, &api.Endpoint,
            &api.CreatedAt, &api.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        
        if tagsStr != "" {
            api.Tags = strings.Split(tagsStr, ",")
        }
        
        apis = append(apis, &api)
    }
    
    return apis, nil
}
