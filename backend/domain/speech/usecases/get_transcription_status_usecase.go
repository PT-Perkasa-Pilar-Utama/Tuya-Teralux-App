package usecases

import (
	"fmt"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/repositories"
)

type GetTranscriptionStatusUseCase interface {
	Execute(taskID string) (interface{}, error)
}

type getTranscriptionStatusUseCase struct {
	taskRepo           repositories.TranscriptionTaskRepository
	whisperProxyUsecase *WhisperProxyUsecase
}

func NewGetTranscriptionStatusUseCase(
	taskRepo repositories.TranscriptionTaskRepository,
	whisperProxyUsecase *WhisperProxyUsecase,
) GetTranscriptionStatusUseCase {
	return &getTranscriptionStatusUseCase{
		taskRepo:           taskRepo,
		whisperProxyUsecase: whisperProxyUsecase,
	}
}

func (uc *getTranscriptionStatusUseCase) Execute(taskID string) (interface{}, error) {
	// 1. Try Short Task Repository
	if status, err := uc.taskRepo.GetShortTask(taskID); err == nil {
		return dtos.AsyncTranscriptionProcessStatusResponseDTO{
			TaskID:     taskID,
			TaskStatus: status,
		}, nil
	}

	// 2. Try Long Task Repository
	if status, err := uc.taskRepo.GetLongTask(taskID); err == nil {
		return dtos.AsyncTranscriptionProcessStatusResponseDTO{
			TaskID:     taskID,
			TaskStatus: status,
		}, nil
	}

	// 3. Try PPU Direct Status (if available)
	if uc.whisperProxyUsecase != nil {
		if status, err := uc.whisperProxyUsecase.GetStatus(taskID); err == nil {
			return dtos.WhisperProxyProcessStatusResponseDTO{
				TaskID:     taskID,
				TaskStatus: status,
			}, nil
		}
	}

	return nil, fmt.Errorf("task not found")
}
