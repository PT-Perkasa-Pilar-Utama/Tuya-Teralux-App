package usecases

import (
	"context"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/services"
	"testing"
	"time"
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

type noopSummaryRenderer struct{}

func (n *noopSummaryRenderer) Render(summary string, path string, meta services.SummaryPDFMeta) error {
	return nil
}

func TestSummaryUseCase_Execute(t *testing.T) {
	cfg := &utils.Config{GeminiModelHigh: "test-model-summary"}
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

	t.Run("Success Indonesian", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Notulen Rapat\n\n## 1. Agenda\nDiskusi fitur RAG.",
		}
		u := NewSummaryUseCase(mockLLM, cfg, nil, store, &noopSummaryRenderer{})

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
		u := NewSummaryUseCase(mockLLM, cfg, nil, store, &noopSummaryRenderer{})

		taskID, err := u.SummarizeText("   ", "id", "", "")
		if err != nil {
			t.Fatalf("expected no error from async call start, got %v", err)
		}
		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Context-aware with valid context", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Meeting Summary\n\n## Decisions\n- Approve budget allocation",
		}
		u := NewSummaryUseCase(mockLLM, cfg, nil, store, &noopSummaryRenderer{})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		taskID, err := u.SummarizeTextWithContext(ctx, "Meeting discussion about Q1 roadmap", "en", "Strategic Planning", "Executive Brief")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Context-aware with cancelled context", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewSummaryUseCase(mockLLM, cfg, nil, store, &noopSummaryRenderer{})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Immediately cancel

		_, err := u.SummarizeTextWithContext(ctx, "Some transcript", "en", "Context", "Brief")
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})
}
