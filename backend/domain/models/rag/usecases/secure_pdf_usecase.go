package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sensio/domain/common/utils"
	"sensio/domain/common/utils/crypto"
	"sensio/domain/infrastructure"

	"github.com/google/uuid"
)

type SecurePDFUseCase interface {
	ProtectAndStore(ctx context.Context, pdfPath string) (string, string, error)
}

type securePDFUseCase struct {
	storage infrastructure.StorageProvider
}

func NewSecurePDFUseCase(storage infrastructure.StorageProvider) SecurePDFUseCase {
	return &securePDFUseCase{storage: storage}
}

func (uc *securePDFUseCase) ProtectAndStore(ctx context.Context, pdfPath string) (string, string, error) {
	if pdfPath == "" {
		return "", "", fmt.Errorf("pdf path is required")
	}

	password, err := crypto.GenerateStrongPassword()
	if err != nil {
		return "", "", fmt.Errorf("generate password: %w", err)
	}

	tmpProtectedPath := filepath.Join(os.TempDir(), fmt.Sprintf("protected_%s.pdf", uuid.New().String()))
	defer os.Remove(tmpProtectedPath)

	protector := crypto.NewPDFProtector()
	if err := protector.Protect(pdfPath, tmpProtectedPath, password); err != nil {
		return "", "", fmt.Errorf("protect PDF: %w", err)
	}

	pdfData, err := os.ReadFile(tmpProtectedPath)
	if err != nil {
		return "", "", fmt.Errorf("read protected PDF: %w", err)
	}

	artifactID := uuid.New().String()
	s3Key := fmt.Sprintf("reports/%s.pdf", artifactID)
	if err := uc.storage.Put(ctx, s3Key, pdfData, "application/pdf"); err != nil {
		return "", "", fmt.Errorf("store protected PDF: %w", err)
	}

	utils.LogInfo("SecurePDFUseCase: protected PDF stored at S3 key %s", s3Key)

	return s3Key, password, nil
}
