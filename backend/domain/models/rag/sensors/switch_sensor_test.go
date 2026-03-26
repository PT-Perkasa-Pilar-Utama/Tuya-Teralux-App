package sensors

import (
	tuyaDtos "sensio/domain/tuya/dtos"
	"testing"
)

// MockTuyaDeviceControlExecutor is a mock implementation for testing
type MockTuyaDeviceControlExecutor struct {
	SendSwitchCommandFunc func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error)
}

func (m *MockTuyaDeviceControlExecutor) SendSwitchCommand(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
	if m.SendSwitchCommandFunc != nil {
		return m.SendSwitchCommandFunc(accessToken, deviceID, commands)
	}
	return true, nil
}

func (m *MockTuyaDeviceControlExecutor) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	return true, nil
}

func TestSwitchSensor_ExecuteControl_AllSwitches(t *testing.T) {
	// Test that when prompt contains "semua" (all), all switches are controlled
	switchSensor := NewSwitchSensor()

	// Create a multi-switch device (kg with switch_1, switch_2, switch_3)
	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "test-switch-device",
		Name:     "Smart Switch Living Room",
		Category: "kg",
		Status: []tuyaDtos.TuyaDeviceStatusDTO{
			{Code: "switch_1", Value: false},
			{Code: "switch_2", Value: false},
			{Code: "switch_3", Value: false},
		},
	}

	var capturedCommands []tuyaDtos.TuyaCommandDTO
	mockExecutor := &MockTuyaDeviceControlExecutor{
		SendSwitchCommandFunc: func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
			capturedCommands = commands
			return true, nil
		},
	}

	// Test with "semua" (all) in Indonesian
	prompt := "nyalakan semua switch"
	result, err := switchSensor.ExecuteControl("test-token", device, prompt, nil, mockExecutor)

	if err != nil {
		t.Fatalf("ExecuteControl returned error: %v", err)
	}

	// Verify all 3 switches were commanded
	if len(capturedCommands) != 3 {
		t.Errorf("Expected 3 commands for all switches, got %d", len(capturedCommands))
	}

	// Verify each switch is commanded to ON (true)
	expectedSwitches := map[string]bool{
		"switch_1": true,
		"switch_2": true,
		"switch_3": true,
	}

	commandedSwitches := make(map[string]bool)
	for _, cmd := range capturedCommands {
		commandedSwitches[cmd.Code] = true
		if cmd.Value != true {
			t.Errorf("Expected switch %s to be ON (true), got %v", cmd.Code, cmd.Value)
		}
	}

	// Verify all expected switches were commanded
	for expectedSwitch := range expectedSwitches {
		if !commandedSwitches[expectedSwitch] {
			t.Errorf("Expected switch %s to be commanded, but it was not", expectedSwitch)
		}
	}

	// Verify response message indicates "semua" was handled
	if result.Message != "Berhasil menyalakan semua switch di Smart Switch Living Room." {
		t.Errorf("Unexpected message: got %q", result.Message)
	}
}

func TestSwitchSensor_ExecuteControl_AllSwitchesOff(t *testing.T) {
	// Test turning off all switches
	switchSensor := NewSwitchSensor()

	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "test-switch-device",
		Name:     "Smart Switch Bedroom",
		Category: "kg",
		Status: []tuyaDtos.TuyaDeviceStatusDTO{
			{Code: "switch_1", Value: true},
			{Code: "switch_2", Value: true},
		},
	}

	var capturedCommands []tuyaDtos.TuyaCommandDTO
	mockExecutor := &MockTuyaDeviceControlExecutor{
		SendSwitchCommandFunc: func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
			capturedCommands = commands
			return true, nil
		},
	}

	prompt := "matikan semuanya"
	result, err := switchSensor.ExecuteControl("test-token", device, prompt, nil, mockExecutor)

	if err != nil {
		t.Fatalf("ExecuteControl returned error: %v", err)
	}

	// Verify both switches were commanded
	if len(capturedCommands) != 2 {
		t.Errorf("Expected 2 commands for all switches, got %d", len(capturedCommands))
	}

	// Verify each switch is commanded to OFF (false)
	for _, cmd := range capturedCommands {
		if cmd.Value != false {
			t.Errorf("Expected switch %s to be OFF (false), got %v", cmd.Code, cmd.Value)
		}
	}

	// Verify response message
	if result.Message != "Berhasil mematikan semua switch di Smart Switch Bedroom." {
		t.Errorf("Unexpected message: got %q", result.Message)
	}
}

