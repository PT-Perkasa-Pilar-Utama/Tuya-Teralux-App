package usecases_test

import (
	"encoding/json"
	"errors"
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

func TestTranscribeGeminiModelUseCase(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "gemini_test_*")
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
			Transcription:    "Halo dunia",
			DetectedLanguage: "id",
			Source:           "Gemini",
		}, nil)

		mockStore.On("Set", mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.AnythingOfType("string"), mock.Anything).Return(nil)
		
		dummyStatus := &dtos.AsyncTranscriptionStatusDTO{StartedAt: time.Now().Add(-5 * time.Second).Format(time.RFC3339)}
		dummyBytes, _ := json.Marshal(dummyStatus)
		mockStore.On("GetWithTTL", mock.AnythingOfType("string")).Return(dummyBytes, 1*time.Hour, nil).Maybe()

		uc := usecases.NewTranscribeGeminiModelUseCase(mockSvc, store, cache, &utils.Config{})
		taskID, err := uc.TranscribeAsync(audioFile, "test.wav", "id")

		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		time.Sleep(50 * time.Millisecond)

		status, ok := store.Get(taskID)
		assert.True(t, ok)
		assert.Equal(t, "completed", status.Status)
		assert.Equal(t, "Halo dunia", status.Result.Transcription)
		assert.NotEmpty(t, status.StartedAt)
		assert.True(t, status.DurationSeconds > 0)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Scenario 2: Validation: Missing Audio File", func(t *testing.T) {
		uc := usecases.NewTranscribeGeminiModelUseCase(nil, nil, nil, nil)
		_, err := uc.TranscribeAsync("non-existent.wav", "none", "id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("Scenario 3: Error: Gemini Service Health Check Failed", func(t *testing.T) {
		mockSvc := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")
		store := tasks.NewStatusStore[dtos.AsyncTranscriptionStatusDTO]()

		mockSvc.On("HealthCheck").Return(false)

		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		uc := usecases.NewTranscribeGeminiModelUseCase(mockSvc, store, cache, &utils.Config{})
		taskID, _ := uc.TranscribeAsync(audioFile, "test.wav", "id")

		time.Sleep(50 * time.Millisecond)
		status, _ := store.Get(taskID)
		assert.Equal(t, "failed", status.Status)
	})

	t.Run("Scenario 7: Validation: Wrong Extension / Corrupt Header", func(t *testing.T) {
		mockSvc := new(speechUtils.GenericMockClient)
		mockStore := new(speechUtils.MockBadgerStore)
		cache := tasks.NewBadgerTaskCache(mockStore, "task:")
		store := tasks.NewStatusStore[dtos.AsyncTranscriptionStatusDTO]()

		mockSvc.On("HealthCheck").Return(true)
		mockSvc.On("Transcribe", audioFile, "id").Return(nil, errors.New("decoding failed"))

		mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("SetPreserveTTL", mock.Anything, mock.Anything).Return(nil)
		mockStore.On("GetWithTTL", mock.Anything).Return(nil, 0*time.Second, nil).Maybe()

		uc := usecases.NewTranscribeGeminiModelUseCase(mockSvc, store, cache, &utils.Config{})
		taskID, _ := uc.TranscribeAsync(audioFile, "test.wav", "id")

		time.Sleep(50 * time.Millisecond)
		status, _ := store.Get(taskID)
		assert.Equal(t, "failed", status.Status)
	})
}
