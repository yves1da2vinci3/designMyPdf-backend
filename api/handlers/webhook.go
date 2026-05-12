package handlers

import (
	"crypto/rand"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/webhook"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// GetWebhookEventDefinitions returns the list of supported event names.
func GetWebhookEventDefinitions(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"events": webhook.AllEvents()})
}

// CreateWebhookSubscription creates a new webhook subscription for the authenticated user.
//
// Body: { "webhook_uri": string, "event_names": []string, "extra_headers": {}, "key_ids": []uint }
// key_ids empty = subscription applies to ALL keys.
func CreateWebhookSubscription(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	var body struct {
		WebhookURI   string            `json:"webhook_uri"`
		EventNames   []string          `json:"event_names"`
		ExtraHeaders map[string]string `json:"extra_headers"`
		KeyIDs       []uint            `json:"key_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid body"})
	}
	if body.WebhookURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "webhook_uri required"})
	}

	// Enforce one-webhook-per-key constraint.
	if len(body.KeyIDs) == 0 {
		// "All keys" scope: user must not have any other active subscription.
		hasOther, err := webhook.UserHasOtherSubscriptions(userID, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		if hasOther {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "user already has active subscriptions; cannot create an all-keys subscription",
			})
		}
	} else {
		for _, kid := range body.KeyIDs {
			linked, err := webhook.KeyAlreadyLinked(kid, "")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			if linked {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"message": fmt.Sprintf("key %d is already linked to another subscription", kid),
				})
			}
		}
	}

	eventNamesJSON, _ := json.Marshal(body.EventNames)
	extraHeadersJSON, _ := json.Marshal(body.ExtraHeaders)

	sub := &entities.WebhookSubscription{
		UserID:       userID,
		WebhookURI:   body.WebhookURI,
		Secret:       generateSecret(),
		IsActive:     true,
		EventNames:   eventNamesJSON,
		ExtraHeaders: extraHeadersJSON,
	}

	if err := webhook.CreateSubscription(sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	for _, kid := range body.KeyIDs {
		sub.Keys = append(sub.Keys, entities.WebhookSubscriptionKey{
			SubscriptionID: sub.ID,
			KeyID:          kid,
		})
	}
	if len(sub.Keys) > 0 {
		if err := webhook.UpdateSubscription(sub); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"subscription": sub,
		"secret":       sub.Secret, // returned only at creation
	})
}

// ListWebhookSubscriptions returns all subscriptions with last delivery metadata.
func ListWebhookSubscriptions(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	subs, err := webhook.GetUserSubscriptionsWithStats(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"subscriptions": subs})
}

// GetWebhookSubscription returns a single subscription owned by the authenticated user.
func GetWebhookSubscription(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	subID := c.Params("id")
	sub, err := webhook.GetSubscriptionByID(subID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "subscription not found"})
	}
	if sub.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "access denied"})
	}

	return c.JSON(fiber.Map{"subscription": sub})
}

// UpdateWebhookSubscription patches an existing subscription.
//
// Body fields are all optional: webhook_uri, event_names, extra_headers, key_ids, is_active.
func UpdateWebhookSubscription(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	subID := c.Params("id")
	sub, err := webhook.GetSubscriptionByID(subID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "subscription not found"})
	}
	if sub.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "access denied"})
	}

	var body struct {
		WebhookURI       *string           `json:"webhook_uri"`
		EventNames       []string          `json:"event_names"`
		ExtraHeaders     map[string]string `json:"extra_headers"`
		KeyIDs           *[]uint           `json:"key_ids"`
		IsActive         *bool             `json:"is_active"`
		RegenerateSecret bool              `json:"regenerate_secret"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid body"})
	}

	if body.WebhookURI != nil {
		sub.WebhookURI = *body.WebhookURI
	}
	if body.EventNames != nil {
		sub.EventNames, _ = json.Marshal(body.EventNames)
	}
	if body.ExtraHeaders != nil {
		sub.ExtraHeaders, _ = json.Marshal(body.ExtraHeaders)
	}
	if body.IsActive != nil {
		sub.IsActive = *body.IsActive
	}

	if body.KeyIDs != nil {
		newKeyIDs := *body.KeyIDs

		if len(newKeyIDs) == 0 {
			// Switching to all-keys scope.
			hasOther, err := webhook.UserHasOtherSubscriptions(userID, subID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			if hasOther {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"message": "cannot switch to all-keys scope: other subscriptions exist",
				})
			}
		} else {
			for _, kid := range newKeyIDs {
				linked, err := webhook.KeyAlreadyLinked(kid, subID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
				}
				if linked {
					return c.Status(fiber.StatusConflict).JSON(fiber.Map{
						"message": fmt.Sprintf("key %d is already linked to another subscription", kid),
					})
				}
			}
		}

		if err := webhook.DeleteSubscriptionKeys(subID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		sub.Keys = nil
		for _, kid := range newKeyIDs {
			sub.Keys = append(sub.Keys, entities.WebhookSubscriptionKey{
				SubscriptionID: subID,
				KeyID:          kid,
			})
		}
	}

	if err := webhook.UpdateSubscription(sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	resp := fiber.Map{"subscription": sub}

	if body.RegenerateSecret {
		newSecret := generateSecret()
		if err := webhook.UpdateSubscriptionSecret(subID, newSecret); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		resp["secret"] = newSecret
	}

	return c.JSON(resp)
}

// GetWebhookDeliveries returns the delivery history for a subscription.
func GetWebhookDeliveries(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	subID := c.Params("id")
	sub, err := webhook.GetSubscriptionByID(subID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "subscription not found"})
	}
	if sub.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "access denied"})
	}

	deliveries, err := webhook.GetDeliveriesForSubscription(subID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"attempts": deliveries, "total": len(deliveries)})
}

// DeleteWebhookSubscription removes a subscription owned by the authenticated user.
func DeleteWebhookSubscription(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c.Locals("userID"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	subID := c.Params("id")
	sub, err := webhook.GetSubscriptionByID(subID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "subscription not found"})
	}
	if sub.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "access denied"})
	}

	if err := webhook.DeleteSubscriptionKeys(subID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	if err := webhook.DeleteSubscription(subID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-secret"
	}
	return hex.EncodeToString(b)
}

// getUserIDFromContext is defined in key.go and accessible across the handlers package.
