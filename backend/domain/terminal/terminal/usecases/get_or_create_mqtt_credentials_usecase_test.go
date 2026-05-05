package usecases

import (
	"errors"
	"regexp"
	"testing"

	"sensio/domain/terminal/terminal/entities"
	"sensio/domain/terminal/terminal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMQTTAuthClient is a mock implementation of MqttAuthClient for testing
type MockMQTTAuthClient struct {
	mock.Mock
}

func (m *MockMQTTAuthClient) GetMQTTCredentials(username string) (*services.MQTTCredentials, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MQTTCredentials), args.Error(1)
}

func (m *MockMQTTAuthClient) CreateMQTTUser(username, password string) (bool, error) {
	args := m.Called(username, password)
	return args.Bool(0), args.Error(1)
}

func (m *MockMQTTAuthClient) DeleteMQTTUser(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

// MockITerminalRepository is a mock implementation of ITerminalRepository for testing
type MockITerminalRepository struct {
	mock.Mock
}

func (m *MockITerminalRepository) Create(terminal *entities.Terminal) error {
	args := m.Called(terminal)
	return args.Error(0)
}

func (m *MockITerminalRepository) GetAll() ([]entities.Terminal, error) {
	args := m.Called()
	return args.Get(0).([]entities.Terminal), args.Error(1)
}

func (m *MockITerminalRepository) GetAllPaginated(offset, limit int, roomID *string) ([]entities.Terminal, int64, error) {
	args := m.Called(offset, limit, roomID)
	return args.Get(0).([]entities.Terminal), args.Get(1).(int64), args.Error(2)
}

func (m *MockITerminalRepository) GetByID(id string) (*entities.Terminal, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Terminal), args.Error(1)
}

func (m *MockITerminalRepository) GetByMacAddress(macAddress string) (*entities.Terminal, error) {
	args := m.Called(macAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Terminal), args.Error(1)
}

func (m *MockITerminalRepository) GetByRoomID(roomID string) ([]entities.Terminal, error) {
	args := m.Called(roomID)
	return args.Get(0).([]entities.Terminal), args.Error(1)
}

func (m *MockITerminalRepository) Update(terminal *entities.Terminal) error {
	args := m.Called(terminal)
	return args.Error(0)
}

func (m *MockITerminalRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockITerminalRepository) InvalidateCache(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockITerminalRepository) CreateMQTTUser(user *entities.MQTTUser) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockITerminalRepository) GetMQTTUserByUsername(username string) (*entities.MQTTUser, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.MQTTUser), args.Error(1)
}

func TestGetOrCreateMQTTCredentials_TerminalExists_MQTTExists(t *testing.T) {
	mockRepo := new(MockITerminalRepository)
	mockMqtt := new(MockMQTTAuthClient)
	uc := NewGetOrCreateMQTTCredentialsUseCase(mockRepo, mockMqtt)

	macAddress := "AA:BB:CC:DD:EE:FF"
	terminal := &entities.Terminal{
		ID:         "terminal-123",
		MacAddress: macAddress,
		RoomID:     "room-1",
		Name:       "Test Terminal",
	}
	creds := &services.MQTTCredentials{
		Username: macAddress,
		Password: "existing-password",
	}

	mockRepo.On("GetByMacAddress", macAddress).Return(terminal, nil)
	mockMqtt.On("GetMQTTCredentials", macAddress).Return(creds, nil)

	result, err := uc.GetOrCreateMQTTCredentials(macAddress)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, macAddress, result.Username)
	assert.Equal(t, "existing-password", result.Password)

	mockRepo.AssertExpectations(t)
	mockMqtt.AssertExpectations(t)
}

