package handlers

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

const maxWithImage = 1
const maxTextOnly = 2

type aiQuotaRequest struct {
	WithImage bool `json:"withImage"`
}

// CheckAndIncrementAiQuota checks daily AI quota for the authenticated user and increments if within limit.
func CheckAndIncrementAiQuota(c *fiber.Ctx) error {
	userIDValue := c.Locals("userID")
	userIDFloat, ok := userIDValue.(float64)
	if !ok {
		err := errors.New("invalid user ID")
		c.Status(http.StatusUnauthorized)
		return c.JSON(fiber.Map{"error": err.Error()})
	}
	userID := uint(userIDFloat)

	var req aiQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{"error": "invalid request body"})
	}

	today := time.Now().UTC().Format("2006-01-02")

	var usage entities.AiGenerationUsage
	result := database.DB.Where("user_id = ? AND date = ?", userID, today).First(&usage)
	if result.Error != nil {
		usage = entities.AiGenerationUsage{
			UserID:         userID,
			Date:           today,
			WithImageCount: 0,
			TextOnlyCount:  0,
		}
	}

	if req.WithImage {
		if usage.WithImageCount >= maxWithImage {
			c.Status(http.StatusTooManyRequests)
			return c.JSON(fiber.Map{"error": "Daily limit reached: 1 image generation per day"})
		}
		usage.WithImageCount++
	} else {
		if usage.TextOnlyCount >= maxTextOnly {
			c.Status(http.StatusTooManyRequests)
			return c.JSON(fiber.Map{"error": "Daily limit reached: 2 text generations per day"})
		}
		usage.TextOnlyCount++
	}

	if err := database.DB.Save(&usage).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "failed to update quota"})
	}

	return c.JSON(fiber.Map{"ok": true})
}
