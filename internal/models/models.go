package models

import "time"

// URLMapping represents the mapping between a short code and an original URL
type URLMapping struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// ClickEvent represents a single access event to a shortened URL
type ClickEvent struct {
	ShortCode string    `json:"short_code"`
	Timestamp time.Time `json:"timestamp"`
	// Future: UserAgent, IP, Referer for extended analytics
}

// Analytics represents aggregated analytics data for a shortened URL
type Analytics struct {
	ShortCode   string     `json:"short_code"`
	ClickCount  int64      `json:"click_count"`
	FirstAccess *time.Time `json:"first_access,omitempty"`
	LastAccess  *time.Time `json:"last_access,omitempty"`
}

// ShortenRequest represents the request body for URL shortening
type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// ShortenResponse represents the response for URL shortening
type ShortenResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
