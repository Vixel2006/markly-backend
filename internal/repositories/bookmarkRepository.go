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

type BookmarkRepository interface {
	Create(ctx context.Context, bm *models.Bookmark) (*models.Bookmark, error)
	Find(ctx context.Context, filter bson.M, limit, page int64) ([]models.Bookmark, error)
	FindOne(ctx context.Context, filter bson.M) (*models.Bookmark, error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error)
}

type bookmarkRepository struct {
	db database.Service
}

func NewBookmarkRepository(db database.Service) BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) Create(ctx context.Context, bm *models.Bookmark) (*models.Bookmark, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.InsertOne(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to add bookmark: %w", err)
	}
	bm.ID = result.InsertedID.(primitive.ObjectID)
	return bm, nil
}

func (r *bookmarkRepository) Find(ctx context.Context, filter bson.M, limit, page int64) ([]models.Bookmark, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	opts := options.Find().SetLimit(limit).SetSkip((page - 1) * limit)

	cursor, err := collection.Find(ctx, filter, opts)

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

func (r *bookmarkRepository) FindOne(ctx context.Context, filter bson.M) (*models.Bookmark, error) {
	var bm models.Bookmark
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	err := collection.FindOne(ctx, filter).Decode(&bm)
	if err != nil {
		return nil, err
	}
	return &bm, nil
}

func (r *bookmarkRepository) UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update bookmark: %w", err)
	}
	return result, nil
}

func (r *bookmarkRepository) DeleteOne(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("bookmarks")
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete bookmark: %w", err)
	}
	return deleteResult, nil
}
