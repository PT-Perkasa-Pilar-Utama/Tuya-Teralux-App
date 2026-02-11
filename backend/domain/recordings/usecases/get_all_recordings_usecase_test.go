package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"teralux_app/domain/recordings/entities"
	"teralux_app/domain/recordings/usecases"
)

func TestGetAllRecordingsUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewGetAllRecordingsUseCase(mockRepo)

		recordings := []entities.Recording{
			{ID: "1", Filename: "rec1.mp3", CreatedAt: time.Now()},
			{ID: "2", Filename: "rec2.mp3", CreatedAt: time.Now()},
		}
		total := int64(2)

		mockRepo.On("GetAll", 1, 10).Return(recordings, total, nil)

		result, err := uc.Execute(1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, total, result.Total)
		assert.Equal(t, 2, len(result.Recordings))
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, "rec1.mp3", result.Recordings[0].Filename)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewGetAllRecordingsUseCase(mockRepo)

		mockRepo.On("GetAll", 1, 10).Return([]entities.Recording{}, int64(0), errors.New("db error"))

		result, err := uc.Execute(1, 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "db error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}
