package usecases_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"teralux_app/domain/recordings/entities"
	"teralux_app/domain/recordings/usecases"
)

func TestDeleteRecordingUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewDeleteRecordingUseCase(mockRepo)

		recording := &entities.Recording{ID: "1", Filename: "rec1.mp3"}
		mockRepo.On("GetByID", "1").Return(recording, nil)
		mockRepo.On("Delete", "1").Return(nil)

		err := uc.Execute("1")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Record Not Found", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewDeleteRecordingUseCase(mockRepo)

		mockRepo.On("GetByID", "1").Return(nil, errors.New("not found"))

		err := uc.Execute("1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recording not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete Metadata Error", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		uc := usecases.NewDeleteRecordingUseCase(mockRepo)

		recording := &entities.Recording{ID: "1", Filename: "rec1.mp3"}
		mockRepo.On("GetByID", "1").Return(recording, nil)
		mockRepo.On("Delete", "1").Return(errors.New("db error"))

		err := uc.Execute("1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete metadata")
		mockRepo.AssertExpectations(t)
	})
}
