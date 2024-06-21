package relational

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
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
	gorm.Model
	CalledAt     time.Time      `json:"called_at"`
	RequestBody  datatypes.JSON `json:"request_body"`
	ResponseBody datatypes.JSON `json:"response_body"`
	StatusCode   StatusCode     `json:"status_code"`
	ErrorMessage string         `json:"error_message,omitempty"`
	TemplateID   uint           `json:"template_id"`
	KeyID        uint           `json:"key_id"`
}
