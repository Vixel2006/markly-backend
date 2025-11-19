package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"markly/internal/database"
	"markly/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AgentService provides business logic for agent-related operations.
type AgentService struct {
	db database.Service
}

// NewAgentService creates a new AgentService.
func NewAgentService(db database.Service) *AgentService {
	return &AgentService{db: db}
}

// GetBookmarkForSummary retrieves a specific bookmark for summarization.
func (s *AgentService) GetBookmarkForSummary(userID, bookmarkID primitive.ObjectID) (*models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to retrieve bookmark for summary")
	var bookmark models.Bookmark
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	err := s.db.Client().Database("markly").
		Collection("bookmarks").
		FindOne(context.Background(), filter).
		Decode(&bookmark)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Failed to retrieve bookmark for summary")
		return nil, err
	}
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Successfully retrieved bookmark for summary")
	return &bookmark, nil
}

// UpdateBookmarkSummary updates the summary of a specific bookmark.
func (s *AgentService) UpdateBookmarkSummary(bookmarkID primitive.ObjectID, userID primitive.ObjectID, summary string) error {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to update bookmark summary")
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": bson.M{"summary": summary}}
	_, err := s.db.Client().Database("markly").
		Collection("bookmarks").
		UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Failed to update bookmark summary")
		return err
	}
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Successfully updated bookmark summary")
	return nil
}

// GetPromptBookmarkInfo fetches bookmark information relevant for AI prompting based on a filter.
func (s *AgentService) GetPromptBookmarkInfo(userID primitive.ObjectID, bookmarkFilter models.PromptBookmarkFilter) ([]models.PromptBookmarkInfo, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("bookmarkFilter", bookmarkFilter).Msg("Attempting to fetch prompt bookmark info")

	// Fetch user's bookmarks from the last 7 days
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

	opts := options.Find().SetLimit(3)

	cursor, err := s.db.Client().Database("markly").Collection("bookmarks").Find(context.Background(), filter, opts)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Interface("filter", filter).Msg("Failed to fetch recent bookmarks")
		return nil, fmt.Errorf("failed to fetch recent bookmarks: %w", err)
	}
	defer cursor.Close(context.Background())

	var recentBookmarks []models.Bookmark
	if err = cursor.All(context.Background(), &recentBookmarks); err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to decode recent bookmarks")
		return nil, fmt.Errorf("failed to decode recent bookmarks: %w", err)
	}
	log.Debug().Int("count", len(recentBookmarks)).Msg("Successfully fetched recent bookmarks")

	// Fetch all categories for the user
	categoryCursor, err := s.db.Client().Database("markly").Collection("categories").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to fetch categories")
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer categoryCursor.Close(context.Background())
	var categories []models.Category
	if err = categoryCursor.All(context.Background(), &categories); err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to decode categories")
		return nil, fmt.Errorf("failed to decode categories: %w", err)
	}
	categoryMap := make(map[primitive.ObjectID]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}
	log.Debug().Int("count", len(categories)).Msg("Successfully fetched categories")

	// Fetch all collections for the user
	collectionCursor, err := s.db.Client().Database("markly").Collection("collections").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to fetch collections")
		return nil, fmt.Errorf("failed to fetch collections: %w", err)
	}
	defer collectionCursor.Close(context.Background())
	var collections []models.Collection
	if err = collectionCursor.All(context.Background(), &collections); err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to decode collections")
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}
	collectionMap := make(map[primitive.ObjectID]string)
	for _, col := range collections {
		collectionMap[col.ID] = col.Name
	}
	log.Debug().Int("count", len(collections)).Msg("Successfully fetched collections")

	// Fetch all tags for the user
	tagCursor, err := s.db.Client().Database("markly").Collection("tags").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to fetch tags")
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer tagCursor.Close(context.Background())
	var tags []models.Tag
	if err = tagCursor.All(context.Background(), &tags); err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to decode tags")
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}
	tagMap := make(map[primitive.ObjectID]string)
	for _, tag := range tags {
		tagMap[tag.ID] = tag.Name
	}
	log.Debug().Int("count", len(tags)).Msg("Successfully fetched tags")

	var promptBookmarks []models.PromptBookmarkInfo
	for _, bm := range recentBookmarks {
		var categoryName string
		if bm.CategoryID != nil {
			categoryName = categoryMap[*bm.CategoryID]
		}

		var collectionName string
		if len(bm.CollectionsID) > 0 {
			// For simplicity, taking the first collection name if multiple exist
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
