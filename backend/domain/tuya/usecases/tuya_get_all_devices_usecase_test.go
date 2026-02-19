package usecases

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGetAllDevices_CachesAndUpsertsVector(t *testing.T) {
	t.Skip("Skipping hanging test environment issue with BadgerDB")
	// Arrange
	gin.SetMode(gin.TestMode)
	// Ensure config is loaded (used by BadgerService initialization)
	utils.LoadConfig()
	tmpDir := t.TempDir()
	cache, err := infrastructure.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("failed to init badger: %v", err)
	}
	defer func() { _ = cache.Close() }()

	vector := infrastructure.NewVectorService("")

	// use the real service in test mode (returns empty list but ok for cache/vector test)
	svc := services.NewTuyaDeviceService()
	deviceState := NewDeviceStateUseCase(cache)
	uc := NewTuyaGetAllDevicesUseCase(svc, deviceState, cache, vector, nil, nil)

	// Act
	resp, err := uc.GetAllDevices("valid_token", "user123", 1, 10, "")
	if err != nil {
		t.Fatalf("GetAllDevices returned error: %v", err)
	}

	// Wait for background vector DB population
	time.Sleep(100 * time.Millisecond)

	// Build expected keys
	cacheKey := "cache:tuya:devices:uid:user123:cat::page:1:limit:10"
	aggID := "tuya:devices:uid:user123"

	// Assert cache exists
	cached, err := cache.Get(cacheKey)
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}
	if cached == nil {
		t.Fatalf("expected cached response under key %s, got nil", cacheKey)
	}

	// Assert vector aggregate exists
	if val, ok := vector.Get(aggID); !ok || val == "" {
		t.Fatalf("expected vector aggregate %s to exist, got ok=%v val='%s'", aggID, ok, val)
	}

	// Also ensure resp is non-nil
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
}
