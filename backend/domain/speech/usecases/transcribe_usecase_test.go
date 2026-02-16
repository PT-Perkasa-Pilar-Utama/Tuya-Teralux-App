package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/usecases"
	"teralux_app/domain/speech/utilities"
	"testing"
)

// MockWhisperClient is a mock implementation of utilities.WhisperClient
type MockWhisperClient struct {
	TranscribeFunc  func(audioPath string, language string) (*utilities.WhisperResult, error)
	HealthCheckFunc func() bool
}

func (m *MockWhisperClient) Transcribe(audioPath string, language string) (*utilities.WhisperResult, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(audioPath, language)
	}
	return &utilities.WhisperResult{
		Transcription:    "test transcription",
		DetectedLanguage: language,
		Source:           "Mock",
	}, nil
}

func (m *MockWhisperClient) HealthCheck() bool {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return true
}

// MockWhisperCppRepository is a mock for local whisper cpp repository
type MockWhisperCppRepository struct {
	TranscribeFullFunc func(wavPath string, modelPath string, lang string) (string, error)
}

func (m *MockWhisperCppRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
	if m.TranscribeFullFunc != nil {
		return m.TranscribeFullFunc(wavPath, modelPath, lang)
	}
	return "mock transcription from local", nil
}

// MockRefineUseCase is a mock implementation of RefineUseCase interface
type MockRefineUseCase struct {
	RefineTextFunc func(text string, lang string) (string, error)
}

func (m *MockRefineUseCase) RefineText(text string, lang string) (string, error) {
	if m.RefineTextFunc != nil {
		return m.RefineTextFunc(text, lang)
	}
	return text, nil
}

func TestNewTranscribeUseCase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	mockClient := &MockWhisperClient{}

	uc := usecases.NewTranscribeUseCase(
		mockClient,
		&MockRefineUseCase{},
		nil,
		cfg,
		nil,
	)
	if uc == nil {
		t.Error("NewTranscribeUseCase returned nil")
	}
}

func TestNewTranscribeWhisperCppUseCase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	mockRepo := &MockWhisperCppRepository{}

	uc := usecases.NewTranscribeWhisperCppUseCase(
		mockRepo,
		&MockRefineUseCase{},
		nil,
		cfg,
	)
	if uc == nil {
		t.Error("NewTranscribeWhisperCppUseCase returned nil")
	}
}

func TestTranscribeWhisperCppUseCase_Execute(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}

	t.Run("File Not Found", func(t *testing.T) {
		mockRepo := &MockWhisperCppRepository{}
		uc := usecases.NewTranscribeWhisperCppUseCase(mockRepo, nil, nil, cfg)

		_, err := uc.TranscribeWhisperCpp("non_existent.mp3", "test.mp3", "en")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("Task Creation Error", func(t *testing.T) {
		// Needs a real file to bypass os.Stat check
		// We can use a dummy file
		// But TranscribeWhisperCppUseCase.Execute triggers a goroutine.
		// Testing async logic here is tricky without waitgroups in the usecase.
		// We can just test that Execute returns a TaskID if file exists.

		// Skip for now or create a temp file
	})
}
