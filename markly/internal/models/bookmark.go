package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bookmark struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title         string               `bson:"title,omitempty" json:"title"`
	Summary       string               `bson:"summary" json:"summary"`
	TagsID        []primitive.ObjectID `bson:"tags" json:"tags"`
	CollectionsID []primitive.ObjectID `bson:"collections" json:"collections"`
	CategoryID    primitive.ObjectID   `bson:"category" json:"category"`
	CreatedAt     primitive.DateTime   `bson:"datetime" json:"datetime"`
	UserID        primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id"`
	IsFav         bool                 `bson:"is_fav" json:"is_fav"`
}

type BookmarkUpdate struct {
	Title *string   `json:"title,omitempty"`
	Tags  *[]string `json:"tags,omitempty"`
}

