package handler

import (
	"ReLink/internal/service"
	"net/http"

	"github.com/gorilla/mux"
)

// AnalyticsHandler handles analytics query requests
type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler instance
func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetAnalytics handles GET /api/analytics/{shortCode}
// Response: {"short_code": "abc123", "click_count": 42, "first_access": "2024-01-01T00:00:00Z", "last_access": "2024-01-15T12:30:00Z"}
func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Extract short code from URL path
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	if shortCode == "" {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Short code is required")
		return
	}

	// Get analytics data
	analytics, err := h.analyticsService.GetAnalytics(shortCode)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "not_found", "Analytics not found for short code")
		return
	}

	respondWithJSON(w, http.StatusOK, analytics)
}
