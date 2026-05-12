package webhook

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetActiveSubscriptionsForUser returns subscriptions that should receive
// eventName for a job triggered by keyID belonging to userID.
//
// A subscription matches when:
//   - is_active = true AND user_id = userID AND event_names contains eventName
//   - AND (no rows in webhook_subscription_keys for this sub  ← "all keys" scope)
//     OR (keyID row exists in webhook_subscription_keys for this sub)
func GetActiveSubscriptionsForUser(userID uint, eventName string, keyID uint) ([]entities.WebhookSubscription, error) {
	var subs []entities.WebhookSubscription

	err := database.DB.
		Preload("Keys").
		Where("is_active = ? AND user_id = ? AND event_names::text LIKE ?",
			true, userID, fmt.Sprintf("%%%s%%", eventName)).
		Find(&subs).Error
	if err != nil {
		return nil, err
	}

	// Filter in Go: subscription matches if it covers all keys (no key rows)
	// OR has a specific entry for keyID.
	var matched []entities.WebhookSubscription
	for _, sub := range subs {
		if len(sub.Keys) == 0 {
			matched = append(matched, sub)
			continue
		}
		for _, k := range sub.Keys {
			if k.KeyID == keyID {
				matched = append(matched, sub)
				break
			}
		}
	}
	return matched, nil
}

func CreateSubscription(sub *entities.WebhookSubscription) error {
	sub.ID = uuid.New().String()
	return database.DB.Create(sub).Error
}

func GetSubscriptionByID(id string) (*entities.WebhookSubscription, error) {
	var sub entities.WebhookSubscription
	err := database.DB.Preload("Keys").First(&sub, "id = ?", id).Error
	return &sub, err
}

func GetUserSubscriptions(userID uint) ([]entities.WebhookSubscription, error) {
	var subs []entities.WebhookSubscription
	err := database.DB.Preload("Keys").Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func UpdateSubscription(sub *entities.WebhookSubscription) error {
	return database.DB.Save(sub).Error
}

func DeleteSubscription(id string) error {
	return database.DB.Where("id = ?", id).Delete(&entities.WebhookSubscription{}).Error
}

// DeleteSubscriptionKeys removes all key associations for a subscription.
func DeleteSubscriptionKeys(subscriptionID string) error {
	return database.DB.Where("subscription_id = ?", subscriptionID).
		Delete(&entities.WebhookSubscriptionKey{}).Error
}

// KeyAlreadyLinked returns true if keyID is already linked to a subscription
// other than excludeSubscriptionID (pass "" to check without exclusion).
func KeyAlreadyLinked(keyID uint, excludeSubscriptionID string) (bool, error) {
	query := database.DB.Model(&entities.WebhookSubscriptionKey{}).Where("key_id = ?", keyID)
	if excludeSubscriptionID != "" {
		query = query.Where("subscription_id != ?", excludeSubscriptionID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// UserHasOtherSubscriptions returns true if userID has active subscriptions
// other than excludeSubscriptionID.
func UserHasOtherSubscriptions(userID uint, excludeSubscriptionID string) (bool, error) {
	query := database.DB.Model(&entities.WebhookSubscription{}).
		Where("user_id = ? AND is_active = ?", userID, true)
	if excludeSubscriptionID != "" {
		query = query.Where("id != ?", excludeSubscriptionID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// DeliveryWithEvent joins a send attempt with its parent webhook event.
type DeliveryWithEvent struct {
	entities.WebhookSendAttempt
	EventName   string `json:"event_name"`
	PayloadJSON string `json:"payload_json"`
}

// GetDeliveriesForSubscription returns the 50 most recent delivery attempts for a subscription.
func GetDeliveriesForSubscription(subscriptionID string) ([]DeliveryWithEvent, error) {
	var results []DeliveryWithEvent
	err := database.DB.
		Table("webhook_send_attempts a").
		Select("a.*, COALESCE(e.event_name,'') as event_name, COALESCE(e.payload_json::text,'') as payload_json").
		Joins("LEFT JOIN webhook_events e ON e.id = a.webhook_event_id").
		Where("a.subscription_id = ?", subscriptionID).
		Order("a.created_at DESC").
		Limit(50).
		Scan(&results).Error
	return results, err
}

// WebhookSubscriptionWithStats extends a subscription with the most recent delivery info.
type WebhookSubscriptionWithStats struct {
	entities.WebhookSubscription
	LastDeliveryStatus int        `json:"last_delivery_status"`
	LastDeliveryAt     *time.Time `json:"last_delivery_at"`
}

// GetUserSubscriptionsWithStats returns subscriptions enriched with last delivery metadata.
func GetUserSubscriptionsWithStats(userID uint) ([]WebhookSubscriptionWithStats, error) {
	subs, err := GetUserSubscriptions(userID)
	if err != nil {
		return nil, err
	}

	results := make([]WebhookSubscriptionWithStats, len(subs))
	for i, sub := range subs {
		results[i] = WebhookSubscriptionWithStats{WebhookSubscription: sub}

		var attempt entities.WebhookSendAttempt
		if err := database.DB.
			Where("subscription_id = ?", sub.ID).
			Order("created_at DESC").
			First(&attempt).Error; err == nil {
			results[i].LastDeliveryStatus = attempt.HTTPStatus
			results[i].LastDeliveryAt = &attempt.CreatedAt
		}
	}
	return results, nil
}

// UpdateSubscriptionSecret sets a new secret on a subscription.
func UpdateSubscriptionSecret(id, secret string) error {
	return database.DB.Model(&entities.WebhookSubscription{}).
		Where("id = ?", id).
		Update("secret", secret).Error
}

func CreateWebhookEvent(event *entities.WebhookEvent) error {
	event.ID = uuid.New().String()
	return database.DB.Create(event).Error
}

func CreateSendAttempt(attempt *entities.WebhookSendAttempt) error {
	attempt.ID = uuid.New().String()
	return database.DB.Create(attempt).Error
}
