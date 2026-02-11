package usecases

import (
	"teralux_app/domain/recordings/dtos"
	"teralux_app/domain/recordings/repositories"
)

type GetAllRecordingsUseCase interface {
	Execute(page, limit int) (*dtos.GetAllRecordingsResponseDto, error)
}

type getAllRecordingsUseCase struct {
	repo repositories.RecordingRepository
}

func NewGetAllRecordingsUseCase(repo repositories.RecordingRepository) GetAllRecordingsUseCase {
	return &getAllRecordingsUseCase{repo: repo}
}

func (uc *getAllRecordingsUseCase) Execute(page, limit int) (*dtos.GetAllRecordingsResponseDto, error) {
	recordings, total, err := uc.repo.GetAll(page, limit)
	if err != nil {
		return nil, err
	}

	var recordingDtos []dtos.RecordingResponseDto
	for _, r := range recordings {
		recordingDtos = append(recordingDtos, dtos.RecordingResponseDto{
			ID:           r.ID,
			Filename:     r.Filename,
			OriginalName: r.OriginalName,
			AudioUrl:     r.AudioUrl,
			CreatedAt:    r.CreatedAt,
		})
	}

	return &dtos.GetAllRecordingsResponseDto{
		Recordings: recordingDtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}, nil
}
