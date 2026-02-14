package tasks

import "sync"

// StatusStore keeps task statuses in memory by task ID.
type StatusStore[T any] struct {
	mu         sync.RWMutex
	taskStatus map[string]*T
}

func NewStatusStore[T any]() *StatusStore[T] {
	return &StatusStore[T]{
		taskStatus: make(map[string]*T),
	}
}

func (s *StatusStore[T]) Set(taskID string, status *T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taskStatus[taskID] = status
}

func (s *StatusStore[T]) Get(taskID string) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.taskStatus[taskID]
	return val, ok
}
