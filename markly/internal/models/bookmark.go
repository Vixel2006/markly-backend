package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bookmark struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title     string               `bson:"title,omitempty" json:"title"`
	Tags      []string             `bson:"tags" json:"tags"`
	CreatedAt primitive.DateTime   `bson:"datetime" json:"datetime"`
	UserID    primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id"`
}

type BookmarkUpdate struct {
	Title *string   `json:"title,omitempty"`
	Tags  *[]string `json:"tags,omitempty"`
}

