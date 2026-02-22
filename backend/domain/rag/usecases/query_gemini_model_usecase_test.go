package usecases_test

import (
	"errors"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryGeminiModelUseCase_Query(t *testing.T) {
	t.Run("Success path - Matches Manual Scenario", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("I'm doing well, thank you for asking! How can I help you today?", nil)

		uc := usecases.NewQueryGeminiModelUseCase(mockLLM)
		
		startTime := time.Now()
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/gemini")
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "completed", res.Status)
		assert.Equal(t, "I'm doing well, thank you for asking! How can I help you today?", res.Result)
		assert.Equal(t, 200, res.HTTPStatusCode)
		assert.Equal(t, "/api/rag/models/gemini", res.Trigger)
		assert.GreaterOrEqual(t, res.DurationSeconds, float64(0))
		
		parsedStartTime, parseErr := time.Parse(time.RFC3339, res.StartedAt)
		assert.NoError(t, parseErr)
		assert.WithinDuration(t, startTime, parsedStartTime, 2*time.Second)

		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Matches Manual Scenario (500 Internal Server Error)", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("", errors.New("failed to call gemini api: timeout"))

		uc := usecases.NewQueryGeminiModelUseCase(mockLLM)
		
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/gemini")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, "failed to call gemini api: timeout", res.Error)
		// Standard error without APIError wrapper defaults to 500 when mapped
		assert.Equal(t, 500, utils.GetErrorStatusCode(err)) 
		assert.Equal(t, "/api/rag/models/gemini", res.Trigger)
		
		mockLLM.AssertExpectations(t)
	})
}
