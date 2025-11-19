package handlers

import (
	"encoding/json"
	"fmt"
	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgentHandler struct {
	db           database.Service
	agentService *services.AgentService
}

func NewAgentHandler(db database.Service) *AgentHandler {
	return &AgentHandler{
		db:           db,
		agentService: services.NewAgentService(db),
	}
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

	bookmark, err := a.agentService.GetBookmarkForSummary(userID, bookmarkID)
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

	if err := a.agentService.UpdateBookmarkSummary(bookmarkID, userID, summary); err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Failed to save summary for bookmark")
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

	var filter models.PromptBookmarkFilter

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

	promptBookmarks, err := a.agentService.GetPromptBookmarkInfo(userID, filter)
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

func (a *AgentHandler) SummarizeURL(w http.ResponseWriter, r *http.Request) {
	var req models.SummarizeURLRequest
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
