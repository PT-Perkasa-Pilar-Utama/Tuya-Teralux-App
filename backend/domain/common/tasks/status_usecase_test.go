package tasks

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type TestStatusWithExpiry struct {
	Status          string `json:"status"`
	Message         string `json:"message"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	ExpiresInSecond int64  `json:"expires_in_seconds,omitempty"`
}

func (s *TestStatusWithExpiry) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSecond = expiresInSeconds
}

type TestStatusNoExpiry struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MockBadgerStore for testing BadgerTaskCache
type MockBadgerStore struct {
	SetFunc            func(key string, value []byte) error
	SetPreserveTTLFunc func(key string, value []byte) error
	GetWithTTLFunc     func(key string) ([]byte, time.Duration, error)
}

func (m *MockBadgerStore) Set(key string, value []byte) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value)
	}
	return nil
}

func (m *MockBadgerStore) SetPreserveTTL(key string, value []byte) error {
	if m.SetPreserveTTLFunc != nil {
		return m.SetPreserveTTLFunc(key, value)
	}
	return nil
}

func (m *MockBadgerStore) GetWithTTL(key string) ([]byte, time.Duration, error) {
	if m.GetWithTTLFunc != nil {
		return m.GetWithTTLFunc(key)
	}
	return nil, 0, nil
}

func TestNewGenericStatusUseCase(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()
	mockStore := &MockBadgerStore{}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	if usecase == nil {
		t.Fatal("expected non-nil usecase")
	}
}

func TestGenericStatusUseCase_GetFromStore(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()
	mockStore := &MockBadgerStore{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			return nil, 3600 * time.Second, nil
		},
	}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	// Set status in store
	status := &TestStatusWithExpiry{
		Status:  "completed",
		Message: "Task finished",
	}
	store.Set("task-123", status)

	// Retrieve status
	retrieved, err := usecase.GetTaskStatus("task-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", retrieved.Status)
	}
}

func TestGenericStatusUseCase_GetFromCache(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()

	mockStore := &MockBadgerStore{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			data, _ := json.Marshal(&TestStatusWithExpiry{
				Status:  "pending",
				Message: "From cache",
			})
			return data, 1800 * time.Second, nil
		},
	}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	retrieved, err := usecase.GetTaskStatus("task-456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", retrieved.Status)
	}
}

func TestGenericStatusUseCase_TaskNotFound(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()
	mockStore := &MockBadgerStore{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			return nil, 0, nil
		},
	}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	_, err := usecase.GetTaskStatus("non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
}

func TestGenericStatusUseCase_CacheError(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()
	mockStore := &MockBadgerStore{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			return nil, 0, errors.New("cache error")
		},
	}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	_, err := usecase.GetTaskStatus("task-789")
	if err == nil {
		t.Fatal("expected error when cache fails")
	}
}

func TestGenericStatusUseCase_NoExpiryInterface(t *testing.T) {
	store := NewStatusStore[TestStatusNoExpiry]()
	mockStore := &MockBadgerStore{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			return nil, 3600 * time.Second, nil
		},
	}
	cache := NewBadgerTaskCache(mockStore, "test:")

	usecase := NewGenericStatusUseCase(cache, store)

	status := &TestStatusNoExpiry{
		Status:  "running",
		Message: "No expiry",
	}
	store.Set("task-no-expiry", status)

	retrieved, err := usecase.GetTaskStatus("task-no-expiry")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", retrieved.Status)
	}
}

func TestGenericStatusUseCase_NilCache(t *testing.T) {
	store := NewStatusStore[TestStatusWithExpiry]()

	usecase := NewGenericStatusUseCase(nil, store)

	status := &TestStatusWithExpiry{
		Status:  "completed",
		Message: "No cache",
	}
	store.Set("task-no-cache", status)

	retrieved, err := usecase.GetTaskStatus("task-no-cache")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", retrieved.Status)
	}
}
