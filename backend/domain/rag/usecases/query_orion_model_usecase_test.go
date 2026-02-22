package usecases_test

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryOrionModelUseCase_Query(t *testing.T) {
	t.Run("Success path - Matches Manual Scenario", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("Hello! As an AI residing on the Orion platform, I am fully operational.", nil)

		uc := usecases.NewQueryOrionModelUseCase(mockLLM)
		
		startTime := time.Now()
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/orion")
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "completed", res.Status)
		assert.Equal(t, "Hello! As an AI residing on the Orion platform, I am fully operational.", res.Result)
		assert.Equal(t, 200, res.HTTPStatusCode)
		assert.Equal(t, "/api/rag/models/orion", res.Trigger)
		assert.GreaterOrEqual(t, res.DurationSeconds, float64(0))
		
		parsedStartTime, parseErr := time.Parse(time.RFC3339, res.StartedAt)
		assert.NoError(t, parseErr)
		assert.WithinDuration(t, startTime, parsedStartTime, 2*time.Second)

		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Matches Manual Scenario (500 Internal Server Error)", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		
		apiErr := utils.NewAPIError(500, "orion api returned status 500: Server Error")
		mockLLM.On("CallModel", "Hello! How are you?", "low").Return("", apiErr)

		uc := usecases.NewQueryOrionModelUseCase(mockLLM)
		
		res, err := uc.Query("Hello! How are you?", "/api/rag/models/orion")
		
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "failed", res.Status)
		assert.Equal(t, apiErr.Error(), res.Error)
		assert.Equal(t, 500, res.HTTPStatusCode) 
		assert.Equal(t, "/api/rag/models/orion", res.Trigger)
		
		mockLLM.AssertExpectations(t)
	})
}
