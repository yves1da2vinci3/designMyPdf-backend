package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// DB is a global variable to hold the SQL database connection
var DB *gorm.DB

// MongoDBClient is a global variable to hold the MongoDB connection
var MongoDBClient *mongo.Client

// DatabaseConfig holds the configuration details for the database
type DatabaseConfig struct {
	DBType   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MongoURI string
}

// Initialize initializes the database connection based on the environment variables
func Initialize() error {
	config := getConfigFromEnv()
	switch strings.ToLower(config.DBType) {
	case "mongodb":
		client, err := InitializeMongo(config.MongoURI)
		if err != nil {
			return err
		}
		MongoDBClient = client
	case "mysql":
		db, err := InitializeSQL("mysql", config.Host, config.Port, config.User, config.Password, config.DBName)
		if err != nil {
			return err
		}
		DB = db
	case "postgresql":
		db, err := InitializeSQL("postgresql", config.Host, config.Port, config.User, config.Password, config.DBName)
		if err != nil {
			return err
		}
		DB = db
	default:
		return fmt.Errorf("unsupported database type: %s", config.DBType)
	}

	return nil
}

// getConfigFromEnv retrieves the database configuration from environment variables
func getConfigFromEnv() DatabaseConfig {
	return DatabaseConfig{
		DBType:   getEnv("DB_TYPE", "MongoDB"),
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", ""),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASS", ""),
		DBName:   getEnv("DB_NAME", ""),
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
	}
}

// getEnv retrieves the value of the environment variable named by the key or returns the fallback value if not present
func getEnv(key, fallback string) string {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
