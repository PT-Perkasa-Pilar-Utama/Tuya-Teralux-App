package usecases_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"teralux_app/domain/recordings/entities"
	"teralux_app/domain/recordings/usecases"
)

func TestGetRecordingByIDUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewGetRecordingByIDUseCase(mockRepo)

		recording := &entities.Recording{ID: "1", Filename: "rec1.mp3"}
		mockRepo.On("GetByID", "1").Return(recording, nil)

		result, err := uc.Execute("1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "rec1.mp3", result.Filename)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewGetRecordingByIDUseCase(mockRepo)

		mockRepo.On("GetByID", "non-existent").Return(nil, errors.New("not found"))

		result, err := uc.Execute("non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "not found", err.Error())

		mockRepo.AssertExpectations(t)
	})
}
