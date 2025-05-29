package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Consumer represents a consumer in the system
type Consumer struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ConsumerStore handles consumer-related database operations
type ConsumerStore struct {
	db *sql.DB
}

// NewConsumerStore creates a new consumer store
func NewConsumerStore(db *sql.DB) *ConsumerStore {
	return &ConsumerStore{db: db}
}

// GetOrCreateConsumerID retrieves the consumer ID for a user, creating one if it doesn't exist
func (s *ConsumerStore) GetOrCreateConsumerID(userID, email string) (string, error) {
	// First, try to get existing consumer
	var consumerID string
	query := `SELECT id FROM consumers WHERE user_id = $1`
	err := s.db.QueryRow(query, userID).Scan(&consumerID)
	
	if err == nil {
		// Consumer exists
		return consumerID, nil
	}
	
	if err != sql.ErrNoRows {
		// Real error
		return "", fmt.Errorf("failed to check for existing consumer: %w", err)
	}
	
	// Consumer doesn't exist, create one
	consumerID = uuid.New().String()
	insertQuery := `
		INSERT INTO consumers (id, user_id, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id
	`
	
	now := time.Now()
	err = s.db.QueryRow(insertQuery, consumerID, userID, email, now, now).Scan(&consumerID)
	if err != nil {
		return "", fmt.Errorf("failed to create consumer: %w", err)
	}
	
	return consumerID, nil
}

// GetConsumerByUserID retrieves a consumer by user ID
func (s *ConsumerStore) GetConsumerByUserID(userID string) (*Consumer, error) {
	var consumer Consumer
	query := `
		SELECT id, user_id, email, name, created_at, updated_at
		FROM consumers
		WHERE user_id = $1
	`
	
	err := s.db.QueryRow(query, userID).Scan(
		&consumer.ID,
		&consumer.UserID,
		&consumer.Email,
		&consumer.Name,
		&consumer.CreatedAt,
		&consumer.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consumer not found")
		}
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}
	
	return &consumer, nil
}

// GetConsumerID retrieves just the consumer ID for a user ID
func (s *ConsumerStore) GetConsumerID(userID string) (string, error) {
	var consumerID string
	query := `SELECT id FROM consumers WHERE user_id = $1`
	
	err := s.db.QueryRow(query, userID).Scan(&consumerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("consumer not found for user")
		}
		return "", fmt.Errorf("failed to get consumer ID: %w", err)
	}
	
	return consumerID, nil
}

// UpdateConsumer updates consumer information
func (s *ConsumerStore) UpdateConsumer(userID string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1
	
	for field, value := range updates {
		switch field {
		case "name", "email":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
			args = append(args, value)
			argIdx++
		}
	}
	
	if len(setClauses) == 0 {
		return fmt.Errorf("no valid fields to update")
	}
	
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++
	
	args = append(args, userID)
	
	query := fmt.Sprintf(`
		UPDATE consumers 
		SET %s
		WHERE user_id = $%d
	`, string(setClauses[0]), argIdx)
	
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update consumer: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("consumer not found")
	}
	
	return nil
}
