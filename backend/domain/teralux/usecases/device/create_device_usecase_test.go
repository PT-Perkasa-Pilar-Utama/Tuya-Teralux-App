package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDeviceUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceRepository)
	statusRepo := new(MockDeviceStatusRepository)
	teraRepo := new(MockTeraluxRepository)
	tuyaAuth := new(MockTuyaAuthUseCase)
	tuyaGetDevice := new(MockTuyaGetDeviceByIDUseCase)
	
	useCase := NewCreateDeviceUseCase(repo, statusRepo, tuyaAuth, tuyaGetDevice, teraRepo)

	// 1. Create Device (Success)
	t.Run("Create Device (Success)", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			ID:        "tuya-device-123",
			Name:      "Kitchen Light",
			TeraluxID: "tx-1",
		}

		teraRepo.On("GetByID", req.TeraluxID).Return(&entities.Teralux{ID: "tx-1"}, nil).Once()
		tuyaAuth.On("Authenticate").Return(&tuya_dtos.TuyaAuthResponseDTO{AccessToken: "token"}, nil).Once()
		tuyaGetDevice.On("GetDeviceByID", "token", req.ID).Return(&tuya_dtos.TuyaDeviceDTO{
			ID: req.ID,
			Name: "Mocked",
			Status: []tuya_dtos.TuyaDeviceStatusDTO{{Code: "s1", Value: "v1"}},
		}, nil).Once()
		
		repo.On("GetByIDUnscoped", req.ID).Return(nil, assert.AnError).Once()
		repo.On("Create", mock.Anything).Return(nil).Once()
		statusRepo.On("UpsertDeviceStatuses", req.ID, mock.Anything).Return(nil).Once()
		teraRepo.On("InvalidateCache", req.TeraluxID).Return(nil).Once()

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
			Name:      "",
			TeraluxID: "",
		}
		_, _, err := useCase.CreateDevice(req)
		assert.Error(t, err)
	})

	// 3. Constraint: Invalid Teralux ID
	t.Run("Constraint: Invalid Teralux ID", func(t *testing.T) {
		req := &dtos.CreateDeviceRequestDTO{
			Name:      "Ghost Device",
			TeraluxID: "tx-999",
		}
		teraRepo.On("GetByID", "tx-999").Return(nil, assert.AnError).Once()
		
		_, _, err := useCase.CreateDevice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid teralux_id")
		teraRepo.AssertExpectations(t)
	})
}

// --- Mocks ---

type MockTeraluxRepository struct {
	mock.Mock
}

func (m *MockTeraluxRepository) Create(teralux *entities.Teralux) error {
	args := m.Called(teralux)
	return args.Error(0)
}

func (m *MockTeraluxRepository) GetAll() ([]entities.Teralux, error) {
	args := m.Called()
	return args.Get(0).([]entities.Teralux), args.Error(1)
}

func (m *MockTeraluxRepository) GetAllPaginated(offset, limit int, roomID *string) ([]entities.Teralux, int64, error) {
	args := m.Called(offset, limit, roomID)
	return args.Get(0).([]entities.Teralux), int64(args.Int(1)), args.Error(2)
}

func (m *MockTeraluxRepository) GetByID(id string) (*entities.Teralux, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Teralux), args.Error(1)
}

func (m *MockTeraluxRepository) GetByMacAddress(macAddress string) (*entities.Teralux, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Teralux), args.Error(1)
}

func (m *MockTeraluxRepository) Update(teralux *entities.Teralux) error {
	args := m.Called(teralux)
	return args.Error(0)
}

func (m *MockTeraluxRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTeraluxRepository) InvalidateCache(id string) error {
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

func (m *MockDeviceRepository) GetByTeraluxID(teraluxID string) ([]entities.Device, error) {
	args := m.Called(teraluxID)
	return args.Get(0).([]entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByTeraluxIDPaginated(teraluxID string, offset, limit int) ([]entities.Device, int64, error) {
	args := m.Called(teraluxID, offset, limit)
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
