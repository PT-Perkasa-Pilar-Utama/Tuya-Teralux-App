package usecases_test

import (
	"os"
	"path/filepath"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"
	speechUtils "teralux_app/domain/speech/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTranscribeGroqModelUseCase(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "groq_test_*")
	defer os.RemoveAll(tmpDir)

	audioFile := filepath.Join(tmpDir, "test.wav")
	_ = os.WriteFile(audioFile, []byte("audio content"), 0644)

	t.Run("Scenario 1: Success", func(t *testing.T) {
		mockSvc := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")
		store := tasks.NewStatusStore[dtos.AsyncTranscriptionStatusDTO]()

		mockSvc.On("HealthCheck").Return(true)
		mockSvc.On("Transcribe", audioFile, "id").Return(&dtos.WhisperResult{
			Transcription:    "Groq result",
			DetectedLanguage: "id",
			Source:           "Groq",
		}, nil)

		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		uc := usecases.NewTranscribeGroqModelUseCase(mockSvc, store, cache, &utils.Config{})
		taskID, err := uc.TranscribeAsync(audioFile, "test.wav", "id")

		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		time.Sleep(50 * time.Millisecond)

		status, _ := store.Get(taskID)
		assert.Equal(t, "completed", status.Status)
		assert.Equal(t, "Groq result", status.Result.Transcription)
	})

	t.Run("Scenario 8: Health Check Failed", func(t *testing.T) {
		mockSvc := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")
		store := tasks.NewStatusStore[dtos.AsyncTranscriptionStatusDTO]()

		mockSvc.On("HealthCheck").Return(false)
		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		uc := usecases.NewTranscribeGroqModelUseCase(mockSvc, store, cache, &utils.Config{})
		taskID, _ := uc.TranscribeAsync(audioFile, "test.wav", "id")

		time.Sleep(50 * time.Millisecond)
		status, _ := store.Get(taskID)
		assert.Equal(t, "failed", status.Status)
	})
}
