package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatusCode int

const (
	Success StatusCode = 200
	Fail    StatusCode = 500
)

type Log struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CalledAt     time.Time          `bson:"called_at"`
	RequestBody  interface{}        `bson:"request_body"`
	ResponseBody interface{}        `bson:"response_body"`
	StatusCode   StatusCode         `bson:"status_code"`
	ErrorMessage string             `bson:"error_message,omitempty"`
	TemplateID   primitive.ObjectID `bson:"template_id"`
	KeyID        primitive.ObjectID `bson:"key_id"`
}
