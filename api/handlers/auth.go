package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/email"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
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
type SessionDTO struct {
	RefreshToken string `json:"refreshToken"`
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

		// Handle session and set cookie
		err = service.SetSession(result.Data.ID, result.RefreshToken)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.UserErrorResponse(err))
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
				"please specify email, password, and username")))
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
		return c.JSON(fiber.Map{"message": "Password reset"})
	}
}

func RefreshToken(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the refresh token from the request
		var requestBody SessionDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}

		refreshToken := requestBody.RefreshToken
		if refreshToken == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "No refresh token provided"})
		}

		session, err := service.GetSessionByToken(refreshToken)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid refresh token"})
		}

		accessToken, err := service.Refresh(session.ID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Could not refresh token"})
		}

		return c.JSON(fiber.Map{
			"accessToken": accessToken,
		})
	}
}

func Logout(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the refresh token from the request
		var requestBody SessionDTO
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.UserErrorResponse(err))
		}

		refreshToken := requestBody.RefreshToken
		if refreshToken == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "No refresh token provided"})
		}

		session, err := service.GetSessionByToken(refreshToken)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid refresh token"})
		}

		err = service.Logout(session.ID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Logout failed"})
		}

		return c.JSON(fiber.Map{"message": "Logout successful"})
	}
}
