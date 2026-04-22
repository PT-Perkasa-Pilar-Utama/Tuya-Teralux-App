package tasks

import "sync"

// StatusStore keeps task statuses in memory by task ID.
type StatusStore[T any] struct {
	mu         sync.RWMutex
	taskStatus map[string]*T
	taskMu     map[string]*sync.Mutex
}

func NewStatusStore[T any]() *StatusStore[T] {
	return &StatusStore[T]{
		taskStatus: make(map[string]*T),
		taskMu:     make(map[string]*sync.Mutex),
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

// GetTaskMutex returns the mutex for a specific taskID, creating one if needed.
// This allows atomic updates per taskID.
func (s *StatusStore[T]) GetTaskMutex(taskID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskMu[taskID] == nil {
		s.taskMu[taskID] = &sync.Mutex{}
	}
	return s.taskMu[taskID]
}
