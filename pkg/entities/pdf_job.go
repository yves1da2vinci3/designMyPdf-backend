package entities

import (
	"time"

	"gorm.io/datatypes"
)

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type PdfGenerationJob struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	KeyID        uint           `json:"key_id" gorm:"not null"`
	Key          Key            `json:"key" gorm:"foreignKey:KeyID"`
	TemplateUUID string         `json:"template_uuid" gorm:"not null"`
	Payload      datatypes.JSON `json:"payload"`
	Format       string         `json:"format" gorm:"default:'A4'"`
	Status       JobStatus      `json:"status" gorm:"default:'queued'"`
	ResultPath   string         `json:"result_path"`
	ErrorMessage string         `json:"error_message"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
