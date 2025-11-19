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
	Category   string   `json:"category"`
	Collection string   `json:"collection"`
	Tags       []string `json:"tags"`
}
