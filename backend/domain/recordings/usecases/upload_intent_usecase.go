package usecases

import (
	"context"
	"fmt"
	"time"

	"sensio/domain/infrastructure"
	"sensio/domain/recordings/dtos"

	"github.com/google/uuid"
)

type UploadIntentUseCase interface {
	CreateUploadIntent(ctx context.Context, req recordings_dtos.CreateUploadIntentRequest) (*recordings_dtos.UploadIntentResponseDTO, error)
}

type uploadIntentUseCase struct {
	storage   infrastructure.StorageProvider
	ttlSeconds int64
}

func NewUploadIntentUseCase(storage infrastructure.StorageProvider, ttlSeconds int64) UploadIntentUseCase {
	if ttlSeconds <= 0 {
		ttlSeconds = 900 // default 15 minutes
	}
	return &uploadIntentUseCase{
		storage:   storage,
		ttlSeconds: ttlSeconds,
	}
}

func (u *uploadIntentUseCase) CreateUploadIntent(ctx context.Context, req recordings_dtos.CreateUploadIntentRequest) (*recordings_dtos.UploadIntentResponseDTO, error) {
	// Validate request
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if req.Size <= 0 {
		return nil, fmt.Errorf("size must be greater than 0")
	}
	if req.ContentType == "" {
		return nil, fmt.Errorf("contentType is required")
	}
	if req.BookingID == "" {
		return nil, fmt.Errorf("bookingId is required")
	}

	// Generate UUID for object key
	uid := uuid.New().String()

	// Build S3Key: Sensio/audio/{YYYY-MM-DD}/{uuid}.zip
	now := time.Now()
	datePath := now.Format("2006-01-02")
	s3Key := fmt.Sprintf("Sensio/audio/%s/%s.zip", datePath, uid)

	// Call StorageProvider.PresignPut
	uploadURL, err := u.storage.PresignPut(ctx, s3Key, req.ContentType, u.ttlSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Calculate expiresAt
	expiresAt := now.Add(time.Duration(u.ttlSeconds) * time.Second)

	return &recordings_dtos.UploadIntentResponseDTO{
		ObjectKey:   s3Key,
		UploadURL:   uploadURL,
		ContentType: req.ContentType,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}