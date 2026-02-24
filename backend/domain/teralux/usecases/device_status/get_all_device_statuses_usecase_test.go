package usecases

import (
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllDeviceStatusesUseCase_UserBehavior(t *testing.T) {
	repo := new(MockDeviceStatusRepository)
	useCase := NewGetAllDeviceStatusesUseCase(repo)

	// 1. Get All Statuses (Success)
	t.Run("Get All Statuses (Success)", func(t *testing.T) {
		expectedStatuses := []entities.DeviceStatus{
			{DeviceID: "d1", Code: "c1", Value: "v1"},
			{DeviceID: "d2", Code: "c2", Value: "v2"},
		}
		
		repo.On("GetAllPaginated", 0, 0).Return(expectedStatuses, int64(2), nil).Once()

		res, err := useCase.ListDeviceStatuses(0, 0)
		assert.NoError(t, err)
		assert.Len(t, res.DeviceStatuses, 2)
		assert.Equal(t, "v1", res.DeviceStatuses[0].Value)
		
		repo.AssertExpectations(t)
	})

	// 2. Get All Statuses (Empty)
	t.Run("Get All Statuses (Empty)", func(t *testing.T) {
		repo.On("GetAllPaginated", 0, 0).Return([]entities.DeviceStatus{}, int64(0), nil).Once()

		res, err := useCase.ListDeviceStatuses(0, 0)
		assert.NoError(t, err)
		assert.Len(t, res.DeviceStatuses, 0)
		
		repo.AssertExpectations(t)
	})
}

// --- Mocks ---

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
	return args.Get(0).([]entities.Device), int64(args.Int(1)), args.Error(2)
}

func (m *MockDeviceRepository) GetByTeraluxID(teraluxID string) ([]entities.Device, error) {
	args := m.Called(teraluxID)
	return args.Get(0).([]entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByTeraluxIDPaginated(teraluxID string, offset, limit int) ([]entities.Device, int64, error) {
	args := m.Called(teraluxID, offset, limit)
	return args.Get(0).([]entities.Device), int64(args.Int(1)), args.Error(1)
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
	return args.Get(0).([]entities.DeviceStatus), args.Get(1).(int64), args.Error(2)
}

func (m *MockDeviceStatusRepository) GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error) {
	args := m.Called(deviceID)
	return args.Get(0).([]entities.DeviceStatus), args.Error(1)
}

func (m *MockDeviceStatusRepository) GetByDeviceIDPaginated(deviceID string, offset, limit int) ([]entities.DeviceStatus, int64, error) {
	args := m.Called(deviceID, offset, limit)
	return args.Get(0).([]entities.DeviceStatus), args.Get(1).(int64), args.Error(2)
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
