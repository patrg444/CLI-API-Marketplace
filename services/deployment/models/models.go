package models

import "time"

type APIModel struct {
    ID          string
    Name        string
    Description string
    CreatorID   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Consumer struct {
    ID        string
    Email     string
    CreatedAt time.Time
}
