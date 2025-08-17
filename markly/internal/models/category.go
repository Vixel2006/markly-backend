package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
}

