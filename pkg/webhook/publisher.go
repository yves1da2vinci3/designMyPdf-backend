package webhook

import (
	"designmypdf/pkg/entities"
	"encoding/json"
	"fmt"
	"time"
)

type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

// Publish records a webhook event and dispatches it to all matching
// subscriptions concurrently. Each delivery is fire-and-forget with retries
// handled inside Send.
func (p *Publisher) Publish(eventName, jobID string, userID uint, keyID uint, extra map[string]interface{}) {
	payload := map[string]interface{}{
		"event":       eventName,
		"job_id":      jobID,
		"key_id":      keyID,
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range extra {
		payload[k] = v
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("webhook: failed to marshal payload: %v\n", err)
		return
	}

	event := &entities.WebhookEvent{
		EventName:   eventName,
		PayloadJSON: payloadJSON,
		JobID:       jobID,
	}
	if err := CreateWebhookEvent(event); err != nil {
		fmt.Printf("webhook: failed to persist event: %v\n", err)
	}

	subs, err := GetActiveSubscriptionsForUser(userID, eventName, keyID)
	if err != nil {
		fmt.Printf("webhook: failed to load subscriptions: %v\n", err)
		return
	}

	for i := range subs {
		sub := subs[i]
		go Send(&sub, event.ID, payloadJSON)
	}
}
