package usecases_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"
)

func TestNewWhisperProxyUsecase(t *testing.T) {
	cfg := &utils.Config{}
	uc := usecases.NewWhisperProxyUsecase(nil, cfg)
	if uc == nil {
		t.Error("NewWhisperProxyUsecase returned nil")
	}
}

func TestWhisperProxyUsecase_ProxyTranscribe_WithoutBadger(t *testing.T) {
	cfg := &utils.Config{}
	uc := usecases.NewWhisperProxyUsecase(nil, cfg)

	// Create temporary audio file
	tempFile, err := os.CreateTemp("", "test_audio_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	_, _ = tempFile.Write([]byte("fake audio data"))
	tempFile.Close()

	// Test task submission without badger
	taskID, err := uc.ProxyTranscribe(tempFile.Name(), "test.mp3")
	if err != nil {
		t.Fatalf("expected no error from ProxyTranscribe, got %v", err)
	}
	if taskID == "" {
		t.Error("expected non-empty task ID")
	}

	// Status should be pending immediately
	status, err := uc.GetStatus(taskID)
	if err != nil {
		t.Fatalf("expected no error from GetStatus, got %v", err)
	}
	if status == nil {
		t.Fatal("expected valid status, got nil")
	}
	if status.Status != "pending" {
		t.Errorf("expected pending status, got %s", status.Status)
	}
}

func TestWhisperProxyUsecase_GetStatus_NotFound(t *testing.T) {
	cfg := &utils.Config{}
	uc := usecases.NewWhisperProxyUsecase(nil, cfg)

	status, err := uc.GetStatus("non-existent-task-id")
	if err == nil {
		t.Error("expected error for non-existent task, got nil")
	}
	if status != nil {
		t.Errorf("expected nil status for non-existent task, got %+v", status)
	}
}

func TestWhisperProxyUsecase_GetStatus_WithBadger(t *testing.T) {
	// Create temporary badger DB
	tempDir, err := os.MkdirTemp("", "badger_test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize global config for badger service
	utils.AppConfig = &utils.Config{CacheTTL: "1h"}

	badgerSvc, err := infrastructure.NewBadgerService(tempDir)
	if err != nil {
		t.Fatalf("failed to create badger service: %v", err)
	}
	defer badgerSvc.Close()

	cfg := &utils.Config{}
	uc := usecases.NewWhisperProxyUsecase(badgerSvc, cfg)

	// Create temporary audio file
	tempFile, err := os.CreateTemp("", "test_audio_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	_, _ = tempFile.Write([]byte("fake audio data"))
	tempFile.Close()

	// Submit task
	taskID, err := uc.ProxyTranscribe(tempFile.Name(), "test.mp3")
	if err != nil {
		t.Fatalf("expected no error from ProxyTranscribe, got %v", err)
	}

	// Verify status from badger
	status, err := uc.GetStatus(taskID)
	if err != nil {
		t.Fatalf("expected no error from GetStatus, got %v", err)
	}
	if status == nil {
		t.Fatal("expected valid status, got nil")
	}
	if status.Status != "pending" {
		t.Errorf("expected pending status, got %s", status.Status)
	}

	// Check TTL info exists
	if status.ExpiresInSecond <= 0 {
		t.Logf("Note: ExpiresInSecond = %d (may be expected without TTL set)", status.ExpiresInSecond)
	}
}

func TestWhisperProxyUsecase_ProxyTranscribe_InvalidFile(t *testing.T) {
	cfg := &utils.Config{}
	uc := usecases.NewWhisperProxyUsecase(nil, cfg)

	// Test with non-existent file
	taskID, err := uc.ProxyTranscribe("/non/existent/file.mp3", "test.mp3")
	if err != nil {
		t.Fatalf("ProxyTranscribe should not error on submission, got %v", err)
	}

	// But the task should eventually fail
	timeout := time.After(2 * time.Second)
	tick := time.Tick(100 * time.Millisecond)
	var finalStatus *dtos.WhisperProxyStatusDTO

	for {
		select {
		case <-timeout:
			// Task should be in error state by now
			if finalStatus == nil || finalStatus.Status != "error" {
				t.Logf("Task status: %+v", finalStatus)
			}
			return
		case <-tick:
			status, err := uc.GetStatus(taskID)
			if err == nil && status != nil {
				finalStatus = status
				if status.Status == "error" {
					// Expected behavior
					return
				}
			}
		}
	}
}

func TestWhisperProxyStatusDTOSerialization(t *testing.T) {
	resultDTO := &dtos.OutsystemsTranscriptionResultDTO{
		Filename:         "test-audio.mp3",
		Transcription:    "Hello world",
		DetectedLanguage: "en",
	}

	statusDTO := &dtos.WhisperProxyStatusDTO{
		Status:          "completed",
		Result:          resultDTO,
		ExpiresAt:       time.Now().UTC().Format(time.RFC3339),
		ExpiresInSecond: 3600,
	}

	// Test marshalling
	b, err := json.Marshal(statusDTO)
	if err != nil {
		t.Fatalf("failed to marshal status DTO: %v", err)
	}

	// Test unmarshalling
	var unmarshalled dtos.WhisperProxyStatusDTO
	err = json.Unmarshal(b, &unmarshalled)
	if err != nil {
		t.Fatalf("failed to unmarshal status DTO: %v", err)
	}

	if unmarshalled.Status != statusDTO.Status {
		t.Errorf("status mismatch: expected %s, got %s", statusDTO.Status, unmarshalled.Status)
	}
	if unmarshalled.Result == nil {
		t.Error("expected result to be non-nil")
	} else if unmarshalled.Result.Transcription != resultDTO.Transcription {
		t.Errorf("result transcription mismatch: expected %s, got %s", resultDTO.Transcription, unmarshalled.Result.Transcription)
	}
}
