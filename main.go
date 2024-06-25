package main

import (
	"designmypdf/api/routes"
	"designmypdf/config/database"
	"fmt"
	"log"
	"os"
	"time"

	_ "designmypdf/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

func SetupFiberServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	app := fiber.New()
	// ** setup CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://transactional-clone-frontend.vercel.app/",
		AllowHeaders:     "Authorization, Content-Type",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowCredentials: true,
	}))
	// ** setup rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))
	// Test firebase upload

	// path := "./designMyPDF-diagram.png"

	// bucket, err := storage.InitializeFirebaseStorage()
	// if err != nil {
	// 	log.Fatalf("Error initializing firebase storage: %v", err)
	// }
	// url, err := storage.UploadFile(bucket, path, "tests/doumgbalolo-diagram.png")

	// fmt.Printf("url %s", url)
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
