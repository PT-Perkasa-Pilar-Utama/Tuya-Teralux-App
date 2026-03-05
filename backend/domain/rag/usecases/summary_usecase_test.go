package usecases

import (
	"context"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/services"
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
	store := tasks.NewStatusStore[dtos.RAGStatusDTO]()

	t.Run("Success", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Notulen Rapat\n\n## 1. Agenda\nDiskusi fitur RAG.",
		}
		u := NewSummaryUseCase(mockLLM, nil, cfg, nil, store, &noopSummaryRenderer{}, nil, nil, &SimpleMockSkill{SkillName: "Summary"})

		taskID, err := u.SummarizeText("Ini adalah transkripsi rapat", "id", "Rapat Teknis", "Professional", "2024-05-20", "Ruang Rapat 1", "Faris, Budi", "", "http://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Empty or Whitespace Input", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewSummaryUseCase(mockLLM, nil, cfg, nil, store, &noopSummaryRenderer{}, nil, nil, &SimpleMockSkill{SkillName: "Summary"})

		taskID, err := u.SummarizeText("   ", "id", "", "", "", "", "", "", "")
		if err != nil {
			t.Fatalf("expected no error for silent audio, got %v", err)
		}
		if taskID != "" {
			t.Error("expected empty taskID for empty input")
		}
	})

	t.Run("Context-aware Success", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{
			ReturnString: "# Meeting Summary\n\n## Decisions\n- Approve budget allocation",
		}
		u := NewSummaryUseCase(mockLLM, nil, cfg, nil, store, &noopSummaryRenderer{}, nil, nil, &SimpleMockSkill{SkillName: "Summary"})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		summary, err := u.SummarizeTextWithContext(ctx, "Text to summarize", "en", "Decision log", "Concise", "2024-05-20", "Boardroom", "CEO, CFO", "", "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if summary == "" {
			t.Error("expected non-empty summary")
		}
	})

	t.Run("Context-aware with cancelled context", func(t *testing.T) {
		mockLLM := &mockLLMForSummary{}
		u := NewSummaryUseCase(mockLLM, nil, cfg, nil, store, &noopSummaryRenderer{}, nil, nil, &SimpleMockSkill{SkillName: "Summary"})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Immediately cancel

		_, err := u.SummarizeTextWithContext(ctx, "Some text", "en", "", "", "", "", "", "", "")
		if err == nil {
			t.Fatal("expected error for cancelled context, got nil")
		}
	})
}
