package repository

import "time"

// URLRepository manages persistence of URL mappings
type URLRepository interface {
	// Save stores a mapping between a short code and original URL
	Save(shortCode, originalURL string) error
	
	// Get retrieves the original URL for a given short code
	Get(shortCode string) (string, error)
	
	// Exists checks if a short code already exists in the repository
	Exists(shortCode string) (bool, error)
}

// ClickRepository manages persistence of click events and analytics aggregation
type ClickRepository interface {
	// RecordClick records a single click event for a short code
	RecordClick(shortCode string, timestamp time.Time) error
	
	// GetClickCount retrieves the total number of clicks for a short code
	GetClickCount(shortCode string) (int64, error)
	
	// GetFirstClick retrieves the timestamp of the first click for a short code
	GetFirstClick(shortCode string) (*time.Time, error)
	
	// GetLastClick retrieves the timestamp of the most recent click for a short code
	GetLastClick(shortCode string) (*time.Time, error)
}
