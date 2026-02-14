package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/common/tasks"
	ragdtos "teralux_app/domain/rag/dtos"
	"time"
)

type RAGStatusUseCase interface {
	GetTaskStatus(taskID string) (*ragdtos.RAGStatusDTO, error)
}

type ragStatusUseCase struct {
	cache *tasks.BadgerTaskCache
	store *tasks.StatusStore[ragdtos.RAGStatusDTO]
}

func NewRAGStatusUseCase(cache *tasks.BadgerTaskCache, store *tasks.StatusStore[ragdtos.RAGStatusDTO]) RAGStatusUseCase {
	return &ragStatusUseCase{
		cache: cache,
		store: store,
	}
}

func (u *ragStatusUseCase) GetTaskStatus(taskID string) (*ragdtos.RAGStatusDTO, error) {
	// First try in-memory map from store
	if s, ok := u.store.Get(taskID); ok {
		// augment with TTL info if available
		if u.cache != nil {
			var cached ragdtos.RAGStatusDTO
			ttl, found, err := u.cache.GetWithTTL(taskID, &cached)
			if err == nil && found && ttl > 0 {
				s.ExpiresInSecond = int64(ttl.Seconds())
				s.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
		}
		return s, nil
	}

	// If not found in-memory, try persistent store (Badger) if configured
	if u.cache != nil {
		var status ragdtos.RAGStatusDTO
		ttl, found, err := u.cache.GetWithTTL(taskID, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to read persistent task: %w", err)
		}
		if found {
			// Cache into memory store for faster subsequent reads
			u.store.Set(taskID, &status)

			if ttl > 0 {
				status.ExpiresInSecond = int64(ttl.Seconds())
				status.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
			utils.LogDebug("RAG Task %s: retrieved from badger, ttl=%v", taskID, ttl)
			return &status, nil
		}
		// Not found in badger either
		utils.LogDebug("RAG Task %s: not found in cache", taskID)
	}

	return nil, fmt.Errorf("task not found")
}
