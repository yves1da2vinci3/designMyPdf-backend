package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/storage"

	"github.com/gofiber/fiber/v2"
)

func UploadRouter(api fiber.Router, b2 *storage.BackblazeStorage) {
	upload := api.Group("/upload", middleware.Protected())
	upload.Post("/cover-image", handlers.UploadCoverImage(b2))
}
