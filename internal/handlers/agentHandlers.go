package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AgentHandler struct {
	db database.Service
}

func NewAgentHandler(db database.Service) *AgentHandler {
	return &AgentHandler{db: db}
}

func (a *AgentHandler) getBookmarkForSummary(userID, bookmarkID primitive.ObjectID) (*models.Bookmark, error) {
	var bookmark models.Bookmark
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	err := a.db.Client().Database("markly").
		Collection("bookmarks").
		FindOne(context.Background(), filter).
		Decode(&bookmark)
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (a *AgentHandler) updateBookmarkSummary(bookmarkID primitive.ObjectID, userID primitive.ObjectID, summary string) error {
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": bson.M{"summary": summary}}
	_, err := a.db.Client().Database("markly").
		Collection("bookmarks").
		UpdateOne(context.Background(), filter, update)
	return err
}

func (a *AgentHandler) GenerateSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	bookmark, err := a.getBookmarkForSummary(userID, bookmarkID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendJSONError(w, "Bookmark not found", http.StatusNotFound)
		} else {
			log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error fetching bookmark")
			utils.SendJSONError(w, "Failed to retrieve bookmark", http.StatusInternalServerError)
		}
		return
	}

	summary, err := services.LLMSummarize(bookmark.URL, bookmark.Title)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error generating summary for bookmark")
		utils.SendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	if err := a.updateBookmarkSummary(bookmarkID, userID, summary); err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Failed to save summary for bookmark")
		utils.SendJSONError(w, "Failed to save summary", http.StatusInternalServerError)
		return
	}

	bookmark.Summary = summary
	utils.RespondWithJSON(w, http.StatusOK, bookmark)
}

type PromptBookmarkFilter struct {
	BookmarkIDs  *[]primitive.ObjectID
	CategoryID   *primitive.ObjectID
	CollectionID *[]primitive.ObjectID
	TagID        *[]primitive.ObjectID
}

func (a *AgentHandler) getPromptBookmarkInfo(userID primitive.ObjectID, bookmarkFilter PromptBookmarkFilter) ([]models.PromptBookmarkInfo, error) {
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
		filter["category"] = *bookmarkFilter.CategoryID
	}

	if bookmarkFilter.CollectionID != nil {
		filter["collections"] = bson.M{"$in": *bookmarkFilter.CollectionID}
	}

	if bookmarkFilter.TagID != nil {
		filter["tags"] = bson.M{"$in": *bookmarkFilter.TagID}
	}

	opts := options.Find().SetLimit(3)

	cursor, err := a.db.Client().Database("markly").Collection("bookmarks").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent bookmarks: %w", err)
	}
	defer cursor.Close(context.Background())

	var recentBookmarks []models.Bookmark
	if err = cursor.All(context.Background(), &recentBookmarks); err != nil {
		return nil, fmt.Errorf("failed to decode recent bookmarks: %w", err)
	}

	// Fetch all categories for the user
	categoryCursor, err := a.db.Client().Database("markly").Collection("categories").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer categoryCursor.Close(context.Background())
	var categories []models.Category
	if err = categoryCursor.All(context.Background(), &categories); err != nil {
		return nil, fmt.Errorf("failed to decode categories: %w", err)
	}
	categoryMap := make(map[primitive.ObjectID]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	// Fetch all collections for the user
	collectionCursor, err := a.db.Client().Database("markly").Collection("collections").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch collections: %w", err)
	}
	defer collectionCursor.Close(context.Background())
	var collections []models.Collection
	if err = collectionCursor.All(context.Background(), &collections); err != nil {
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}
	collectionMap := make(map[primitive.ObjectID]string)
	for _, col := range collections {
		collectionMap[col.ID] = col.Name
	}

	// Fetch all tags for the user
	tagCursor, err := a.db.Client().Database("markly").Collection("tags").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer tagCursor.Close(context.Background())
	var tags []models.Tag
	if err = tagCursor.All(context.Background(), &tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
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

	return promptBookmarks, nil
}

func (a *AgentHandler) GenerateAISuggestions(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var filter PromptBookmarkFilter

	bookmarkParams := r.URL.Query().Get("bookmarks")

	if bookmarkParams != "" {
		bookmarkIDs, err := utils.ParseObjectIDs(bookmarkParams)
		if err != nil {
			log.Error().Err(err).Str("bookmark_params", bookmarkParams).Msg("Invalid bookmark ID format")
			utils.SendJSONError(w, "Invalid bookmark ID format", http.StatusBadRequest)
			return
		}
		filter.BookmarkIDs = &bookmarkIDs
	}

	categoryParam := r.URL.Query().Get("category")

	if categoryParam != "" {
		categoryID, err := primitive.ObjectIDFromHex(categoryParam)
		if err != nil {
			log.Error().Err(err).Str("category_param", categoryParam).Msg("Invalid category ID format")
			utils.SendJSONError(w, "Invalid category ID format", http.StatusBadRequest)
			return
		}
		filter.CategoryID = &categoryID
	}

	collectionParam := r.URL.Query().Get("collection")

	if collectionParam != "" {
		collectionID, err := utils.ParseObjectIDs(collectionParam)
		if err != nil {
			log.Error().Err(err).Str("collection_param", collectionParam).Msg("Invalid collection ID format")
			utils.SendJSONError(w, "Invalid collection ID format", http.StatusBadRequest)
			return
		}
		filter.CollectionID = &collectionID
	}

	tagParam := r.URL.Query().Get("tag")

	if tagParam != "" {
		tagID, err := utils.ParseObjectIDs(tagParam)
		if err != nil {
			log.Error().Err(err).Str("tag_param", tagParam).Msg("Invalid tag ID format")
			utils.SendJSONError(w, "Invalid tag ID format", http.StatusBadRequest)
			return
		}
		filter.TagID = &tagID
	}

	promptBookmarks, err := a.getPromptBookmarkInfo(userID, filter)
	if err != nil {
		log.Error().Err(err).Msg("Error preparing prompt bookmark info")
		utils.SendJSONError(w, fmt.Sprintf("Failed to prepare AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	if len(promptBookmarks) == 0 {
		log.Info().Msg("No recent bookmarks found to generate suggestions from")
		utils.SendJSONError(w, "No recent bookmarks found to generate suggestions from. Please add some bookmarks first.", http.StatusOK)
		return
	}

	// Generate suggestions using LLM
	suggestions, err := services.LLMGenerateSuggestions(promptBookmarks)
	if err != nil {
		log.Error().Err(err).Msg("Error generating AI suggestions")
		utils.SendJSONError(w, fmt.Sprintf("Failed to generate AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, suggestions)
}

type SummarizeURLRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (a *AgentHandler) SummarizeURL(w http.ResponseWriter, r *http.Request) {
	var req SummarizeURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request payload for SummarizeURL")
		utils.SendJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		log.Error().Msg("URL is required for SummarizeURL")
		utils.SendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	summary, err := services.LLMSummarize(req.URL, req.Title)
	if err != nil {
		log.Error().Err(err).Str("url", req.URL).Msg("Error generating summary for URL")
		utils.SendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"summary": summary})
}
