package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryGroqModelUseCase_Query(t *testing.T) {
	t.Run("Success path - Matches Manual Scenario", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("Grateful to help! What do you need today?", nil)

		uc := usecases.NewQueryGroqModelUseCase(mockLLM)
		
		startTime := time.Now()
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/groq")
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "completed", res.Status)
		assert.Equal(t, "Grateful to help! What do you need today?", res.Result)
		assert.Equal(t, 200, res.HTTPStatusCode)
		assert.Equal(t, "/api/rag/models/groq", res.Trigger)
		assert.GreaterOrEqual(t, res.DurationSeconds, float64(0))
		
		parsedStartTime, parseErr := time.Parse(time.RFC3339, res.StartedAt)
		assert.NoError(t, parseErr)
		assert.WithinDuration(t, startTime, parsedStartTime, 2*time.Second)

		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Matches Manual Scenario (429 Too Many Requests)", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		
		apiErr := utils.NewAPIError(429, "groq api returned status 429: Rate limit exceeded")
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("", apiErr)

		uc := usecases.NewQueryGroqModelUseCase(mockLLM)
		
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/groq")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, apiErr.Error(), res.Error)
		assert.Equal(t, 429, res.HTTPStatusCode) 
		assert.Equal(t, "/api/rag/models/groq", res.Trigger)
		
		mockLLM.AssertExpectations(t)
	})
}
