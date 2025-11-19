package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

// TagService defines the interface for tag-related business logic.
type TagService interface {
	AddTag(ctx context.Context, userID primitive.ObjectID, tag models.Tag) (*models.Tag, error)
	GetTagsByID(ctx context.Context, userID primitive.ObjectID, ids []string) ([]models.Tag, error)
	GetUserTags(ctx context.Context, userID primitive.ObjectID) ([]models.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID primitive.ObjectID) (bool, error)
	UpdateTag(ctx context.Context, userID, tagID primitive.ObjectID, updatePayload models.TagUpdate) (*models.Tag, error)
}

// tagServiceImpl implements the TagService interface.
type tagServiceImpl struct {
	db database.Service
}

// NewTagService creates a new TagService.
func NewTagService(db database.Service) TagService {
	return &tagServiceImpl{db: db}
}

func (s *tagServiceImpl) AddTag(ctx context.Context, userID primitive.ObjectID, tag models.Tag) (*models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("tagName", tag.Name).Msg("Attempting to add tag")
	tag.ID = primitive.NewObjectID()
	tag.UserID = userID
	tag.WeeklyCount = 0
	tag.PrevCount = 0
	tag.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	collection := s.db.Client().Database("markly").Collection("tags")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Tag name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("tagName", tag.Name).Msg("Tag name already exists during index creation")
			return nil, fmt.Errorf("tag name already exists")
		} else {
			log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to create index for tag")
			return nil, fmt.Errorf("failed to set up tag collection")
		}
	}

	_, err := collection.InsertOne(ctx, tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("tagName", tag.Name).Msg("Tag name already exists for this user")
			return nil, fmt.Errorf("tag name already exists for this user")
		} else {
			log.Error().Err(err).Str("tag_name", tag.Name).Str("user_id", userID.Hex()).Msg("Failed to insert tag")
			return nil, fmt.Errorf("failed to insert tag")
		}
	}
	log.Info().Str("userID", userID.Hex()).Str("tagID", tag.ID.Hex()).Interface("tagName", tag.Name).Msg("Tag added successfully")
	return &tag, nil
}

func (s *tagServiceImpl) fetchTagByID(ctx context.Context, userID, tagID primitive.ObjectID) (*models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Attempting to fetch tag by ID")
	var tag models.Tag
	filter := bson.M{"_id": tagID, "user_id": userID}
	collection := s.db.Client().Database("markly").Collection("tags")
	err := collection.FindOne(ctx, filter).Decode(&tag)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Failed to fetch tag by ID")
		return nil, err
	}
	log.Debug().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Successfully fetched tag by ID")
	return &tag, nil
}

func (s *tagServiceImpl) GetTagsByID(ctx context.Context, userID primitive.ObjectID, ids []string) ([]models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("tagIDs", ids).Msg("Attempting to retrieve tags by IDs")
	if len(ids) == 0 {
		log.Debug().Str("userID", userID.Hex()).Msg("No tag IDs provided, returning empty list")
		return []models.Tag{}, nil
	}

	type result struct {
		Tag models.Tag
		Err error
	}

	resultsChan := make(chan result, len(ids))
	var wg sync.WaitGroup
	wg.Add(len(ids))

	for _, idStr := range ids {
		idStr := idStr
		go func() {
			defer wg.Done()

			objID, err := primitive.ObjectIDFromHex(strings.TrimSpace(idStr))
			if err != nil {
				log.Error().Err(err).Str("tag_id_string", idStr).Msg("Invalid tag ID format")
				resultsChan <- result{Err: err}
				return
			}

			tag, err := s.fetchTagByID(ctx, userID, objID)
			if err != nil {
				log.Error().Err(err).Str("tag_id", idStr).Str("user_id", userID.Hex()).Msg("Error finding tag")
				resultsChan <- result{Err: err}
				return
			}

			resultsChan <- result{Tag: *tag, Err: nil}
		}()
	}

	wg.Wait()
	close(resultsChan)

	var tags []models.Tag
	for r := range resultsChan {
		if r.Err == nil {
			tags = append(tags, r.Tag)
		}
	}
	log.Debug().Str("userID", userID.Hex()).Int("count", len(tags)).Msg("Successfully retrieved tags by IDs")
	return tags, nil
}

func (s *tagServiceImpl) GetUserTags(ctx context.Context, userID primitive.ObjectID) ([]models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to retrieve user tags")
	var tags []models.Tag
	collection := s.db.Client().Database("markly").Collection("tags")

	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error finding tags for user")
		return nil, fmt.Errorf("failed to retrieve tags")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &tags); err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error decoding tags")
		return nil, fmt.Errorf("error decoding tags")
	}
	log.Debug().Str("userID", userID.Hex()).Int("count", len(tags)).Msg("Successfully retrieved user tags")
	return tags, nil
}

func (s *tagServiceImpl) DeleteTag(ctx context.Context, userID, tagID primitive.ObjectID) (bool, error) {
	log.Debug().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Attempting to delete tag")
	collection := s.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"_id": tagID, "user_id": userID}

	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to delete tag")
		return false, fmt.Errorf("failed to delete tag")
	}

	if deleteResult.DeletedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag not found or unauthorized to delete")
		return false, fmt.Errorf("tag not found or unauthorized to delete")
	}
	log.Info().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag deleted successfully")
	return true, nil
}

func (s *tagServiceImpl) buildTagUpdateFields(updatePayload models.TagUpdate) (bson.M, error) {
	log.Debug().Interface("updatePayload", updatePayload).Msg("Building tag update fields")
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	if updatePayload.WeeklyCount != nil {
		updateFields["weekly_count"] = *updatePayload.WeeklyCount
	}
	if updatePayload.PrevCount != nil {
		updateFields["prev_count"] = *updatePayload.PrevCount
	}
	log.Debug().Interface("updateFields", updateFields).Msg("Tag update fields built successfully")
	return updateFields, nil
}

func (s *tagServiceImpl) UpdateTag(ctx context.Context, userID, tagID primitive.ObjectID, updatePayload models.TagUpdate) (*models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Interface("updatePayload", updatePayload).Msg("Attempting to update tag")
	updateFields, err := s.buildTagUpdateFields(updatePayload)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Failed to build tag update fields")
		return nil, err
	}

	if len(updateFields) == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("No fields to update for tag")
		return nil, fmt.Errorf("no fields to update")
	}

	filter := bson.M{"_id": tagID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := s.db.Client().Database("markly").Collection("tags")

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag name already exists for this user during update")
			return nil, fmt.Errorf("tag name already exists for this user")
		} else {
			log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update tag")
			return nil, fmt.Errorf("failed to update tag")
		}
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag not found or unauthorized to update")
		return nil, fmt.Errorf("tag not found or unauthorized to update")
	}

	var updatedTag models.Tag
	err = collection.FindOne(ctx, filter).Decode(&updatedTag)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated tag")
		return nil, fmt.Errorf("failed to retrieve the updated tag")
	}
	log.Info().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag updated successfully")
	return &updatedTag, nil
}
