package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bookmark struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UserID        primitive.ObjectID   `json:"user_id" bson:"user_id"`
	URL           string               `json:"url" bson:"url"`
	Title         string               `json:"title" bson:"title"`
	Summary       string               `json:"summary,omitempty" bson:"summary,omitempty"`
	TagsID        []primitive.ObjectID `json:"tags,omitempty" bson:"tagsid,omitempty"`
	CollectionsID []primitive.ObjectID `json:"collections,omitempty" bson:"collectionsid,omitempty"`
	CategoryID    *primitive.ObjectID  `json:"category,omitempty" bson:"categoryid,omitempty"`
	IsFav         bool                 `json:"is_fav" bson:"is_fav"`
	CreatedAt     primitive.DateTime   `json:"created_at" bson:"created_at"`
}

type BookmarkUpdate struct {
	URL           *string               `json:"url,omitempty" bson:"url,omitempty"`
	Title         *string               `json:"title,omitempty" bson:"title,omitempty"`
	Summary       *string               `json:"summary,omitempty" bson:"summary,omitempty"`
	TagsID        *[]primitive.ObjectID `json:"tags,omitempty" bson:"tagsid,omitempty"`
	CollectionsID *[]primitive.ObjectID `json:"collections,omitempty" bson:"collectionsid,omitempty"`
	CategoryID    *primitive.ObjectID   `json:"category,omitempty" bson:"categoryid,omitempty"`
	IsFav         *bool                 `json:"is_fav,omitempty" bson:"is_fav,omitempty"`
}

type AddBookmarkRequestBody struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Collections []string `json:"collections,omitempty"`
	CategoryID  *string  `json:"category_id,omitempty"`
	IsFav       bool     `json:"is_fav"`
}
