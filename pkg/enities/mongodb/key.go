package mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Key struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	Value        string             `bson:"value"`
	KeyCount     int                `bson:"key_count"`
	KeyCountUsed int                `bson:"key_count_used"`
	Logs         []Log              `bson:"logs"`
	UserID       primitive.ObjectID `bson:"user_id"`
}
