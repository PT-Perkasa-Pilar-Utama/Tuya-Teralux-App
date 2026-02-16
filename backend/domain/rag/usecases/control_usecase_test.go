package usecases

import (
	"encoding/json"
	"fmt"
	"os"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMForControl is a mock for LLMClient
type mockLLMForControl struct {
	mock.Mock
}

func (m *mockLLMForControl) CallModel(prompt string, model string) (string, error) {
	args := m.Called(prompt, model)
	return args.String(0), args.Error(1)
}

func TestControlUseCase_ProcessControl(t *testing.T) {
	mockLLM := new(mockLLMForControl)
	cfg := &utils.Config{LLMModel: "test-model"}
	
	// Setup Vector Service
	vectorFile := "test_vector_control.json"
	defer fmt.Println("Cleaning up test vector file")
	defer func() { _ = os.Remove(vectorFile) }()
	vector := infrastructure.NewVectorService(vectorFile)
	
	// Setup Badger
	badger, _ := infrastructure.NewBadgerService("") // In-memory

	uc := NewControlUseCase(mockLLM, cfg, vector, badger)

	uid := "user-123"
	teraluxID := "tx-1"

	// 1. Setup Mock User Devices in Vector DB
	devices := []tuyaDtos.TuyaDeviceDTO{
		{ID: "dev-ac-1", Name: "AC Kamar Utama"},
		{ID: "dev-ac-2", Name: "AC Ruang Tamu"},
		{ID: "dev-lamp-1", Name: "Lampu Teras"},
	}
	resp := tuyaDtos.TuyaDevicesResponseDTO{Devices: devices}
	respJSON, _ := json.Marshal(resp)
	_ = vector.Upsert(fmt.Sprintf("tuya:devices:uid:%s", uid), string(respJSON), nil)
	
	for _, d := range devices {
		dJSON, _ := json.Marshal(d)
		_ = vector.Upsert("tuya:device:"+d.ID, string(dJSON), nil)
	}

	t.Run("Single Match", func(t *testing.T) {
		res, err := uc.ProcessControl(uid, teraluxID, "Nyalakan Lampu Teras")
		assert.NoError(t, err)
		assert.Contains(t, res, "Running command for **Lampu Teras**")
	})

	t.Run("Multiple Matches (Ambiguity)", func(t *testing.T) {
		res, err := uc.ProcessControl(uid, teraluxID, "Matikan AC")
		assert.NoError(t, err)
		assert.Contains(t, res, "I found 2 matching devices")
		assert.Contains(t, res, "AC Kamar Utama")
		assert.Contains(t, res, "AC Ruang Tamu")
	})

	t.Run("No Direct Match - Reconcile via LLM (Context)", func(t *testing.T) {
		// Use a prompt that won't trigger the 2-char vector match (AC, TV, etc.)
		prompt := "The main one"
		localMock := new(mockLLMForControl)
		localUC := NewControlUseCase(localMock, cfg, vector, badger)

		localMock.On("CallModel", mock.Anything, "test-model").Return("EXECUTE:dev-ac-1", nil).Once()

		res, err := localUC.ProcessControl(uid, teraluxID, prompt)
		assert.NoError(t, err)
		assert.Contains(t, res, "Running command for **AC Kamar Utama**")
		localMock.AssertExpectations(t)
	})

	t.Run("No Match - Not Found", func(t *testing.T) {
		prompt := "Turn on the Fridge"
		localMock := new(mockLLMForControl)
		localUC := NewControlUseCase(localMock, cfg, vector, badger)

		localMock.On("CallModel", mock.Anything, "test-model").Return("NOT_FOUND", nil).Once()

		res, err := localUC.ProcessControl(uid, teraluxID, prompt)
		assert.NoError(t, err)
		assert.Contains(t, res, "I'm sorry, I couldn't find any device matching")
		localMock.AssertExpectations(t)
	})
}
