package usecases

import (
	"sensio/domain/common/infrastructure"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/repositories"
)

type GetAllRecordingsUseCase interface {
	ListRecordings(page, limit int) (*recordings_dtos.GetAllRecordingsResponseDto, error)
}

type getAllRecordingsUseCase struct {
	repo      repositories.RecordingRepository
	s3Service *infrastructure.S3Service
}

func NewGetAllRecordingsUseCase(repo repositories.RecordingRepository, s3Service *infrastructure.S3Service) GetAllRecordingsUseCase {
	return &getAllRecordingsUseCase{repo: repo, s3Service: s3Service}
}

func (uc *getAllRecordingsUseCase) ListRecordings(page, limit int) (*recordings_dtos.GetAllRecordingsResponseDto, error) {
	recordings, total, err := uc.repo.GetAll(page, limit)
	if err != nil {
		return nil, err
	}

	recordingDtos := make([]recordings_dtos.RecordingResponseDto, 0, len(recordings))
	for _, r := range recordings {
		audioURL := r.AudioUrl
		if r.S3ObjectKey != "" && uc.s3Service != nil {
			presignedURL, err := uc.s3Service.GeneratePresignedGetURL(r.S3ObjectKey)
			if err != nil {
				return nil, err
			}
			if presignedURL != "" {
				audioURL = presignedURL
			}
		}

		recordingDtos = append(recordingDtos, recordings_dtos.RecordingResponseDto{
			ID:           r.ID,
			Filename:     r.Filename,
			OriginalName: r.OriginalName,
			AudioUrl:     audioURL,
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
