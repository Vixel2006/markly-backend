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

type CollectionRepository interface {
	Create(ctx context.Context, col *models.Collection) (*models.Collection, error)
	FindByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error)
	Update(ctx context.Context, userID, collectionID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID, collectionID primitive.ObjectID) (*mongo.DeleteResult, error)
}

type collectionRepository struct {
	db database.Service
}

func NewCollectionRepository(db database.Service) CollectionRepository {
	return &collectionRepository{db: db}
}

func (r *collectionRepository) Create(ctx context.Context, col *models.Collection) (*models.Collection, error) {
	queryType := "create"
	repository := "collection"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("collections")
	_, err := collection.InsertOne(ctx, col)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("failed to insert collection: %w", err)
	}
	return col, nil
}

func (r *collectionRepository) FindByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error) {
	queryType := "findByID"
	repository := "collection"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	var col models.Collection
	filter := bson.M{"_id": collectionID, "user_id": userID}
	collection := r.db.Client().Database("markly").Collection("collections")
	err := collection.FindOne(ctx, filter).Decode(&col)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, err
	}
	return &col, nil
}

func (r *collectionRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error) {
	queryType := "findByUser"
	repository := "collection"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	var results []models.Collection
	collection := r.db.Client().Database("markly").Collection("collections")
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("database error fetching collections: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("error decoding collection results: %w", err)
	}
	return results, nil
}

func (r *collectionRepository) Update(ctx context.Context, userID, collectionID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	queryType := "update"
	repository := "collection"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("collections")
	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}
	return result, nil
}

func (r *collectionRepository) Delete(ctx context.Context, userID, collectionID primitive.ObjectID) (*mongo.DeleteResult, error) {
	queryType := "delete"
	repository := "collection"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("collections")
	filter := bson.M{"_id": collectionID, "user_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, fmt.Errorf("database error deleting collection: %w", err)
	}
	return result, nil
}