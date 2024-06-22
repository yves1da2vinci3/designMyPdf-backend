package user

import (
	"designmypdf/config/database"
	"fmt"
)

func NewUserRepository() UserRepository {
	config := database.GetConfigFromEnv()
	if database.DB != nil {
		return &gormRepository{db: database.DB}
	} else if database.MongoDBClient != nil {
		collection := database.MongoDBClient.Database(config.DBName).Collection("users")
		return &mongoRepository{collection}
	} else {
		fmt.Println("No database connection established")
		return nil
	}
}
