package usecases

import (
	"fmt"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/speech/dtos"
)

type GetTranscriptionStatusUseCase interface {
	GetTranscriptionStatus(taskID string) (interface{}, error)
}

type getTranscriptionStatusUseCase struct {
	shortCache          *tasks.BadgerTaskCache
	longCache           *tasks.BadgerTaskCache
	whisperProxyUsecase WhisperProxyUsecase
}

func NewGetTranscriptionStatusUseCase(
	shortCache *tasks.BadgerTaskCache,
	longCache *tasks.BadgerTaskCache,
	whisperProxyUsecase WhisperProxyUsecase,
) GetTranscriptionStatusUseCase {
	return &getTranscriptionStatusUseCase{
		shortCache:          shortCache,
		longCache:           longCache,
		whisperProxyUsecase: whisperProxyUsecase,
	}
}

func (uc *getTranscriptionStatusUseCase) GetTranscriptionStatus(taskID string) (interface{}, error) {
	// 1. Try Short Task
	if uc.shortCache != nil {
		var st dtos.AsyncTranscriptionStatusDTO
		if ttl, found, err := uc.shortCache.GetWithTTL(taskID, &st); err == nil && found {
			st.ExpiresInSecond = int64(ttl.Seconds())
			return dtos.AsyncTranscriptionProcessStatusResponseDTO{
				TaskID:     taskID,
				TaskStatus: &st,
			}, nil
		}
	}

	// 2. Try Long Task
	if uc.longCache != nil {
		var st dtos.AsyncTranscriptionLongStatusDTO
		if ttl, found, err := uc.longCache.GetWithTTL(taskID, &st); err == nil && found {
			st.ExpiresInSecond = int64(ttl.Seconds())
			return dtos.AsyncTranscriptionProcessStatusResponseDTO{
				TaskID:     taskID,
				TaskStatus: &st,
			}, nil
		}
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
