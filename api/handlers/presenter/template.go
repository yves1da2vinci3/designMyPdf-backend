package presenter

import (
	"designmypdf/pkg/entities"

	"github.com/gofiber/fiber/v2"
)

// Handler
func TemplateSuccessResponse(Template *entities.Template) *fiber.Map {

	return &fiber.Map{
		"status":   true,
		"template": Template,
		"error":    nil,
	}
}

// TemplatesSuccessResponse is the list SuccessResponse that will be passed in the response by Handler
func TemplatesSuccessResponse(data *[]entities.Template) *fiber.Map {
	return &fiber.Map{
		"status":    true,
		"templates": data,
		"error":     nil,
	}
}

// UserErrorResponse is the ErrorResponse that will be passed in the response by Handler
func TemplateErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status":   false,
		"template": "",
		"error":    err.Error(),
	}
}
