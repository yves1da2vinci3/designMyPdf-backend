package handlers

import (
	"designmypdf/pkg/entities"
	"designmypdf/pkg/marketplace"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PublishRequest struct {
	TemplateID    uint                 `json:"templateId"`
	Price         int                  `json:"price"`
	Description   string               `json:"description"`
	Category      string               `json:"category"`
	Features      entities.MultiString `json:"features"`
	CoverImageURL string               `json:"coverImageURL"`
}

type CopyRequest struct {
	NamespaceID uint `json:"namespaceId"`
}

func ListMarketplace(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		category := c.Query("category", "")
		templates, err := svc.GetAll(category)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": true, "templates": templates})
	}
}

func GetMarketplaceListing(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "invalid id"})
		}
		template, err := svc.GetByID(uint(id))
		if err != nil {
			c.Status(http.StatusNotFound)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": true, "template": template})
	}
}

func PublishToMarketplace(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			c.Status(http.StatusUnauthorized)
			return c.JSON(fiber.Map{"status": false, "error": "invalid user"})
		}
		userID := uint(userIDFloat)

		var req PublishRequest
		if err := c.BodyParser(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		if req.TemplateID == 0 {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "templateId required"})
		}

		template, err := svc.Publish(req.TemplateID, userID, req.Description, req.Price, req.Category, req.Features, req.CoverImageURL)
		if err != nil {
			if err.Error() == "unauthorized: template does not belong to user" {
				c.Status(http.StatusForbidden)
			} else {
				c.Status(http.StatusInternalServerError)
			}
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": true, "template": template})
	}
}

func CopyMarketplaceTemplate(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			c.Status(http.StatusUnauthorized)
			return c.JSON(fiber.Map{"status": false, "error": "invalid user"})
		}
		userID := uint(userIDFloat)

		idStr := c.Params("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "invalid id"})
		}

		var req CopyRequest
		if err := c.BodyParser(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		if req.NamespaceID == 0 {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": errors.New("namespaceId required").Error()})
		}

		template, err := svc.CopyToNamespace(uint(id), req.NamespaceID, userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": true, "template": template})
	}
}

func PurchaseMarketplaceTemplate(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Payment stub — always succeeds for now
		return c.JSON(fiber.Map{
			"status":  true,
			"message": "payment processed",
			"success": true,
		})
	}
}

func GetMyListings(svc marketplace.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			c.Status(http.StatusUnauthorized)
			return c.JSON(fiber.Map{"status": false, "error": "invalid user"})
		}
		userID := uint(userIDFloat)

		templates, err := svc.GetUserListings(userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"status": false, "error": err.Error()})
		}

		type ListingWithStats struct {
			ID            uint                 `json:"id"`
			Name          string               `json:"name"`
			Description   string               `json:"description"`
			CoverImageURL string               `json:"cover_image_url"`
			Price         int                  `json:"price"`
			IsPublished   bool                 `json:"is_published"`
			Category      string               `json:"category"`
			Features      entities.MultiString `json:"features"`
			UsesCount     int                  `json:"uses_count"`
			Revenue       int                  `json:"revenue"`
		}

		var listings []ListingWithStats
		for _, t := range templates {
			listings = append(listings, ListingWithStats{
				ID:            t.ID,
				Name:          t.Name,
				Description:   t.Description,
				CoverImageURL: t.CoverImageURL,
				Price:         t.Price,
				IsPublished:   t.IsPublished,
				Category:      t.Category,
				Features:      t.Features,
				UsesCount:     t.UsesCount,
				Revenue:       t.UsesCount * t.Price,
			})
		}

		if listings == nil {
			listings = []ListingWithStats{}
		}

		return c.JSON(fiber.Map{"status": true, "listings": listings})
	}
}
