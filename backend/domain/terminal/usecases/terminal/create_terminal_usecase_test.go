package usecases

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/entities"
	"sensio/domain/terminal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newMqttTestServer spins up a test HTTP server that simulates the Rust EMQX Auth Service.
// If mqttExists is true, POST /mqtt/create returns 409 (already exists).
// GET /mqtt/credentials/{username} always returns the test credentials.
func newMqttTestServer(mqttExists bool) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/mqtt/create", func(w http.ResponseWriter, r *http.Request) {
		if mqttExists {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/mqtt/credentials/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "ok",
			"data": map[string]string{
				"username": "test-username",
				"password": "decrypted-pass",
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestCreateTerminal_UserBehavior(t *testing.T) {
	repo := new(MockTerminalRepository)
	extSvc := new(MockTerminalExternalService)

	// Create useCase with a test MQTT server (new user scenario: 200 OK)
	mqttSrv := newMqttTestServer(false)
	defer mqttSrv.Close()
	mqttClient := services.NewMqttAuthClient(mqttSrv.URL, "test-api-key")
	useCase := NewCreateTerminalUseCase(repo, extSvc, mqttClient)

	// 1. Create Terminal (Success Condition)
	t.Run("Create Terminal (Success Condition)", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Master Bedroom Hub",
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", req.MacAddress).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(terminal *entities.Terminal) bool {
			return terminal.Name == req.Name && terminal.MacAddress == req.MacAddress && terminal.RoomID == req.RoomID
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, req.MacAddress, 1).Return(nil).Once()

		res, _, err := useCase.CreateTerminal(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TerminalID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 1c. Create Terminal (Scenario C: MQTT User already exists)
	t.Run("Create Terminal (Scenario C: MQTT User already exists)", func(t *testing.T) {
		// Create a separate useCase with MQTT server returning 409
		mqttSrvC := newMqttTestServer(true)
		defer mqttSrvC.Close()
		mqttClientC := services.NewMqttAuthClient(mqttSrvC.URL, "test-api-key")
		useCaseC := NewCreateTerminalUseCase(repo, extSvc, mqttClientC)

		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Duplicate MQTT Hub",
			MacAddress:   "DE:AD:BE:EF:00:11",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", req.MacAddress).Return(nil, nil).Once()
		repo.On("Create", mock.Anything).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, req.MacAddress, 1).Return(nil).Once()

		res, _, err := useCaseC.CreateTerminal(req)
		assert.NoError(t, err)
		assert.Empty(t, res.MQTTPassword) // existing user: password returned empty, must call GET /credentials
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 1b. Create Terminal with Android ID (Success Condition)
	t.Run("Create Terminal with Android ID", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Android Device",
			MacAddress:   "C756630F2F039D27", // 16 chars hex
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		normMAC := "C756630F2F039D27"
		repo.On("GetByMacAddress", normMAC).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(terminal *entities.Terminal) bool {
			return terminal.MacAddress == normMAC
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, normMAC, 1).Return(nil).Once()

		res, _, err := useCase.CreateTerminal(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TerminalID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 2. Validation: Empty Fields
	t.Run("Validation: Empty Fields", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "",
			MacAddress:   "",
			RoomID:       "",
			DeviceTypeID: "",
		}

		_, _, err := useCase.CreateTerminal(req)
		assert.Error(t, err)
		var valErr *utils.ValidationError
		if assert.ErrorAs(t, err, &valErr) {
			assert.Equal(t, "Validation Error", valErr.Message)
			assert.Len(t, valErr.Details, 4)
		}
	})

	// 3. Validation: Invalid MAC Address Format
	t.Run("Validation: Invalid MAC Address Format", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Living Room",
			MacAddress:   "INVALID-MAC",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		_, _, err := useCase.CreateTerminal(req)
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
		req := &dtos.CreateTerminalRequestDTO{
			Name:         string(make([]byte, 256)),
			MacAddress:   "AA:BB:CC:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		_, _, err := useCase.CreateTerminal(req)
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
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "New Hub",
			MacAddress:   "DD:EE:FF:11:22:33",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", "DD:EE:FF:11:22:33").Return(&entities.Terminal{ID: "existing-id"}, nil).Once()
		req.MacAddress = "dd:ee:ff:11:22:33"

		_, _, err := useCase.CreateTerminal(req)
		assert.Error(t, err)
		var apiErr *utils.APIError
		if assert.ErrorAs(t, err, &apiErr) {
			assert.Equal(t, 409, apiErr.StatusCode)
		}
		repo.AssertExpectations(t)
	})

	// 5b. Create Terminal with 12-digit RAW MAC (Success)
	t.Run("Create Terminal with 12-digit RAW MAC", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Raw Hub",
			MacAddress:   "aabbcc112233",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		normMAC := "AABBCC112233"
		repo.On("GetByMacAddress", normMAC).Return(nil, nil).Once()
		repo.On("Create", mock.MatchedBy(func(terminal *entities.Terminal) bool {
			return terminal.MacAddress == normMAC
		})).Return(nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, normMAC, 1).Return(nil).Once()

		res, _, err := useCase.CreateTerminal(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.TerminalID)
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
	})

	// 6. External Service Error
	t.Run("External Service Error", func(t *testing.T) {
		req := &dtos.CreateTerminalRequestDTO{
			Name:         "Error Hub",
			MacAddress:   "11:22:33:44:55:66",
			RoomID:       "1",
			DeviceTypeID: "1",
		}

		repo.On("GetByMacAddress", "11:22:33:44:55:66").Return(nil, nil).Once()
		extSvc.On("ProcInsertMacAddress", 1, "11:22:33:44:55:66", 1).Return(errors.New("API error")).Once()

		_, _, err := useCase.CreateTerminal(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		repo.AssertExpectations(t)
		extSvc.AssertExpectations(t)
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
	return args.Get(0).([]entities.Terminal), args.Get(1).(int64), args.Error(2)
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

func (m *MockTerminalRepository) CreateMQTTUser(user *entities.MQTTUser) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockTerminalRepository) GetMQTTUserByUsername(username string) (*entities.MQTTUser, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.MQTTUser), args.Error(1)
}

type MockTerminalExternalService struct {
	mock.Mock
}

func (m *MockTerminalExternalService) ProcInsertMacAddress(roomID int, macAddress string, deviceTypeID int) error {
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

func (m *MockDeviceRepository) GetByTerminalID(terminalID string) ([]entities.Device, error) {
	args := m.Called(terminalID)
	return args.Get(0).([]entities.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByTerminalIDPaginated(terminalID string, offset, limit int) ([]entities.Device, int64, error) {
	args := m.Called(terminalID, offset, limit)
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
