package mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FrameworkType string

const (
	Bootstrap FrameworkType = "bootstrap"
	Tailwind  FrameworkType = "tailwind"
)

type Template struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Content     string             `bson:"content"`
	Framework   FrameworkType      `bson:"framework"`
	Variables   interface{}        `bson:"variables"`
	Fonts       []string           `bson:"fonts"`
	Logs        []Log              `bson:"logs"`
	NamespaceID primitive.ObjectID `bson:"namespace_id"`
}
