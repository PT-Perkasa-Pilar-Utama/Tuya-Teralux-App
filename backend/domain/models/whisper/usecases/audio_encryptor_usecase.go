package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sensio/domain/common/utils"
	"sensio/domain/crypto"
	"sensio/domain/infrastructure"

	"github.com/google/uuid"
)

type AudioEncryptorUseCase interface {
	EncryptAndStore(ctx context.Context, sessionID string, mergedPath string) (string, string, error)
}

type audioEncryptorUseCase struct {
	storage   infrastructure.StorageProvider
	uploadDir string
}

func NewAudioEncryptorUseCase(storage infrastructure.StorageProvider, uploadDir string) AudioEncryptorUseCase {
	return &audioEncryptorUseCase{
		storage:   storage,
		uploadDir: uploadDir,
	}
}

func (u *audioEncryptorUseCase) EncryptAndStore(ctx context.Context, sessionID string, mergedPath string) (string, string, error) {
	if mergedPath == "" {
		return "", "", fmt.Errorf("merged path is required")
	}

	password, err := crypto.GenerateStrongPassword()
	if err != nil {
		return "", "", fmt.Errorf("generate password: %w", err)
	}

	artifactID := uuid.New().String()
	tmpZipPath := filepath.Join(u.uploadDir, fmt.Sprintf("%s_encrypted.zip", artifactID))
	defer os.Remove(tmpZipPath)

	encryptor := crypto.NewZIPEncryptor()
	if err := encryptor.EncryptFiles(tmpZipPath, []string{mergedPath}, password); err != nil {
		return "", "", fmt.Errorf("encrypt files: %w", err)
	}

	zipData, err := os.ReadFile(tmpZipPath)
	if err != nil {
		return "", "", fmt.Errorf("read encrypted zip: %w", err)
	}

	s3Key := fmt.Sprintf("audio/%s.zip", artifactID)
	if err := u.storage.Put(ctx, s3Key, zipData, "application/zip"); err != nil {
		return "", "", fmt.Errorf("store encrypted zip: %w", err)
	}

	utils.LogInfo("AudioEncryptorUseCase: encrypted audio stored at S3 key %s", s3Key)

	return s3Key, password, nil
}

type FinalizedAudio struct {
	S3Key       string
	Password    string
	FinalizedAt time.Time
}
