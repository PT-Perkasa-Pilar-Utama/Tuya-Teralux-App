package usecases

import (
	recordings_dtos "teralux_app/domain/recordings/dtos"
	"teralux_app/domain/recordings/repositories"
)

type GetAllRecordingsUseCase interface {
	ListRecordings(page, limit int) (*recordings_dtos.GetAllRecordingsResponseDto, error)
}

type getAllRecordingsUseCase struct {
	repo repositories.RecordingRepository
}

func NewGetAllRecordingsUseCase(repo repositories.RecordingRepository) GetAllRecordingsUseCase {
	return &getAllRecordingsUseCase{repo: repo}
}

func (uc *getAllRecordingsUseCase) ListRecordings(page, limit int) (*recordings_dtos.GetAllRecordingsResponseDto, error) {
	recordings, total, err := uc.repo.GetAll(page, limit)
	if err != nil {
		return nil, err
	}

	var recordingDtos []recordings_dtos.RecordingResponseDto
	for _, r := range recordings {
		recordingDtos = append(recordingDtos, recordings_dtos.RecordingResponseDto{
			ID:           r.ID,
			Filename:     r.Filename,
			OriginalName: r.OriginalName,
			AudioUrl:     r.AudioUrl,
			CreatedAt:    r.CreatedAt,
		})
	}

	return &recordings_dtos.GetAllRecordingsResponseDto{
		Recordings: recordingDtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}, nil
}
