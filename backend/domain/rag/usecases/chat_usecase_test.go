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

type mockControlUseCaseForChat struct {
	mock.Mock
}

func (m *mockControlUseCaseForChat) ProcessControl(uid, teraluxID, prompt string) (*dtos.ControlResultDTO, error) {
	args := m.Called(uid, teraluxID, prompt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ControlResultDTO), args.Error(1)
}

type mockTranslateUseCaseForChat struct {
	mock.Mock
}

func (m *mockTranslateUseCaseForChat) TranslateText(text, targetLang string) (string, error) {
	args := m.Called(text, targetLang)
	return args.String(0), args.Error(1)
}

func (m *mockTranslateUseCaseForChat) TranslateTextSync(text, targetLang string) (string, error) {
	args := m.Called(text, targetLang)
	return args.String(0), args.Error(1)
}

func TestChatUseCase_Chat(t *testing.T) {
	mockLLM := new(mockLLMForChat)
	mockControl := new(mockControlUseCaseForChat)
	mockTranslate := new(mockTranslateUseCaseForChat)
	cfg := &utils.Config{LLMModel: "test-model"}
	uc := NewChatUseCase(mockLLM, cfg, nil, mockControl, mockTranslate)
	uid := "test-user"
	teraluxID := "teralux-1"

	t.Run("Empty prompt", func(t *testing.T) {
		res, err := uc.Chat(uid, "teralux-1", "", "id")
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Control command classification - Indonesian", func(t *testing.T) {
		prompt := "Nyalakan AC"
		mockLLM.On("CallModel", mock.AnythingOfType("string"), "test-model").Return("CONTROL", nil).Once()
		mockControl.On("ProcessControl", uid, teraluxID, prompt).Return(&dtos.ControlResultDTO{
			Message:  "Sure! Running command for **AC**.",
			DeviceID: "ac-1",
		}, nil).Once()

		mockTranslate.On("TranslateTextSync", "Sure! Running command for **AC**.", "id").Return("Tentu! Menjalankan perintah untuk **AC**.", nil).Once()

		res, err := uc.Chat(uid, teraluxID, prompt, "id")
		assert.NoError(t, err)
		assert.True(t, res.IsControl)
		assert.Equal(t, "/api/rag/control", res.Redirect.Endpoint)
		assert.Equal(t, prompt, res.Redirect.Body.(dtos.RAGControlRequestDTO).Prompt)
		assert.Equal(t, "Tentu! Menjalankan perintah untuk **AC**.", res.Response)
		mockLLM.AssertExpectations(t)
		mockControl.AssertExpectations(t)
		mockTranslate.AssertExpectations(t)
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
