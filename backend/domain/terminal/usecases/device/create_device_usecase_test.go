package usecases

import (
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	tuya_dtos "sensio/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDeviceUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	statusRepo := new(MockDeviceStatusRepository)
	teraRepo := new(MockTerminalRepository)
	tuyaAuth := new(MockTuyaAuthUseCase)
	tuyaGetDevice := new(MockTuyaGetDeviceByIDUseCase)

	useCase := NewCreateDeviceUseCase(repo, statusRepo, tuyaAuth, tuyaGetDevice, teraRepo)

	// 1. Create Device (Success)
	t.Run("Create Device (Success)", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			ID:         "tuya-device-123",
			Name:       "Kitchen Light",
			TerminalID: "tx-1",
		}

		teraRepo.On("GetByID", req.TerminalID).Return(&entities.Terminal{ID: "tx-1"}, nil).Once()
		tuyaAuth.On("Authenticate").Return(&tuya_dtos.TuyaAuthResponseDTO{AccessToken: "token"}, nil).Once()
		tuyaGetDevice.On("GetDeviceByID", "token", req.ID).Return(&tuya_dtos.TuyaDeviceDTO{
			ID:     req.ID,
			Name:   "Mocked",
			Status: []tuya_dtos.TuyaDeviceStatusDTO{{Code: "s1", Value: "v1"}},
		}, nil).Once()

		repo.On("GetByIDUnscoped", req.ID).Return(nil, assert.AnError).Once()
		repo.On("Create", mock.Anything).Return(nil).Once()
		statusRepo.On("UpsertDeviceStatuses", req.ID, mock.Anything).Return(nil).Once()
		teraRepo.On("InvalidateCache", req.TerminalID).Return(nil).Once()

		res, _, err := useCase.CreateDevice(req)
		assert.NoError(t, err)
		assert.Equal(t, req.ID, res.DeviceID)

		repo.AssertExpectations(t)
		statusRepo.AssertExpectations(t)
		teraRepo.AssertExpectations(t)
		tuyaAuth.AssertExpectations(t)
		tuyaGetDevice.AssertExpectations(t)
	})

	// 2. Validation: Missing Required Fields
	t.Run("Validation: Missing Required Fields", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			Name:       "",
			TerminalID: "",
		}
		_, _, err := useCase.CreateDevice(req)
		assert.Error(t, err)
	})

	// 3. Constraint: Invalid Terminal ID
	t.Run("Constraint: Invalid Terminal ID", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			Name:       "Ghost Device",
			TerminalID: "tx-999",
		}
		teraRepo.On("GetByID", "tx-999").Return(nil, assert.AnError).Once()

		_, _, err := useCase.CreateDevice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid terminal_id")
		teraRepo.AssertExpectations(t)
	})
}

// --- Mocks ---

type MockTerminalRepository struct {
	mock.Mock
}

func (m *MockTerminalRepository) Create(terminal *entities.Terminal) error {
	args := m.Called(terminal)
	return args.Error(0)
}

func (m *MockTerminalRepository) GetAll() ([]entities.Terminal, error) {
	args := m.Called()
	return args.Get(0).([]entities.Terminal), args.Error(1)
}

func (m *MockTerminalRepository) GetAllPaginated(offset, limit int, roomID *string) ([]entities.Terminal, int64, error) {
	args := m.Called(offset, limit, roomID)
	return args.Get(0).([]entities.Terminal), int64(args.Int(1)), args.Error(2)
}

func (m *MockTerminalRepository) GetByID(id string) (*entities.Terminal, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Terminal), args.Error(1)
}

func (m *MockTerminalRepository) GetByMacAddress(macAddress string) (*entities.Terminal, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Terminal), args.Error(1)
}

func (m *MockTerminalRepository) Update(terminal *entities.Terminal) error {
	args := m.Called(terminal)
	return args.Error(0)
}

func (m *MockTerminalRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTerminalRepository) InvalidateCache(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(device *entities.Device) error {
	args := m.Called(device)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetAll() ([]entities.Device, error) {
	args := m.Called()
	return args.Get(0).([]entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetAllPaginated(offset, limit int) ([]entities.Device, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]entities.Device), args.Get(1).(int64), args.Error(2)
}

func (m *MockDeviceRepository) GetByTerminalID(terminalID string) ([]entities.Device, error) {
	args := m.Called(terminalID)
	return args.Get(0).([]entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByTerminalIDPaginated(terminalID string, offset, limit int) ([]entities.Device, int64, error) {
	args := m.Called(terminalID, offset, limit)
	return args.Get(0).([]entities.Device), args.Get(1).(int64), args.Error(2)
}

func (m *MockDeviceRepository) GetByIDUnscoped(id string) (*entities.Device, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByID(id string) (*entities.Device, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) Update(device *entities.Device) error {
	args := m.Called(device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByRemoteID(remoteID string) (*entities.Device, error) {
	args := m.Called(remoteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Device), args.Error(1)
}

type MockTuyaAuthUseCase struct {
	mock.Mock
}

func (m *MockTuyaAuthUseCase) Authenticate() (*tuya_dtos.TuyaAuthResponseDTO, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tuya_dtos.TuyaAuthResponseDTO), args.Error(1)
}

type MockTuyaGetDeviceByIDUseCase struct {
	mock.Mock
}

func (m *MockTuyaGetDeviceByIDUseCase) GetDeviceByID(token, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error) {
	args := m.Called(token, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tuya_dtos.TuyaDeviceDTO), args.Error(1)
}

type MockDeviceStatusRepository struct {
	mock.Mock
}

func (m *MockDeviceStatusRepository) Create(status *entities.DeviceStatus) error {
	args := m.Called(status)
	return args.Error(0)
}

func (m *MockDeviceStatusRepository) GetAll() ([]entities.DeviceStatus, error) {
	args := m.Called()
	return args.Get(0).([]entities.DeviceStatus), args.Error(1)
}

func (m *MockDeviceStatusRepository) GetAllPaginated(offset, limit int) ([]entities.DeviceStatus, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]entities.DeviceStatus), int64(args.Int(1)), args.Error(2)
}

func (m *MockDeviceStatusRepository) GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error) {
	args := m.Called(deviceID)
	return args.Get(0).([]entities.DeviceStatus), args.Error(1)
}

func (m *MockDeviceStatusRepository) GetByDeviceIDPaginated(deviceID string, offset, limit int) ([]entities.DeviceStatus, int64, error) {
	args := m.Called(deviceID, offset, limit)
	return args.Get(0).([]entities.DeviceStatus), int64(args.Int(1)), args.Error(2)
}

func (m *MockDeviceStatusRepository) GetByDeviceIDAndCode(deviceID, code string) (*entities.DeviceStatus, error) {
	args := m.Called(deviceID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DeviceStatus), args.Error(1)
}

func (m *MockDeviceStatusRepository) UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error {
	args := m.Called(deviceID, statuses)
	return args.Error(0)
}

func (m *MockDeviceStatusRepository) Upsert(status *entities.DeviceStatus) error {
	args := m.Called(status)
	return args.Error(0)
}

func (m *MockDeviceStatusRepository) DeleteByDeviceIDAndCode(deviceID, code string) error {
	args := m.Called(deviceID, code)
	return args.Error(0)
}

func (m *MockDeviceStatusRepository) DeleteByDeviceID(deviceID string) error {
	args := m.Called(deviceID)
	return args.Error(0)
}
