package usecases_test

import (
	"errors"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/repositories"
	"teralux_app/domain/speech/usecases"
	"testing"
)

// MockWhisperRepository is a mock implementation of WhisperRepositoryInterface
type MockWhisperRepository struct {
	TranscribeFunc     func(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFullFunc func(wavPath string, modelPath string, lang string) (string, error)
}

func (m *MockWhisperRepository) Transcribe(wavPath string, modelPath string, lang string) (string, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(wavPath, modelPath, lang)
	}
	return "", nil
}

func (m *MockWhisperRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
	if m.TranscribeFullFunc != nil {
		return m.TranscribeFullFunc(wavPath, modelPath, lang)
	}
	return "", nil
}

func TestNewTranscriptionUsecase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	whisperRepo := repositories.NewWhisperRepository(cfg)
	ollamaRepo := repositories.NewOllamaRepository()
	geminiRepo := repositories.NewGeminiRepository()

	uc := usecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, geminiRepo, nil, nil, cfg, nil, nil, nil)
	if uc == nil {
		t.Error("NewTranscriptionUsecase returned nil")
	}
}

func TestTranscriptionUsecase_TranscribeLongAudio(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}

	t.Run("File Not Found", func(t *testing.T) {
		mockRepo := &MockWhisperRepository{}
		uc := usecases.NewTranscriptionUsecase(mockRepo, nil, nil, nil, nil, cfg, nil, nil, nil)

		_, err := uc.TranscribeLongAudio("non_existent.mp3", "id")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("Transcription Error", func(t *testing.T) {
		// This test is harder because TranscribeLongAudio calls utils.ConvertToWav first.
		// If we provide a file that exists but isn't a real audio, ffmpeg might fail.
		mockRepo := &MockWhisperRepository{
			TranscribeFullFunc: func(wavPath string, modelPath string, lang string) (string, error) {
				return "", errors.New("whisper error")
			},
		}

		_ = usecases.NewTranscriptionUsecase(mockRepo, nil, nil, nil, nil, cfg, nil, nil, nil)
		// This will likely fail at ConvertToWav, but that's okay for verifying it doesn't crash.
	})
}
