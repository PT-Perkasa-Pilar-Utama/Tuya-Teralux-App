package orchestrator

import (
	"context"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	tuyaDtos "sensio/domain/tuya/dtos"
	"strconv"
	"testing"
)

// MockTuyaAuthUseCase is a mock implementation for testing
type MockTuyaAuthUseCase struct{}

func (m *MockTuyaAuthUseCase) Authenticate() (*tuyaDtos.TuyaAuthResponseDTO, error) {
	return &tuyaDtos.TuyaAuthResponseDTO{AccessToken: "test-token"}, nil
}

func (m *MockTuyaAuthUseCase) GetTuyaAccessToken() (string, error) {
	return "test-token", nil
}

// MockTuyaDeviceControlExecutor is a mock implementation for testing
type MockTuyaDeviceControlExecutor struct{}

func (m *MockTuyaDeviceControlExecutor) SendSwitchCommand(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
	return true, nil
}

func (m *MockTuyaDeviceControlExecutor) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	return true, nil
}

func TestGetDevices_AllLightsIntent_FiltersPanelDevices(t *testing.T) {
	// Create orchestrator with mock dependencies
	orchestrator := NewControlOrchestrator(
		&MockTuyaDeviceControlExecutor{},
		&MockTuyaAuthUseCase{},
	)

	// Create mock vector store response with mixed device types
	// Including: lamp (dj), Smart Switch Lamp (kg), and Teralux panel (dgnzk)
	devices := []tuyaDtos.TuyaDeviceDTO{
		{
			ID:       "lamp-1",
			Name:     "Living Room Lamp",
			Category: "dj",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
		{
			ID:       "switch-lamp-1",
			Name:     "Smart Switch Lamp",
			Category: "kg",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch_1", Value: false}},
		},
		{
			ID:       "teralux-panel-1",
			Name:     "Teralux Receptionist Panel",
			Category: "dgnzk",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
		{
			ID:       "light-2",
			Name:     "Bedroom Light",
			Category: "xdd",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "bright_value", Value: 100}},
		},
		{
			ID:       "outlet-1",
			Name:     "Smart Outlet",
			Category: "cz",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
	}

	// Marshal devices to JSON
	devicesJSON := `{"devices": [`
	for i, d := range devices {
		if i > 0 {
			devicesJSON += ","
		}
		devicesJSON += `{
			"id": "` + d.ID + `",
			"name": "` + d.Name + `",
			"category": "` + d.Category + `",
			"status": [`
		for j, s := range d.Status {
			if j > 0 {
				devicesJSON += ","
			}
			devicesJSON += `{"code": "` + s.Code + `", "value": `
			switch v := s.Value.(type) {
			case bool:
				if v {
					devicesJSON += "true"
				} else {
					devicesJSON += "false"
				}
			case int:
				devicesJSON += strconv.Itoa(v)
			default:
				devicesJSON += `"` + s.Code + `"`
			}
			devicesJSON += `}`
		}
		devicesJSON += `]}`
	}
	devicesJSON += `], "total_devices": 5, "current_page_count": 5, "page": 1, "per_page": 50, "total": 5}`

	// Create a real VectorService and populate it with test data
	vectorService := infrastructure.NewVectorService("")
	_ = vectorService.Upsert("tuya:devices:uid:test-user", devicesJSON, nil)

	// Test with "all lights" prompt in Indonesian
	ctx := &skills.SkillContext{
		Ctx:        context.Background(),
		UID:        "test-user",
		TerminalID: "test-terminal",
		Prompt:     "nyalakan semua lampu",
		Language:   "id",
		Vector:     vectorService,
		Config:     &utils.Config{},
	}

	returnedDevices, _, err := orchestrator.getDevices(ctx)
	if err != nil {
		t.Fatalf("getDevices returned error: %v", err)
	}

	// Verify panel device (dgnzk) was filtered out
	for _, d := range returnedDevices {
		if d.Category == "dgnzk" {
			t.Errorf("Panel device (dgnzk) should have been filtered out, but was included: %s", d.Name)
		}
	}

	// Verify lamp devices were kept
	expectedCategories := map[string]bool{
		"dj":  true,
		"kg":  true,
		"xdd": true,
		"cz":  true,
	}

	returnedCategories := make(map[string]bool)
	for _, d := range returnedDevices {
		returnedCategories[d.Category] = true
	}

	for expectedCat := range expectedCategories {
		if !returnedCategories[expectedCat] {
			t.Errorf("Expected category %s to be included, but it was filtered out", expectedCat)
		}
	}

	// Verify we got 4 devices (all except the panel)
	if len(returnedDevices) != 4 {
		t.Errorf("Expected 4 devices (excluding panel), got %d", len(returnedDevices))
	}
}

func TestGetDevices_NormalPrompt_IncludesAllDevices(t *testing.T) {
	// Test that normal prompts (without "all lights" intent) include all devices
	orchestrator := NewControlOrchestrator(
		&MockTuyaDeviceControlExecutor{},
		&MockTuyaAuthUseCase{},
	)

	devices := []tuyaDtos.TuyaDeviceDTO{
		{
			ID:       "lamp-1",
			Name:     "Living Room Lamp",
			Category: "dj",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
		{
			ID:       "teralux-panel-1",
			Name:     "Teralux Receptionist Panel",
			Category: "dgnzk",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
		{
			ID:       "ac-1",
			Name:     "Living Room AC",
			Category: "cl",
			Status:   []tuyaDtos.TuyaDeviceStatusDTO{{Code: "switch", Value: false}},
		},
	}

	devicesJSON := `{"devices": [`
	for i, d := range devices {
		if i > 0 {
			devicesJSON += ","
		}
		devicesJSON += `{
			"id": "` + d.ID + `",
			"name": "` + d.Name + `",
			"category": "` + d.Category + `",
			"status": [{"code": "` + d.Status[0].Code + `", "value": false}]
		}`
	}
	devicesJSON += `], "total_devices": 3, "current_page_count": 3, "page": 1, "per_page": 50, "total": 3}`

	// Create a real VectorService and populate it with test data
	vectorService := infrastructure.NewVectorService("")
	_ = vectorService.Upsert("tuya:devices:uid:test-user", devicesJSON, nil)

	// Test with normal prompt (no "all lights" intent)
	ctx := &skills.SkillContext{
		Ctx:        context.Background(),
		UID:        "test-user",
		TerminalID: "test-terminal",
		Prompt:     "nyalakan lampu ruang tamu",
		Language:   "id",
		Vector:     vectorService,
		Config:     &utils.Config{},
	}

	returnedDevices, _, err := orchestrator.getDevices(ctx)
	if err != nil {
		t.Fatalf("getDevices returned error: %v", err)
	}

	// Verify all devices were returned (no filtering)
	if len(returnedDevices) != 3 {
		t.Errorf("Expected 3 devices (no filtering for normal prompt), got %d", len(returnedDevices))
	}

	// Verify panel device was included
	panelFound := false
	for _, d := range returnedDevices {
		if d.Category == "dgnzk" {
			panelFound = true
			break
		}
	}
	if !panelFound {
		t.Error("Panel device (dgnzk) should have been included for normal prompt")
	}
}

func TestIsAllLightsIntent(t *testing.T) {
	orchestrator := NewControlOrchestrator(
		&MockTuyaDeviceControlExecutor{},
		&MockTuyaAuthUseCase{},
	)

	testCases := []struct {
		name     string
		prompt   string
		expected bool
	}{
		// Indonesian "all lights" patterns - MUST contain both quantifier AND light word
		{"Indonesian semua lampu", "nyalakan semua lampu", true},
		{"Indonesian lampu semua", "matikan lampu semua", true},
		{"Indonesian semua switch", "hidupkan semua switch", true},
		// Generic "all" without light word should be FALSE (tightened behavior)
		{"Indonesian nyalakan semua", "nyalakan semua", false},
		{"Indonesian matikan semua", "matikan semua", false},

		// English "all lights" patterns - MUST contain both quantifier AND light word
		{"English all lights", "turn on all lights", true},
		{"English all light", "turn off all light", true},
		{"English every light", "turn on every light", true},
		// Generic "all" without light word should be FALSE (tightened behavior)
		{"English turn on all", "turn on all", false},
		{"English turn off all", "turn off all", false},

		// Mixed patterns
		{"Mixed semua light", "nyalakan semua light", true},

		// Non-"all lights" patterns (should be false)
		{"Specific lamp", "nyalakan lampu ruang tamu", false},
		{"Specific switch", "matikan switch 1", false},
		{"AC command", "set ac to 24", false},
		{"Discovery", "what devices do i have", false},
		{"Empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := orchestrator.isAllLightsIntent(tc.prompt)
			if result != tc.expected {
				t.Errorf("isAllLightsIntent(%q) = %v, want %v", tc.prompt, result, tc.expected)
			}
		})
	}
}

func TestFilterLampDevices(t *testing.T) {
	orchestrator := NewControlOrchestrator(
		&MockTuyaDeviceControlExecutor{},
		&MockTuyaAuthUseCase{},
	)

	// Create test devices with various categories
	devices := []tuyaDtos.TuyaDeviceDTO{
		{ID: "dj-1", Name: "Light 1", Category: "dj"},
		{ID: "xdd-1", Name: "Light 2", Category: "xdd"},
		{ID: "fwd-1", Name: "Light 3", Category: "fwd"},
		{ID: "ty-1", Name: "Light 4", Category: "ty"},
		{ID: "kg-1", Name: "Switch Lamp 1", Category: "kg"},
		{ID: "cz-1", Name: "Outlet 1", Category: "cz"},
		{ID: "pc-1", Name: "PC 1", Category: "pc"},
		{ID: "dlq-1", Name: "Breaker 1", Category: "dlq"},
		{ID: "dgnzk-1", Name: "Panel 1", Category: "dgnzk"},
		{ID: "cl-1", Name: "AC 1", Category: "cl"},
		{ID: "ws-1", Name: "Sensor 1", Category: "ws"},
	}

	filtered := orchestrator.filterLampDevices(devices)

	// Expected lamp-relevant categories
	expectedIDs := map[string]bool{
		"dj-1":  true,
		"xdd-1": true,
		"fwd-1": true,
		"ty-1":  true,
		"kg-1":  true,
		"cz-1":  true,
		"pc-1":  true,
		"dlq-1": true,
	}

	// Verify filtered devices
	filteredIDs := make(map[string]bool)
	for _, d := range filtered {
		filteredIDs[d.ID] = true
	}

	// Check all expected devices are included
	for expectedID := range expectedIDs {
		if !filteredIDs[expectedID] {
			t.Errorf("Expected device %s to be included in filtered results", expectedID)
		}
	}

	// Check excluded devices are not included
	excludedIDs := []string{"dgnzk-1", "cl-1", "ws-1"}
	for _, excludedID := range excludedIDs {
		if filteredIDs[excludedID] {
			t.Errorf("Device %s should have been excluded from filtered results", excludedID)
		}
	}

	// Verify count
	if len(filtered) != 8 {
		t.Errorf("Expected 8 filtered devices, got %d", len(filtered))
	}
}

func TestGetDevices_EnglishAllLights(t *testing.T) {
	// Test with English "all lights" prompt
	orchestrator := NewControlOrchestrator(
		&MockTuyaDeviceControlExecutor{},
		&MockTuyaAuthUseCase{},
	)

	devicesJSON := `{"devices": [
		{"id": "lamp-1", "name": "Living Room Lamp", "category": "dj", "status": [{"code": "switch", "value": false}]},
		{"id": "teralux-panel-1", "name": "Teralux Receptionist Panel", "category": "dgnzk", "status": [{"code": "switch", "value": false}]}
	], "total_devices": 2, "current_page_count": 2, "page": 1, "per_page": 50, "total": 2}`

	// Create a real VectorService and populate it with test data
	vectorService := infrastructure.NewVectorService("")
	_ = vectorService.Upsert("tuya:devices:uid:test-user", devicesJSON, nil)

	ctx := &skills.SkillContext{
		Ctx:        context.Background(),
		UID:        "test-user",
		TerminalID: "test-terminal",
		Prompt:     "turn on all lights",
		Language:   "en",
		Vector:     vectorService,
		Config:     &utils.Config{},
	}

	returnedDevices, _, err := orchestrator.getDevices(ctx)
	if err != nil {
		t.Fatalf("getDevices returned error: %v", err)
	}

	// Verify panel device was filtered out
	for _, d := range returnedDevices {
		if d.Category == "dgnzk" {
			t.Errorf("Panel device (dgnzk) should have been filtered out, but was included: %s", d.Name)
		}
	}

	// Verify only lamp device remains
	if len(returnedDevices) != 1 {
		t.Errorf("Expected 1 device (lamp only), got %d", len(returnedDevices))
	}
}
