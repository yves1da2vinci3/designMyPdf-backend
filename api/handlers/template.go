package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/template"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

type TemplateRequest struct {
	Name      string               `json:"name"`
	Content   string               `json:"content"`
	Variables datatypes.JSON       `json:"variables" gorm:"type:json"`
	Fonts     entities.MultiString `json:"fonts" gorm:"type:text"`
}

func CreateTemplate(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody TemplateRequest
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		if requestBody.Name == "" {
			err = errors.New("template name cannot be empty")
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		namespaceIDString := c.Params("namespaceID")
		namespaceID, err := strconv.ParseUint(namespaceIDString, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid namespace ID")))
		}
		result, err := templateService.Create(requestBody.Name, requestBody.Content, requestBody.Variables, requestBody.Fonts, uint(namespaceID))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplateSuccessResponse(result))
	}
}

func DeleteTemplate(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		templateIDString := c.Params("templateID")
		templateID, err := strconv.ParseUint(templateIDString, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid template ID")))
		}
		result, err := templateService.Delete(uint(templateID))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplateSuccessResponse(result))
	}
}

func UpdateTemplate(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		templateIDString := c.Params("templateID")
		templateID, err := strconv.ParseUint(templateIDString, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid template ID")))
		}
		var requestBody TemplateRequest
		err = c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		if requestBody.Name == "" {
			err = errors.New("template name cannot be empty")
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		result, err := templateService.Update(uint(templateID), requestBody.Name, requestBody.Content, requestBody.Variables, requestBody.Fonts)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplateSuccessResponse(result))
	}
}
func ChangeTemplateNamespace(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		templateIDString := c.Params("templateID")
		namespaceIDString := c.Params("namespaceID")

		templateID, err := strconv.ParseUint(templateIDString, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid template ID")))
		}
		namespaceID, err := strconv.ParseUint(namespaceIDString, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid namespace ID")))
		}

		err = templateService.ChangeTemplateNamespace(uint(templateID), uint(namespaceID))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(fiber.Map{
			"message": "Template namespace changed",
		})
	}
}

func GetTemplates(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			err := errors.New("invalid user ID type")
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		userID := uint(userIDFloat)
		result, err := templateService.GetUserTemplates(userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplatesSuccessResponse(result))
	}
}
