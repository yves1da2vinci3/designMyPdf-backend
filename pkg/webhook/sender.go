package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	maxAttempts    = 3
	retryBackoff   = 2 * time.Second
	requestTimeout = 10 * time.Second
)

var httpClient = &http.Client{Timeout: requestTimeout}

// Send delivers the webhook payload to a single subscription with up to
// maxAttempts retries. Each attempt is recorded in webhook_send_attempts.
// After maxAttempts consecutive failures the subscription is deactivated.
func Send(sub *entities.WebhookSubscription, eventID string, payloadJSON []byte) {
	var consecutiveFails int

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		status, snippet, sendErr := doPost(sub, payloadJSON)

		errStr := ""
		if sendErr != nil {
			errStr = sendErr.Error()
			consecutiveFails++
		} else {
			consecutiveFails = 0
		}

		_ = CreateSendAttempt(&entities.WebhookSendAttempt{
			WebhookEventID:  eventID,
			SubscriptionID:  sub.ID,
			HTTPStatus:      status,
			ResponseSnippet: snippet,
			Error:           errStr,
			AttemptNo:       attempt,
		})

		if sendErr == nil && status >= 200 && status < 300 {
			return
		}

		if attempt < maxAttempts {
			time.Sleep(retryBackoff)
		}
	}

	if consecutiveFails >= maxAttempts {
		deactivateSubscription(sub.ID)
	}
}

func doPost(sub *entities.WebhookSubscription, payloadJSON []byte) (statusCode int, responseSnippet string, err error) {
	req, err := http.NewRequest(http.MethodPost, sub.WebhookURI, bytes.NewReader(payloadJSON))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Dmp-Webhook-Signature", signPayload(payloadJSON, sub.Secret))

	// Merge extra headers from subscription.
	if len(sub.ExtraHeaders) > 0 {
		var extra map[string]string
		if json.Unmarshal(sub.ExtraHeaders, &extra) == nil {
			for k, v := range extra {
				req.Header.Set(k, v)
			}
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return resp.StatusCode, string(body), nil
}

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func deactivateSubscription(id string) {
	if err := database.DB.Model(&entities.WebhookSubscription{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		fmt.Printf("warning: failed to deactivate subscription %s: %v\n", id, err)
	}
}
