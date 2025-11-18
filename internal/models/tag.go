package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Tag struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	WeeklyCount int                `json:"weeklyCount" bson:"weekly_count"`
	PrevCount   int                `json:"prevCount" bson:"prev_count"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"created_at"`
}

type TagUpdate struct {
	Name        *string `json:"name,omitempty" bson:"name,omitempty"`
	WeeklyCount *int    `json:"weeklyCount,omitempty" bson:"weekly_count,omitempty"`
	PrevCount   *int    `json:"prevCount,omitempty" bson:"prev_count,omitempty"`
}
