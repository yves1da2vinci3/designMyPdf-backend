package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/usercredit"

	"github.com/gofiber/fiber/v2"
)

func AiCreditsRouter(api fiber.Router, svc *usercredit.Service) {
	g := api.Group("/ai/credits", middleware.Protected())
	g.Get("/", handlers.GetAiCredits(svc))
	g.Post("/consume", handlers.ConsumeAiCredits(svc))
}
