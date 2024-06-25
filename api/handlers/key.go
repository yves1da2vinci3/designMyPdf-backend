package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/key"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// KeyRequest represents the request payload for creating/updating keys.
type KeyRequest struct {
	Name     string `json:"name"`
	KeyCount int    `json:"key_count"`
}

// CreateKey handles the creation of a new key.
func CreateKey(service key.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userID, err := getUserIDFromContext(userIDValue)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}

		var request KeyRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}
		key, err := service.Create(request.Name, userID, request.KeyCount)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		return c.Status(fiber.StatusOK).JSON(presenter.KeySuccessResponse(key))
	}
}

// GetAllUserKeys handles retrieving all keys for the authenticated user.
func GetAllUserKeys(service key.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userID, err := getUserIDFromContext(userIDValue)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}

		keys, err := service.GetUserKeys(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		return c.Status(fiber.StatusOK).JSON(presenter.KeysSuccessResponse(keys))
	}
}

// UpdateKey handles updating an existing key.
func UpdateKey(service key.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		keyID, err := c.ParamsInt("keyID")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}
		var requestBody KeyRequest
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}
		key, err := service.Update(uint(keyID), requestBody.Name, requestBody.KeyCount)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		return c.Status(fiber.StatusOK).JSON(presenter.KeySuccessResponse(key))
	}
}

// DeleteKey handles deleting a key by its ID.
func DeleteKey(service key.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("keyID")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(presenter.KeyErrorResponse(err))
		}
		key, err := service.Delete(uint(id))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		return c.JSON(presenter.KeySuccessResponse(key))
	}
}

// CheckKey is middleware that checks the validity of the provided key.
func CheckKey(service key.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		keyValue := c.Get("dmp_KEY")
		if keyValue == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "No key provided"})
		}

		key, err := service.GetKeyByValue(keyValue)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid key"})
		}

		c.Locals("key", key.ID)
		c.Locals("userID", key.UserID)
		return c.Next()
	}
}

// getUserIDFromContext safely converts the user ID from context to uint.
func getUserIDFromContext(userIDValue interface{}) (uint, error) {
	switch v := userIDValue.(type) {
	case float64:
		return uint(v), nil
	case uint:
		return v, nil
	default:
		return 0, errors.New("invalid user ID type")
	}
}
