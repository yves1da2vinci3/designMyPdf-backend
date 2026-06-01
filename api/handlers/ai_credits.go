package handlers

import (
	"designmypdf/pkg/usercredit"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func GetAiCredits(svc *usercredit.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserIDFromContext(c.Locals("userID"))
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return c.JSON(fiber.Map{"error": err.Error()})
		}

		used, limit, remaining, creditsUsed, creditsLimit, creditsRemaining, month, err := svc.GetBalance(userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"error": "failed to get credit balance"})
		}

		return c.JSON(fiber.Map{
			"used":             used,
			"limit":            limit,
			"remaining":        remaining,
			"creditsUsed":      creditsUsed,
			"creditsLimit":     creditsLimit,
			"creditsRemaining": creditsRemaining,
			"month":            month,
		})
	}
}

type consumeCreditsRequest struct {
	Model        string `json:"model"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
	AllowPartial bool   `json:"allowPartial"`
}

func ConsumeAiCredits(svc *usercredit.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserIDFromContext(c.Locals("userID"))
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return c.JSON(fiber.Map{"error": err.Error()})
		}

		var req consumeCreditsRequest
		if err := c.BodyParser(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"error": "invalid request body"})
		}

		consumeReq := usercredit.ConsumeRequest{
			Model:        req.Model,
			InputTokens:  req.InputTokens,
			OutputTokens: req.OutputTokens,
		}

		if req.AllowPartial {
			result, err := svc.ConsumeUpToLimit(userID, consumeReq)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return c.JSON(fiber.Map{"error": "failed to consume credits"})
			}
			return c.JSON(fiber.Map{
				"ok":               true,
				"remaining":        result.RemainingMicro,
				"creditsRemaining": result.CreditsRemaining,
				"deducted":         result.DeductedMicro,
				"capped":           result.Capped,
			})
		}

		remaining, creditsRemaining, err := svc.Consume(userID, consumeReq)
		if err != nil {
			if err.Error() == "monthly credit limit reached" {
				c.Status(http.StatusTooManyRequests)
				return c.JSON(fiber.Map{"error": "Monthly credit limit reached", "capped": true})
			}
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"error": "failed to consume credits"})
		}

		return c.JSON(fiber.Map{"ok": true, "remaining": remaining, "creditsRemaining": creditsRemaining, "capped": false})
	}
}
