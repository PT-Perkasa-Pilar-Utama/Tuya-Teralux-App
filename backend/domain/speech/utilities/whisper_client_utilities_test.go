package utilities

import (
	"errors"
	speechdtos "teralux_app/domain/speech/dtos"
	"testing"
)

// MockWhisperClient for testing
type MockWhisperClient struct {
	TranscribeFunc  func(audioPath string, language string) (*WhisperResult, error)
	HealthCheckFunc func() bool
}

func (m *MockWhisperClient) Transcribe(audioPath string, language string) (*WhisperResult, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(audioPath, language)
	}
	return &WhisperResult{
		Transcription:    "mock transcription",
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

func TestNewWhisperClientWithFallback(t *testing.T) {
	primary := &MockWhisperClient{}
	secondary := &MockWhisperClient{}
	tertiary := &MockWhisperClient{}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	fallback, ok := client.(*WhisperClientWithFallback)
	if !ok {
		t.Fatal("expected *WhisperClientWithFallback type")
	}

	if fallback.primary != primary {
		t.Error("primary client not set correctly")
	}
	if fallback.secondary != secondary {
		t.Error("secondary client not set correctly")
	}
	if fallback.tertiary != tertiary {
		t.Error("tertiary client not set correctly")
	}
}

func TestWhisperClientFallback_PrimarySucceeds(t *testing.T) {
	primary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return &WhisperResult{
				Transcription:    "primary transcription",
				DetectedLanguage: language,
				Source:           "PPU",
			}, nil
		},
		HealthCheckFunc: func() bool { return true },
	}
	secondary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			t.Error("secondary should not be called")
			return nil, errors.New("should not reach here")
		},
	}
	tertiary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			t.Error("tertiary should not be called")
			return nil, errors.New("should not reach here")
		},
	}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)
	result, err := client.Transcribe("test.wav", "en")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transcription != "primary transcription" {
		t.Errorf("expected 'primary transcription', got '%s'", result.Transcription)
	}
	if result.Source != "PPU" {
		t.Errorf("expected source 'PPU', got '%s'", result.Source)
	}
}

func TestWhisperClientFallback_PrimaryFailsSecondarySucceeds(t *testing.T) {
	primary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("primary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}
	secondary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return &WhisperResult{
				Transcription:    "secondary transcription",
				DetectedLanguage: language,
				Source:           "Orion",
			}, nil
		},
		HealthCheckFunc: func() bool { return true },
	}
	tertiary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			t.Error("tertiary should not be called")
			return nil, errors.New("should not reach here")
		},
	}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)
	result, err := client.Transcribe("test.wav", "en")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transcription != "secondary transcription" {
		t.Errorf("expected 'secondary transcription', got '%s'", result.Transcription)
	}
	if result.Source != "Orion" {
		t.Errorf("expected source 'Orion', got '%s'", result.Source)
	}
}

func TestWhisperClientFallback_PrimaryUnhealthySecondarySucceeds(t *testing.T) {
	primary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			t.Error("primary should not be called when unhealthy")
			return nil, errors.New("should not reach here")
		},
		HealthCheckFunc: func() bool { return false },
	}
	secondary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return &WhisperResult{
				Transcription:    "secondary transcription",
				DetectedLanguage: language,
				Source:           "Orion",
			}, nil
		},
		HealthCheckFunc: func() bool { return true },
	}
	tertiary := &MockWhisperClient{}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)
	result, err := client.Transcribe("test.wav", "en")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transcription != "secondary transcription" {
		t.Errorf("expected 'secondary transcription', got '%s'", result.Transcription)
	}
}

func TestWhisperClientFallback_AllFailTertiarySucceeds(t *testing.T) {
	primary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("primary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}
	secondary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("secondary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}
	tertiary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return &WhisperResult{
				Transcription:    "tertiary transcription",
				DetectedLanguage: language,
				Source:           "Local",
			}, nil
		},
		HealthCheckFunc: func() bool { return true },
	}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)
	result, err := client.Transcribe("test.wav", "en")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transcription != "tertiary transcription" {
		t.Errorf("expected 'tertiary transcription', got '%s'", result.Transcription)
	}
	if result.Source != "Local" {
		t.Errorf("expected source 'Local', got '%s'", result.Source)
	}
}

func TestWhisperClientFallback_AllFail(t *testing.T) {
	primary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("primary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}
	secondary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("secondary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}
	tertiary := &MockWhisperClient{
		TranscribeFunc: func(audioPath, language string) (*WhisperResult, error) {
			return nil, errors.New("tertiary failed")
		},
		HealthCheckFunc: func() bool { return true },
	}

	client := NewWhisperClientWithFallback(primary, secondary, tertiary)
	_, err := client.Transcribe("test.wav", "en")

	if err == nil {
		t.Fatal("expected error when all clients fail")
	}
	if err.Error() != "tertiary failed" {
		t.Errorf("expected last error 'tertiary failed', got '%v'", err)
	}
}

func TestWhisperClientFallback_NoClients(t *testing.T) {
	client := NewWhisperClientWithFallback(nil, nil, nil)
	_, err := client.Transcribe("test.wav", "en")

	if err == nil {
		t.Fatal("expected error when no clients available")
	}
}

// Test adapters
func TestPPUWhisperClient_Transcribe(t *testing.T) {
	mockUsecase := &MockWhisperProxyUsecase{
		FetchToOutsystemsFunc: func(filePath, fileName, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error) {
			return &speechdtos.OutsystemsTranscriptionResultDTO{
				Filename:         fileName,
				Transcription:    "PPU transcription",
				DetectedLanguage: language,
			}, nil
		},
	}

	client := NewPPUWhisperClient(mockUsecase)
	result, err := client.Transcribe("test.wav", "en")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transcription != "PPU transcription" {
		t.Errorf("expected 'PPU transcription', got '%s'", result.Transcription)
	}
	if result.Source != "PPU (Outsystems)" {
		t.Errorf("expected source 'PPU (Outsystems)', got '%s'", result.Source)
	}
}

func TestPPUWhisperClient_HealthCheck(t *testing.T) {
	mockUsecase := &MockWhisperProxyUsecase{
		HealthCheckFunc: func() error { return nil },
	}

	client := NewPPUWhisperClient(mockUsecase)
	if ppu, ok := client.(*PPUWhisperClient); ok {
		if !ppu.HealthCheck() {
			t.Error("expected health check to return true")
		}
	}
}

// Mock for testing adapters
type MockWhisperProxyUsecase struct {
	FetchToOutsystemsFunc func(filePath, fileName, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error)
	HealthCheckFunc       func() error
}

func (m *MockWhisperProxyUsecase) FetchToOutsystems(filePath, fileName, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error) {
	if m.FetchToOutsystemsFunc != nil {
		return m.FetchToOutsystemsFunc(filePath, fileName, language)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWhisperProxyUsecase) HealthCheck() error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return nil
}
