package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/template"
	"designmypdf/utils"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type TemplateRequest struct {
	Name               string               `json:"name"`
	Content            string               `json:"content"`
	Variables          datatypes.JSON       `json:"variables" gorm:"type:json"`
	Fonts              entities.MultiString `json:"fonts" gorm:"type:text"`
	PdfBackgroundColor string               `json:"pdf_background_color"`
	PdfContentPadding  string               `json:"pdf_content_padding"`
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
		if !utils.IsValidPdfContentPadding(requestBody.PdfContentPadding) {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid pdf_content_padding")))
		}
		result, err := templateService.Update(uint(templateID), requestBody.Name, requestBody.Content, requestBody.Variables, requestBody.Fonts, requestBody.PdfBackgroundColor, requestBody.PdfContentPadding)
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

		pageStr := c.Query("page")
		limitStr := c.Query("limit")
		namespaceStr := c.Query("namespace_id")
		q := strings.TrimSpace(c.Query("q"))

		if pageStr != "" || limitStr != "" || namespaceStr != "" || q != "" {
			page := 1
			limit := 12
			if pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
					page = p
				}
			}
			if limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
					limit = l
					if limit > 50 {
						limit = 50
					}
				}
			}
			var namespaceID *uint
			if namespaceStr != "" {
				if ns, err := strconv.ParseUint(namespaceStr, 10, 32); err == nil {
					v := uint(ns)
					namespaceID = &v
				}
			}
			result, err := templateService.ListUserTemplates(userID, namespaceID, q, page, limit)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return c.JSON(presenter.TemplateErrorResponse(err))
			}
			return c.JSON(presenter.TemplatesPaginatedSuccessResponse(result.Items, result.Total, page, limit))
		}

		result, err := templateService.GetUserTemplates(userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplatesSuccessResponse(result))
	}
}
func GetTemplate(templateService template.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		templateID := strings.TrimSpace(c.Params("templateID"))
		if templateID == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid template ID")))
		}

		var result *entities.Template
		var err error
		if _, parseErr := uuid.Parse(templateID); parseErr == nil {
			result, err = templateService.GetByUUID(templateID)
		} else if idNum, parseErr := strconv.ParseUint(templateID, 10, 32); parseErr == nil {
			result, err = templateService.Get(uint(idNum))
		} else {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.TemplateErrorResponse(errors.New("invalid template ID")))
		}
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.TemplateErrorResponse(err))
		}
		return c.JSON(presenter.TemplateSuccessResponse(result))
	}
}
