package usecases

import (
	"os"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/repositories"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTuyaExecutor is a mock for TuyaDeviceControlExecutor
type MockTuyaExecutor struct {
	mock.Mock
}

func (m *MockTuyaExecutor) SendCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error) {
	args := m.Called(accessToken, deviceID, commands)
	return args.Bool(0), args.Error(1)
}

func (m *MockTuyaExecutor) SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error) {
	args := m.Called(accessToken, infraredID, remoteID, code, value)
	return args.Bool(0), args.Error(1)
}

// MockMqttPublisher is a mock for MqttPublisher
type MockMqttPublisher struct {
	mock.Mock
}

func (m *MockMqttPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	args := m.Called(topic, qos, retained, payload)
	return args.Error(0)
}

func TestSceneUseCases(t *testing.T) {
	// Initialize config for BadgerService
	utils.AppConfig = &utils.Config{
		CacheTTL: "1h",
	}

	dbPath := "/tmp/scene_test_badger"
	os.RemoveAll(dbPath)
	badger, _ := infrastructure.NewBadgerService(dbPath)
	defer func() {
		badger.Close()
		os.RemoveAll(dbPath)
	}()
	repo := repositories.NewSceneRepository(badger)

	// 1. Add Scene
	addUC := NewAddSceneUseCase(repo)
	req := &dtos.CreateSceneRequestDTO{
		Name: "Test Scene",
		Actions: []dtos.ActionDTO{
			{DeviceID: "dev1", Code: "switch", Value: true},
			{Topic: "test/topic", Value: "hello"},
		},
	}
	id, err := addUC.Execute(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// 2. Get All
	getAllUC := NewGetAllScenesUseCase(repo)
	scenes, err := getAllUC.Execute()
	assert.NoError(t, err)
	assert.Len(t, scenes, 1)
	assert.Equal(t, "Test Scene", scenes[0].Name)

	// 3. Control Scene
	mockTuya := new(MockTuyaExecutor)
	mockMqtt := new(MockMqttPublisher)
	controlUC := NewControlSceneUseCase(repo, mockTuya, mockMqtt)

	accessToken := "token123"
	mockTuya.On("SendCommand", accessToken, "dev1", []tuya_dtos.TuyaCommandDTO{{Code: "switch", Value: true}}).Return(true, nil)
	mockMqtt.On("Publish", "test/topic", byte(0), false, "hello").Return(nil)

	err = controlUC.Execute(id, accessToken)
	assert.NoError(t, err)
	mockTuya.AssertExpectations(t)
	mockMqtt.AssertExpectations(t)

	// 4. Update Scene
	updateUC := NewUpdateSceneUseCase(repo)
	updateReq := &dtos.UpdateSceneRequestDTO{
		Name: "Updated Scene",
		Actions: []dtos.ActionDTO{
			{DeviceID: "dev2", Code: "switch", Value: false},
		},
	}
	err = updateUC.Execute(id, updateReq)
	assert.NoError(t, err)

	updatedScenes, _ := getAllUC.Execute()
	assert.Equal(t, "Updated Scene", updatedScenes[0].Name)

	// 5. Delete Scene
	deleteUC := NewDeleteSceneUseCase(repo)
	err = deleteUC.Execute(id)
	assert.NoError(t, err)

	finalScenes, _ := getAllUC.Execute()
	assert.Len(t, finalScenes, 0)
}
