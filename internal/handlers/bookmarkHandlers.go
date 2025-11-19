package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
)

type BookmarkHandler struct {
	service services.BookmarkService
}

func NewBookmarksHandler(service services.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{service: service}
}

func (h *BookmarkHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarks, err := h.service.GetBookmarks(r.Context(), userID, r)
	if err != nil {
		log.Error().Err(err).Msg("Error getting bookmarks from service")
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var reqBody models.AddBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error().Err(err).Msg("Error decoding request body for AddBookmark")
		utils.SendJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Debug().Interface("request_body", reqBody).Msg("Received bookmark request")

	bm, err := h.service.AddBookmark(r.Context(), userID, reqBody)
	if err != nil {
		log.Error().Err(err).Msg("Error adding bookmark via service")
		statusCode := http.StatusInternalServerError
		if err.Error() == "URL and Title are required" ||
			(err.Error() == "invalid tag ID format" || err.Error() == "invalid collection ID format" || err.Error() == "invalid category ID format") ||
			(err.Error() == "invalid reference: "+err.Error()) { // This part needs to be more specific if possible
			statusCode = http.StatusBadRequest
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("bookmark_id", bm.ID.Hex()).Msg("Successfully created bookmark")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
}

func (h *BookmarkHandler) GetBookmarkByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	bm, err := h.service.GetBookmarkByID(r.Context(), userID, bookmarkID)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error getting bookmark by ID from service")
		if err.Error() == "bookmark not found" {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bm)
}

func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	deleted, err := h.service.DeleteBookmark(r.Context(), userID, bookmarkID)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error deleting bookmark via service")
		if err.Error() == "bookmark not found or not authorized to delete" {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if deleted {
		log.Info().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark deleted successfully")
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookmarkHandler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.UpdateBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON for UpdateBookmark")
		utils.SendJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedBookmark, err := h.service.UpdateBookmark(r.Context(), userID, bookmarkID, updatePayload)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error updating bookmark via service")
		statusCode := http.StatusInternalServerError
		if err.Error() == "no valid fields provided for update" ||
			(err.Error() == "invalid tag ID format" || err.Error() == "invalid collection ID format" || err.Error() == "invalid category ID format") ||
			(err.Error() == "invalid tag reference" || err.Error() == "invalid collection reference" || err.Error() == "invalid category reference") {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "bookmark not found or not authorized to update" {
			statusCode = http.StatusNotFound
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedBookmark)
}
