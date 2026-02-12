package usecases

import (
	"errors"
	"strings"
	"teralux_app/domain/common/utils"
	"testing"
)

// mockLLMForSummary allows setting return values and inspecting calls
type mockLLMForSummary struct {
	CapturedPrompt string
	CapturedModel  string
	ReturnString   string
	ReturnError    error
}

func (m *mockLLMForSummary) CallModel(prompt string, model string) (string, error) {
	m.CapturedPrompt = prompt
	m.CapturedModel = model
	return m.ReturnString, m.ReturnError
}

func TestRAGUsecase_Summary(t *testing.T) {
	cfg := &utils.Config{LLMModel: "test-model-summary"}

	t.Run("Success Indonesian", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Notulen Rapat\n\n## 1. Agenda\nDiskusi fitur RAG.",
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		got, err := u.Summary("Ini adalah transkripsi rapat", "id", "Rapat Teknis", "Professional")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(got.Summary, "# Notulen Rapat") {
			t.Errorf("expected summary to contain '# Notulen Rapat', got '%s'", got.Summary)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "Indonesian") {
			t.Errorf("expected prompt to contain 'Indonesian', got '%s'", mockLLM.CapturedPrompt)
		}
	})

	t.Run("Success English", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Meeting Minutes\n\n## 1. Agenda\nRAG feature discussion.",
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		got, err := u.Summary("This is a meeting transcript", "en", "Technical Meeting", "Professional")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(got.Summary, "# Meeting Minutes") {
			t.Errorf("expected summary to contain '# Meeting Minutes', got '%s'", got.Summary)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "English") {
			t.Errorf("expected prompt to contain 'English', got '%s'", mockLLM.CapturedPrompt)
		}
	})

	t.Run("LLM Error", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnError: errors.New("llm failure"),
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		_, err := u.Summary("test", "id", "", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "llm failure" {
			t.Errorf("expected 'llm failure', got '%v'", err)
		}
	})

	t.Run("Empty or Whitespace Input", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		_, err := u.Summary("   ", "id", "", "")
		if err == nil {
			t.Fatal("expected error for whitespace input, got nil")
		}
		if !strings.Contains(err.Error(), "text is empty") {
			t.Errorf("expected 'text is empty' error, got '%v'", err)
		}
	})

	t.Run("Invalid Language Fallback to ID", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "Summary in ID",
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		_, err := u.Summary("test text", "alien-lang", "", "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(mockLLM.CapturedPrompt, "Indonesian") {
			t.Errorf("expected prompt to default to Indonesian for invalid lang, got %s", mockLLM.CapturedPrompt)
		}
	})
}
