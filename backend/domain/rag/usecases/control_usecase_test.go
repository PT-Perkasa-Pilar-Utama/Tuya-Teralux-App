package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/rag/skills"
	tuyaDtos "sensio/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMForControl is a mock for skills.LLMClient
type mockLLMForControl struct {
	mock.Mock
}

func (m *mockLLMForControl) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	args := m.Called(ctx, prompt, model)
	return args.String(0), args.Error(1)
}

// mockTuyaAuthForControl is a mock for TuyaAuthUseCase
type mockTuyaAuthForControl struct {
	mock.Mock
}

func (m *mockTuyaAuthForControl) Authenticate() (*tuyaDtos.TuyaAuthResponseDTO, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tuyaDtos.TuyaAuthResponseDTO), args.Error(1)
}

func (m *mockTuyaAuthForControl) GetTuyaAccessToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// mockTuyaExecutorForControl is a mock for TuyaDeviceControlExecutor
type mockTuyaExecutorForControl struct {
	mock.Mock
}

func (m *mockTuyaExecutorForControl) SendSwitchCommand(accessToken, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
	args := m.Called(accessToken, deviceID, commands)
	return args.Bool(0), args.Error(1)
}

func (m *mockTuyaExecutorForControl) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	args := m.Called(accessToken, infraredID, remoteID, params)
	return args.Bool(0), args.Error(1)
}

