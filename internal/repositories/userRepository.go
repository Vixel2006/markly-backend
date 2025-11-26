package repositories

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	Update(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error)
	CountAll(ctx context.Context) (int64, error)
	CountUsersCreatedBetween(ctx context.Context, startDate, endDate interface{}) (int64, error)
}

type userRepository struct {
	db database.Service
}

func NewUserRepository(db database.Service) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	queryType := "create"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Str("email", user.Email).Msg("Failed to insert user into database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	queryType := "findByEmail"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	queryType := "findById"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, err // Can be mongo.ErrNoDocuments
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	queryType := "update"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error updating user profile")
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return result, nil
}

func (r *userRepository) Delete(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error) {
	queryType := "delete"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	filter := bson.M{"_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error deleting user account")
		return nil, fmt.Errorf("failed to delete account: %w", err)
	}
	return result, nil
}

func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	queryType := "countAll"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Msg("Failed to count total users")
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}
	return count, nil
}

func (r *userRepository) CountUsersCreatedBetween(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	queryType := "countUsersCreatedBetween"
	repository := "user"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("users")
	filter := bson.M{
		"createdAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Msg("Failed to count users created between dates")
		return 0, fmt.Errorf("failed to count users created between dates: %w", err)
	}
	return count, nil
}
