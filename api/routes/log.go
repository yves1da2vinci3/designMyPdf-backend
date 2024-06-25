package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/logs"

	"github.com/gofiber/fiber/v2"
)

func LogRouter(api fiber.Router, logService logs.Service) {
	// log
	logRouter := api.Group("/logs", middleware.Protected())
	logRouter.Get("/", handlers.GetLogsByUserID(logService))
	logRouter.Get("/stats", handlers.GetLogStats(logService))
}
