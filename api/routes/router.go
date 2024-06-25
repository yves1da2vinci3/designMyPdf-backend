package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/namespace"
	"designmypdf/pkg/template"
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
	// Template
	templateService := template.NewService(template.Repository{})
	TemplateRouter(api, templateService)
	// key
	keyService := key.NewService(key.Repository{})
	KeyRouter(api, keyService)
	// Logs
	logService := logs.NewService(logs.Repository{})
	LogRouter(api, logService)
}
