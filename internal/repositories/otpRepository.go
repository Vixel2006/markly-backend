package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/models"
)

type OTPRepository interface {
	Create(ctx context.Context, otp *models.OTP) (*models.OTP, error)
	FindByUserIDAndOTPCode(ctx context.Context, userID primitive.ObjectID, otpCode string, purpose string) (*models.OTP, error)
	FindByUserEmailAndOTPCodeAndPurpose(ctx context.Context, email string, otpCode string, purpose string) (*models.OTP, error)
	MarkAsUsed(ctx context.Context, otpID primitive.ObjectID) error
	DeleteExpiredOTPs(ctx context.Context) error
}

type otpRepository struct {
	collection *mongo.Collection
	userRepo   UserRepository
}

func NewOTPRepository(db *mongo.Database, userRepo UserRepository) OTPRepository {
	return &otpRepository{collection: db.Collection("otps"), userRepo: userRepo}
}

func (r *otpRepository) Create(ctx context.Context, otp *models.OTP) (*models.OTP, error) {
	otp.ID = primitive.NewObjectID()
	otp.CreatedAt = time.Now()
	otp.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, otp)
	if err != nil {
		return nil, err
	}
	return otp, nil
}

func (r *otpRepository) FindByUserIDAndOTPCode(ctx context.Context, userID primitive.ObjectID, otpCode string, purpose string) (*models.OTP, error) {
	var otp models.OTP
	filter := bson.M{"user_id": userID, "otp_code": otpCode, "purpose": purpose, "is_used": false, "expires_at": bson.M{"$gt": time.Now()}}
	err := r.collection.FindOne(ctx, filter).Decode(&otp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) MarkAsUsed(ctx context.Context, otpID primitive.ObjectID) error {
	filter := bson.M{"_id": otpID}
	update := bson.M{"$set": bson.M{"is_used": true, "updated_at": time.Now()}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *otpRepository) DeleteExpiredOTPs(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}, "is_used": false}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *otpRepository) FindByUserEmailAndOTPCodeAndPurpose(ctx context.Context, email string, otpCode string, purpose string) (*models.OTP, error) {
	user, err := r.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	var otp models.OTP
	filter := bson.M{"user_id": user.ID, "otp_code": otpCode, "purpose": purpose, "is_used": false, "expires_at": bson.M{"$gt": time.Now()}}
	err = r.collection.FindOne(ctx, filter).Decode(&otp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &otp, nil
}
