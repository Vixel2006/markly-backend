package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
)

type TagHandler struct {
	service services.TagService
}

func NewTagHandler(service services.TagService) *TagHandler {
	return &TagHandler{service: service}
}
func (h *TagHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		log.Error().Err(err).Msg("Invalid JSON input for AddTag")
		utils.SendJSONError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	addedTag, err := h.service.AddTag(r.Context(), userID, tag)
	if err != nil {
		log.Error().Err(err).Msg("Error adding tag via service")
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("tag_id", addedTag.ID.Hex()).Str("tag_name", addedTag.Name).Msg("Tag added successfully")
	utils.RespondWithJSON(w, http.StatusCreated, addedTag)
}

func (h *TagHandler) GetTagsByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	ids := r.URL.Query()["id"]

	tags, err := h.service.GetTagsByID(r.Context(), userID, ids)
	if err != nil {
		log.Error().Err(err).Msg("Error getting tags by ID from service")
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(tags)).Str("user_id", userID.Hex()).Msg("Tags retrieved by ID successfully")
	utils.RespondWithJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) GetUserTags(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	tags, err := h.service.GetUserTags(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting user tags from service")
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(tags)).Str("user_id", userID.Hex()).Msg("User tags retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	tagID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	deleted, err := h.service.DeleteTag(r.Context(), userID, tagID)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Error deleting tag via service")
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if deleted {
		log.Info().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag deleted successfully")
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	tagID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.TagUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateTag")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedTag, err := h.service.UpdateTag(r.Context(), userID, tagID, updatePayload)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Error updating tag via service")
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "no fields to update") || strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedTag)
}
