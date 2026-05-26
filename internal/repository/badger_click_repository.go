package repository

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerClickRepository implements ClickRepository using BadgerDB
type BadgerClickRepository struct {
	db *badger.DB
}

// NewBadgerClickRepository creates a new BadgerClickRepository instance
func NewBadgerClickRepository(db *badger.DB) *BadgerClickRepository {
	return &BadgerClickRepository{
		db: db,
	}
}

// RecordClick records a single click event for a short code
// Uses key format: click:{shortCode}:{timestamp}
// Also updates aggregated count key: count:{shortCode}
// Also updates first/last click keys if needed
func (r *BadgerClickRepository) RecordClick(shortCode string, timestamp time.Time) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Store individual click event
		clickKey := fmt.Sprintf("click:%s:%d", shortCode, timestamp.UnixNano())
		if err := txn.Set([]byte(clickKey), []byte{}); err != nil {
			return fmt.Errorf("failed to record click event: %w", err)
		}

		// Update aggregated count
		countKey := []byte(fmt.Sprintf("count:%s", shortCode))
		var count int64
		item, err := txn.Get(countKey)
		if err == nil {
			// Count exists, increment it
			err = item.Value(func(val []byte) error {
				if len(val) == 8 {
					count = int64(binary.BigEndian.Uint64(val))
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read click count: %w", err)
			}
		} else if err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get click count: %w", err)
		}
		// Increment count (or set to 1 if it didn't exist)
		count++
		countBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(countBytes, uint64(count))
		if err := txn.Set(countKey, countBytes); err != nil {
			return fmt.Errorf("failed to update click count: %w", err)
		}

		// Update first click if this is the first one
		firstKey := []byte(fmt.Sprintf("first:%s", shortCode))
		_, err = txn.Get(firstKey)
		if err == badger.ErrKeyNotFound {
			// First click doesn't exist, set it
			timestampBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp.UnixNano()))
			if err := txn.Set(firstKey, timestampBytes); err != nil {
				return fmt.Errorf("failed to set first click: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check first click: %w", err)
		}

		// Always update last click
		lastKey := []byte(fmt.Sprintf("last:%s", shortCode))
		timestampBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp.UnixNano()))
		if err := txn.Set(lastKey, timestampBytes); err != nil {
			return fmt.Errorf("failed to update last click: %w", err)
		}

		return nil
	})
}

// GetClickCount retrieves the total number of clicks for a short code
// Uses aggregated count key: count:{shortCode}
func (r *BadgerClickRepository) GetClickCount(shortCode string) (int64, error) {
	var count int64
	err := r.db.View(func(txn *badger.Txn) error {
		countKey := []byte(fmt.Sprintf("count:%s", shortCode))
		item, err := txn.Get(countKey)
		if err == badger.ErrKeyNotFound {
			// No clicks recorded yet, return 0
			count = 0
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get click count: %w", err)
		}

		return item.Value(func(val []byte) error {
			if len(val) == 8 {
				count = int64(binary.BigEndian.Uint64(val))
			}
			return nil
		})
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetFirstClick retrieves the timestamp of the first click for a short code
// Uses key: first:{shortCode}
func (r *BadgerClickRepository) GetFirstClick(shortCode string) (*time.Time, error) {
	var firstClick *time.Time
	err := r.db.View(func(txn *badger.Txn) error {
		firstKey := []byte(fmt.Sprintf("first:%s", shortCode))
		item, err := txn.Get(firstKey)
		if err == badger.ErrKeyNotFound {
			// No first click recorded
			firstClick = nil
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get first click: %w", err)
		}

		return item.Value(func(val []byte) error {
			if len(val) == 8 {
				nanos := int64(binary.BigEndian.Uint64(val))
				t := time.Unix(0, nanos)
				firstClick = &t
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return firstClick, nil
}

// GetLastClick retrieves the timestamp of the most recent click for a short code
// Uses key: last:{shortCode}
func (r *BadgerClickRepository) GetLastClick(shortCode string) (*time.Time, error) {
	var lastClick *time.Time
	err := r.db.View(func(txn *badger.Txn) error {
		lastKey := []byte(fmt.Sprintf("last:%s", shortCode))
		item, err := txn.Get(lastKey)
		if err == badger.ErrKeyNotFound {
			// No last click recorded
			lastClick = nil
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get last click: %w", err)
		}

		return item.Value(func(val []byte) error {
			if len(val) == 8 {
				nanos := int64(binary.BigEndian.Uint64(val))
				t := time.Unix(0, nanos)
				lastClick = &t
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return lastClick, nil
}
