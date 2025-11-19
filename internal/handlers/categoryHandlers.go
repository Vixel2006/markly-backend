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

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) AddCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		log.Error().Err(err).Msg("Invalid JSON for AddCategory")
		utils.SendJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	addedCategory, err := h.service.AddCategory(r.Context(), userID, category)
	if err != nil {
		log.Error().Err(err).Msg("Error adding category via service")
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	log.Info().Str("category_id", addedCategory.ID.Hex()).Str("category_name", addedCategory.Name).Msg("Category added successfully")
	utils.RespondWithJSON(w, http.StatusCreated, addedCategory)
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categories, err := h.service.GetCategories(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting categories from service")
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(categories)).Str("user_id", userID.Hex()).Msg("Categories retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	category, err := h.service.GetCategoryByID(r.Context(), userID, categoryID)
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Error getting category by ID from service")
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Info().Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Category retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	deleted, err := h.service.DeleteCategory(r.Context(), userID, categoryID)
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Error deleting category via service")
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONError(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if deleted {
		log.Info().Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Category deleted successfully")
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateCategory")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedCategory, err := h.service.UpdateCategory(r.Context(), userID, categoryID, updatePayload)
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Error updating category via service")
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

	log.Info().Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Category updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedCategory)
}
