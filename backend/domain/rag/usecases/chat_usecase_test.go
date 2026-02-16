package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMForChat is a mock of utilities.LLMClient for Chat test
type mockLLMForChat struct {
	mock.Mock
}

func (m *mockLLMForChat) CallModel(prompt string, model string) (string, error) {
	args := m.Called(prompt, model)
	return args.String(0), args.Error(1)
}

func TestChatUseCase_Chat(t *testing.T) {
	mockLLM := new(mockLLMForChat)
	cfg := &utils.Config{LLMModel: "test-model"}
	uc := NewChatUseCase(mockLLM, cfg, nil)
	uid := "test-user"

	t.Run("Empty prompt", func(t *testing.T) {
		res, err := uc.Chat(uid, "teralux-1", "", "id")
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Control command classification", func(t *testing.T) {
		prompt := "Nyalakan AC"
		mockLLM.On("CallModel", mock.AnythingOfType("string"), "test-model").Return("CONTROL", nil).Once()

		res, err := uc.Chat(uid, "teralux-1", prompt, "id")
		assert.NoError(t, err)
		assert.True(t, res.IsControl)
		assert.Equal(t, "/api/rag/control", res.Redirect.Endpoint)
		assert.Equal(t, prompt, res.Redirect.Body.(dtos.RAGControlRequestDTO).Prompt)
		mockLLM.AssertExpectations(t)
	})

	t.Run("General chat classification", func(t *testing.T) {
		prompt := "Halo siapa kamu?"
		mockLLM.On("CallModel", mock.MatchedBy(func(p string) bool {
			return strings.Contains(p, "CONTROL") || strings.Contains(p, "CHAT")
		}), "test-model").Return("CHAT", nil).Once()
		
		mockLLM.On("CallModel", mock.MatchedBy(func(p string) bool {
			return strings.Contains(p, "Sensio AI Assistant")
		}), "test-model").Return("Saya adalah asisten AI Sensio.", nil).Once()

		res, err := uc.Chat(uid, "teralux-1", prompt, "id")
		assert.NoError(t, err)
		assert.False(t, res.IsControl)
		assert.Equal(t, "Saya adalah asisten AI Sensio.", res.Response)
		mockLLM.AssertExpectations(t)
	})

	t.Run("LLM error in classification", func(t *testing.T) {
		mockLLM.On("CallModel", mock.Anything, "test-model").Return("", fmt.Errorf("LLM error")).Once()

		res, err := uc.Chat(uid, "teralux-1", "p", "id")
		assert.Error(t, err)
		assert.Nil(t, res)
		mockLLM.AssertExpectations(t)
	})
}
