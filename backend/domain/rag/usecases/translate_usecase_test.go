package usecases

import (
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	ragdtos "sensio/domain/rag/dtos"
	"testing"
)

// mockLLMForTranslate allows setting return values and inspecting calls
type mockLLMForTranslate struct {
	CapturedPrompt string
	CapturedModel  string
	ReturnString   string
	ReturnError    error
}

func (m *mockLLMForTranslate) CallModel(prompt string, model string) (string, error) {
	m.CapturedPrompt = prompt
	m.CapturedModel = model
	return m.ReturnString, m.ReturnError
}

func TestTranslateUseCase_Execute(t *testing.T) {
	utils.LoadConfig()
	cfg := utils.GetConfig()
	cfg.GeminiModelLow = "test-model-v1"
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

	t.Run("Success", func(t *testing.T) {
		mockLLM := &mockLLMForTranslate{
			ReturnString: "  Hello World  ", // Intentionally padded to test trim
		}
		u := NewTranslateUseCase(mockLLM, nil, cfg, nil, store, nil)

		taskID, err := u.TranslateText("hallo welt", "en")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Empty Config Model fallback", func(t *testing.T) {
		emptyCfg := &utils.Config{GeminiModelLow: ""}
		mockLLM := &mockLLMForTranslate{}
		u := NewTranslateUseCase(mockLLM, nil, emptyCfg, nil, store, nil)

		taskID, _ := u.TranslateText("test", "en")
		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})
}