func TestControlUseCase_ProcessControl(t *testing.T) {
	mockLLM := new(mockLLMForControl)
	cfg := &utils.Config{GeminiModelHigh: "test-model-control"}

	// Setup Vector Service
	vectorFile := "test_vector_control.json"
	defer fmt.Println("Cleaning up test vector file")
	defer func() { _ = os.Remove(vectorFile) }()
	vector := infrastructure.NewVectorService(vectorFile)

	// Setup Badger
	badger, _ := infrastructure.NewBadgerService("") // In-memory

	// Setup mocks
	mockTuyaAuth := new(mockTuyaAuthForControl)
	mockTuyaExecutor := new(mockTuyaExecutorForControl)

	// Setup mock responses
	mockTuyaAuth.On("GetTuyaAccessToken").Return("mock-token", nil)
	mockTuyaExecutor.On("SendSwitchCommand", "mock-token", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mockTuyaExecutor.On("SendIRACCommand", "mock-token", mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Maybe()

	mockControlSkill := new(mockSkill)
	uc := NewControlUseCase(mockLLM, nil, cfg, vector, badger, mockTuyaExecutor, mockTuyaAuth, mockControlSkill)

	uid := "user-123"
	terminalID := "tx-1"

	// 1. Setup Mock User Devices in Vector DB
	devices := []tuyaDtos.TuyaDeviceDTO{
		{
			ID:       "dev-ac-1",
			Name:     "AC Kamar Utama",
			RemoteID: "remote-ac-1",
			Status: []tuyaDtos.TuyaDeviceStatusDTO{
				{Code: "power", Value: 0},
				{Code: "temp", Value: 24},
			},
		},
		{
			ID:       "dev-ac-2",
			Name:     "AC Ruang Tamu",
			RemoteID: "remote-ac-2",
			Status: []tuyaDtos.TuyaDeviceStatusDTO{
				{Code: "power", Value: 0},
				{Code: "temp", Value: 24},
			},
		},
		{
			ID:   "dev-lamp-1",
			Name: "Lampu Teras",
			Status: []tuyaDtos.TuyaDeviceStatusDTO{
				{Code: "switch_1", Value: false},
			},
		},
	}
	resp := tuyaDtos.TuyaDevicesResponseDTO{Devices: devices}
	respJSON, _ := json.Marshal(resp)
	_ = vector.Upsert(fmt.Sprintf("tuya:devices:uid:%s", uid), string(respJSON), nil)

	for _, d := range devices {
		searchDoc := fmt.Sprintf("Device: %s | Category: %s | Room: %s | Product: %s | Hub: %s | ID: %s",
			d.Name, "Light / Switch / Socket", "Bedroom", "Test Product", "Test Hub", d.ID)
		_ = vector.Upsert("tuya:device:"+d.ID, searchDoc, nil)
	}

	t.Run("Single Match", func(t *testing.T) {
		mockControlSkill.On("Execute", mock.Anything).Return(&skills.SkillResult{
			Message:        "Lampu Teras has been turned on.",
			Data:           map[string]interface{}{"device_id": "dev-lamp-1"},
			HTTPStatusCode: 200,
		}, nil).Once()

		res, err := uc.ProcessControl(uid, terminalID, "Nyalakan Lampu Teras")
		assert.NoError(t, err)
		assert.Contains(t, res.Message, "turned on")
		assert.Contains(t, res.Message, "Lampu Teras")
		assert.Equal(t, "dev-lamp-1", res.DeviceID)
	})

	t.Run("Multiple Matches (Ambiguity)", func(t *testing.T) {
		prompt := "Matikan AC"
		localMock := new(mockLLMForControl)
		// Re-setup usecase with local mock to avoid interference
		// Expect LLM call for disambiguation
		expectedResponse := "I found 2 matching devices: AC Kamar Utama, AC Ruang Tamu. Which one?"
		mSkill := new(mockSkill)
		localUC := NewControlUseCase(localMock, nil, cfg, vector, badger, mockTuyaExecutor, mockTuyaAuth, mSkill)
		mSkill.On("Execute", mock.Anything).Return(&skills.SkillResult{
			Message: expectedResponse,
		}, nil).Once()

		res, err := localUC.ProcessControl(uid, terminalID, prompt)
		assert.NoError(t, err)
		assert.Contains(t, res.Message, "I found 2 matching devices")
		assert.Contains(t, res.Message, "AC Kamar Utama")
		assert.Contains(t, res.Message, "AC Ruang Tamu")
		assert.Empty(t, res.DeviceID)
	})

	t.Run("No Direct Match - Reconcile via LLM (Context)", func(t *testing.T) {
		// Use a prompt that won't trigger the 2-char vector match (AC, TV, etc.)
		prompt := "The main one"
		localMock := new(mockLLMForControl)
		mockAuth := new(mockTuyaAuthForControl)
		mockExec := new(mockTuyaExecutorForControl)
		mockAuth.On("GetTuyaAccessToken").Return("mock-token", nil)
		mockExec.On("SendIRACCommand", "mock-token", "dev-ac-1", "remote-ac-1", mock.Anything).Return(true, nil)
		mSkill := new(mockSkill)
		localUC := NewControlUseCase(localMock, nil, cfg, vector, badger, mockExec, mockAuth, mSkill)

		mSkill.On("Execute", mock.Anything).Return(&skills.SkillResult{
			Message:        "Control AC Kamar Utama successful",
			Data:           map[string]interface{}{"device_id": "dev-ac-1"},
			HTTPStatusCode: 200,
		}, nil).Once()

		res, err := localUC.ProcessControl(uid, terminalID, prompt)
		assert.NoError(t, err)
		// AC commands use IR API which always sets mode/temp/wind defaults, so message contains specific settings
		assert.Contains(t, res.Message, "AC Kamar Utama")
		assert.Equal(t, "dev-ac-1", res.DeviceID)
		localMock.AssertExpectations(t)
	})

	t.Run("Quote-wrapped LLM response (regression)", func(t *testing.T) {
		// Regression: OpenAI sometimes returns `"EXECUTE:id"` with surrounding quotes.
		// This must be treated the same as EXECUTE:id without quotes.
		prompt := "The main one"
		localMock := new(mockLLMForControl)
		mockAuth := new(mockTuyaAuthForControl)
		mockExec := new(mockTuyaExecutorForControl)
		mockAuth.On("GetTuyaAccessToken").Return("mock-token", nil)
		mockExec.On("SendIRACCommand", "mock-token", "dev-ac-1", "remote-ac-1", mock.Anything).Return(true, nil)
		mockExec.On("SendIRACCommand", "mock-token", "dev-ac-1", "remote-ac-1", mock.Anything).Return(true, nil)
		mSkill := new(mockSkill)
		localUC := NewControlUseCase(localMock, nil, cfg, vector, badger, mockExec, mockAuth, mSkill)

		mSkill.On("Execute", mock.Anything).Return(&skills.SkillResult{
			Message:        "Control AC Kamar Utama successful",
			Data:           map[string]interface{}{"device_id": "dev-ac-1"},
			HTTPStatusCode: 200,
		}, nil).Once()

		res, err := localUC.ProcessControl(uid, terminalID, prompt)
		assert.NoError(t, err)
		assert.Contains(t, res.Message, "AC Kamar Utama")
		assert.Equal(t, "dev-ac-1", res.DeviceID)
		localMock.AssertExpectations(t)
	})

	t.Run("No Match - Not Found", func(t *testing.T) {
		prompt := "Turn on the Fridge"
		localMock := new(mockLLMForControl)
		mSkill := new(mockSkill)
		localUC := NewControlUseCase(localMock, nil, cfg, vector, badger, nil, nil, mSkill)

		expectedResponse := "I'm sorry, I couldn't find any device matching 'Turn on the Fridge'."
		mSkill.On("Execute", mock.Anything).Return(&skills.SkillResult{
			Message: expectedResponse,
		}, nil).Once()

		res, err := localUC.ProcessControl(uid, terminalID, prompt)
		assert.NoError(t, err)
		assert.Contains(t, res.Message, "I'm sorry, I couldn't find any device matching")
		assert.Empty(t, res.DeviceID)
		localMock.AssertExpectations(t)
	})
}
