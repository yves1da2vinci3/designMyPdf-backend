package mongodb

import "go.mongodb.org/mongo-driver/bson/primitive"

type Namespace struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Templates []Template         `bson:"templates"`
	UserID    primitive.ObjectID `bson:"user_id"`
}
