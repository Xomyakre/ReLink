package handler

import (
	"ReLink/internal/service"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RedirectHandler handles redirect requests and triggers analytics
type RedirectHandler struct {
	shortenerService *service.URLShortenerService
	analyticsService *service.AnalyticsService
}

// NewRedirectHandler creates a new RedirectHandler instance
func NewRedirectHandler(shortenerService *service.URLShortenerService, analyticsService *service.AnalyticsService) *RedirectHandler {
	return &RedirectHandler{
		shortenerService: shortenerService,
		analyticsService: analyticsService,
	}
}

// Redirect handles GET /{shortCode}
// Redirects to original URL with 302 status
func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Extract short code from URL path
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	if shortCode == "" {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Short code is required")
		return
	}

	// Get original URL
	originalURL, err := h.shortenerService.GetOriginalURL(shortCode)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "not_found", "Short code not found")
		return
	}

	// Record click event asynchronously (non-blocking, Requirement 3.3)
	h.analyticsService.RecordClickAsync(shortCode, time.Now())

	// Perform redirect (HTTP 302 - temporary redirect)
	http.Redirect(w, r, originalURL, http.StatusFound)
}
