package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryOpenAIModelUseCase_Query(t *testing.T) {
	t.Run("Success path - Matches Manual Scenario", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("I am just a computer program, but I'm functioning perfectly.", nil)

		uc := usecases.NewQueryOpenAIModelUseCase(mockLLM)
		
		startTime := time.Now()
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/openai")
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "completed", res.Status)
		assert.Equal(t, "I am just a computer program, but I'm functioning perfectly.", res.Result)
		assert.Equal(t, 200, res.HTTPStatusCode)
		assert.Equal(t, "/api/rag/models/openai", res.Trigger)
		assert.GreaterOrEqual(t, res.DurationSeconds, float64(0))
		
		parsedStartTime, parseErr := time.Parse(time.RFC3339, res.StartedAt)
		assert.NoError(t, parseErr)
		assert.WithinDuration(t, startTime, parsedStartTime, 2*time.Second)

		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Matches Manual Scenario (401 Unauthorized)", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		
		// Simulate a wrapped utils.APIError as often returned by services
		apiErr := utils.NewAPIError(401, "openai api returned status 401: unauthorized")
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("", apiErr)

		uc := usecases.NewQueryOpenAIModelUseCase(mockLLM)
		
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/openai")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, apiErr.Error(), res.Error)
		assert.Equal(t, 401, res.HTTPStatusCode) 
		assert.Equal(t, "/api/rag/models/openai", res.Trigger)
		
		mockLLM.AssertExpectations(t)
	})
}
