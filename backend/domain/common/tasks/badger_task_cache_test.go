package tasks

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// MockBadgerService for testing BadgerTaskCache
type MockBadgerService struct {
	SetFunc            func(key string, value []byte) error
	SetPreserveTTLFunc func(key string, value []byte) error
	GetWithTTLFunc     func(key string) ([]byte, time.Duration, error)
	data               map[string][]byte
	ttls               map[string]time.Duration
}

func NewMockBadgerService() *MockBadgerService {
	return &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}
}

func (m *MockBadgerService) Set(key string, value []byte) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value)
	}
	m.data[key] = value
	m.ttls[key] = 3600 * time.Second // Default TTL
	return nil
}

func (m *MockBadgerService) SetPreserveTTL(key string, value []byte) error {
	if m.SetPreserveTTLFunc != nil {
		return m.SetPreserveTTLFunc(key, value)
	}
	m.data[key] = value
	// Preserve existing TTL
	if _, exists := m.ttls[key]; !exists {
		m.ttls[key] = 3600 * time.Second
	}
	return nil
}

func (m *MockBadgerService) GetWithTTL(key string) ([]byte, time.Duration, error) {
	if m.GetWithTTLFunc != nil {
		return m.GetWithTTLFunc(key)
	}
	data, ok := m.data[key]
	if !ok {
		return nil, 0, nil
	}
	ttl := m.ttls[key]
	return data, ttl, nil
}

type CacheTestStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TestNewBadgerTaskCache(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	if cache.keyPrefix != "test:task:" {
		t.Errorf("expected keyPrefix 'test:task:', got '%s'", cache.keyPrefix)
	}
}

func TestBadgerTaskCache_Key(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	key := cache.key("123")
	if key != "test:task:123" {
		t.Errorf("expected key 'test:task:123', got '%s'", key)
	}
}

func TestBadgerTaskCache_Set(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	status := &CacheTestStatus{
		Status:  "pending",
		Message: "Task pending",
	}

	err := cache.Set("task-1", status)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify data was stored
	data, ok := mockBadger.data["test:task:task-1"]
	if !ok {
		t.Fatal("expected data to be stored")
	}

	var retrieved CacheTestStatus
	if err := json.Unmarshal(data, &retrieved); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if retrieved.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", retrieved.Status)
	}
}

func TestBadgerTaskCache_SetPreserveTTL(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	// Set initial data with TTL
	initialStatus := &CacheTestStatus{Status: "pending", Message: "Initial"}
	cache.Set("task-2", initialStatus)

	// Store original TTL
	originalTTL := mockBadger.ttls["test:task:task-2"]

	// Update with SetPreserveTTL
	updatedStatus := &CacheTestStatus{Status: "completed", Message: "Updated"}
	err := cache.SetPreserveTTL("task-2", updatedStatus)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify TTL was preserved
	currentTTL := mockBadger.ttls["test:task:task-2"]
	if currentTTL != originalTTL {
		t.Errorf("expected TTL to be preserved")
	}

	// Verify data was updated
	data := mockBadger.data["test:task:task-2"]
	var retrieved CacheTestStatus
	json.Unmarshal(data, &retrieved)
	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", retrieved.Status)
	}
}

func TestBadgerTaskCache_GetWithTTL(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	// Set data
	status := &CacheTestStatus{Status: "running", Message: "In progress"}
	cache.Set("task-3", status)

	// Get with TTL
	var retrieved CacheTestStatus
	ttl, found, err := cache.GetWithTTL("task-3", &retrieved)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !found {
		t.Fatal("expected data to be found")
	}
	if ttl == 0 {
		t.Error("expected non-zero TTL")
	}
	if retrieved.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", retrieved.Status)
	}
}

func TestBadgerTaskCache_GetWithTTL_NotFound(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	var retrieved CacheTestStatus
	ttl, found, err := cache.GetWithTTL("non-existent", &retrieved)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found {
		t.Error("expected data not to be found")
	}
	if ttl != 0 {
		t.Error("expected zero TTL for non-existent key")
	}
}

func TestBadgerTaskCache_SetError(t *testing.T) {
	mockBadger := &MockBadgerService{
		SetFunc: func(key string, value []byte) error {
			return errors.New("badger error")
		},
	}
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	status := &CacheTestStatus{Status: "pending"}
	err := cache.Set("task-error", status)

	if err == nil {
		t.Fatal("expected error from badger")
	}
	if err.Error() != "badger error" {
		t.Errorf("expected 'badger error', got '%v'", err)
	}
}

func TestBadgerTaskCache_GetWithTTL_Error(t *testing.T) {
	mockBadger := &MockBadgerService{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			return nil, 0, errors.New("get error")
		},
	}
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	var retrieved CacheTestStatus
	_, _, err := cache.GetWithTTL("task-error", &retrieved)

	if err == nil {
		t.Fatal("expected error from badger")
	}
}

func TestBadgerTaskCache_NilBadger(t *testing.T) {
	cache := NewBadgerTaskCache(nil, "test:task:")

	status := &CacheTestStatus{Status: "pending"}

	// All operations should handle nil badger gracefully
	err := cache.Set("task-nil", status)
	if err != nil {
		t.Errorf("expected no error with nil badger, got %v", err)
	}

	err = cache.SetPreserveTTL("task-nil", status)
	if err != nil {
		t.Errorf("expected no error with nil badger, got %v", err)
	}

	var retrieved CacheTestStatus
	ttl, found, err := cache.GetWithTTL("task-nil", &retrieved)
	if err != nil {
		t.Errorf("expected no error with nil badger, got %v", err)
	}
	if found {
		t.Error("expected not found with nil badger")
	}
	if ttl != 0 {
		t.Error("expected zero TTL with nil badger")
	}
}

func TestBadgerTaskCache_NilCache(t *testing.T) {
	var cache *BadgerTaskCache

	status := &CacheTestStatus{Status: "pending"}

	// All operations should handle nil cache gracefully
	err := cache.Set("task-nil", status)
	if err != nil {
		t.Errorf("expected no error with nil cache, got %v", err)
	}

	err = cache.SetPreserveTTL("task-nil", status)
	if err != nil {
		t.Errorf("expected no error with nil cache, got %v", err)
	}

	var retrieved CacheTestStatus
	ttl, found, err := cache.GetWithTTL("task-nil", &retrieved)
	if err != nil {
		t.Errorf("expected no error with nil cache, got %v", err)
	}
	if found {
		t.Error("expected not found with nil cache")
	}
	if ttl != 0 {
		t.Error("expected zero TTL with nil cache")
	}
}

func TestBadgerTaskCache_MarshalError(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	// Use a type that can't be marshaled
	invalidData := make(chan int)

	err := cache.Set("task-invalid", invalidData)
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestBadgerTaskCache_UnmarshalError(t *testing.T) {
	mockBadger := &MockBadgerService{
		GetWithTTLFunc: func(key string) ([]byte, time.Duration, error) {
			// Return invalid JSON
			return []byte("invalid json"), 3600 * time.Second, nil
		},
	}
	cache := NewBadgerTaskCache(mockBadger, "test:task:")

	var retrieved CacheTestStatus
	_, found, err := cache.GetWithTTL("task-invalid", &retrieved)

	if err == nil {
		t.Fatal("expected unmarshal error")
	}
	if found {
		t.Error("expected not found on unmarshal error")
	}
}
