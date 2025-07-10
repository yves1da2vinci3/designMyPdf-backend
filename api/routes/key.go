package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/key"

	"github.com/gofiber/fiber/v2"
)

func KeyRouter(api fiber.Router, keyService key.Service) {
	// key
	keyRouter := api.Group("/keys", middleware.Protected())
	keyRouter.Post("/", handlers.CreateKey(keyService))
	keyRouter.Delete("/:keyID", handlers.DeleteKey(keyService))
	keyRouter.Put("/:keyID", handlers.UpdateKey(keyService))
	keyRouter.Get("/", handlers.GetAllUserKeys(keyService))
}