func TestSwitchSensor_ExecuteControl_SpecificSwitch(t *testing.T) {
	// Test controlling a specific switch
	switchSensor := NewSwitchSensor()

	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "test-switch-device",
		Name:     "Smart Switch Kitchen",
		Category: "kg",
		Status: []tuyaDtos.TuyaDeviceStatusDTO{
			{Code: "switch_1", Value: false},
			{Code: "switch_2", Value: false},
			{Code: "switch_3", Value: false},
		},
	}

	var capturedCommands []tuyaDtos.TuyaCommandDTO
	mockExecutor := &MockTuyaDeviceControlExecutor{
		SendSwitchCommandFunc: func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
			capturedCommands = commands
			return true, nil
		},
	}

	// Test with specific switch number
	prompt := "nyalakan switch 2"
	result, err := switchSensor.ExecuteControl("test-token", device, prompt, nil, mockExecutor)

	if err != nil {
		t.Fatalf("ExecuteControl returned error: %v", err)
	}

	// Verify only 1 switch was commanded
	if len(capturedCommands) != 1 {
		t.Errorf("Expected 1 command for specific switch, got %d", len(capturedCommands))
	}

	// Verify the correct switch was commanded
	if capturedCommands[0].Code != "switch_2" {
		t.Errorf("Expected switch_2 to be commanded, got %s", capturedCommands[0].Code)
	}

	// Verify switch is commanded to ON (true)
	if capturedCommands[0].Value != true {
		t.Errorf("Expected switch to be ON (true), got %v", capturedCommands[0].Value)
	}

	// Verify response message indicates specific switch
	if result.Message != "Berhasil menyalakan switch 2 di Smart Switch Kitchen." {
		t.Errorf("Unexpected message: got %q", result.Message)
	}
}

func TestSwitchSensor_ExecuteControl_EnglishAll(t *testing.T) {
	// Test with English "all" quantifier
	switchSensor := NewSwitchSensor()

	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "test-switch-device",
		Name:     "Smart Switch",
		Category: "kg",
		Status: []tuyaDtos.TuyaDeviceStatusDTO{
			{Code: "switch_1", Value: false},
			{Code: "switch_2", Value: false},
		},
	}

	var capturedCommands []tuyaDtos.TuyaCommandDTO
	mockExecutor := &MockTuyaDeviceControlExecutor{
		SendSwitchCommandFunc: func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
			capturedCommands = commands
			return true, nil
		},
	}

	prompt := "turn on all switches"
	result, err := switchSensor.ExecuteControl("test-token", device, prompt, nil, mockExecutor)

	if err != nil {
		t.Fatalf("ExecuteControl returned error: %v", err)
	}

	// Verify both switches were commanded
	if len(capturedCommands) != 2 {
		t.Errorf("Expected 2 commands for all switches, got %d", len(capturedCommands))
	}

	// Verify response message
	if result.Message != "Berhasil menyalakan semua switch di Smart Switch." {
		t.Errorf("Unexpected message: got %q", result.Message)
	}
}

func TestSwitchSensor_ExecuteControl_NoSwitches(t *testing.T) {
	// Test device with no switch codes - when "semua" is used, it will execute with empty commands
	// This is existing behavior - the "all" path doesn't validate switch presence
	switchSensor := NewSwitchSensor()

	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "test-device",
		Name:     "Non-Switch Device",
		Category: "dj",
		Status: []tuyaDtos.TuyaDeviceStatusDTO{
			{Code: "bright_value", Value: 100},
		},
	}

	var capturedCommands []tuyaDtos.TuyaCommandDTO
	mockExecutor := &MockTuyaDeviceControlExecutor{
		SendSwitchCommandFunc: func(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
			capturedCommands = commands
			// Return success even with empty commands (existing behavior)
			return true, nil
		},
	}

	prompt := "nyalakan semua"
	result, err := switchSensor.ExecuteControl("test-token", device, prompt, nil, mockExecutor)

	if err != nil {
		t.Fatalf("ExecuteControl returned error: %v", err)
	}

	// With "semua" path, empty commands are sent (existing behavior)
	if len(capturedCommands) != 0 {
		t.Errorf("Expected 0 commands for device with no switches, got %d", len(capturedCommands))
	}

	// The result will still be success (existing behavior)
	if result.HTTPStatusCode != 0 {
		t.Errorf("Expected HTTP status 0 (success), got %d", result.HTTPStatusCode)
	}
}
