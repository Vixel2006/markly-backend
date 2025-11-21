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

	"markly/internal/models"
	"markly/internal/repositories"
)

type TagService interface {
	AddTag(ctx context.Context, userID primitive.ObjectID, tag models.Tag) (*models.Tag, error)
	GetTagsByID(ctx context.Context, userID primitive.ObjectID, ids []string) ([]models.Tag, error)
	GetUserTags(ctx context.Context, userID primitive.ObjectID) ([]models.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID primitive.ObjectID) (bool, error)
	UpdateTag(ctx context.Context, userID, tagID primitive.ObjectID, updatePayload models.TagUpdate) (*models.Tag, error)
}

type tagServiceImpl struct {
	tagRepo repositories.TagRepository
}

func NewTagService(tagRepo repositories.TagRepository) TagService {
	return &tagServiceImpl{tagRepo: tagRepo}
}

func (s *tagServiceImpl) AddTag(ctx context.Context, userID primitive.ObjectID, tag models.Tag) (*models.Tag, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("tagName", tag.Name).Msg("Attempting to add tag")
	tag.ID = primitive.NewObjectID()
	tag.UserID = userID
	tag.WeeklyCount = 0
	tag.PrevCount = 0
	tag.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	createdTag, err := s.tagRepo.Create(ctx, &tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("tagName", tag.Name).Msg("Tag name already exists for this user")
			return nil, fmt.Errorf("tag name already exists for this user")
		}
		return nil, err
	}
	log.Info().Str("userID", userID.Hex()).Str("tagID", createdTag.ID.Hex()).Interface("tagName", createdTag.Name).Msg("Tag added successfully")
	return createdTag, nil
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

			tag, err := s.tagRepo.FindByID(ctx, userID, objID)
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
	tags, err := s.tagRepo.FindByUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error finding tags for user")
		return nil, err
	}
	log.Debug().Str("userID", userID.Hex()).Int("count", len(tags)).Msg("Successfully retrieved user tags")
	return tags, nil
}

func (s *tagServiceImpl) DeleteTag(ctx context.Context, userID, tagID primitive.ObjectID) (bool, error) {
	log.Debug().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Attempting to delete tag")
	result, err := s.tagRepo.Delete(ctx, userID, tagID)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to delete tag")
		return false, err
	}

	if result.DeletedCount == 0 {
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

	result, err := s.tagRepo.Update(ctx, userID, tagID, updateFields)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag name already exists for this user during update")
			return nil, fmt.Errorf("tag name already exists for this user")
		}
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update tag")
		return nil, fmt.Errorf("failed to update tag")
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag not found or unauthorized to update")
		return nil, fmt.Errorf("tag not found or unauthorized to update")
	}

	updatedTag, err := s.tagRepo.FindByID(ctx, userID, tagID)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated tag")
		return nil, fmt.Errorf("failed to retrieve the updated tag")
	}
	log.Info().Str("userID", userID.Hex()).Str("tagID", tagID.Hex()).Msg("Tag updated successfully")
	return updatedTag, nil
}
