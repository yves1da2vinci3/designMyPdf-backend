package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"

	"github.com/gofiber/fiber/v2"
)

func WebhookRouter(api fiber.Router) {
	api.Get("/webhook-events/definitions", middleware.Protected(), handlers.GetWebhookEventDefinitions)

	wh := api.Group("/webhook-subscriptions", middleware.Protected())
	wh.Post("/", handlers.CreateWebhookSubscription)
	wh.Get("/", handlers.ListWebhookSubscriptions)
	wh.Get("/:id", handlers.GetWebhookSubscription)
	wh.Patch("/:id", handlers.UpdateWebhookSubscription)
	wh.Delete("/:id", handlers.DeleteWebhookSubscription)
	wh.Get("/:id/deliveries", handlers.GetWebhookDeliveries)
}
