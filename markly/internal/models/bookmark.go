package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bookmark struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	URL           string               `bson:"url,omitempty" json:"url"`
	Title         string               `bson:"title,omitempty" json:"title"`
	Summary       string               `bson:"summary" json:"summary"`
	TagsID        []primitive.ObjectID `bson:"tags" json:"tags"`
	CollectionsID []primitive.ObjectID `bson:"collections" json:"collections"`
	CategoryID    *primitive.ObjectID  `bson:"category" json:"category"`
	CreatedAt     primitive.DateTime   `bson:"created_at" json:"created_at"`
	UserID        primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id"`
	IsFav         bool                 `bson:"is_fav" json:"is_fav"`
}

type BookmarkUpdate struct {
	URL           *string               `json:"url,omitempty" bson:"url,omitempty"`
	Title         *string               `json:"title,omitempty" bson:"title,omitempty"`
	Summary       *string               `json:"summary,omitempty" bson:"summary,omitempty"`
	TagsID        *[]primitive.ObjectID `json:"tags,omitempty" bson:"tags,omitempty"`
	CollectionsID *[]primitive.ObjectID `json:"collections,omitempty" bson:"collections,omitempty"`
	CategoryID    *primitive.ObjectID   `json:"category,omitempty" bson:"category,omitempty"`
	IsFav         *bool                 `json:"is_fav,omitempty" bson:"is_fav,omitempty"`
}
