package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

type UserProfileUpdate struct {
	Username string  `json:"username,omitempty" bson:"username,omitempty"`
	Email    *string `json:"email,omitempty" bson:"email,omitempty"`
	Password *string `json:"password,omitempty" bson:"password,omitempty"`
}
