package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sensio/domain/common/dtos"
	"sensio/domain/common/services"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/entities"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTerminalRepository for controller test
type MockTerminalRepository struct {
	mock.Mock
}

func (m *MockTerminalRepository) Create(terminal *entities.Terminal) error { return nil }
func (m *MockTerminalRepository) GetAll() ([]entities.Terminal, error)     { return nil, nil }
func (m *MockTerminalRepository) GetAllPaginated(offset, limit int, roomID *string) ([]entities.Terminal, int64, error) {
	return nil, 0, nil
}
func (m *MockTerminalRepository) GetByID(id string) (*entities.Terminal, error) { return nil, nil }
func (m *MockTerminalRepository) GetByMacAddress(macAddress string) (*entities.Terminal, error) {
	return nil, nil
}
func (m *MockTerminalRepository) GetByRoomID(roomID string) ([]entities.Terminal, error) {
	args := m.Called(roomID)
	return args.Get(0).([]entities.Terminal), args.Error(1)
}
func (m *MockTerminalRepository) Update(terminal *entities.Terminal) error     { return nil }
func (m *MockTerminalRepository) Delete(id string) error                       { return nil }
func (m *MockTerminalRepository) InvalidateCache(id string) error              { return nil }
func (m *MockTerminalRepository) CreateMQTTUser(user *entities.MQTTUser) error { return nil }
func (m *MockTerminalRepository) GetMQTTUserByUsername(username string) (*entities.MQTTUser, error) {
	return nil, nil
}

// MockMqttService for controller test
type MockMqttService struct {
	mock.Mock
}

func (m *MockMqttService) Connect() error { return nil }
func (m *MockMqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	return nil
}
func (m *MockMqttService) Unsubscribe(topic string) error { return nil }
func (m *MockMqttService) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	args := m.Called(topic, qos, retained, payload)
	return args.Error(0)
}
func (m *MockMqttService) IsConnected() bool { return true }
func (m *MockMqttService) Close()            {}

func TestPublishToRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := services.NewNotificationExternalService(mockRepo, mockMqtt)
		controller := NewNotificationExternalController(service)

		roomID := "room-123"
		dateTimeEnd := "2026-03-17T14:00:00+07:00"
		intervalTime := 15

		terminals := []entities.Terminal{
			{MacAddress: "AA:BB:CC:DD:EE:FF", RoomID: roomID},
		}

		mockRepo.On("GetByRoomID", roomID).Return(terminals, nil)
		mockMqtt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			DateTimeEnd:  dateTimeEnd,
			IntervalTime: intervalTime,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/notification/publish", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.PublishToRoom(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dtos.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Status)
		assert.Equal(t, "Notification published successfully", resp.Message)
	})

	t.Run("Validation Error - Missing RoomID", func(t *testing.T) {
		controller := NewNotificationExternalController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := terminal_dtos.NotificationPublishRequest{
			RoomID:       "",
			DateTimeEnd:  "2026-03-17T14:00:00+07:00",
			IntervalTime: 15,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/notification/publish", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.PublishToRoom(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp dtos.StandardResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Status)
		assert.Contains(t, resp.Message, "room_id is required")
	})

	t.Run("Service Error - Not Found", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := services.NewNotificationExternalService(mockRepo, mockMqtt)
		controller := NewNotificationExternalController(service)

		roomID := "room-empty"
		mockRepo.On("GetByRoomID", roomID).Return([]entities.Terminal{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			DateTimeEnd:  "2026-03-17T14:00:00+07:00",
			IntervalTime: 15,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/notification/publish", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.PublishToRoom(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Service Error - Internal", func(t *testing.T) {
		mockRepo := new(MockTerminalRepository)
		mockMqtt := new(MockMqttService)
		service := services.NewNotificationExternalService(mockRepo, mockMqtt)
		controller := NewNotificationExternalController(service)

		roomID := "room-error"
		mockRepo.On("GetByRoomID", roomID).Return([]entities.Terminal{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := terminal_dtos.NotificationPublishRequest{
			RoomID:       roomID,
			DateTimeEnd:  "2026-03-17T14:00:00+07:00",
			IntervalTime: 15,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/notification/publish", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.PublishToRoom(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
