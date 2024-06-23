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

		email.SendSignupEmail(requestBody.Email, requestBody.UserName)
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
