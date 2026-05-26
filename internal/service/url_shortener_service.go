package service

import (
	"ReLink/internal/repository"
	"fmt"
	"net/url"
	"sync"
)

// URLShortenerService provides business logic for URL shortening and retrieval
type URLShortenerService struct {
	urlRepo       repository.URLRepository
	codeGenerator *ShortCodeGenerator
	mu            sync.RWMutex
	maxRetries    int
}

// NewURLShortenerService creates a new URLShortenerService instance
func NewURLShortenerService(urlRepo repository.URLRepository, codeGenerator *ShortCodeGenerator) *URLShortenerService {
	return &URLShortenerService{
		urlRepo:       urlRepo,
		codeGenerator: codeGenerator,
		maxRetries:    3, // Maximum attempts to generate unique short code
	}
}

// ShortenURL creates a shortened URL for the given original URL
// Returns the generated short code or an error
func (s *URLShortenerService) ShortenURL(originalURL string) (string, error) {
	// Validate URL
	if err := s.ValidateURL(originalURL); err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Try to generate a unique short code with retries
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		shortCode, err := s.codeGenerator.Generate()
		if err != nil {
			return "", fmt.Errorf("failed to generate short code: %w", err)
		}

		// Check if the short code already exists
		exists, err := s.urlRepo.Exists(shortCode)
		if err != nil {
			return "", fmt.Errorf("failed to check short code existence: %w", err)
		}

		if !exists {
			// Short code is unique, save the mapping
			if err := s.urlRepo.Save(shortCode, originalURL); err != nil {
				return "", fmt.Errorf("failed to save URL mapping: %w", err)
			}
			return shortCode, nil
		}

		// Collision detected, retry
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", s.maxRetries)
}

// GetOriginalURL retrieves the original URL for a given short code
func (s *URLShortenerService) GetOriginalURL(shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("short code cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, err := s.urlRepo.Get(shortCode)
	if err != nil {
		return "", fmt.Errorf("short code not found: %w", err)
	}

	return originalURL, nil
}

// ValidateURL checks if a URL is valid and well-formed
// Requirements: URL must have http or https scheme and a valid host
func (s *URLShortenerService) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check for required scheme (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: must be http or https, got %s", parsedURL.Scheme)
	}

	// Check for valid host
	if parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}

	return nil
}
