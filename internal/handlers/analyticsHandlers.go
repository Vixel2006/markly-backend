package handlers

import (
	"net/http"
	"time"

	"markly/internal/services"
	"markly/internal/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalyticsHandlers struct {
	AnalyticsService *services.AnalyticsService
}

func NewAnalyticsHandlers(analyticsService *services.AnalyticsService) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		AnalyticsService: analyticsService,
	}
}

func (h *AnalyticsHandlers) GetUserGrowth(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if startDateStr == "" || endDateStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "startDate and endDate are required query parameters")
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid startDate format. Use RFC3339.")
		return
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid endDate format. Use RFC3339.")
		return
	}

	userGrowth, err := h.AnalyticsService.GetUserGrowth(r.Context(), startDate, endDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user growth data")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, userGrowth)
}

func (h *AnalyticsHandlers) GetBookmarkActivity(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if startDateStr == "" || endDateStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "startDate and endDate are required query parameters")
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid startDate format. Use RFC3339.")
		return
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid endDate format. Use RFC3339.")
		return
	}

	bookmarkActivity, err := h.AnalyticsService.GetBookmarkActivity(r.Context(), startDate, endDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve bookmark activity data")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bookmarkActivity)
}

func (h *AnalyticsHandlers) GetBookmarkEngagement(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(primitive.ObjectID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	bookmarkEngagement, err := h.AnalyticsService.GetBookmarkEngagement(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve bookmark engagement data")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bookmarkEngagement)
}

func (h *AnalyticsHandlers) GetTagTrends(w http.ResponseWriter, r *http.Request) {
	tagTrends, err := h.AnalyticsService.GetTagTrends(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag trends data")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, tagTrends)
}

func (h *AnalyticsHandlers) GetTrendingItems(w http.ResponseWriter, r *http.Request) {
	trendingItems, err := h.AnalyticsService.GetTrendingItems(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve trending items data")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, trendingItems)
}
