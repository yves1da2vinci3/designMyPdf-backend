package main

import (
	"designmypdf/config/database"
	"fmt"
	"log"
)

func main() {

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
