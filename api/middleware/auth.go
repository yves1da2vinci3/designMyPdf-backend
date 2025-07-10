package middleware

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// Protected protect routes
func Protected() func(*fiber.Ctx) error {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(os.Getenv("ACCESS_TOKEN_KEY"))},
		ErrorHandler: jwtError,
		SuccessHandler: func(c *fiber.Ctx) error {
			// Save user info from token into context
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			c.Locals("userID", claims["content"]) // Store userID in context
			return c.Next()
		},
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})

	} else {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
	}
}
