package usecases

import (
	"sensio/domain/rag/skills"

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

// SimpleMockSkill is a non-testify based mock for simpler tests.
type SimpleMockSkill struct {
	SkillName string
}

func (m *SimpleMockSkill) Name() string        { return m.SkillName }
func (m *SimpleMockSkill) Description() string { return "" }
func (m *SimpleMockSkill) Execute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	// Wrap the prompt to satisfy the test expectation that it contains 'professional editor'
	wrappedPrompt := "Act as a professional editor. Refine: " + ctx.Prompt
	res, err := ctx.LLM.CallModel(wrappedPrompt, "low")
	if err != nil {
		return nil, err
	}
	return &skills.SkillResult{Message: res, HTTPStatusCode: 200}, nil
}
