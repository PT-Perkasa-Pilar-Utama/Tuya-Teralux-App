package usecases

import (
	"errors"
	"sensio/domain/terminal/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMqttAuthClient mocks the MqttAuthClient for tests
type MockMqttAuthClient struct {
	credentials    *services.MQTTCredentials
	credentialsErr error
	createExists   bool
	createErr      error
}

func (m *MockMqttAuthClient) GetMQTTCredentials(_ string) (*services.MQTTCredentials, error) {
	return m.credentials, m.credentialsErr
}

func (m *MockMqttAuthClient) CreateMQTTUser(_, _ string) (bool, error) {
	return m.createExists, m.createErr
}

func TestGetTerminalByMACUseCase(t *testing.T) {
	mockClient := services.NewMqttAuthClient("http://localhost:5500", "test-key")

	// 1. Get Terminal By MAC (Not Found)
	t.Run("Get Terminal By MAC (Not Found)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewGetTerminalByMACUseCase(repo, mockClient)

		repo.On("GetByMacAddress", "AA:BB:CC:DD:EE:FF").Return(nil, errors.New("record not found")).Once()

		_, err := useCase.GetTerminalByMAC("AA:BB:CC:DD:EE:FF")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terminal not found")
		repo.AssertExpectations(t)
	})

	// 2. Get Terminal By MAC (Invalid Format)
	t.Run("Get Terminal By MAC (Invalid Format)", func(t *testing.T) {
		repo := new(MockTerminalRepository)
		useCase := NewGetTerminalByMACUseCase(repo, mockClient)

		_, err := useCase.GetTerminalByMAC("invalid-mac")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mac address")
	})
}
