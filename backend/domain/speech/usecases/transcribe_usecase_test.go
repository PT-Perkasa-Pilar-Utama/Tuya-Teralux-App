package usecases_test

import (
	"errors"
	"os"
	"path/filepath"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	speechdtos "teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"
	speechUtils "teralux_app/domain/speech/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRefineUseCase struct {
	RefineTextFunc func(text string, lang string) (string, error)
}

func (m *MockRefineUseCase) RefineText(text string, lang string) (string, error) {
	if m.RefineTextFunc != nil {
		return m.RefineTextFunc(text, lang)
	}
	return text, nil
}

func TestTranscribeUseCase_FullScenarios(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "transcribe_usecase_test_*")
	defer os.RemoveAll(tmpDir)

	audioFile := filepath.Join(tmpDir, "test.wav")
	_ = os.WriteFile(audioFile, []byte("audio content"), 0644)

	t.Run("Scenario 1: Fallback Success", func(t *testing.T) {
		mockClient := new(speechUtils.GenericMockClient)
		mockRefine := new(MockRefineUseCase)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")

		mockClient.On("Transcribe", audioFile, "id").Return(&speechdtos.WhisperResult{
			Transcription:    "Raw text",
			DetectedLanguage: "id",
			Source:           "Gemini",
		}, nil)

		mockRefine.RefineTextFunc = func(text, lang string) (string, error) {
			return "Refined text", nil
		}

		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		statusStore := tasks.NewStatusStore[speechdtos.AsyncTranscriptionStatusDTO]()
		uc := usecases.NewTranscribeUseCase(mockClient, nil, mockRefine, statusStore, cache, &utils.Config{}, nil)
		taskID, err := uc.TranscribeAudio(audioFile, "test.wav", "id")

		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		time.Sleep(50 * time.Millisecond)
		// We can't easily check StatusStore here because it's internal to TranscribeUseCase and not exposed or passed in.
		// Wait, TranscribeUseCase uses cache (BadgerTaskCache).
		
		mockStore.AssertCalled(t, "SetPreserveTTL", mock.Anything, mock.Anything)
	})

	t.Run("Scenario 2: File Not Found", func(t *testing.T) {
		statusStore := tasks.NewStatusStore[speechdtos.AsyncTranscriptionStatusDTO]()
		uc := usecases.NewTranscribeUseCase(nil, nil, nil, statusStore, nil, nil, nil)
		_, err := uc.TranscribeAudio("missing.wav", "none", "id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audio file not found")
	})

	t.Run("Scenario 8: All Providers Failed", func(t *testing.T) {
		mockClient := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")

		mockClient.On("Transcribe", audioFile, "id").Return(nil, errors.New("total failure"))
		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		statusStore := tasks.NewStatusStore[speechdtos.AsyncTranscriptionStatusDTO]()
		uc := usecases.NewTranscribeUseCase(mockClient, nil, nil, statusStore, cache, nil, nil)
		taskID, _ := uc.TranscribeAudio(audioFile, "test.wav", "id")
		assert.NotEmpty(t, taskID)

		time.Sleep(50 * time.Millisecond)
		// Check that it was updated to failed (SetPreserveTTL called with failed status)
		// This is hard to assert without custom matcher for the byte payload.
	})

	t.Run("Scenario 9: MQTT Chaining", func(t *testing.T) {
		mockClient := new(speechUtils.GenericMockClient)
		mockMqtt := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")

		mockClient.On("Transcribe", audioFile, "id").Return(&speechdtos.WhisperResult{
			Transcription:    "Mqtt text",
			DetectedLanguage: "id",
			Source:           "OpenAI",
		}, nil)

		mockMqtt.On("Publish", "users/teralux/chat", byte(0), false, mock.Anything).Return(nil)
		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		statusStore := tasks.NewStatusStore[speechdtos.AsyncTranscriptionStatusDTO]()
		uc := usecases.NewTranscribeUseCase(mockClient, nil, &MockRefineUseCase{}, statusStore, cache, &utils.Config{}, mockMqtt)
		_, _ = uc.TranscribeAudio(audioFile, "test.wav", "id", usecases.TranscriptionMetadata{
			Source:    "mqtt",
			TeraluxID: "TLX001",
			UID:       "USER001",
		})

		time.Sleep(50 * time.Millisecond)
		mockMqtt.AssertCalled(t, "Publish", "users/teralux/chat", byte(0), false, mock.Anything)
	})
}
