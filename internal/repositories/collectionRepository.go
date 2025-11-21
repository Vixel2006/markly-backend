package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
)

// CollectionRepository defines the interface for collection-related database operations.
type CollectionRepository interface {
	Create(ctx context.Context, col *models.Collection) (*models.Collection, error)
	FindByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error)
	Update(ctx context.Context, userID, collectionID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID, collectionID primitive.ObjectID) (*mongo.DeleteResult, error)
}

// collectionRepository implements the CollectionRepository interface.
type collectionRepository struct {
	db database.Service
}

// NewCollectionRepository creates a new CollectionRepository.
func NewCollectionRepository(db database.Service) CollectionRepository {
	return &collectionRepository{db: db}
}

// Create inserts a new collection into the database.
func (r *collectionRepository) Create(ctx context.Context, col *models.Collection) (*models.Collection, error) {
	collection := r.db.Client().Database("markly").Collection("collections")
	_, err := collection.InsertOne(ctx, col)
	if err != nil {
		return nil, fmt.Errorf("failed to insert collection: %w", err)
	}
	return col, nil
}

// FindByID finds a collection by its ID for a specific user.
func (r *collectionRepository) FindByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error) {
	var col models.Collection
	filter := bson.M{"_id": collectionID, "user_id": userID}
	collection := r.db.Client().Database("markly").Collection("collections")
	err := collection.FindOne(ctx, filter).Decode(&col)
	if err != nil {
		return nil, err
	}
	return &col, nil
}

// FindByUser finds all collections for a specific user.
func (r *collectionRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error) {
	var results []models.Collection
	collection := r.db.Client().Database("markly").Collection("collections")
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("database error fetching collections: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("error decoding collection results: %w", err)
	}
	return results, nil
}

// Update updates a collection's information in the database.
func (r *collectionRepository) Update(ctx context.Context, userID, collectionID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("collections")
	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}
	return result, nil
}

// Delete removes a collection from the database.
func (r *collectionRepository) Delete(ctx context.Context, userID, collectionID primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("collections")
	filter := bson.M{"_id": collectionID, "user_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("database error deleting collection: %w", err)
	}
	return result, nil
}
