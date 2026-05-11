package handlers

import (
	"context"
	"designmypdf/pkg/storage"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UploadCoverImage(b2 *storage.BackblazeStorage) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if b2 == nil {
			c.Status(http.StatusServiceUnavailable)
			return c.JSON(fiber.Map{"status": false, "error": "storage not configured"})
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "file is required"})
		}

		// Validate content type
		ct := file.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "image/") {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "only image files are allowed"})
		}

		// Limit 10MB
		if file.Size > 10*1024*1024 {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{"status": false, "error": "file too large (max 10MB)"})
		}

		ext := filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".jpg"
		}

		id := uuid.New().String()
		tmpPath := fmt.Sprintf("/tmp/cover-%s%s", id, ext)
		objectName := fmt.Sprintf("covers/%s%s", id, ext)

		if err := c.SaveFile(file, tmpPath); err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"status": false, "error": "failed to save file"})
		}
		defer os.Remove(tmpPath)

		url, err := b2.UploadFile(context.Background(), tmpPath, objectName)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{"status": false, "error": "upload failed: " + err.Error()})
		}

		return c.JSON(fiber.Map{"status": true, "url": url})
	}
}
