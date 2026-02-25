package usecases

import (
	"errors"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTeralux_UserBehavior(t *testing.T) {
	repo := new(MockTeraluxRepository)
	extSvc := new(MockTeraluxExternalService)
	useCase := NewCreateTeraluxUseCase(repo, extSvc)

	// 1. Create Teralux (Success Condition)
	t.Run("Create Teralux (Success Condition)", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Master Bedroom Hub",
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", req.MacAddress).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(teralux *entities.Teralux) bool {
			return teralux.Name == req.Name && teralux.MacAddress == req.MacAddress && teralux.RoomID == req.RoomID
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, req.MacAddress, 1).Return(nil).Once()

		res, _, err := useCase.CreateTeralux(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TeraluxID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 1b. Create Teralux with Android ID (Success Condition)
	t.Run("Create Teralux with Android ID", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Android Device",
			MacAddress:   "C756630F2F039D27", // 16 chars hex
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", req.MacAddress).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(teralux *entities.Teralux) bool {
			return teralux.Name == req.Name && teralux.MacAddress == req.MacAddress && teralux.RoomID == req.RoomID
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, req.MacAddress, 1).Return(nil).Once()

		res, _, err := useCase.CreateTeralux(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TeraluxID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 2. Validation: Empty Fields
	t.Run("Validation: Empty Fields", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "",
			MacAddress:   "",
			RoomID:       "",
			DeviceTypeID: "",
		}

		_, _, err := useCase.CreateTeralux(req)
		assert.Error(t, err)
		var valErr *utils.ValidationError
		if assert.ErrorAs(t, err, &valErr) {
			assert.Equal(t, "Validation Error", valErr.Message)
			assert.Len(t, valErr.Details, 4)
		}
	})

	// 3. Validation: Invalid MAC Address Format
	t.Run("Validation: Invalid MAC Address Format", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Living Room",
			MacAddress:   "INVALID-MAC",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		_, _, err := useCase.CreateTeralux(req)
		assert.Error(t, err)
		var valErr *utils.ValidationError
		if assert.ErrorAs(t, err, &valErr) {
			found := false
			for _, d := range valErr.Details {
				if d.Field == "mac_address" && d.Message == "invalid mac address format" {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected detail for mac_address invalid format")
		}
	})

	// 4. Validation: Name Too Long
	t.Run("Validation: Name Too Long", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         string(make([]byte, 256)),
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		_, _, err := useCase.CreateTeralux(req)
		assert.Error(t, err)
		var valErr *utils.ValidationError
		if assert.ErrorAs(t, err, &valErr) {
			found := false
			for _, d := range valErr.Details {
				if d.Field == "name" && d.Message == "name must be 255 characters or less" {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected detail for name too long")
		}
	})

	// 5. Conflict: Duplicate MAC Address
	t.Run("Conflict: Duplicate MAC Address", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "New Hub",
			MacAddress:   "DD:EE:FF:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		// Mock normalization: "dd:ee:ff:11:22:33" -> "DD:EE:FF:11:22:33"
		repo.On("GetByMacAddress", "DD:EE:FF:11:22:33").Return(&entities.Teralux{ID: "existing-id"}, nil).Once()

		req.MacAddress = "dd:ee:ff:11:22:33"

		_, _, err := useCase.CreateTeralux(req)
		assert.Error(t, err)
		var apiErr *utils.APIError
		if assert.ErrorAs(t, err, &apiErr) {
			assert.Equal(t, 409, apiErr.StatusCode)
		}
		repo.AssertExpectations(t)
	})

	// 5b. Create Teralux with 12-digit RAW MAC (Success)
	t.Run("Create Teralux with 12-digit RAW MAC", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Raw Hub",
			MacAddress:   "aabbcc112233",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		// Normalized MAC
		normMAC := "AABBCC112233"
		repo.On("GetByMacAddress", normMAC).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(teralux *entities.Teralux) bool {
			return teralux.MacAddress == normMAC
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, normMAC, 1).Return(nil).Once()

		res, _, err := useCase.CreateTeralux(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TeraluxID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 6. External Service Error
	t.Run("External Service Error", func(t *testing.T) {
		req := &dtos.CreateTeraluxRequestDTO{
			Name:         "Error Hub",
			MacAddress:   "11:22:33:44:55:66",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", "11:22:33:44:55:66").Return(nil, nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, "11:22:33:44:55:66", 1).Return(errors.New("API error")).Once()

		_, _, err := useCase.CreateTeralux(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
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
	return args.Get(0).([]entities.Teralux), args.Get(1).(int64), args.Error(2)
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

type MockTeraluxExternalService struct {
	mock.Mock
}

func (m *MockTeraluxExternalService) ProcInsertMacAddress(roomID int, macAddress string, deviceTypeID int) error {
	args := m.Called(roomID, macAddress, deviceTypeID)
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
