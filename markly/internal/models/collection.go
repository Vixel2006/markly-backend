package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Collection struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
}

