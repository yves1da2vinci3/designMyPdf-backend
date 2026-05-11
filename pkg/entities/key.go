package entities

import (
	"time"

	"gorm.io/gorm"
)

// Key represents an API key in the database
type Key struct {
	gorm.Model
	Name         string     `json:"name"`
	Value        string     `json:"value"`
	KeyCount     int        `json:"key_count"`
	KeyCountUsed int        `json:"key_count_used"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	Logs         []Log      `json:"logs"`
	UserID       uint       `json:"user_id"`
}
