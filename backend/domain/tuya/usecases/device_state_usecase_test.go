package usecases

import (
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
	"testing"
)

func TestDeviceStateUseCase_SaveAndGet(t *testing.T) {
	// Setup Config
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}

	// Setup Badger
	tmpDir := t.TempDir()
	cache, err := persistence.NewBadgerService(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create Badger: %v", err)
	}
	defer cache.Close()

	useCase := NewDeviceStateUseCase(cache)

	// 1. Test Save
	commands := []dtos.DeviceStateCommandDTO{
		{Code: "power", Value: true},
		{Code: "temp", Value: 24},
	}

	err = useCase.SaveDeviceState("dev-1", commands)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// 2. Test Get
	state, err := useCase.GetDeviceState("dev-1")
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if state == nil {
		t.Fatal("State should not be nil")
	}
	if state.DeviceID != "dev-1" {
		t.Errorf("Expected dev-1, got %s", state.DeviceID)
	}
	if len(state.LastCommands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(state.LastCommands))
	}

	// Verify persistence key directly (optional, but ensures "as is" logic)
	// Key format: device_state:{id}
	val, _ := cache.Get("device_state:dev-1")
	if val == nil {
		t.Error("Direct key lookup failed")
	}
}

func TestDeviceStateUseCase_Merge(t *testing.T) {
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}
	// Setup Badger
	tmpDir := t.TempDir()
	cache, _ := persistence.NewBadgerService(tmpDir)
	defer cache.Close()

	useCase := NewDeviceStateUseCase(cache)

	// Initial State
	initial := []dtos.DeviceStateCommandDTO{
		{Code: "power", Value: true},
		{Code: "mode", Value: "cool"},
	}
	useCase.SaveDeviceState("dev-merge", initial)

	// Update (Merge)
	update := []dtos.DeviceStateCommandDTO{
		{Code: "power", Value: false}, // Changed
		{Code: "temp", Value: 20},     // New
	}
	useCase.SaveDeviceState("dev-merge", update)

	// Verify
	state, _ := useCase.GetDeviceState("dev-merge")
	if len(state.LastCommands) != 3 {
		t.Errorf("Expected 3 commands after merge, got %d", len(state.LastCommands))
	}

	// check values
	cmdMap := make(map[string]interface{})
	for _, c := range state.LastCommands {
		cmdMap[c.Code] = c.Value
	}

	if cmdMap["power"] != false {
		t.Error("Expected power to be updated to false")
	}
	if cmdMap["mode"] != "cool" {
		t.Error("Expected mode to persist as cool")
	}
	if cmdMap["temp"] != float64(20) && cmdMap["temp"] != 20 { // json unmarshal might be float
		t.Error("Expected temp to be added")
	}
}

func TestDeviceStateUseCase_CleanupOrphanedStates(t *testing.T) {
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}
	// Setup Badger
	tmpDir := t.TempDir()
	cache, _ := persistence.NewBadgerService(tmpDir)
	defer cache.Close()

	useCase := NewDeviceStateUseCase(cache)

	// Seed states
	useCase.SaveDeviceState("valid-1", []dtos.DeviceStateCommandDTO{{Code: "a", Value: 1}})
	useCase.SaveDeviceState("orphan-1", []dtos.DeviceStateCommandDTO{{Code: "b", Value: 1}})

	// valid list
	validIDs := []string{"valid-1"}

	err := useCase.CleanupOrphanedStates(validIDs)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify valid exists
	s1, _ := useCase.GetDeviceState("valid-1")
	if s1 == nil {
		t.Error("Valid state should remain")
	}

	// Verify orphan removed
	s2, _ := useCase.GetDeviceState("orphan-1")
	if s2 != nil {
		t.Error("Orphan state should be removed")
	}
}
