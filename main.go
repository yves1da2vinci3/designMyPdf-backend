package main

import (
	"designmypdf/api/routes"
	"designmypdf/config/database"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func SetupFiberServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	app := fiber.New()

	// Set up routes
	routes.SetupRoutes(app)
	if err := app.Listen(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func main() {

	// Initialize Fiber
	defer SetupFiberServer()
	// Initialize the database
	err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if database.DB != nil {
		fmt.Println("Connected to SQL database:", database.DB)
	} else if database.MongoDBClient != nil {
		fmt.Println("Connected to MongoDB:", database.MongoDBClient)
	} else {
		fmt.Println("No database connection established")
	}
}
