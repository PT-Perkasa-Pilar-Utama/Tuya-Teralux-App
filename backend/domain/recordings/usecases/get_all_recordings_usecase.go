package usecases

import (
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/repositories"
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

	recordingDtos := make([]recordings_dtos.RecordingResponseDto, 0, len(recordings))
	for _, r := range recordings {
		recordingDtos = append(recordingDtos, recordings_dtos.RecordingResponseDto{
			ID:           r.ID,
			Filename:     r.Filename,
			OriginalName: r.OriginalName,
			AudioURL:     r.AudioUrl,
			CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &recordings_dtos.GetAllRecordingsResponseDto{
		Recordings: recordingDtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}, nil
}
