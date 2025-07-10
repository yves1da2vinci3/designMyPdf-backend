package database

import (
	"designmypdf/pkg/entities"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitializeSQL initializes the SQL database based on the provided type
func InitializeSQL(dbType, host, port, user, password, dbName string) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

	switch strings.ToLower(dbType) {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbName)
		dialector = mysql.Open(dsn)
	case "postgresql":
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
		sslmode := "disable"
		stage := os.Getenv("GO_ENV")
		if stage == "production" {
			sslmode = "require"
		}
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai", host, user, password, dbName, port, sslmode)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Make migration
	db.AutoMigrate(&entities.User{}, &entities.Namespace{}, &entities.Template{}, &entities.Key{}, &entities.Log{}, &entities.Session{})

	return db, nil
}
