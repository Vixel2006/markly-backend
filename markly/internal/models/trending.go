package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type TrendingItem struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Count int                `json:"count" bson:"count"`
}

type AISuggestion struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"` // Category name
	Collection string   `json:"collection"` // Collection name
	Tags       []string `json:"tags"`       // Array of tag names
}

// PromptBookmarkInfo is a temporary struct to hold bookmark data with resolved names for LLM prompting
type PromptBookmarkInfo struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Category   string   `json:"category"`
	Collection string   `json:"collection"`
	Tags       []string `json:"tags"`
}
