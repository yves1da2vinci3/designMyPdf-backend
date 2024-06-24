package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/namespace"
	"designmypdf/pkg/user"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	api.Get("/", handlers.HelloWorld)

	// Auth
	authService := auth.NewService(user.Repository{})
	AuthRouter(api, authService)
	// Namepsace
	namepsaceService := namespace.NewService(namespace.Repository{})
	NampesaceRouter(api, namepsaceService)
}
