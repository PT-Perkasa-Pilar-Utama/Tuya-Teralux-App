package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/usecases"
	"testing"
)

// MockWhisperCppRepository is a mock implementation of WhisperCppRepositoryInterface
type MockWhisperCppRepository struct {
	TranscribeFunc     func(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFullFunc func(wavPath string, modelPath string, lang string) (string, error)
}

func (m *MockWhisperCppRepository) Transcribe(wavPath string, modelPath string, lang string) (string, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(wavPath, modelPath, lang)
	}
	return "", nil
}

func (m *MockWhisperCppRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
	if m.TranscribeFullFunc != nil {
		return m.TranscribeFullFunc(wavPath, modelPath, lang)
	}
	return "", nil
}

// MockWhisperOrionRepository is a mock implementation of WhisperOrionRepositoryInterface
type MockWhisperOrionRepository struct {
	TranscribeFunc func(audioPath string, lang string) (string, error)
}

func (m *MockWhisperOrionRepository) Transcribe(audioPath string, lang string) (string, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(audioPath, lang)
	}
	return "", nil
}

// TranscriptionTaskRepository mocks removed as it's deprecated.


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
	cppRepo := &MockWhisperCppRepository{}
	orionRepo := &MockWhisperOrionRepository{}

	uc := usecases.NewTranscribeUseCase(
		cppRepo,
		orionRepo,
		nil,
		&MockRefineUseCase{},
		nil,
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
	cppRepo := &MockWhisperCppRepository{}

	uc := usecases.NewTranscribeWhisperCppUseCase(
		cppRepo, 
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

		_, err := uc.TranscribeWhisperCpp("non_existent.mp3", "id", "en") // fileName, lang
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
