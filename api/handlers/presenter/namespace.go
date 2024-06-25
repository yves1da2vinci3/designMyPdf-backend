package presenter

import (
	"designmypdf/pkg/entities"

	"github.com/gofiber/fiber/v2"
)

type Namespace struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	ID     uint
}

// Handler
func NamespaceSuccessResponse(namespace *entities.Namespace) *fiber.Map {
	NamespaceData := Namespace{
		UserID: namespace.UserID,
		Name:   namespace.Name,
		ID:     namespace.ID,
	}
	return &fiber.Map{
		"status":    true,
		"namespace": NamespaceData,
		"error":     nil,
	}
}

// NamespacesSuccessResponse is the list SuccessResponse that will be passed in the response by Handler
func NamespacesSuccessResponse(data *[]entities.Namespace) *fiber.Map {
	return &fiber.Map{
		"status":     true,
		"namepsaces": data,
		"error":      nil,
	}
}

// UserErrorResponse is the ErrorResponse that will be passed in the response by Handler
func NamespaceErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status":    false,
		"namespace": "",
		"error":     err.Error(),
	}
}
