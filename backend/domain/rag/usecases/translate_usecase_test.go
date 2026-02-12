package usecases

import (
	"errors"
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

		got, err := u.Translate("hallo welt", "en")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "Hello World" {
			t.Errorf("expected 'Hello World', got '%s'", got)
		}

		if mockLLM.CapturedModel != "test-model-v1" {
			t.Errorf("expected model 'test-model-v1', got '%s'", mockLLM.CapturedModel)
		}
	})

	t.Run("LLM Error", func(t *testing.T) {
		mockLLM := &mockLLMForTranslate{
			ReturnError: errors.New("llm failure"),
		}
		u := NewRAGUsecase(nil, mockLLM, cfg, nil, nil)

		_, err := u.Translate("fail me", "en")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "llm failure" {
			t.Errorf("expected 'llm failure', got '%v'", err)
		}
	})

	t.Run("Empty Config Model fallback", func(t *testing.T) {
		emptyCfg := &utils.Config{LLMModel: ""}
		mockLLM := &mockLLMForTranslate{}
		u := NewRAGUsecase(nil, mockLLM, emptyCfg, nil, nil)

		_, _ = u.Translate("test", "en")
		if mockLLM.CapturedModel != "default" {
			t.Errorf("expected model 'default', got '%s'", mockLLM.CapturedModel)
		}
	})
}
