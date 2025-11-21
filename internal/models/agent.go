package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PromptBookmarkFilter struct {
	BookmarkIDs  *[]primitive.ObjectID
	CategoryID   *primitive.ObjectID
	CollectionID *[]primitive.ObjectID
	TagID        *[]primitive.ObjectID
}

type SummarizeURLRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type PromptBookmarkInfo struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Collection string   `json:"collection"`
	Tags       []string `json:"tags"`
}
