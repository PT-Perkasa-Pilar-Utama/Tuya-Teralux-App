package usecases

import (
	"encoding/json"
	"fmt"
	"time"

	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
)

func (u *RAGUsecase) GetStatus(taskID string) (*ragdtos.RAGStatusDTO, error) {
	// First try in-memory map with read lock
	u.mu.RLock()
	if s, ok := u.taskStatus[taskID]; ok {
		u.mu.RUnlock()
		// augment with TTL info if available
		if u.badger != nil {
			key := "rag:task:" + taskID
			_, ttl, err := u.badger.GetWithTTL(key)
			if err == nil && ttl > 0 {
				s.ExpiresInSecond = int64(ttl.Seconds())
				s.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
		}
		return s, nil
	}
	u.mu.RUnlock()

	// If not found in-memory, try persistent store (Badger) if configured
	if u.badger != nil {
		key := "rag:task:" + taskID
		b, ttl, err := u.badger.GetWithTTL(key)
		if err != nil {
			return nil, fmt.Errorf("failed to read persistent task: %w", err)
		}
		if b != nil {
			var status ragdtos.RAGStatusDTO
			if err := json.Unmarshal(b, &status); err == nil {
				// Cache into memory for faster subsequent reads
				u.mu.Lock()
				u.taskStatus[taskID] = &status
				u.mu.Unlock()
				if ttl > 0 {
					status.ExpiresInSecond = int64(ttl.Seconds())
					status.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
				}
				utils.LogDebug("RAG Task %s: retrieved from badger, ttl=%v", taskID, ttl)
				return &status, nil
			}
		}
		// Not found in badger either
		utils.LogDebug("RAG Task %s: not found in cache", taskID)
	}

	return nil, fmt.Errorf("task not found")
}
