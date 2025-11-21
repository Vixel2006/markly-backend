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

// CategoryRepository defines the interface for category-related database operations.
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) (*models.Category, error)
	FindByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error)
	Update(ctx context.Context, userID, categoryID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID, categoryID primitive.ObjectID) (*mongo.DeleteResult, error)
}

// categoryRepository implements the CategoryRepository interface.
type categoryRepository struct {
	db database.Service
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db database.Service) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create inserts a new category into the database.
func (r *categoryRepository) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	collection := r.db.Client().Database("markly").Collection("categories")
	_, err := collection.InsertOne(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to insert category: %w", err)
	}
	return category, nil
}

// FindByID finds a category by its ID for a specific user.
func (r *categoryRepository) FindByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error) {
	var category models.Category
	filter := bson.M{"_id": categoryID, "user_id": userID}
	collection := r.db.Client().Database("markly").Collection("categories")
	err := collection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// FindByUser finds all categories for a specific user.
func (r *categoryRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error) {
	var categories []models.Category
	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error fetching categories: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("error decoding categories: %w", err)
	}
	return categories, nil
}

// Update updates a category's information in the database.
func (r *categoryRepository) Update(ctx context.Context, userID, categoryID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"_id": categoryID, "user_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return result, nil
}

// Delete removes a category from the database.
func (r *categoryRepository) Delete(ctx context.Context, userID, categoryID primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("categories")
	filter := bson.M{"user_id": userID, "_id": categoryID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}
	return result, nil
}
