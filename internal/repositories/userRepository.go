package repositories

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
)

// UserRepository defines the interface for user-related database operations.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	Update(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error)
	CountAll(ctx context.Context) (int64, error)
}

// userRepository implements the UserRepository interface.
type userRepository struct {
	db database.Service
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db database.Service) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("Failed to insert user into database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// FindByEmail finds a user by their email address.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err // Can be mongo.ErrNoDocuments
	}
	return &user, nil
}

// FindByID finds a user by their ID.
func (r *userRepository) FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err // Can be mongo.ErrNoDocuments
	}
	return &user, nil
}

// Update updates a user's information in the database.
func (r *userRepository) Update(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error updating user profile")
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return result, nil
}

// Delete removes a user from the database.
func (r *userRepository) Delete(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	filter := bson.M{"_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error deleting user account")
		return nil, fmt.Errorf("failed to delete account: %w", err)
	}
	return result, nil
}

// CountAll counts the total number of users.
func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	collection := r.db.Client().Database("markly").Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to count total users")
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}
	return count, nil
}
