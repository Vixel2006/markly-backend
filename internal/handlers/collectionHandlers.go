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

type CollectionHandler struct {
	service services.CollectionService
}

func NewCollectionHandler(service services.CollectionService) *CollectionHandler {
	return &CollectionHandler{service: service}
}

func (h *CollectionHandler) AddCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var col models.Collection
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		log.Error().Err(err).Msg("Invalid JSON input for AddCollection")
		utils.SendJSONError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	addedCollection, err := h.service.AddCollection(r.Context(), userID, col)
	if err != nil {
		log.Error().Err(err).Msg("Error adding collection via service")
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("collection_id", addedCollection.ID.Hex()).Str("collection_name", addedCollection.Name).Msg("Collection added successfully")
	utils.RespondWithJSON(w, http.StatusCreated, addedCollection)
}

func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collections, err := h.service.GetCollections(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting collections from service")
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(collections)).Str("user_id", userID.Hex()).Msg("Collections retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, collections)
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	col, err := h.service.GetCollectionByID(r.Context(), userID, collectionID)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Error getting collection by ID from service")
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, col)
}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	deleted, err := h.service.DeleteCollection(r.Context(), userID, collectionID)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Error deleting collection via service")
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if deleted {
		log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection deleted successfully")
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *CollectionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.CollectionUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateCollection")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedCollection, err := h.service.UpdateCollection(r.Context(), userID, collectionID, updatePayload)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Error updating collection via service")
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

	log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedCollection)
}
