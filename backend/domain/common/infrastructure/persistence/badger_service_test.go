package persistence

import (
	"os"
	"teralux_app/domain/common/utils"
	"testing"
)

func TestNewBadgerService(t *testing.T) {
	tmpDir := t.TempDir()

	// Set cache TTL in environment
	os.Setenv("CACHE_TTL", "30m")
	defer os.Unsetenv("CACHE_TTL")

	// Initialize config to avoid nil pointer
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	if service == nil {
		t.Fatal("Expected service instance, got nil")
	}
	if service.db == nil {
		t.Fatal("Expected database instance, got nil")
	}
}

func TestBadgerService_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	key := "test_key"
	value := []byte("test_value")

	// Set value
	err = service.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value
	retrieved, err := service.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestBadgerService_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	retrieved, err := service.Get("non_existent_key")
	if err != nil {
		t.Fatalf("Get should not error on non-existent key, got: %v", err)
	}
	if retrieved != nil {
		t.Errorf("Expected nil for non-existent key, got %v", retrieved)
	}
}

func TestBadgerService_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	key := "delete_test"
	value := []byte("value")

	// Set then delete
	service.Set(key, value)
	err = service.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	retrieved, _ := service.Get(key)
	if retrieved != nil {
		t.Error("Expected key to be deleted")
	}
}

func TestBadgerService_SetPersistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	key := "persistent_key"
	value := []byte("persistent_value")

	err = service.SetPersistent(key, value)
	if err != nil {
		t.Fatalf("SetPersistent failed: %v", err)
	}

	retrieved, err := service.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestBadgerService_ClearWithPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	// Set multiple keys with same prefix
	prefix := "test_prefix:"
	service.Set(prefix+"key1", []byte("value1"))
	service.Set(prefix+"key2", []byte("value2"))
	service.Set("other_key", []byte("other_value"))

	// Clear with prefix
	err = service.ClearWithPrefix(prefix)
	if err != nil {
		t.Fatalf("ClearWithPrefix failed: %v", err)
	}

	// Verify prefix keys are deleted
	val1, _ := service.Get(prefix + "key1")
	val2, _ := service.Get(prefix + "key2")
	if val1 != nil || val2 != nil {
		t.Error("Expected prefix keys to be deleted")
	}

	// Verify other key still exists
	otherVal, _ := service.Get("other_key")
	if otherVal == nil {
		t.Error("Expected other_key to still exist")
	}
}

func TestBadgerService_GetAllKeysWithPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	prefix := "list_test:"
	service.Set(prefix+"a", []byte("1"))
	service.Set(prefix+"b", []byte("2"))
	service.Set("other", []byte("3"))

	keys, err := service.GetAllKeysWithPrefix(prefix)
	if err != nil {
		t.Fatalf("GetAllKeysWithPrefix failed: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestBadgerService_FlushAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}
	defer service.Close()

	// Set cache and persistent data
	service.Set("cache:key1", []byte("cache_value"))
	service.SetPersistent("persistent:key1", []byte("persistent_value"))

	// Flush cache
	err = service.FlushAll()
	if err != nil {
		t.Fatalf("FlushAll failed: %v", err)
	}

	// Verify cache is cleared
	cacheVal, _ := service.Get("cache:key1")
	if cacheVal != nil {
		t.Error("Expected cache key to be flushed")
	}

	// Verify persistent data still exists
	persistentVal, _ := service.Get("persistent:key1")
	if persistentVal == nil {
		t.Error("Expected persistent key to remain after flush")
	}
}

func TestBadgerService_Close(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	utils.AppConfig = nil
	_ = utils.GetConfig()

	service, err := NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("NewBadgerService failed: %v", err)
	}

	err = service.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Closing again should not error
	err = service.Close()
	if err != nil {
		t.Errorf("Second Close should not error, got: %v", err)
	}
}
