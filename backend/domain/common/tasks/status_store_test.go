package tasks

import (
	"sync"
	"testing"
)

type TestStatusDTO struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TestNewStatusStore(t *testing.T) {
	store := NewStatusStore[TestStatusDTO]()

	if store == nil {
		t.Fatal("expected non-nil store")
	}

	if store.taskStatus == nil {
		t.Error("expected taskStatus map to be initialized")
	}
}

func TestStatusStore_SetAndGet(t *testing.T) {
	store := NewStatusStore[TestStatusDTO]()

	status := &TestStatusDTO{
		Status:  "completed",
		Message: "Task finished successfully",
	}

	// Test Set
	store.Set("task-123", status)

	// Test Get
	retrieved, ok := store.Get("task-123")
	if !ok {
		t.Fatal("expected task to be found")
	}

	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", retrieved.Status)
	}
	if retrieved.Message != "Task finished successfully" {
		t.Errorf("expected message 'Task finished successfully', got '%s'", retrieved.Message)
	}
}

func TestStatusStore_GetNonExistent(t *testing.T) {
	store := NewStatusStore[TestStatusDTO]()

	_, ok := store.Get("non-existent")
	if ok {
		t.Error("expected task not to be found")
	}
}

func TestStatusStore_UpdateExisting(t *testing.T) {
	store := NewStatusStore[TestStatusDTO]()

	// Set initial status
	initial := &TestStatusDTO{
		Status:  "pending",
		Message: "Task is pending",
	}
	store.Set("task-456", initial)

	// Update status
	updated := &TestStatusDTO{
		Status:  "completed",
		Message: "Task completed",
	}
	store.Set("task-456", updated)

	// Verify update
	retrieved, ok := store.Get("task-456")
	if !ok {
		t.Fatal("expected task to be found")
	}

	if retrieved.Status != "completed" {
		t.Errorf("expected updated status 'completed', got '%s'", retrieved.Status)
	}
	if retrieved.Message != "Task completed" {
		t.Errorf("expected updated message 'Task completed', got '%s'", retrieved.Message)
	}
}

func TestStatusStore_Concurrency(t *testing.T) {
	store := NewStatusStore[TestStatusDTO]()
	var wg sync.WaitGroup

	// Spawn multiple goroutines to set values concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := string(rune('A' + (id % 26)))
			status := &TestStatusDTO{
				Status:  "completed",
				Message: "Concurrent write",
			}
			store.Set(taskID, status)
		}(i)
	}

	// Spawn multiple goroutines to read values concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := string(rune('A' + (id % 26)))
			store.Get(taskID)
		}(i)
	}

	wg.Wait()

	// Verify some values were set
	count := 0
	for i := 0; i < 26; i++ {
		taskID := string(rune('A' + i))
		if _, ok := store.Get(taskID); ok {
			count++
		}
	}

	if count == 0 {
		t.Error("expected some tasks to be set during concurrent writes")
	}
}

func TestStatusStore_MultipleTypes(t *testing.T) {
	// Test with different types
	stringStore := NewStatusStore[string]()
	stringStore.Set("task-1", ptr("string status"))

	retrieved, ok := stringStore.Get("task-1")
	if !ok || *retrieved != "string status" {
		t.Error("string store failed")
	}

	// Test with int
	intStore := NewStatusStore[int]()
	intStore.Set("task-2", ptr(42))

	retrievedInt, ok := intStore.Get("task-2")
	if !ok || *retrievedInt != 42 {
		t.Error("int store failed")
	}
}

// Helper function to create pointers
func ptr[T any](v T) *T {
	return &v
}
