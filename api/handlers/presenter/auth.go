package presenter

import (
	"designmypdf/pkg/entities"

	"github.com/gofiber/fiber/v2"
)

// UserSuccessResponse is the singular SuccessResponse that will be passed in the response by

type LoginResponse struct {
	Status       bool           `json:"status"`
	Data         *entities.User `json:"data"`
	Error        error          `json:"error"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
}

// Handler
func UserSuccessResponse(user *entities.User) *fiber.Map {

	userData := entities.User{
		UserName: user.UserName,
		Email:    user.Email,
	}
	return &fiber.Map{
		"status": true,
		"data":   userData,
		"error":  nil,
	}
}
func LoginSuccessResponse(loginResponse *LoginResponse) *fiber.Map {
	userData := entities.User{
		UserName: loginResponse.Data.UserName,
		Email:    loginResponse.Data.Email,
	}
	return &fiber.Map{
		"status":       true,
		"data":         userData,
		"accessToken":  loginResponse.AccessToken,
		"refreshToken": loginResponse.RefreshToken,
		"error":        nil,
	}
}

// UsersSuccessResponse is the list SuccessResponse that will be passed in the response by Handler
func UsersSuccessResponse(data *[]entities.User) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data":   data,
		"error":  nil,
	}
}

// UserErrorResponse is the ErrorResponse that will be passed in the response by Handler
func UserErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}
