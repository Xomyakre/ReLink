package handler

import (
	"ReLink/internal/models"
	"ReLink/internal/service"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// ShortenerHandler handles URL shortening requests
type ShortenerHandler struct {
	shortenerService *service.URLShortenerService
	validator        *validator.Validate
	baseURL          string
}

// NewShortenerHandler creates a new ShortenerHandler instance
func NewShortenerHandler(shortenerService *service.URLShortenerService, baseURL string) *ShortenerHandler {
	return &ShortenerHandler{
		shortenerService: shortenerService,
		validator:        validator.New(),
		baseURL:          baseURL,
	}
}

// CreateShortURL handles POST /api/shorten
// Request: {"url": "https://example.com/very/long/url"}
// Response: {"short_url": "http://localhost:8080/abc123", "short_code": "abc123"}
func (h *ShortenerHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON format")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", "URL is required")
		return
	}

	// Shorten URL
	shortCode, err := h.shortenerService.ShortenURL(req.URL)
	if err != nil {
		// Check if it's a validation error
		respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Build response
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortCode)
	response := models.ShortenResponse{
		ShortURL:  shortURL,
		ShortCode: shortCode,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, errorType, message string) {
	response := models.ErrorResponse{
		Error:   errorType,
		Message: message,
	}
	respondWithJSON(w, statusCode, response)
}
