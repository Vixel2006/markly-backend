package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bookmark struct {
    ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
    Title     string               `bson:"title,omitempty" json:"title"`
    Tags      []string             `bson:"tags" json:"tags"`
    CreatedAt primitive.DateTime   `bson:"datetime" json:"datetime"`
}

