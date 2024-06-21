package routes

import (
	"designmypdf/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRouter(api fiber.Router) {
	// auth
	authRouter := api.Group("/auth")
	authRouter.Post("/login", handlers.Login)
}
