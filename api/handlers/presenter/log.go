package presenter

import (
	"designmypdf/pkg/entities"

	"github.com/gofiber/fiber/v2"
)

// Log represents a log entry response structure
type Log struct {
	ID           uint   `json:"id"`
	CalledAt     string `json:"called_at"`
	RequestBody  string `json:"request_body"`
	ResponseBody string `json:"response_body"`
	StatusCode   int    `json:"status_code"`
	ErrorMessage string `json:"error_message,omitempty"`
	TemplateID   uint   `json:"template_id"`
	KeyID        uint   `json:"key_id"`
}

// LogSuccessResponse returns a success response for a single log entry
func LogSuccessResponse(log *entities.Log) *fiber.Map {
	response := Log{
		ID:           log.ID,
		CalledAt:     log.CalledAt.String(), // Example formatting
		RequestBody:  string(log.RequestBody),
		ResponseBody: string(log.ResponseBody),
		StatusCode:   int(log.StatusCode),
		ErrorMessage: log.ErrorMessage,
		TemplateID:   log.TemplateID,
		KeyID:        log.KeyID,
	}
	return &fiber.Map{
		"status": true,
		"log":    response,
		"error":  nil,
	}
}

// LogsSuccessResponse returns a success response for a list of log entries
func LogsSuccessResponse(logs *[]entities.Log) *fiber.Map {

	return &fiber.Map{
		"status": true,
		"logs":   logs,
		"error":  nil,
	}
}

// LogErrorResponse returns an error response for log operations
func LogErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"log":    nil,
		"error":  err.Error(),
	}
}
