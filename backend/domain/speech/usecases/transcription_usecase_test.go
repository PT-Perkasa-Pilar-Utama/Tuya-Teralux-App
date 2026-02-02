package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/repositories"
	"teralux_app/domain/speech/usecases"
	"testing"
)

// Mocking isn't strictly necessary if we test the repo directly,
// but since we don't want to actually connect to MQTT for unit tests,
// we should probably mock the repo or use an integration test style
// IF we had interfaces. Since we are using concrete structs, unit testing
// logic with external dependencies is hard without interfaces.
//
// For now, I will create a basic test that ensures structs are initialized correctly.
// A proper unit test would require refactoring to Repository Interfaces.

func TestNewTranscriptionUsecase(t *testing.T) {
	cfg := &utils.Config{
		WhisperModelPath: "test_model",
	}
	whisperRepo := repositories.NewWhisperRepository()
	ollamaRepo := repositories.NewOllamaRepository()
	// We refrain from creating a real MqttRepository here because it tries to connect on init.
	// This shows a design limitation (side effect in constructor).
	// For this test, I will pass nil for mqttRepo and verify it doesn't crash
	// unless we call a method using it.

	// Wait, we can't really test much without mocking.
	// I'll skip deep logic tests and just check instantiation.

	geminiRepo := repositories.NewGeminiRepository()

	uc := usecases.NewTranscriptionUsecase(whisperRepo, ollamaRepo, geminiRepo, nil, cfg, nil, nil)
	if uc == nil {
		t.Error("NewTranscriptionUsecase returned nil")
	}
}
