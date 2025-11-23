package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OTP struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	OTPCode   string             `bson:"otp_code" json:"otp_code"`
	Purpose   string             `bson:"purpose" json:"purpose"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	IsUsed    bool               `bson:"is_used" json:"is_used"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
