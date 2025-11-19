package services

import (
	"context"
	"fmt"
	"strings"


	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

// CollectionService defines the interface for collection-related business logic.
type CollectionService interface {
	AddCollection(ctx context.Context, userID primitive.ObjectID, col models.Collection) (*models.Collection, error)
	GetCollections(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error)
	GetCollectionByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error)
	DeleteCollection(ctx context.Context, userID, collectionID primitive.ObjectID) (bool, error)
	UpdateCollection(ctx context.Context, userID, collectionID primitive.ObjectID, updatePayload models.CollectionUpdate) (*models.Collection, error)
}

// collectionServiceImpl implements the CollectionService interface.
type collectionServiceImpl struct {
	db database.Service
}

// NewCollectionService creates a new CollectionService.
func NewCollectionService(db database.Service) CollectionService {
	return &collectionServiceImpl{db: db}
}

func (s *collectionServiceImpl) AddCollection(ctx context.Context, userID primitive.ObjectID, col models.Collection) (*models.Collection, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("collectionName", col.Name).Msg("Attempting to add collection")
	col.UserID = userID
	col.ID = primitive.NewObjectID()


	collection := s.db.Client().Database("markly").Collection("collections")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Collection name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("collectionName", col.Name).Msg("Collection name already exists during index creation")
			return nil, fmt.Errorf("collection name already exists")
		} else {
			log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to create index for collection")
			return nil, fmt.Errorf("failed to set up collection")
		}
	}

	_, err := collection.InsertOne(ctx, col)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("collectionName", col.Name).Msg("Collection name already exists for this user")
			return nil, fmt.Errorf("collection name already exists for this user")
		} else {
			log.Error().Err(err).Str("collection_name", col.Name).Str("user_id", userID.Hex()).Msg("Failed to insert collection")
			return nil, fmt.Errorf("failed to insert collection")
		}
	}
	log.Info().Str("userID", userID.Hex()).Str("collectionID", col.ID.Hex()).Interface("collectionName", col.Name).Msg("Collection added successfully")
	return &col, nil
}

func (s *collectionServiceImpl) GetCollections(ctx context.Context, userID primitive.ObjectID) ([]models.Collection, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to retrieve collections")
	collection := s.db.Client().Database("markly").Collection("collections")

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Database error fetching collections")
		return nil, fmt.Errorf("database error fetching collections")
	}
	defer cursor.Close(ctx)

	var results []models.Collection
	if err := cursor.All(ctx, &results); err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error decoding collection results")
		return nil, fmt.Errorf("error decoding collection results")
	}
	log.Debug().Str("userID", userID.Hex()).Int("count", len(results)).Msg("Successfully retrieved collections")
	return results, nil
}

func (s *collectionServiceImpl) GetCollectionByID(ctx context.Context, userID, collectionID primitive.ObjectID) (*models.Collection, error) {
	log.Debug().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Attempting to retrieve collection by ID")
	collection := s.db.Client().Database("markly").Collection("collections")

	var col models.Collection
	filter := bson.M{"_id": collectionID, "user_id": userID}
	err := collection.FindOne(ctx, filter).Decode(&col)
	if err == mongo.ErrNoDocuments {
		log.Warn().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection not found or unauthorized")
		return nil, fmt.Errorf("collection not found or unauthorized")
	} else if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Database error finding collection")
		return nil, fmt.Errorf("database error finding collection")
	}
	log.Debug().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Successfully retrieved collection by ID")
	return &col, nil
}

func (s *collectionServiceImpl) DeleteCollection(ctx context.Context, userID, collectionID primitive.ObjectID) (bool, error) {
	log.Debug().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Attempting to delete collection")
	collection := s.db.Client().Database("markly").Collection("collections")

	filter := bson.M{"_id": collectionID, "user_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Database error deleting collection")
		return false, fmt.Errorf("database error deleting collection")
	}
	if result.DeletedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection not found or unauthorized to delete")
		return false, fmt.Errorf("collection not found or unauthorized to delete")
	}
	log.Info().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection deleted successfully")
	return true, nil
}

func (s *collectionServiceImpl) buildCollectionUpdateFields(updatePayload models.CollectionUpdate) (bson.M, error) {
	log.Debug().Interface("updatePayload", updatePayload).Msg("Building collection update fields")
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	log.Debug().Interface("updateFields", updateFields).Msg("Collection update fields built successfully")
	return updateFields, nil
}

func (s *collectionServiceImpl) UpdateCollection(ctx context.Context, userID, collectionID primitive.ObjectID, updatePayload models.CollectionUpdate) (*models.Collection, error) {
	log.Debug().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Interface("updatePayload", updatePayload).Msg("Attempting to update collection")
	updateFields, err := s.buildCollectionUpdateFields(updatePayload)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Failed to build collection update fields")
		return nil, err
	}

	if len(updateFields) == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("No fields to update for collection")
		return nil, fmt.Errorf("no fields to update")
	}

	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := s.db.Client().Database("markly").Collection("collections")

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection name already exists for this user during update")
			return nil, fmt.Errorf("collection name already exists for this user")
		}
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update collection")
		return nil, fmt.Errorf("failed to update collection")
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection not found or unauthorized to update")
		return nil, fmt.Errorf("collection not found or unauthorized to update")
	}

	var updatedCollection models.Collection
	err = collection.FindOne(ctx, filter).Decode(&updatedCollection)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated collection")
		return nil, fmt.Errorf("failed to retrieve the updated collection")
	}
	log.Info().Str("userID", userID.Hex()).Str("collectionID", collectionID.Hex()).Msg("Collection updated successfully")
	return &updatedCollection, nil
}
