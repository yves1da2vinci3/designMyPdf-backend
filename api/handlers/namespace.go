package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/namespace"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NamespaceRequest struct {
	Name string `json:"name"`
}

func CreateNamespace(namepsaceService namespace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody NamespaceRequest
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		if requestBody.Name == "" {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			err := errors.New("invalid user ID type")
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		userID := uint(userIDFloat)
		result, err := namepsaceService.Create(requestBody.Name, userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		return c.JSON(presenter.NamespaceSuccessResponse(result))
	}
}

func DeleteNamespace(namespaceService namespace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespaceIDString := c.Params("namespaceID")
		namespaceID, err := strconv.Atoi(namespaceIDString)
		if err != nil {
			return errors.New("Error converting string to integer:")
		}
		result, err := namespaceService.Delete(uint(namespaceID))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		return c.JSON(presenter.NamespaceSuccessResponse(result))
	}
}

func UpdateNamespace(namespaceService namespace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		namespaceIDString := c.Params("namespaceID")
		namespaceID, err := strconv.Atoi(namespaceIDString)
		if err != nil {
			return errors.New("Error converting string to integer:")
		}
		var requestBody NamespaceRequest
		err = c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		if requestBody.Name == "" {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		result, err := namespaceService.Update(uint(namespaceID), requestBody.Name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		return c.JSON(presenter.NamespaceSuccessResponse(result))
	}
}

func GetNamespaces(namespaceService namespace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			err := errors.New("invalid user ID type")
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		userID := uint(userIDFloat)
		result, err := namespaceService.GetUserNamespaces(userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NamespaceErrorResponse(err))
		}
		return c.JSON(presenter.NamespacesSuccessResponse(result))
	}
}
