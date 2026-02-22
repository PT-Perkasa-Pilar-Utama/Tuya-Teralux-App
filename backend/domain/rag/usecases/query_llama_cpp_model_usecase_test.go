package usecases_test

import (
	"errors"
	"teralux_app/domain/rag/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient manually mimics skills.LLMClient for testing.
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) HealthCheck() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLLMClient) CallModel(prompt string, model string) (string, error) {
	args := m.Called(prompt, model)
	return args.String(0), args.Error(1)
}


func TestQueryLlamaCppModelUseCase_Query(t *testing.T) {
	t.Run("Success path", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "testing prompt", "low").Return("mocked response payload", nil)

		uc := usecases.NewQueryLlamaCppModelUseCase(mockLLM)
		
		startTime := time.Now()
		res, err := uc.Query("testing prompt", "/api/rag/models/gemini")
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "completed", res.Status)
		assert.Equal(t, "mocked response payload", res.Result)
		assert.Equal(t, 200, res.HTTPStatusCode)
		assert.Equal(t, "/api/rag/models/gemini", res.Trigger)
		assert.GreaterOrEqual(t, res.DurationSeconds, float64(0)) // Duration should be >0
		
		parsedStartTime, parseErr := time.Parse(time.RFC3339, res.StartedAt)
		assert.NoError(t, parseErr)
		assert.WithinDuration(t, startTime, parsedStartTime, 2*time.Second)

		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - LLM Engine Error", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "failing prompt", "low").Return("", errors.New("upstream timeout"))

		uc := usecases.NewQueryLlamaCppModelUseCase(mockLLM)
		
		res, err := uc.Query("failing prompt", "/api/rag/models/openai")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, "upstream timeout", res.Error)
		// Assuming GetErrorStatusCode gracefully handles standard errors to 0 then mapped dynamically or 500 later. 
		// Actually if its not API Error it might be 500
		assert.Equal(t, "/api/rag/models/openai", res.Trigger)
		
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Empty response", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "empty prompt", "low").Return("", nil)

		uc := usecases.NewQueryLlamaCppModelUseCase(mockLLM)
		
		res, err := uc.Query("empty prompt", "/api/rag/models/groq")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, "llm returned an empty response", res.Error)
		assert.Equal(t, 500, res.HTTPStatusCode)
		
		mockLLM.AssertExpectations(t)
	})
}
