package main

import (
	"ReLink/internal/config"
	"ReLink/internal/handler"
	"ReLink/internal/repository"
	"ReLink/internal/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting ReLink URL Shortener on port %d", cfg.ServerPort)
	log.Printf("Base URL: %s", cfg.BaseURL)
	log.Printf("Database path: %s", cfg.DBPath)

	// Initialize BadgerDB
	db, err := initBadgerDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize BadgerDB: %v", err)
	}
	defer db.Close()
	log.Println("BadgerDB initialized successfully")

	// Create repository instances
	urlRepo := repository.NewBadgerURLRepository(db)
	clickRepo := repository.NewBadgerClickRepository(db)

	// Create service instances
	codeGenerator := service.NewShortCodeGenerator(cfg.ShortCodeLength)
	shortenerService := service.NewURLShortenerService(urlRepo, codeGenerator)
	analyticsService := service.NewAnalyticsService(clickRepo, cfg.AnalyticsQueueSize)

	// Start analytics service background worker
	analyticsService.Start()
	defer analyticsService.Stop()
	log.Println("Analytics service started")

	// Create handler instances
	shortenerHandler := handler.NewShortenerHandler(shortenerService, cfg.BaseURL)
	redirectHandler := handler.NewRedirectHandler(shortenerService, analyticsService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	// Set up HTTP router
	router := mux.NewRouter()

	// Serve static files (HTML, CSS, JS)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	// Serve index.html at root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")

	// Register API routes
	router.HandleFunc("/api/shorten", shortenerHandler.CreateShortURL).Methods("POST")
	router.HandleFunc("/api/analytics/{shortCode}", analyticsHandler.GetAnalytics).Methods("GET")
	router.HandleFunc("/health", createHealthHandler(db, analyticsService)).Methods("GET")
	
	// Redirect route (must be last to avoid conflicts)
	router.HandleFunc("/{shortCode}", redirectHandler.Redirect).Methods("GET")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on http://localhost:%d", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// initBadgerDB initializes and returns a BadgerDB instance
func initBadgerDB(dbPath string) (*badger.DB, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open BadgerDB
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable BadgerDB logging for cleaner output

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	return db, nil
}

// createHealthHandler creates a health check handler
func createHealthHandler(db *badger.DB, analyticsService *service.AnalyticsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected"
		}

		// Get analytics queue size
		queueSize := analyticsService.GetQueueSize()

		// Build response
		response := map[string]interface{}{
			"status":          "healthy",
			"database":        dbStatus,
			"analytics_queue": queueSize,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"%s","database":"%s","analytics_queue":%d}`, 
			response["status"], response["database"], response["analytics_queue"])
	}
}
