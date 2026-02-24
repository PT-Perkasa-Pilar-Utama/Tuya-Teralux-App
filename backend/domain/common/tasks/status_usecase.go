package tasks

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"time"
)

// StatusWithExpiry is an interface for status DTOs that support expiration tracking.
// Implement this interface to enable automatic TTL info population.
type StatusWithExpiry interface {
	SetExpiry(expiresAt string, expiresInSeconds int64)
}

// GenericStatusUseCase provides common task status retrieval logic.
// It handles both in-memory and persistent storage with TTL support.
type GenericStatusUseCase[T any] interface {
	GetTaskStatus(taskID string) (*T, error)
}

type genericStatusUseCase[T any] struct {
	cache *BadgerTaskCache
	store *StatusStore[T]
}

// NewGenericStatusUseCase creates a new generic status usecase.
// This can be used by any domain to retrieve task status from both
// in-memory store and persistent cache (Badger).
//
// Parameters:
//   - cache: BadgerTaskCache for persistent storage (can be nil)
//   - store: StatusStore for in-memory storage (required)
//
// Returns:
//   - GenericStatusUseCase: A generic status usecase instance
func NewGenericStatusUseCase[T any](cache *BadgerTaskCache, store *StatusStore[T]) GenericStatusUseCase[T] {
	return &genericStatusUseCase[T]{
		cache: cache,
		store: store,
	}
}

// GetTaskStatus retrieves task status from storage.
// It first checks the in-memory store, then falls back to persistent cache.
// For status DTOs that implement StatusWithExpiry interface, TTL info is automatically populated.
//
// Parameters:
//   - taskID: The unique task identifier
//
// Returns:
//   - *T: The task status DTO
//   - error: Error if task not found or retrieval fails
func (u *genericStatusUseCase[T]) GetTaskStatus(taskID string) (*T, error) {
	// First try in-memory map from store
	if s, ok := u.store.Get(taskID); ok {
		// Augment with TTL info if available and if type supports it
		if u.cache != nil {
			var cached T
			ttl, found, err := u.cache.GetWithTTL(taskID, &cached)
			if err == nil && found && ttl > 0 {
				// Try to set expiry info if the type supports it
				if expirable, ok := any(s).(StatusWithExpiry); ok {
					expiresAt := time.Now().Add(ttl).UTC().Format(time.RFC3339)
					expiresInSeconds := int64(ttl.Seconds())
					expirable.SetExpiry(expiresAt, expiresInSeconds)
				}
			}
		}
		return s, nil
	}

	// If not found in-memory, try persistent store (Badger) if configured
	if u.cache != nil {
		var status T
		ttl, found, err := u.cache.GetWithTTL(taskID, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to read persistent task: %w", err)
		}
		if found {
			// Cache into memory store for faster subsequent reads
			u.store.Set(taskID, &status)

			// Set expiry info if type supports it
			if ttl > 0 {
				if expirable, ok := any(&status).(StatusWithExpiry); ok {
					expiresAt := time.Now().Add(ttl).UTC().Format(time.RFC3339)
					expiresInSeconds := int64(ttl.Seconds())
					expirable.SetExpiry(expiresAt, expiresInSeconds)
				}
			}
			utils.LogDebug("Task %s: retrieved from badger, ttl=%v", taskID, ttl)
			return &status, nil
		}
		// Not found in badger either
		utils.LogDebug("Task %s: not found in cache", taskID)
	}

	return nil, fmt.Errorf("task not found")
}
