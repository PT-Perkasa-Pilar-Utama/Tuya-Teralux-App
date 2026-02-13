package usecases

import (
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

		taskID, err := u.Summary("Ini adalah transkripsi rapat", "id", "Rapat Teknis", "Professional")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Empty or Whitespace Input", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		taskID, err := u.Summary("   ", "id", "", "")
		if err != nil {
			t.Fatalf("expected no error from async call start, got %v", err)
		}
		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})
}
