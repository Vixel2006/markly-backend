package repositories

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus" // Added for Prometheus
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils" // Added for Prometheus metrics
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) (*models.Category, error)
	FindByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error)
	Update(ctx context.Context, userID, categoryID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID, categoryID primitive.ObjectID) (*mongo.DeleteResult, error)
}

type categoryRepository struct {
	db database.Service
}

func NewCategoryRepository(db database.Service) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	queryType := "create"
	repository := "category"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("categories")
	_, err := collection.InsertOne(ctx, category)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("failed to insert category: %w", err)
	}
	return category, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error) {
	queryType := "findByID"
	repository := "category"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	var category models.Category
	filter := bson.M{"_id": categoryID, "user_id": userID}
	collection := r.db.Client().Database("markly").Collection("categories")
	err := collection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error) {
	queryType := "findByUser"
	repository := "category"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	var categories []models.Category
	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("error fetching categories: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &categories); err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("error decoding categories: %w", err)
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, userID, categoryID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	queryType := "update"
	repository := "category"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"_id": categoryID, "user_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return result, nil
}

func (r *categoryRepository) Delete(ctx context.Context, userID, categoryID primitive.ObjectID) (*mongo.DeleteResult, error) {
	queryType := "delete"
	repository := "category"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"user_id": userID, "_id": categoryID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}
	return result, nil
}