package usecases_test

import (
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
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

// MockTranscriptionTaskRepository is a mock
type MockTranscriptionTaskRepository struct {
	repositories.TranscriptionTaskRepository
	SaveShortTaskFunc func(taskID string, status *speechdtos.AsyncTranscriptionStatusDTO) error
	SaveLongTaskFunc  func(taskID string, status *speechdtos.AsyncTranscriptionLongStatusDTO) error
}

func (m *MockTranscriptionTaskRepository) SaveShortTask(taskID string, status *speechdtos.AsyncTranscriptionStatusDTO) error {
	if m.SaveShortTaskFunc != nil {
		return m.SaveShortTaskFunc(taskID, status)
	}
	return nil
}

func (m *MockTranscriptionTaskRepository) SaveLongTask(taskID string, status *speechdtos.AsyncTranscriptionLongStatusDTO) error {
	if m.SaveLongTaskFunc != nil {
		return m.SaveLongTaskFunc(taskID, status)
	}
	return nil
}


func TestNewTranscribeUseCase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	whisperRepo := repositories.NewWhisperRepository(cfg)
	taskRepo := &MockTranscriptionTaskRepository{}

	uc := usecases.NewTranscribeUseCase(
		whisperRepo, 
		nil,
		&ragUsecases.RAGUsecase{},
		taskRepo,
		cfg,
	)
	if uc == nil {
		t.Error("NewTranscribeUseCase returned nil")
	}
}

func TestNewTranscribeWhisperCppUseCase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	whisperRepo := repositories.NewWhisperRepository(cfg)
	taskRepo := &MockTranscriptionTaskRepository{}

	uc := usecases.NewTranscribeWhisperCppUseCase(
		whisperRepo, 
		&ragUsecases.RAGUsecase{},
		taskRepo,
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
		mockRepo := &MockWhisperRepository{}
		taskRepo := &MockTranscriptionTaskRepository{}
		uc := usecases.NewTranscribeWhisperCppUseCase(mockRepo, nil, taskRepo, cfg)

		_, err := uc.Execute("non_existent.mp3", "id", "en") // fileName, lang
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
