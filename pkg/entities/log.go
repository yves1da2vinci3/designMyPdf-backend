package entities

import (
	"time"

	"gorm.io/datatypes"
)

// StatusCode represents the HTTP status code type
type StatusCode int

// Define status code constants
const (
	Success StatusCode = 200
	Fail    StatusCode = 500
)

// Log represents a log entry in the database
type Log struct {
	ID           uint           `gorm:"primaryKey"`
	KeyID        uint           `gorm:"not null"`
	Key          Key            `gorm:"foreignKey:KeyID" json:"key"`
	TemplateID   uint           `gorm:"not null"`
	Template     Template       `gorm:"foreignKey:TemplateID" json:"template"`
	CalledAt     time.Time      `json:"called_at"`
	RequestBody  datatypes.JSON `json:"request_body"`
	ResponseBody datatypes.JSON `json:"response_body"`
	StatusCode   StatusCode     `json:"status_code"`
	ErrorMessage string         `json:"error_message"`
}
