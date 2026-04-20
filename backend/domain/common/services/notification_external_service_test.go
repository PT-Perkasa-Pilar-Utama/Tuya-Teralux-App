package services

import (
	"errors"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/entities"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTerminalRepository is a mock implementation of ITerminalRepository
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

func (m *MockTerminalRepository) GetByRoomID(roomID string) ([]entities.Terminal, error) {
	args := m.Called(roomID)
	return args.Get(0).([]entities.Terminal), args.Error(1)
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

// MockMqttService is a mock implementation of infrastructure.IMqttService
type MockMqttService struct {
	mock.Mock
}

func (m *MockMqttService) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	args := m.Called(topic, qos, handler)
	return args.Error(0)
}

func (m *MockMqttService) Unsubscribe(topic string) error {
	args := m.Called(topic)
	return args.Error(0)
}

func (m *MockMqttService) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	args := m.Called(topic, qos, retained, payload)
	return args.Error(0)
}

func (m *MockMqttService) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttService) Close() {
	m.Called()
}

func ptr(s string) *string { return &s }

func TestPublishNotificationToRoom(t *testing.T) {
	roomID := "room-123"
	scheduledAt := "2026-03-17T14:00:00+07:00"
	env := utils.GetConfig().ApplicationEnvironment

	t.Run("Success - Multiple Terminals", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		terminals := []entities.Terminal{
			{MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: roomID},
			{MacAddress: "11:22:33:44:55:66", RoomID: roomID},
		}

		mockRepo.On("GetByRoomID", roomID).Return(terminals, nil)
		mockMqtt.On("IsConnected").Return(true)
		mockMqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr(scheduledAt),
			PhoneNumbers: []string{"+6281234567890"},
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, roomID, resp.RoomID)
		assert.Equal(t, 2, resp.PublishedCount)
		assert.ElementsMatch(t, []string{
			"users/AA:BB:CC:DD:EE:FF/" + env + "/notification",
			"users/11:22:33:44:55:66/" + env + "/notification",
		}, resp.PublishedTopics)

		mockRepo.AssertExpectations(t)
		mockMqtt.AssertExpectations(t)
	})

	t.Run("Failure - No Terminals Found", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		mockRepo.On("GetByRoomID", roomID).Return([]entities.Terminal{}, nil)

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr(scheduledAt),
			PhoneNumbers: []string{"+6281234567890"},
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		var apiErr *utils.APIError
		assert.True(t, errors.As(err, &apiErr))
		assert.Equal(t, 404, apiErr.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Invalid ScheduledAt Format", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr("invalid-format"),
			PhoneNumbers: []string{"+6281234567890"},
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		var apiErr *utils.APIError
		assert.True(t, errors.As(err, &apiErr))
		assert.Equal(t, 400, apiErr.StatusCode)
		assert.Contains(t, apiErr.Message, "Invalid scheduled_at format")
	})

	t.Run("Success - With PhoneNumbers and Template", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		terminals := []entities.Terminal{
			{MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: roomID},
		}

		mockRepo.On("GetByRoomID", roomID).Return(terminals, nil)
		mockMqtt.On("IsConnected").Return(true)
		mockMqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr(scheduledAt),
			PhoneNumbers: []string{"+6281234567890"},
			Template:     "start_meeting",
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, roomID, resp.RoomID)
		assert.Equal(t, 1, resp.PublishedCount)

		mockRepo.AssertExpectations(t)
		mockMqtt.AssertExpectations(t)
	})

	t.Run("Failure - MQTT Publish Error", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		terminals := []entities.Terminal{
			{MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: roomID},
		}

		mockRepo.On("GetByRoomID", roomID).Return(terminals, nil)
		mockMqtt.On("IsConnected").Return(true)
		mockMqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mqtt error"))

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr(scheduledAt),
			PhoneNumbers: []string{"+6281234567890"},
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "mqtt error")
	})

	t.Run("Success - Different Valid RFC3339 Timezone Offset", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := NewNotificationExternalService(mockRepo, mockMqtt)

		terminals := []entities.Terminal{
			{MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: roomID},
		}

		mockRepo.On("GetByRoomID", roomID).Return(terminals, nil)
		mockMqtt.On("IsConnected").Return(true)
		mockMqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			ScheduledAt:  ptr("2026-03-17T23:00:00Z"),
			PhoneNumbers: []string{"+6281234567890"},
		}

		resp, err := service.PublishNotificationToRoom(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		mockRepo.AssertExpectations(t)
		mockMqtt.AssertExpectations(t)
	})
}
