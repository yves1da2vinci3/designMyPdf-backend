package entities

import (
	"gorm.io/gorm"
)

// Key represents an API key in the database
type Key struct {
	gorm.Model
	Name         string `json:"name"`
	Value        string `json:"value"`
	KeyCount     int    `json:"key_count"`
	KeyCountUsed int    `json:"key_count_used"`
	Logs         []Log  `json:"logs" `
	UserID       uint   `json:"user_id"`
}
