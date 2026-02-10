package infrastructure

import (
	"errors"
	"testing"
	"time"

	"teralux_app/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMqttClient is a mock of mqtt.Client
type MockMqttClient struct {
	mock.Mock
}

func (m *MockMqttClient) Connect() mqtt.Token {
	args := m.Called()
	return args.Get(0).(mqtt.Token)
}

func (m *MockMqttClient) Disconnect(quiesce uint) {
	m.Called(quiesce)
}

func (m *MockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	args := m.Called(topic, qos, retained, payload)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(topic, qos, callback)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMqttClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(filters, callback)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	args := m.Called(topics)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMqttClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttClient) IsConnectionOpen() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	m.Called(topic, callback)
}

func (m *MockMqttClient) OptionsReader() mqtt.ClientOptionsReader {
	args := m.Called()
	return args.Get(0).(mqtt.ClientOptionsReader)
}

// MockToken is a mock of mqtt.Token
type MockToken struct {
	mock.Mock
}

func (m *MockToken) Wait() bool {
	return m.Called().Bool(0)
}

func (m *MockToken) WaitTimeout(d time.Duration) bool {
	return m.Called(d).Bool(0)
}

func (m *MockToken) Done() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}

func (m *MockToken) Error() error {
	return m.Called().Error(0)
}

func TestMqttService_Connect(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)
	
	s := &MqttService{
		client: mockClient,
		config: &utils.Config{},
	}

	// Success case
	doneChan := make(chan struct{})
	close(doneChan)
	mockToken.On("Wait").Return(true)
	mockToken.On("Error").Return(nil)
	mockToken.On("Done").Return((<-chan struct{})(doneChan))
	mockClient.On("Connect").Return(mockToken)

	err := s.Connect()
	assert.NoError(t, err)

	// Error case
	mockToken2 := new(MockToken)
	mockToken2.On("Wait").Return(true)
	mockToken2.On("Error").Return(errors.New("connect error"))
	mockClient.On("Connect").Unset()
	mockClient.On("Connect").Return(mockToken2)

	err = s.Connect()
	assert.Error(t, err)
	assert.Equal(t, "connect error", err.Error())
}

func TestMqttService_Publish(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)
	
	s := &MqttService{
		client: mockClient,
		config: &utils.Config{},
	}

	topic := "test/topic"
	payload := "hello"

	// Success case
	mockClient.On("IsConnected").Return(true)
	mockToken.On("Wait").Return(true)
	mockToken.On("Error").Return(nil)
	mockClient.On("Publish", topic, byte(0), false, payload).Return(mockToken)

	err := s.Publish(topic, 0, false, payload)
	assert.NoError(t, err)

	// Not connected case
	mockClient.On("IsConnected").Unset()
	mockClient.On("IsConnected").Return(false)
	err = s.Publish(topic, 0, false, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestMqttService_Subscribe(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)
	
	s := &MqttService{
		client: mockClient,
		config: &utils.Config{},
	}

	topic := "test/topic"
	handler := func(client mqtt.Client, msg mqtt.Message) {}

	mockToken.On("Wait").Return(true)
	mockToken.On("Error").Return(nil)
	mockClient.On("Subscribe", topic, byte(0), mock.Anything).Return(mockToken)

	err := s.Subscribe(topic, 0, handler)
	assert.NoError(t, err)
}
