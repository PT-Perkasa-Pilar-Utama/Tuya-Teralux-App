package usecases

import (
	"context"
	"errors"
	"sensio/domain/rag/skills"
	"strings"

	"github.com/stretchr/testify/mock"
)

// UseCaseMockSkill is a shared mock implementation of the Skill interface for usecase tests.
type UseCaseMockSkill struct {
	mock.Mock
}

func (m *UseCaseMockSkill) Name() string {
	return m.Called().String(0)
}

func (m *UseCaseMockSkill) Description() string {
	return m.Called().String(0)
}

func (m *UseCaseMockSkill) Execute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*skills.SkillResult), args.Error(1)
}

// MockLLM is a mock implementation of the LLM interface for testing.
type MockLLM struct {
	mock.Mock
}

func (m *MockLLM) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	args := m.Called(ctx, prompt, model)
	return args.String(0), args.Error(1)
}

// SimpleMockSkill is a non-testify based mock for simpler tests.
type SimpleMockSkill struct {
	SkillName string
}

func (m *SimpleMockSkill) Name() string        { return m.SkillName }
func (m *SimpleMockSkill) Description() string { return "" }
func (s *SimpleMockSkill) Execute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	if s.SkillName == "Translation" {
		return &skills.SkillResult{Message: "Translated!"}, nil
	}
	// For Refine
	lowerPrompt := strings.ToLower(ctx.Prompt)
	if lowerPrompt == "test" {
		return nil, errors.New("llm failure")
	}
	if strings.Contains(lowerPrompt, "mamam") {
		return &skills.SkillResult{Message: "Saya sedang makan nasi."}, nil
	}
	if strings.Contains(lowerPrompt, "eating") {
		return &skills.SkillResult{Message: "I am eating rice."}, nil
	}
	return &skills.SkillResult{Message: "Mocked!"}, nil
}
