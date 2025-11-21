package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"markly/internal/database"
	"markly/internal/models"
)

// BookmarkRepository defines the interface for bookmark-related database operations.
type BookmarkRepository interface {
	Create(ctx context.Context, bm *models.Bookmark) (*models.Bookmark, error)
	Find(ctx context.Context, filter bson.M) ([]models.Bookmark, error)
	FindOne(ctx context.Context, filter bson.M) (*models.Bookmark, error)
	FindWithLimit(ctx context.Context, filter bson.M, limit int64) ([]models.Bookmark, error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error)
}

// bookmarkRepository implements the BookmarkRepository interface.
type bookmarkRepository struct {
	db database.Service
}

// NewBookmarkRepository creates a new BookmarkRepository.
func NewBookmarkRepository(db database.Service) BookmarkRepository {
	return &bookmarkRepository{db: db}
}

// Create inserts a new bookmark into the database.
func (r *bookmarkRepository) Create(ctx context.Context, bm *models.Bookmark) (*models.Bookmark, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.InsertOne(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to add bookmark: %w", err)
	}
	bm.ID = result.InsertedID.(primitive.ObjectID)
	return bm, nil
}

// Find retrieves bookmarks from the database based on a filter.
func (r *bookmarkRepository) Find(ctx context.Context, filter bson.M) ([]models.Bookmark, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bookmarks: %w", err)
	}
	defer cursor.Close(ctx)

	var bookmarks []models.Bookmark
	if err := cursor.All(ctx, &bookmarks); err != nil {
		return nil, fmt.Errorf("error decoding bookmarks: %w", err)
	}
	return bookmarks, nil
}

// FindOne retrieves a single bookmark from the database.
func (r *bookmarkRepository) FindOne(ctx context.Context, filter bson.M) (*models.Bookmark, error) {
	var bm models.Bookmark
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	err := collection.FindOne(ctx, filter).Decode(&bm)
	if err != nil {
		return nil, err
	}
	return &bm, nil
}

// FindWithLimit retrieves bookmarks with a limit.
func (r *bookmarkRepository) FindWithLimit(ctx context.Context, filter bson.M, limit int64) ([]models.Bookmark, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	opts := options.Find().SetLimit(limit)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent bookmarks: %w", err)
	}
	defer cursor.Close(ctx)

	var bookmarks []models.Bookmark
	if err = cursor.All(ctx, &bookmarks); err != nil {
		return nil, fmt.Errorf("failed to decode recent bookmarks: %w", err)
	}
	return bookmarks, nil
}

// UpdateOne updates a single bookmark in the database.
func (r *bookmarkRepository) UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update bookmark: %w", err)
	}
	return result, nil
}

// DeleteOne deletes a single bookmark from the database.
func (r *bookmarkRepository) DeleteOne(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete bookmark: %w", err)
	}
	return deleteResult, nil
}
