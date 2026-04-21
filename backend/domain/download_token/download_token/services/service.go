package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sensio/domain/download_token/download_token/entities"
	"sensio/domain/download_token/download_token/repositories"
	"sensio/domain/infrastructure"
)

type DownloadTokenService struct {
	store            *repositories.Store
	storageProvider  infrastructure.StorageProvider
	presignTTLSecond int64
	now              func() time.Time
}

func NewDownloadTokenService(store *repositories.Store, storageProvider infrastructure.StorageProvider) *DownloadTokenService {
	if store == nil {
		store = repositories.NewStore()
	}

	return &DownloadTokenService{
		store:            store,
		storageProvider:  storageProvider,
		presignTTLSecond: 900,
		now:              func() time.Time { return time.Now().UTC() },
	}
}

func (s *DownloadTokenService) CreateToken(recipient, objectKey, purpose string, password ...string) (*entities.Token, error) {
	recipient = strings.TrimSpace(recipient)
	objectKey = strings.TrimSpace(objectKey)
	purpose = strings.TrimSpace(purpose)

	if recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	if objectKey == "" {
		return nil, fmt.Errorf("object key is required")
	}
	if purpose != "audio_zip" && purpose != "summary_pdf" {
		return nil, fmt.Errorf("invalid purpose")
	}

	passwd := ""
	if len(password) > 0 && password[0] != "" {
		passwd = password[0]
	}

	now := s.now()
	token := &entities.Token{
		TokenID:   uuid.NewString(),
		Recipient: recipient,
		ObjectKey: objectKey,
		Purpose:   purpose,
		Password:  passwd,
		ExpiresAt: now.AddDate(1, 0, 0),
	}

	s.store.Save(token)

	return token, nil
}

func (s *DownloadTokenService) ResolveToken(tokenID string) (string, error) {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return "", entities.ErrTokenNotFound
	}

	token, ok := s.store.Get(tokenID)
	if !ok || token == nil {
		return "", entities.ErrTokenNotFound
	}
	if token.RevokedAt != nil {
		return "", entities.ErrTokenRevoked
	}
	if token.ConsumedAt != nil {
		return "", entities.ErrTokenConsumed
	}
	if token.ExpiresAt.Before(s.now()) {
		return "", entities.ErrTokenExpired
	}

	if s.storageProvider == nil {
		return "", fmt.Errorf("storage provider is not configured")
	}

	signedURL, err := s.storageProvider.PresignPut(context.Background(), token.ObjectKey, "application/octet-stream", s.presignTTLSecond)
	if err != nil {
		return "", fmt.Errorf("presign object key %s: %w", token.ObjectKey, err)
	}

	if err := s.store.MarkConsumed(tokenID); err != nil {
		return "", fmt.Errorf("mark token consumed: %w", err)
	}

	return signedURL, nil
}

func (s *DownloadTokenService) RevokeToken(tokenID string) error {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return entities.ErrTokenNotFound
	}

	token, ok := s.store.Get(tokenID)
	if !ok || token == nil {
		return entities.ErrTokenNotFound
	}

	return s.store.Revoke(tokenID)
}
