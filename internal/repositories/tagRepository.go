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

type TagRepository interface {
	Create(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	FindByID(ctx context.Context, userID, tagID primitive.ObjectID) (*models.Tag, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Tag, error)
	Update(ctx context.Context, userID, tagID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, userID, tagID primitive.ObjectID) (*mongo.DeleteResult, error)
}

type tagRepository struct {
	db database.Service
}

func NewTagRepository(db database.Service) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	collection := r.db.Client().Database("markly").Collection("tags")
	_, err := collection.InsertOne(ctx, tag)
	if err != nil {
		log.Error().Err(err).Str("tag_name", tag.Name).Str("user_id", tag.UserID.Hex()).Msg("Failed to insert tag")
		return nil, fmt.Errorf("failed to insert tag: %w", err)
	}
	return tag, nil
}

func (r *tagRepository) FindByID(ctx context.Context, userID, tagID primitive.ObjectID) (*models.Tag, error) {
	var tag models.Tag
	filter := bson.M{"_id": tagID, "user_id": userID}
	collection := r.db.Client().Database("markly").Collection("tags")
	err := collection.FindOne(ctx, filter).Decode(&tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Tag, error) {
	var tags []models.Tag
	collection := r.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tags: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &tags); err != nil {
		return nil, fmt.Errorf("error decoding tags: %w", err)
	}
	return tags, nil
}

func (r *tagRepository) Update(ctx context.Context, userID, tagID primitive.ObjectID, updateFields bson.M) (*mongo.UpdateResult, error) {
	collection := r.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"_id": tagID, "user_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}
	return result, nil
}

func (r *tagRepository) Delete(ctx context.Context, userID, tagID primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := r.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"_id": tagID, "user_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete tag: %w", err)
	}
	return result, nil
}