func TestGetOrCreateMQTTCredentials_TerminalExists_MQTTMissing(t *testing.T) {
	mockRepo := new(MockITerminalRepository)
	mockMqtt := new(MockMQTTAuthClient)
	uc := NewGetOrCreateMQTTCredentialsUseCase(mockRepo, mockMqtt)

	macAddress := "AA:BB:CC:DD:EE:FF"
	terminal := &entities.Terminal{
		ID:         "terminal-123",
		MacAddress: macAddress,
		RoomID:     "room-1",
		Name:       "Test Terminal",
	}

	mockRepo.On("GetByMacAddress", macAddress).Return(terminal, nil)
	mockMqtt.On("GetMQTTCredentials", macAddress).Return(nil, nil) // MQTT user not found
	mockMqtt.On("CreateMQTTUser", macAddress, mock.Anything).Return(false, nil)

	result, err := uc.GetOrCreateMQTTCredentials(macAddress)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, macAddress, result.Username)
	assert.NotEmpty(t, result.Password)

	mockRepo.AssertExpectations(t)
	mockMqtt.AssertExpectations(t)
}

func TestGetOrCreateMQTTCredentials_TerminalNotFound(t *testing.T) {
	mockRepo := new(MockITerminalRepository)
	mockMqtt := new(MockMQTTAuthClient)
	uc := NewGetOrCreateMQTTCredentialsUseCase(mockRepo, mockMqtt)

	macAddress := "AA:BB:CC:DD:EE:FF"

	mockRepo.On("GetByMacAddress", macAddress).Return(nil, errors.New("terminal not found"))

	result, err := uc.GetOrCreateMQTTCredentials(macAddress)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockMqtt.AssertExpectations(t)
}

func TestGetOrCreateMQTTCredentials_RaceCondition_409Conflict(t *testing.T) {
	mockRepo := new(MockITerminalRepository)
	mockMqtt := new(MockMQTTAuthClient)
	uc := NewGetOrCreateMQTTCredentialsUseCase(mockRepo, mockMqtt)

	macAddress := "AA:BB:CC:DD:EE:FF"
	terminal := &entities.Terminal{
		ID:         "terminal-123",
		MacAddress: macAddress,
		RoomID:     "room-1",
		Name:       "Test Terminal",
	}
	creds := &services.MQTTCredentials{
		Username: macAddress,
		Password: "password-from-race",
	}

	mockRepo.On("GetByMacAddress", macAddress).Return(terminal, nil)
	// First call: GetMQTTCredentials returns nil (user doesn't exist)
	mockMqtt.On("GetMQTTCredentials", macAddress).Return(nil, nil).Once()
	// CreateMQTTUser returns 409 Conflict (alreadyExists = true)
	mockMqtt.On("CreateMQTTUser", macAddress, mock.Anything).Return(true, nil).Once()
	// Second call: GetMQTTCredentials returns the newly created creds
	mockMqtt.On("GetMQTTCredentials", macAddress).Return(creds, nil).Once()

	result, err := uc.GetOrCreateMQTTCredentials(macAddress)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, macAddress, result.Username)
	assert.Equal(t, "password-from-race", result.Password)

	mockRepo.AssertExpectations(t)
	mockMqtt.AssertExpectations(t)
}

func TestGetOrCreateMQTTCredentials_ValidationError_EmptyMac(t *testing.T) {
	mockRepo := new(MockITerminalRepository)
	mockMqtt := new(MockMQTTAuthClient)
	uc := NewGetOrCreateMQTTCredentialsUseCase(mockRepo, mockMqtt)

	validMAC := regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$|^[0-9a-fA-F]{16}$`)

	invalidMacAddresses := []string{
		"",
		"invalid",
		"AA:BB:CC:DD:EE",    // incomplete
		"GG:HH:II:JJ:KK:LL", // invalid hex
		"AA-BB-CC-DD-EE-FF", // wrong separator
	}

	for _, mac := range invalidMacAddresses {
		if validMAC.MatchString(mac) {
			continue // Skip if it's actually valid (e.g., Android ID format)
		}

		result, err := uc.GetOrCreateMQTTCredentials(mac)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid mac address")
	}

	// No mocks should be called for invalid input
	mockRepo.AssertNotCalled(t, "GetByMacAddress")
	mockMqtt.AssertNotCalled(t, "GetMQTTCredentials")
}
