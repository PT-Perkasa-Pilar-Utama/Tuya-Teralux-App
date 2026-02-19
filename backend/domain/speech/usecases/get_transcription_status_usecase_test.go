package usecases_test

import (
	"fmt"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"
	"testing"
)

type MockWhisperProxyUsecase struct {
	ProxyTranscribeFunc   func(filePath string, fileName string, language string) (string, error)
	GetStatusFunc         func(taskID string) (*dtos.WhisperProxyStatusDTO, error)
	HealthCheckFunc       func() error
	FetchToOutsystemsFunc func(filePath string, fileName string, language string) (*dtos.OutsystemsTranscriptionResultDTO, error)
	TranscribeFunc        func(audioPath string, language string) (*dtos.WhisperResult, error)
}

func (m *MockWhisperProxyUsecase) ProxyTranscribe(filePath string, fileName string, language string) (string, error) {
	if m.ProxyTranscribeFunc != nil {
		return m.ProxyTranscribeFunc(filePath, fileName, language)
	}
	return "", nil
}

func (m *MockWhisperProxyUsecase) GetStatus(taskID string) (*dtos.WhisperProxyStatusDTO, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(taskID)
	}
	return nil, fmt.Errorf("task not found")
}

func (m *MockWhisperProxyUsecase) HealthCheck() error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return nil
}

func (m *MockWhisperProxyUsecase) FetchToOutsystems(filePath string, fileName string, language string) (*dtos.OutsystemsTranscriptionResultDTO, error) {
	if m.FetchToOutsystemsFunc != nil {
		return m.FetchToOutsystemsFunc(filePath, fileName, language)
	}
	return nil, nil
}

func (m *MockWhisperProxyUsecase) Transcribe(audioPath string, language string) (*dtos.WhisperResult, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(audioPath, language)
	}
	return nil, nil // Default return
}

func TestGetTranscriptionStatusUseCase_Execute(t *testing.T) {
	t.Run("Task Found in Whisper Proxy", func(t *testing.T) {
		mockProxy := &MockWhisperProxyUsecase{
			GetStatusFunc: func(taskID string) (*dtos.WhisperProxyStatusDTO, error) {
				return &dtos.WhisperProxyStatusDTO{
					Status: "completed",
					Result: &dtos.OutsystemsTranscriptionResultDTO{
						Transcription: "Hello world",
					},
				}, nil
			},
		}

		uc := usecases.NewGetTranscriptionStatusUseCase(nil, nil, mockProxy)
		resp, err := uc.GetTranscriptionStatus("task-123")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		proxyResp, ok := resp.(dtos.WhisperProxyProcessStatusResponseDTO)
		if !ok {
			t.Fatalf("Expected WhisperProxyProcessStatusResponseDTO, got %T", resp)
		}

		if proxyResp.TaskID != "task-123" {
			t.Errorf("Expected TaskID task-123, got %s", proxyResp.TaskID)
		}

		if proxyResp.TaskStatus.Status != "completed" {
			t.Errorf("Expected status completed, got %s", proxyResp.TaskStatus.Status)
		}

		if proxyResp.TaskStatus.Result.Transcription != "Hello world" {
			t.Errorf("Expected transcription 'Hello world', got %s", proxyResp.TaskStatus.Result.Transcription)
		}
	})

	t.Run("Task Not Found", func(t *testing.T) {
		mockProxy := &MockWhisperProxyUsecase{
			GetStatusFunc: func(taskID string) (*dtos.WhisperProxyStatusDTO, error) {
				return nil, fmt.Errorf("not found")
			},
		}

		uc := usecases.NewGetTranscriptionStatusUseCase(nil, nil, mockProxy)
		_, err := uc.GetTranscriptionStatus("unknown-task")

		if err == nil {
			t.Error("Expected error for unknown task, got nil")
		}
	})
}
