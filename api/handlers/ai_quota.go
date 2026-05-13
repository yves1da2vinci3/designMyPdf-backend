package handlers

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func getQuotaLimits() (withImage int, textOnly int) {
	withImage = 1
	textOnly = 2
	if v, err := strconv.Atoi(os.Getenv("AI_QUOTA_WITH_IMAGE")); err == nil && v > 0 {
		withImage = v
	}
	if v, err := strconv.Atoi(os.Getenv("AI_QUOTA_TEXT_ONLY")); err == nil && v > 0 {
		textOnly = v
	}
	return
}

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

	maxWithImage, maxTextOnly := getQuotaLimits()
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
			return c.JSON(fiber.Map{"error": fmt.Sprintf("Daily limit reached: %d image generation(s) per day", maxWithImage)})
		}
		usage.WithImageCount++
	} else {
		if usage.TextOnlyCount >= maxTextOnly {
			c.Status(http.StatusTooManyRequests)
			return c.JSON(fiber.Map{"error": fmt.Sprintf("Daily limit reached: %d text generation(s) per day", maxTextOnly)})
		}
		usage.TextOnlyCount++
	}

	if err := database.DB.Save(&usage).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "failed to update quota"})
	}

	return c.JSON(fiber.Map{"ok": true})
}
