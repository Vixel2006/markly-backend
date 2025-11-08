package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AgentHandler struct {
	db database.Service
}

func NewAgentHandler(db database.Service) *AgentHandler {
	return &AgentHandler{db: db}
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

	var bookmark models.Bookmark
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	err = a.db.Client().Database("markly").
		Collection("bookmarks").
		FindOne(context.Background(), filter).
		Decode(&bookmark)
	if err != nil {
		utils.SendJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	summary, err := services.LLMSummarize(bookmark.URL, bookmark.Title)
	if err != nil {
		log.Printf("Error generating summary for bookmark %s: %v", bookmarkID.Hex(), err)
		utils.SendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	update := bson.M{"$set": bson.M{"summary": summary}}
	_, err = a.db.Client().Database("markly").
		Collection("bookmarks").
		UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.SendJSONError(w, "Failed to save summary", http.StatusInternalServerError)
		return
	}

	bookmark.Summary = summary
	utils.RespondWithJSON(w, http.StatusOK, bookmark)
}

func (a *AgentHandler) GenerateAISuggestions(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	// Fetch user's bookmarks from the last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	filter := bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": sevenDaysAgo},
	}

	cursor, err := a.db.Client().Database("markly").Collection("bookmarks").Find(context.Background(), filter)
	if err != nil {
		utils.SendJSONError(w, "Failed to fetch recent bookmarks", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var recentBookmarks []models.Bookmark
	if err = cursor.All(context.Background(), &recentBookmarks); err != nil {
		utils.SendJSONError(w, "Failed to decode recent bookmarks", http.StatusInternalServerError)
		return
	}

	// Fetch all categories for the user
	categoryCursor, err := a.db.Client().Database("markly").Collection("categories").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		utils.SendJSONError(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer categoryCursor.Close(context.Background())
	var categories []models.Category
	if err = categoryCursor.All(context.Background(), &categories); err != nil {
		utils.SendJSONError(w, "Failed to decode categories", http.StatusInternalServerError)
		return
	}
	categoryMap := make(map[primitive.ObjectID]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	// Fetch all collections for the user
	collectionCursor, err := a.db.Client().Database("markly").Collection("collections").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		utils.SendJSONError(w, "Failed to fetch collections", http.StatusInternalServerError)
		return
	}
	defer collectionCursor.Close(context.Background())
	var collections []models.Collection
	if err = collectionCursor.All(context.Background(), &collections); err != nil {
		utils.SendJSONError(w, "Failed to decode collections", http.StatusInternalServerError)
		return
	}
	collectionMap := make(map[primitive.ObjectID]string)
	for _, col := range collections {
		collectionMap[col.ID] = col.Name
	}

	// Fetch all tags for the user
	tagCursor, err := a.db.Client().Database("markly").Collection("tags").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		utils.SendJSONError(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer tagCursor.Close(context.Background())
	var tags []models.Tag
	if err = tagCursor.All(context.Background(), &tags); err != nil {
		utils.SendJSONError(w, "Failed to decode tags", http.StatusInternalServerError)
		return
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

	if len(promptBookmarks) == 0 {
		utils.SendJSONError(w, "No recent bookmarks found to generate suggestions from. Please add some bookmarks first.", http.StatusOK)
		return
	}

	// Generate suggestions using LLM
	suggestions, err := services.LLMGenerateSuggestions(promptBookmarks)
	if err != nil {
		log.Printf("Error generating AI suggestions: %v", err)
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
		utils.SendJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		utils.SendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	summary, err := services.LLMSummarize(req.URL, req.Title)
	if err != nil {
		log.Printf("Error generating summary for URL %s: %v", req.URL, err)
		utils.SendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"summary": summary})
}
