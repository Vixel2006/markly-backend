package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PromptBookmarkFilter defines the filter criteria for fetching bookmarks for AI prompting.
type PromptBookmarkFilter struct {
	BookmarkIDs  *[]primitive.ObjectID
	CategoryID   *primitive.ObjectID
	CollectionID *[]primitive.ObjectID
	TagID        *[]primitive.ObjectID
}

// SummarizeURLRequest represents the request body for summarizing a URL.
type SummarizeURLRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// PromptBookmarkInfo represents a simplified bookmark structure for LLM prompting.
type PromptBookmarkInfo struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Collection string   `json:"collection"`
	Tags       []string `json:"tags"`
}
