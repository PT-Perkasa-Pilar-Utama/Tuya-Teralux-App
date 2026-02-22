package usecases

import (
	"errors"
	"strings"
	"teralux_app/domain/common/utils"
	"testing"
)

// mockLLMForRefine allows setting return values and inspecting calls
type mockLLMForRefine struct {
	CapturedPrompt string
	CapturedModel  string
	ReturnString   string
	ReturnError    error
}

func (m *mockLLMForRefine) CallModel(prompt string, model string) (string, error) {
	m.CapturedPrompt = prompt
	m.CapturedModel = model
	return m.ReturnString, m.ReturnError
}

func TestRefineUseCase_Execute(t *testing.T) {
	cfg := &utils.Config{GeminiModelLow: "test-model-refine"}

	t.Run("Refine Indonesian (KBBI)", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnString: "Saya sedang makan nasi.",
		}
		u := NewRefineUseCase(mockLLM, nil, cfg)

		got, err := u.RefineText("aku lagi mamam nasi", "id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "Saya sedang makan nasi." {
			t.Errorf("expected 'Saya sedang makan nasi.', got '%s'", got)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "KBBI") {
			t.Errorf("expected prompt to contain 'KBBI', got '%s'", mockLLM.CapturedPrompt)
		}
	})

	t.Run("Refine English", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnString: "I am eating rice.",
		}
		u := NewRefineUseCase(mockLLM, nil, cfg)

		got, err := u.RefineText("i is eating rice", "en")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "I am eating rice." {
			t.Errorf("expected 'I am eating rice.', got '%s'", got)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "English editor") {
			t.Errorf("expected prompt to contain 'English editor', got '%s'", mockLLM.CapturedPrompt)
		}
	})

	t.Run("LLM Error", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnError: errors.New("llm failure"),
		}
		u := NewRefineUseCase(mockLLM, nil, cfg)

		_, err := u.RefineText("test", "id")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "llm failure") {
			t.Errorf("expected error to contain 'llm failure', got '%v'", err)
		}

	})

	t.Run("Empty or Whitespace Input (Silent Audio)", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{}
		u := NewRefineUseCase(mockLLM, nil, cfg)

		got, err := u.RefineText("   ", "id")
		if err != nil {
			t.Fatalf("expected no error for silent audio, got %v", err)
		}
		if got != "" {
			t.Errorf("expected empty string result, got '%s'", got)
		}
	})

	t.Run("Unknown Language Default to English", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnString: "Refined English",
		}
		u := NewRefineUseCase(mockLLM, nil, cfg)

		_, err := u.RefineText("some text", "xyz")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "English editor") {
			t.Errorf("expected prompt to default to English editor for unknown lang, got %s", mockLLM.CapturedPrompt)
		}
	})
}
