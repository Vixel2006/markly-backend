package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Name   string             `json:"name" bson:"name"`
	Emoji  string             `json:"emoji,omitempty" bson:"emoji,omitempty"`
}

type CategoryUpdate struct {
	Name  *string `json:"name,omitempty" bson:"name,omitempty"`
	Emoji *string `json:"emoji,omitempty" bson:"emoji,omitempty"`
}
