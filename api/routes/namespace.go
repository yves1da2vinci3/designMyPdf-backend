package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/namespace"

	"github.com/gofiber/fiber/v2"
)

func NampesaceRouter(api fiber.Router, namepsaceService namespace.Service) {
	// auth
	namespaceRouter := api.Group("/namespaces", middleware.Protected())
	namespaceRouter.Post("/", handlers.CreateNamespace(namepsaceService))
	namespaceRouter.Delete("/:namespaceID", handlers.DeleteNamespace(namepsaceService))
	namespaceRouter.Put("/:namespaceID", handlers.UpdateNamespace(namepsaceService))
	namespaceRouter.Get("/", handlers.GetNamespaces(namepsaceService))
}
