package user

import (
	"context"
	"designmypdf/pkg/enities/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB implementation of UserRepository
type mongoRepository struct {
	db *mongo.Collection
}

func (r *mongoRepository) Create(user interface{}) error {
	_, err := r.db.InsertOne(context.Background(), user)
	return err
}

func (r *mongoRepository) Get(id interface{}) (interface{}, error) {
	var user mongodb.User
	objID, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		return nil, err
	}
	err = r.db.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoRepository) Update(user interface{}) error {
	usr := user.(*mongodb.User)
	_, err := r.db.UpdateOne(context.Background(), bson.M{"_id": usr.ID}, bson.M{"$set": usr})
	return err
}

func (r *mongoRepository) Delete(user interface{}) error {
	usr := user.(*mongodb.User)
	_, err := r.db.DeleteOne(context.Background(), bson.M{"_id": usr.ID})
	return err
}

func (r *mongoRepository) GetAll() ([]interface{}, error) {
	cursor, err := r.db.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []mongodb.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	result := make([]interface{}, len(users))
	for i, user := range users {
		result[i] = user
	}
	return result, nil
}

func (r *mongoRepository) GetByEmail(email string) (interface{}, error) {
	var user mongodb.User
	err := r.db.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoRepository) GetByUserName(userName string) (interface{}, error) {
	var user mongodb.User
	err := r.db.FindOne(context.Background(), bson.M{"user_name": userName}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoRepository) GetByUserNameAndPassword(userName string, password string) (interface{}, error) {
	var user mongodb.User
	err := r.db.FindOne(context.Background(), bson.M{"user_name": userName, "password": password}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
