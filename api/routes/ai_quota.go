package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"

	"github.com/gofiber/fiber/v2"
)

func AiQuotaRouter(api fiber.Router) {
	aiRouter := api.Group("/ai", middleware.Protected())
	aiRouter.Post("/quota/check", handlers.CheckAndIncrementAiQuota)
}
