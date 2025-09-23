package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Collection struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Name   string             `json:"name" bson:"name"`
}

type CollectionUpdate struct {
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
}
