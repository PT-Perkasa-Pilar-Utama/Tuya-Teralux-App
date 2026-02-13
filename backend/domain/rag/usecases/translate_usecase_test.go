package usecases

import (
	"teralux_app/domain/common/utils"
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

func TestRAGUsecase_Translate(t *testing.T) {
	utils.LoadConfig()
	cfg := utils.GetConfig()
	cfg.LLMModel = "test-model-v1"

	t.Run("Success", func(t *testing.T) {
		mockLLM := &mockLLMForTranslate{
			ReturnString: "  Hello World  ", // Intentionally padded to test trim
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		taskID, err := u.Translate("hallo welt", "en")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})

	t.Run("Empty Config Model fallback", func(t *testing.T) {
		emptyCfg := &utils.Config{LLMModel: ""}
		mockLLM := &mockLLMForTranslate{}
		u := NewRAGUsecase(nil, mockLLM, emptyCfg, nil, nil)

		taskID, _ := u.Translate("test", "en")
		if taskID == "" {
			t.Error("expected non-empty taskID")
		}
	})
}
