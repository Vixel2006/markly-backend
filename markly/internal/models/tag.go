package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Tag struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	WeeklyCount int                `bson:"weekly_count" json:"weekly_count"`
	PrevCount   int                `bson:"prev_count" json:"prev_count"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at"`
}
