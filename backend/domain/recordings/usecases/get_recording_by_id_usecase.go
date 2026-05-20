package usecases

import (
	"sensio/domain/common/infrastructure"
	recordings_dtos "sensio/domain/recordings/dtos"
	"sensio/domain/recordings/repositories"
)

type GetRecordingByIDUseCase interface {
	GetRecordingByID(id string) (*recordings_dtos.RecordingResponseDto, error)
}

type getRecordingByIDUseCase struct {
	repo      repositories.RecordingRepository
	s3Service *infrastructure.S3Service
}

func NewGetRecordingByIDUseCase(repo repositories.RecordingRepository, s3Service *infrastructure.S3Service) GetRecordingByIDUseCase {
	return &getRecordingByIDUseCase{repo: repo, s3Service: s3Service}
}

func (uc *getRecordingByIDUseCase) GetRecordingByID(id string) (*recordings_dtos.RecordingResponseDto, error) {
	recording, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	audioURL := recording.AudioUrl
	if recording.S3ObjectKey != "" && uc.s3Service != nil {
		presignedURL, err := uc.s3Service.GeneratePresignedGetURL(recording.S3ObjectKey)
		if err != nil {
			return nil, err
		}
		if presignedURL != "" {
			audioURL = presignedURL
		}
	}

	return &recordings_dtos.RecordingResponseDto{
		ID:           recording.ID,
		Filename:     recording.Filename,
		OriginalName: recording.OriginalName,
		AudioUrl:     audioURL,
		CreatedAt:    recording.CreatedAt,
	}, nil
}
