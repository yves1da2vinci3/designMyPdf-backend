package main

import (
	"designmypdf/api/routes"
	"designmypdf/config/database"
	"designmypdf/pkg/auth"
	"fmt"
	"log"
	"os"

	_ "designmypdf/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

func SetupFiberServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	app := fiber.New()

	//** setup session middleware
	// Initialize the session database schema
	auth.InitSessionDB()

	// Get the session store
	sessionStore := auth.GetSessionStore()
	// Set up session middleware
	store := session.New(session.Config{
		Storage: sessionStore,
	})
	// Middleware to use session
	app.Use(func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return err
		}
		c.Locals("session", sess)
		return c.Next()
	})
	// Setup Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	app.Get("/swagger/*", swagger.New(swagger.Config{ // custom
		URL:         "./docs/swagger.json",
		DeepLinking: true,
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "none",
	}))

	// Set up routes
	routes.SetupRoutes(app)
	if err := app.Listen(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

// @title			DesignMyPdf API
// @version		1.0
// @description	This is the first version of the design
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.email	fiber@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:5000
// @BasePath		/api
func main() {
	// Initialize the database
	err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if database.DB != nil {
		fmt.Println("Connected to SQL database:", database.DB)

	} else {
		fmt.Println("No database connection established")
	}

	// Initialize Fiber server
	SetupFiberServer()
}
