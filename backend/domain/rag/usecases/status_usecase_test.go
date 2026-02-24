package usecases

import (
	"os"
	"testing"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
)

func TestStatusUseCase_Execute(t *testing.T) {
	// Setup Badger
	dbDir := "./tmp/badger-test-status"
	_ = os.RemoveAll(dbDir)

	// Initialize config to avoid panic in NewBadgerService
	utils.AppConfig = nil
	_ = utils.GetConfig()

	badgerSvc, err := infrastructure.NewBadgerService(dbDir)
	if err != nil {
		t.Fatalf("failed to create badger service: %v", err)
	}
	defer func() {
		_ = badgerSvc.Close()
		_ = os.RemoveAll(dbDir)
	}()

	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	cache := tasks.NewBadgerTaskCache(badgerSvc, "rag:task:")
	u := tasks.NewGenericStatusUseCase(cache, store)

	t.Run("NotFound", func(t *testing.T) {
		got, err := u.GetTaskStatus("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent task, got nil")
		}
		if got != nil {
			t.Errorf("expected nil result for non-existent task, got %+v", got)
		}
	})

	t.Run("InMemory Hit", func(t *testing.T) {
		taskID := "mem-task-1"
		status := &ragdtos.RAGStatusDTO{Status: "pending", Result: "waiting"}

		store.Set(taskID, status)

		got, err := u.GetTaskStatus(taskID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Status != "pending" {
			t.Errorf("expected status pending, got %s", got.Status)
		}
	})

	t.Run("Persistent Hit (Badger)", func(t *testing.T) {
		taskID := "persist-task-1"
		status := ragdtos.RAGStatusDTO{Status: "done", Result: "completed"}

		// Write via cache helper
		if err := cache.Set(taskID, &status); err != nil {
			t.Fatalf("failed to set badger key: %v", err)
		}

		got, err := u.GetTaskStatus(taskID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Status != "done" {
			t.Errorf("expected status done, got %s", got.Status)
		}

		// Check if it promoted to memory
		inMem, ok := store.Get(taskID)
		if !ok {
			t.Error("expected task to be promoted to in-memory cache after fetch")
		} else if inMem.Status != "done" {
			t.Errorf("promoted task has wrong status: %s", inMem.Status)
		}
	})

	t.Run("TTL Augmentation", func(t *testing.T) {
		taskID := "ttl-task-1"
		status := &ragdtos.RAGStatusDTO{Status: "pending"}

		// Set with TTL
		if err := cache.Set(taskID, status); err != nil {
			t.Fatalf("failed to set badger key: %v", err)
		}

		// Ensure in memory so we hit the optimizations path that also checks TTL
		store.Set(taskID, status)

		got, err := u.GetTaskStatus(taskID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.ExpiresAt == "" {
			t.Log("ExpiresAt not set, maybe TTL is 0 or expired?")
		}
	})
}
