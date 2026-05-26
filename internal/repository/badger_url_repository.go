package repository

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// BadgerURLRepository implements URLRepository using BadgerDB
type BadgerURLRepository struct {
	db *badger.DB
}

// NewBadgerURLRepository creates a new BadgerURLRepository instance
func NewBadgerURLRepository(db *badger.DB) *BadgerURLRepository {
	return &BadgerURLRepository{
		db: db,
	}
}

// Save stores a mapping between a short code and original URL
// Key format: url:{shortCode} -> {originalURL}
func (r *BadgerURLRepository) Save(shortCode, originalURL string) error {
	if shortCode == "" {
		return fmt.Errorf("short code cannot be empty")
	}
	if originalURL == "" {
		return fmt.Errorf("original URL cannot be empty")
	}

	key := []byte(fmt.Sprintf("url:%s", shortCode))
	value := []byte(originalURL)

	err := r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})

	if err != nil {
		return fmt.Errorf("failed to save URL mapping for short code %s: %w", shortCode, err)
	}

	return nil
}

// Get retrieves the original URL for a given short code
func (r *BadgerURLRepository) Get(shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("short code cannot be empty")
	}

	key := []byte(fmt.Sprintf("url:%s", shortCode))
	var originalURL string

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			originalURL = string(val)
			return nil
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return "", fmt.Errorf("short code %s not found: %w", shortCode, err)
		}
		return "", fmt.Errorf("failed to retrieve URL for short code %s: %w", shortCode, err)
	}

	return originalURL, nil
}

// Exists checks if a short code already exists in the repository
func (r *BadgerURLRepository) Exists(shortCode string) (bool, error) {
	if shortCode == "" {
		return false, fmt.Errorf("short code cannot be empty")
	}

	key := []byte(fmt.Sprintf("url:%s", shortCode))

	err := r.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence of short code %s: %w", shortCode, err)
	}

	return true, nil
}

// Close closes the BadgerDB connection
func (r *BadgerURLRepository) Close() error {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			return fmt.Errorf("failed to close BadgerDB: %w", err)
		}
	}
	return nil
}
