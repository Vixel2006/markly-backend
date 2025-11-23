package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"markly/internal/models"
	"markly/internal/repositories"
)

type AgentService struct {
	bookmarkRepo   repositories.BookmarkRepository
	categoryRepo   repositories.CategoryRepository
	collectionRepo repositories.CollectionRepository
	tagRepo        repositories.TagRepository
}

func NewAgentService(
	bookmarkRepo repositories.BookmarkRepository,
	categoryRepo repositories.CategoryRepository,
	collectionRepo repositories.CollectionRepository,
	tagRepo repositories.TagRepository,
) *AgentService {
	return &AgentService{
		bookmarkRepo:   bookmarkRepo,
		categoryRepo:   categoryRepo,
		collectionRepo: collectionRepo,
		tagRepo:        tagRepo,
	}
}

func (s *AgentService) GetBookmarkForSummary(userID, bookmarkID primitive.ObjectID) (*models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to retrieve bookmark for summary")
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	bookmark, err := s.bookmarkRepo.FindOne(context.Background(), filter)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Failed to retrieve bookmark for summary")
		return nil, err
	}
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Successfully retrieved bookmark for summary")
	return bookmark, nil
}

func (s *AgentService) UpdateBookmarkSummary(bookmarkID primitive.ObjectID, userID primitive.ObjectID, summary string) error {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to update bookmark summary")
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": bson.M{"summary": summary}}
	_, err := s.bookmarkRepo.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Failed to update bookmark summary")
		return err
	}
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Successfully updated bookmark summary")
	return nil
}

func (s *AgentService) GetPromptBookmarkInfo(userID primitive.ObjectID, bookmarkFilter models.PromptBookmarkFilter) ([]models.PromptBookmarkInfo, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("bookmarkFilter", bookmarkFilter).Msg("Attempting to fetch prompt bookmark info")

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	filter := bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": sevenDaysAgo},
	}

	if bookmarkFilter.BookmarkIDs != nil {
		filter["_id"] = bson.M{"$in": *bookmarkFilter.BookmarkIDs}
	}
	if bookmarkFilter.CategoryID != nil {
		filter["category_id"] = *bookmarkFilter.CategoryID
	}
	if bookmarkFilter.CollectionID != nil {
		filter["collections_id"] = bson.M{"$in": *bookmarkFilter.CollectionID}
	}
	if bookmarkFilter.TagID != nil {
		filter["tags_id"] = bson.M{"$in": *bookmarkFilter.TagID}
	}

	recentBookmarks, err := s.bookmarkRepo.Find(context.Background(), filter, 3, 0)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Interface("filter", filter).Msg("Failed to fetch recent bookmarks")
		return nil, err
	}
	log.Debug().Int("count", len(recentBookmarks)).Msg("Successfully fetched recent bookmarks")

	categories, err := s.categoryRepo.FindByUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	categoryMap := make(map[primitive.ObjectID]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	collections, err := s.collectionRepo.FindByUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch collections: %w", err)
	}
	collectionMap := make(map[primitive.ObjectID]string)
	for _, col := range collections {
		collectionMap[col.ID] = col.Name
	}

	tags, err := s.tagRepo.FindByUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	tagMap := make(map[primitive.ObjectID]string)
	for _, tag := range tags {
		tagMap[tag.ID] = tag.Name
	}

	var promptBookmarks []models.PromptBookmarkInfo
	for _, bm := range recentBookmarks {
		var categoryName string
		if bm.CategoryID != nil {
			categoryName = categoryMap[*bm.CategoryID]
		}

		var collectionName string
		if len(bm.CollectionsID) > 0 {
			collectionName = collectionMap[bm.CollectionsID[0]]
		}

		var tagNames []string
		for _, tagID := range bm.TagsID {
			if tagName, ok := tagMap[tagID]; ok {
				tagNames = append(tagNames, tagName)
			}
		}

		promptBookmarks = append(promptBookmarks, models.PromptBookmarkInfo{
			URL:        bm.URL,
			Title:      bm.Title,
			Summary:    bm.Summary,
			Category:   categoryName,
			Collection: collectionName,
			Tags:       tagNames,
		})
	}
	log.Debug().Int("count", len(promptBookmarks)).Msg("Successfully prepared prompt bookmark info")

	return promptBookmarks, nil
}
