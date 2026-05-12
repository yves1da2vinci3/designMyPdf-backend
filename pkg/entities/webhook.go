package entities

import (
	"time"

	"gorm.io/datatypes"
)

type WebhookSubscription struct {
	ID           string                  `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID       uint                    `json:"user_id" gorm:"not null;index"`
	WebhookURI   string                  `json:"webhook_uri" gorm:"not null"`
	Secret       string                  `json:"-" gorm:"not null"`
	IsActive     bool                    `json:"is_active" gorm:"default:true"`
	EventNames   datatypes.JSON          `json:"event_names"`
	ExtraHeaders datatypes.JSON          `json:"extra_headers"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	Keys         []WebhookSubscriptionKey `json:"keys" gorm:"foreignKey:SubscriptionID"`
}

// WebhookSubscriptionKey links a subscription to specific API keys.
// No rows for a subscription means it applies to ALL keys of the user.
// key_id has a unique index: one key can belong to at most one subscription.
type WebhookSubscriptionKey struct {
	SubscriptionID string `json:"subscription_id" gorm:"primaryKey;type:varchar(36)"`
	KeyID          uint   `json:"key_id" gorm:"primaryKey;uniqueIndex"`
}

type WebhookEvent struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	EventName   string         `json:"event_name" gorm:"not null"`
	PayloadJSON datatypes.JSON `json:"payload_json"`
	JobID       string         `json:"job_id" gorm:"index;type:varchar(36)"`
	CreatedAt   time.Time      `json:"created_at"`
}

type WebhookSendAttempt struct {
	ID              string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	WebhookEventID  string    `json:"webhook_event_id" gorm:"index;type:varchar(36)"`
	SubscriptionID  string    `json:"subscription_id" gorm:"index;type:varchar(36)"`
	HTTPStatus      int       `json:"http_status"`
	ResponseSnippet string    `json:"response_snippet"`
	Error           string    `json:"error"`
	AttemptNo       int       `json:"attempt_no"`
	CreatedAt       time.Time `json:"created_at"`
}
