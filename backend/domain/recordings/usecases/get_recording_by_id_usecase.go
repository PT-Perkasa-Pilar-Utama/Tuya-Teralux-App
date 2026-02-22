package usecases

import (
	recordings_dtos "teralux_app/domain/recordings/dtos"
	"teralux_app/domain/recordings/repositories"
)

type GetRecordingByIDUseCase interface {
	GetRecordingByID(id string) (*recordings_dtos.RecordingResponseDto, error)
}

type getRecordingByIDUseCase struct {
	repo repositories.RecordingRepository
}

func NewGetRecordingByIDUseCase(repo repositories.RecordingRepository) GetRecordingByIDUseCase {
	return &getRecordingByIDUseCase{repo: repo}
}

func (uc *getRecordingByIDUseCase) GetRecordingByID(id string) (*recordings_dtos.RecordingResponseDto, error) {
	recording, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &recordings_dtos.RecordingResponseDto{
		ID:           recording.ID,
		Filename:     recording.Filename,
		OriginalName: recording.OriginalName,
		AudioUrl:     recording.AudioUrl,
		CreatedAt:    recording.CreatedAt,
	}, nil
}
