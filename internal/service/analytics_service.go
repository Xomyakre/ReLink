package service

import (
	"ReLink/internal/models"
	"ReLink/internal/repository"
	"fmt"
	"sync"
	"time"
)

// AnalyticsService manages click event recording and analytics queries
type AnalyticsService struct {
	clickRepo  repository.ClickRepository
	eventQueue chan models.ClickEvent
	wg         sync.WaitGroup
	stopChan   chan struct{}
	mu         sync.Mutex
	running    bool
}

// NewAnalyticsService creates a new AnalyticsService instance
// queueSize: the size of the buffered channel for async event recording
func NewAnalyticsService(clickRepo repository.ClickRepository, queueSize int) *AnalyticsService {
	if queueSize <= 0 {
		queueSize = 1000 // Default queue size
	}

	return &AnalyticsService{
		clickRepo:  clickRepo,
		eventQueue: make(chan models.ClickEvent, queueSize),
		stopChan:   make(chan struct{}),
		running:    false,
	}
}

// Start begins the background worker that processes click events
// This method should be called before using RecordClickAsync
func (s *AnalyticsService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return // Already running
	}

	s.running = true
	s.wg.Add(1)

	go s.processEvents()
}

// Stop gracefully shuts down the analytics service
// It waits for all queued events to be processed before returning
func (s *AnalyticsService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return // Not running
	}
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

// processEvents is the background worker that processes click events from the queue
func (s *AnalyticsService) processEvents() {
	defer s.wg.Done()

	for {
		select {
		case event := <-s.eventQueue:
			// Process the click event
			if err := s.clickRepo.RecordClick(event.ShortCode, event.Timestamp); err != nil {
				// Log error but don't block (requirement 3.4: recording failures should not block redirects)
				fmt.Printf("Error recording click event for %s: %v\n", event.ShortCode, err)
			}

		case <-s.stopChan:
			// Drain remaining events before stopping
			for {
				select {
				case event := <-s.eventQueue:
					if err := s.clickRepo.RecordClick(event.ShortCode, event.Timestamp); err != nil {
						fmt.Printf("Error recording click event for %s during shutdown: %v\n", event.ShortCode, err)
					}
				default:
					// Queue is empty, exit
					return
				}
			}
		}
	}
}

// RecordClickAsync queues a click event for asynchronous recording
// This method is non-blocking to satisfy Requirement 3.3
func (s *AnalyticsService) RecordClickAsync(shortCode string, timestamp time.Time) {
	event := models.ClickEvent{
		ShortCode: shortCode,
		Timestamp: timestamp,
	}

	// Non-blocking send to queue
	select {
	case s.eventQueue <- event:
		// Event queued successfully
	default:
		// Queue is full, log error but don't block
		fmt.Printf("Warning: analytics queue is full, dropping click event for %s\n", shortCode)
	}
}

// GetAnalytics retrieves analytics data for a given short code
// Returns Analytics struct with click count and timestamps
func (s *AnalyticsService) GetAnalytics(shortCode string) (*models.Analytics, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("short code cannot be empty")
	}

	// Get click count
	clickCount, err := s.clickRepo.GetClickCount(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get click count: %w", err)
	}

	// Get first access timestamp
	firstAccess, err := s.clickRepo.GetFirstClick(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get first access: %w", err)
	}

	// Get last access timestamp
	lastAccess, err := s.clickRepo.GetLastClick(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get last access: %w", err)
	}

	// If no clicks recorded, return analytics with zero count
	if clickCount == 0 {
		return &models.Analytics{
			ShortCode:   shortCode,
			ClickCount:  0,
			FirstAccess: nil,
			LastAccess:  nil,
		}, nil
	}

	return &models.Analytics{
		ShortCode:   shortCode,
		ClickCount:  clickCount,
		FirstAccess: firstAccess,
		LastAccess:  lastAccess,
	}, nil
}

// GetQueueSize returns the current number of events in the queue
// Useful for health checks and monitoring
func (s *AnalyticsService) GetQueueSize() int {
	return len(s.eventQueue)
}
