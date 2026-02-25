package usecases

import (
	"errors"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	tuya_dtos "sensio/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTuyaDeviceControlExecutor mocks the TuyaDeviceControlExecutor interface
type MockTuyaDeviceControlExecutor struct {
	mock.Mock
}

func (m *MockTuyaDeviceControlExecutor) SendSwitchCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error) {
	args := m.Called(accessToken, deviceID, commands)
	return args.Bool(0), args.Error(1)
}

func (m *MockTuyaDeviceControlExecutor) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	args := m.Called(accessToken, infraredID, remoteID, params)
	return args.Bool(0), args.Error(1)
}

func TestUpdateDeviceStatus_UserBehavior(t *testing.T) {
	repo := new(MockDeviceStatusRepository)
	devRepo := new(MockDeviceRepository)
	mockTuya := new(MockTuyaDeviceControlExecutor)

	useCase := NewUpdateDeviceStatusUseCase(repo, devRepo, mockTuya)

	// 1. Update Status (Success)
	t.Run("Update Status (Success - Command)", func(t *testing.T) {
		deviceID := "d1"
		token := "valid_token"
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}

		devRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID, RemoteID: "r1"}, nil).Once()

		mockTuya.On("SendSwitchCommand", token, deviceID, mock.MatchedBy(func(cmds []tuya_dtos.TuyaCommandDTO) bool {
			return len(cmds) == 1 && cmds[0].Code == "switch_1" && cmds[0].Value == true
		})).Return(true, nil).Once()

		repo.On("Upsert", mock.MatchedBy(func(s *entities.DeviceStatus) bool {
			return s.DeviceID == deviceID && s.Code == "switch_1" && s.Value == "true"
		})).Return(nil).Once()

		err := useCase.UpdateDeviceStatus(deviceID, req, token)
		assert.NoError(t, err)

		devRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
		mockTuya.AssertExpectations(t)
	})

	// 2. Update Status (Not Found - Device)
	t.Run("Update Status (Not Found - Device)", func(t *testing.T) {
		devRepo.On("GetByID", "unknown").Return(nil, errors.New("record not found")).Once()

		err := useCase.UpdateDeviceStatus("unknown", &dtos.UpdateDeviceStatusRequestDTO{Code: "c"}, "t")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Device not found")
	})

	// 3. Update Status (Not Found - Invalid Code)
	t.Run("Update Status (Not Found - Invalid Code)", func(t *testing.T) {
		deviceID := "d1"
		token := "valid_token"
		devRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID}, nil).Once()

		err := useCase.UpdateDeviceStatus(deviceID, &dtos.UpdateDeviceStatusRequestDTO{Code: "nuclear_launch"}, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid status code")
	})

	// 5. Command Failure
	t.Run("Command Failure", func(t *testing.T) {
		deviceID := "d1"
		token := "valid_token"
		req := &dtos.UpdateDeviceStatusRequestDTO{Code: "switch_1", Value: true}

		devRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID, RemoteID: "r1"}, nil).Once()

		mockTuya.On("SendSwitchCommand", token, deviceID, mock.Anything).Return(false, errors.New("tuya error")).Once()

		err := useCase.UpdateDeviceStatus(deviceID, req, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tuya error")
	})

	// 6. IR Command Success
	t.Run("IR Command Success", func(t *testing.T) {
		deviceID := "d1"
		token := "valid_token"
		remoteID := "ir_remote_1"
		req := &dtos.UpdateDeviceStatusRequestDTO{
			Code:     "temp",
			Value:    24,
			RemoteID: remoteID,
		}

		devRepo.On("GetByID", deviceID).Return(&entities.Device{ID: deviceID}, nil).Once()

		mockTuya.On("SendIRACCommand", token, deviceID, remoteID, mock.MatchedBy(func(p map[string]int) bool {
			return p["temp"] == 24
		})).Return(true, nil).Once()

		repo.On("Upsert", mock.MatchedBy(func(s *entities.DeviceStatus) bool {
			return s.DeviceID == deviceID && s.Code == "temp" && s.Value == "24"
		})).Return(nil).Once()

		err := useCase.UpdateDeviceStatus(deviceID, req, token)
		assert.NoError(t, err)

		devRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
		mockTuya.AssertExpectations(t)
	})
}
