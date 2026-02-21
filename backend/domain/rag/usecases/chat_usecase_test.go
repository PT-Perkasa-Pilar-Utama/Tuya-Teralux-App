package usecases

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/skills"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMForChat is a mock of skills.LLMClient for Chat test
type mockLLMForChat struct {
	mock.Mock
}

func (m *mockLLMForChat) CallModel(prompt string, model string) (string, error) {
	args := m.Called(prompt, model)
	return args.String(0), args.Error(1)
}

// mockSkill is a mock implementation of the Skill interface.
type mockSkill struct {
	mock.Mock
}

func (m *mockSkill) Name() string {
	return m.Called().String(0)
}

func (m *mockSkill) Description() string {
	return m.Called().String(0)
}

func (m *mockSkill) Execute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*skills.SkillResult), args.Error(1)
}

func TestChatUseCase_Chat(t *testing.T) {
	mockLLM := new(mockLLMForChat)
	registry := skills.NewSkillRegistry()
	mockControlSkill := new(mockSkill)
	mockIdentitySkill := new(mockSkill)

	mockControlSkill.On("Name").Return("Control")
	mockControlSkill.On("Description").Return("Control devices")
	mockIdentitySkill.On("Name").Return("Identity")
	mockIdentitySkill.On("Description").Return("Persona")

	registry.Register(mockControlSkill)
	registry.Register(mockIdentitySkill)

	orchestrator := skills.NewOrchestrator(registry, nil)
	cfg := &utils.Config{GeminiModelHigh: "test-model-chat"}
	uc := NewChatUseCase(mockLLM, cfg, nil, nil, orchestrator)
	uid := "test-user"
	teraluxID := "teralux-1"

	t.Run("Empty prompt", func(t *testing.T) {
		res, err := uc.Chat(uid, "teralux-1", "", "id")
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Orchestration to Control", func(t *testing.T) {
		prompt := "Nyalakan AC"
		// Orchestrator logic: first it calls LLM to route
		mockLLM.On("CallModel", mock.Anything, "high").Return("Control", nil).Once()

		// Then it calls the chosen skill's Execute
		mockControlSkill.On("Execute", mock.MatchedBy(func(ctx *skills.SkillContext) bool {
			return ctx.Prompt == prompt
		})).Return(&skills.SkillResult{
			Message:   "Sure! Running command for AC.",
			IsControl: true,
		}, nil).Once()

		res, err := uc.Chat(uid, teraluxID, prompt, "id")
		assert.NoError(t, err)
		assert.True(t, res.IsControl)
		assert.Equal(t, "Sure! Running command for AC.", res.Response)
		mockLLM.AssertExpectations(t)
		mockControlSkill.AssertExpectations(t)
	})

	t.Run("Orchestration to Identity", func(t *testing.T) {
		prompt := "Siapa kamu?"
		mockLLM.On("CallModel", mock.Anything, "high").Return("Identity", nil).Once()

		mockIdentitySkill.On("Execute", mock.MatchedBy(func(ctx *skills.SkillContext) bool {
			return ctx.Prompt == prompt
		})).Return(&skills.SkillResult{
			Message: "Saya adalah asisten AI Sensio.",
		}, nil).Once()

		res, err := uc.Chat(uid, teraluxID, prompt, "id")
		assert.NoError(t, err)
		assert.False(t, res.IsControl)
		assert.Equal(t, "Saya adalah asisten AI Sensio.", res.Response)
		mockLLM.AssertExpectations(t)
		mockIdentitySkill.AssertExpectations(t)
	})
}
