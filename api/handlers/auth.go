package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/email"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type SignupDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"userName"`
}
type UpdateDTO struct {
	Password string `json:"password"`
	UserName string `json:"userName"`
}

type ForgotPasswordDTO struct {
	Email string `json:"email"`
}

type ResetPasswordDTO struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Login is handler/controller which
//	@Summary		Login user
//	@Description	Login user with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginDTO											true	"Login request body"
//	@Success		200		{object}	presenter.UserSuccessResponse
//	@Failure		400		{object}	presenter.UserErrorResponse
//	@Failure		500		{object}	presenter.UserErrorResponse
//	@Router			/auth/login [post]

func Login(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody LoginDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		if requestBody.Email == "" || requestBody.Password == "" {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(errors.New(
				"please specify email and password")))
		}
		result, err := service.Login(requestBody.Email, requestBody.Password)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		// Store refresh token in session
		sess := c.Locals("session").(*session.Session)
		sess.Set("refreshToken", result.RefreshToken)
		sess.Set("userID", result.Data.ID)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not save session",
			})
		}
		return c.JSON(presenter.LoginSuccessResponse(result))
	}
}

func Register(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody SignupDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		if requestBody.Email == "" || requestBody.Password == "" || requestBody.UserName == "" {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(errors.New(
				"please specify email and password")))
		}
		result, err := service.Register(requestBody.UserName, requestBody.Email, requestBody.Password)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(err))
		}

		err = email.SendSignupEmail(requestBody.Email, requestBody.UserName)
		if err != nil {
			return c.JSON(presenter.UserErrorResponse(err))
		}

		return c.JSON(presenter.UserSuccessResponse(result))
	}
}

func Update(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody UpdateDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		userID := c.Locals("userID").(float64)
		result, err := service.Update(userID, requestBody.UserName, requestBody.Password)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		return c.JSON(presenter.UserSuccessResponse(result))
	}
}

func ForgotPassword(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody ForgotPasswordDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		err = service.ForgotPassword(requestBody.Email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{"message": "Password reset email sent"})
	}
}
func ResetPassword(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody ResetPasswordDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}
		err = service.ResetPassword(requestBody.Token, requestBody.Password)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(presenter.UserErrorResponse(err))
		}
		return c.JSON(fiber.Map{"message": "Password reseted "})
	}
}

func RefreshToken(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session
		sess := c.Locals("session").(*session.Session)
		refreshToken := sess.Get("refreshToken")
		userID := sess.Get("userID")
		if refreshToken == nil || userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No refresh token or user ID found",
			})
		}

		// Validate and decode refresh token
		claims, err := auth.DecodeRefreshToken(refreshToken.(string))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid refresh token",
			})
		}

		// Check if the user ID matches
		if claims.Content != userID {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user ID",
			})
		}

		// Generate new access token
		accessToken, err := auth.GenerateAccessToken(claims.Content)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not generate access token",
			})
		}

		return c.JSON(fiber.Map{
			"accessToken": accessToken,
		})
	}
}
