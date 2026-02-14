package usecases

import (
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
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

func TestSummaryUseCase_Execute(t *testing.T) {
	cfg := &utils.Config{LLMModel: "test-model-summary"}
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

	t.Run("Success Indonesian", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Notulen Rapat\n\n## 1. Agenda\nDiskusi fitur RAG.",
		}
		u := NewSummaryUseCase(mockLLM, cfg, nil, store)

		taskID, err := u.SummarizeText("Ini adalah transkripsi rapat", "id", "Rapat Teknis", "Professional")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Empty or Whitespace Input", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewSummaryUseCase(mockLLM, cfg, nil, store)

		taskID, err := u.SummarizeText("   ", "id", "", "")
		if err != nil {
			t.Fatalf("expected no error from async call start, got %v", err)
		}
		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})
}
