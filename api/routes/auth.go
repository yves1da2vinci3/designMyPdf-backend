package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthRouter(api fiber.Router, authService auth.Service) {
	// auth
	authRouter := api.Group("/auth")
	authRouter.Post("/login", handlers.Login(authService))
	authRouter.Post("/register", handlers.Register(authService))
}
