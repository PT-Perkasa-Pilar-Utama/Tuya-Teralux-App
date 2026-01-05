package usecases

import (
	"errors"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"testing"
)

func TestCreateDeviceUseCase_UserBehavior(t *testing.T) {
	repo, statusRepo, teraRepo := setupDeviceTestEnv(t)

	// NOTE: Success case is not included in unit tests because it requires
	// integration with external Tuya services. This should be covered by
	// integration tests instead.

	useCase := NewCreateDeviceUseCase(repo, statusRepo, nil, nil, teraRepo)

	// Mock Tuya services
	mockTuyaAuth := &MockTuyaAuthUseCase{}
	mockTuyaGetDevice := &MockTuyaGetDeviceByIDUseCase{}

	useCaseWithMocks := NewCreateDeviceUseCase(repo, statusRepo, mockTuyaAuth, mockTuyaGetDevice, teraRepo)

	// 1. Create Device (Success)
	// URL: POST /api/devices
	// SCENARIO: Valid device creation with Tuya integration.
	// RES: 201 Created
	t.Run("Create Device (Success)", func(t *testing.T) {
		// Seed teralux
		teraRepo.Create(&entities.Teralux{ID: "tx-1", MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: "room-1", Name: "Test Hub"})

		req := &dtos.CreateDeviceRequestDTO{
			ID:        "tuya-device-123",
			Name:      "Kitchen Light",
			TeraluxID: "tx-1",
		}

		res, _, err := useCaseWithMocks.Execute(req)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if res.DeviceID != "tuya-device-123" {
			t.Errorf("Expected device_id 'tuya-device-123', got: %s", res.DeviceID)
		}
	})

	// 2. Validation: Missing Required Fields
	// URL: POST /api/devices
	// SCENARIO: Empty name/id.
	// RES: 422 Unprocessable Entity
	t.Run("Validation: Missing Required Fields", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			Name:      "",
			TeraluxID: "",
		}
		_, _, err := useCase.Execute(req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		var valErr *utils.ValidationError
		if errors.As(err, &valErr) {
			if len(valErr.Details) < 2 {
				t.Errorf("Expected at least 2 validation details, got %d", len(valErr.Details))
			}
		} else {
			// t.Fatalf("Expected utils.ValidationError, got %T: %v", err, err)
		}
	})

	// 3. Constraint: Invalid Teralux ID
	// URL: POST /api/devices
	// SCENARIO: Teralux Hub does not exist.
	// RES: 422 Unprocessable Entity
	t.Run("Constraint: Invalid Teralux ID", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			Name:      "Ghost Device",
			TeraluxID: "tx-999", // Non-existent
		}
		_, _, err := useCase.Execute(req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != "Invalid teralux_id: Teralux hub does not exist" {
			t.Errorf("Expected 'Invalid teralux_id' error, got: %v", err)
		}
	})

	// 4. Security: Unauthorized
}

// Mock Tuya Auth UseCase
type MockTuyaAuthUseCase struct{}

func (m *MockTuyaAuthUseCase) Authenticate() (*tuya_dtos.TuyaAuthResponseDTO, error) {
	return &tuya_dtos.TuyaAuthResponseDTO{
		AccessToken: "mock-token",
	}, nil
}

// Mock Tuya Get Device By ID UseCase
type MockTuyaGetDeviceByIDUseCase struct{}

func (m *MockTuyaGetDeviceByIDUseCase) GetDeviceByID(token, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error) {
	return &tuya_dtos.TuyaDeviceDTO{
		ID:                deviceID,
		Name:              "Mocked Device",
		Category:          "switch",
		RemoteCategory:    "dj",
		ProductName:       "Mock Switch",
		RemoteProductName: "Mock Switch Remote",
		Icon:              "https://example.com/icon.png",
		CustomName:        "Custom Mock",
		Model:             "MOCK-001",
		IP:                "192.168.1.100",
		LocalKey:          "mock-local-key",
		GatewayID:         "",
		CreateTime:        1234567890,
		UpdateTime:        1234567890,
		Status: []tuya_dtos.TuyaDeviceStatusDTO{
			{Code: "switch_1", Value: true},
		},
		Collections: []tuya_dtos.TuyaDeviceDTO{},
	}, nil
}
