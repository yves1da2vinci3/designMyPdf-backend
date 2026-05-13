package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

type FirebaseLoginDTO struct {
	IDToken string `json:"idToken"`
}

// FirebaseLogin exchanges a Firebase ID token for application JWTs (same response as email login).
func FirebaseLogin(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body FirebaseLoginDTO
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(presenter.UserErrorResponse(err))
		}
		if strings.TrimSpace(body.IDToken) == "" {
			return c.Status(http.StatusBadRequest).JSON(presenter.UserErrorResponse(
				errors.New("idToken is required")))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		result, err := service.LoginWithFirebaseIDToken(ctx, body.IDToken)
		if err != nil {
			status := http.StatusInternalServerError
			msg := err.Error()
			switch {
			case strings.Contains(msg, "not configured"):
				status = http.StatusServiceUnavailable
			case strings.Contains(msg, "invalid firebase token"),
				strings.Contains(msg, "invalid token"),
				strings.Contains(msg, "email must be verified"),
				strings.Contains(msg, "email is required"):
				status = http.StatusUnauthorized
			}
			return c.Status(status).JSON(presenter.UserErrorResponse(err))
		}

		if err := service.SetSession(result.Data.ID, result.RefreshToken); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(presenter.UserErrorResponse(err))
		}

		return c.JSON(presenter.LoginSuccessResponse(result))
	}
}
