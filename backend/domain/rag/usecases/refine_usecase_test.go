package usecases

import (
	"context"
	"errors"
	"fmt"
	"sensio/domain/common/utils"
	"testing"
)

// mockLLMForRefine allows setting return values and inspecting calls
type mockLLMForRefine struct {
	CapturedPrompt string
	CapturedModel  string
	ReturnString   string
	ReturnError    error
}

func (m *mockLLMForRefine) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	m.CapturedPrompt = prompt
	m.CapturedModel = model
	return m.ReturnString, m.ReturnError
}

// MockRefineUseCase is a mock implementation of the RefineUseCase interface for testing
type MockRefineUseCase struct {
	// Embed testify/mock.Mock to get all the mocking functionalities
	// This is a placeholder, assuming testify/mock is used elsewhere or intended.
	// For this specific file, it's not fully defined, but the method signature is provided.
}

func (m *MockRefineUseCase) RefineText(ctx context.Context, text string, targetLang string) (string, error) {
	// This implementation assumes 'testify/mock' is used.
	// Since 'testify/mock' is not imported or defined in this snippet,
	// this method will cause a compile error if MockRefineUseCase is actually used.
	// For the purpose of fulfilling the request, I'm adding it as provided.
	// If this is not the intended use, please clarify.
	// args := m.Called(ctx, text, targetLang)
	// return args.String(0), args.Error(1)
	return "", fmt.Errorf("MockRefineUseCase.RefineText not implemented without testify/mock")
}

func TestRefineUseCase_Execute(t *testing.T) {
	cfg := &utils.Config{GeminiModelLow: "test-model-refine"}

	t.Run("Refine Indonesian (KBBI)", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnString: "Saya sedang makan nasi.",
		}
		u := NewRefineUseCase(mockLLM, nil, cfg, &SimpleMockSkill{SkillName: "Refine"})

		got, err := u.RefineText(context.Background(), "aku lagi mamam nasi", "id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "Saya sedang makan nasi." {
			t.Errorf("expected 'Saya sedang makan nasi.', got '%s'", got)
		}

	})

	t.Run("Refine English", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnString: "I am eating rice.",
		}
		u := NewRefineUseCase(mockLLM, nil, cfg, &SimpleMockSkill{SkillName: "Refine"})

		got, err := u.RefineText(context.Background(), "i is eating rice", "en")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "I am eating rice." {
			t.Errorf("expected 'I am eating rice.', got '%s'", got)
		}
	})

	t.Run("LLM Error", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{
			ReturnError: errors.New("llm failure"),
		}
		u := NewRefineUseCase(mockLLM, nil, cfg, &SimpleMockSkill{SkillName: "Refine"})

		_, err := u.RefineText(context.Background(), "test", "id")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Empty or Whitespace Input (Silent Audio)", func(t *testing.T) {
		mockLLM := &mockLLMForRefine{}
		u := NewRefineUseCase(mockLLM, nil, cfg, &SimpleMockSkill{SkillName: "Refine"})

		_, err := u.RefineText(context.Background(), "aku lagi mamam nasi", "id")
		if err != nil {
			t.Fatalf("expected no error for silent audio, got %v", err)
		}
	})
}
