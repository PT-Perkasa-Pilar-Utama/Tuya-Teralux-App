package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sensio/domain/common/utils"
	"sensio/domain/download_token/entities"
	"sensio/domain/infrastructure"
)

// DownloadClaims represents the JWT claims for download token
type DownloadClaims struct {
	jwt.RegisteredClaims
	State     string `json:"state"`
	Client    string `json:"client"`
	Purpose   string `json:"purpose"`
	ObjectKey string `json:"object_key"`
	Recipient string `json:"recipient"`
}

type DownloadTokenService struct {
	storageProvider  infrastructure.StorageProvider
	presignTTLSecond int64
	now              func() time.Time
}

func NewDownloadTokenService(storageProvider infrastructure.StorageProvider) *DownloadTokenService {
	return &DownloadTokenService{
		storageProvider:  storageProvider,
		presignTTLSecond: 300,
		now:              func() time.Time { return time.Now().UTC() },
	}
}

func (s *DownloadTokenService) CreateToken(recipient, objectKey, purpose string, password ...string) (string, error) {
	recipient = strings.TrimSpace(recipient)
	objectKey = strings.TrimSpace(objectKey)
	purpose = strings.TrimSpace(purpose)

	if recipient == "" {
		return "", fmt.Errorf("recipient is required")
	}
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}
	if purpose != "audio_zip" && purpose != "summary_pdf" {
		return "", fmt.Errorf("invalid purpose")
	}

	config := utils.GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not configured")
	}

	now := s.now()
	claims := DownloadClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.AddDate(1, 0, 0)),
		},
		State:     recipient,
		Client:    "default",
		Purpose:   purpose,
		ObjectKey: objectKey,
		Recipient: recipient,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *DownloadTokenService) ResolveToken(tokenString, client, purpose string) (string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", entities.ErrTokenNotFound
	}

	config := utils.GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &DownloadClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", entities.ErrTokenExpired
		}
		return "", entities.ErrTokenNotFound
	}

	claims, ok := token.Claims.(*DownloadClaims)
	if !ok || !token.Valid {
		return "", entities.ErrTokenNotFound
	}

	if purpose != "" && claims.Purpose != purpose {
		return "", fmt.Errorf("purpose mismatch")
	}
	if client != "" && claims.Client != client {
		return "", fmt.Errorf("client mismatch")
	}

	if s.storageProvider == nil {
		return "", fmt.Errorf("storage provider is not configured")
	}

	ctx := context.Background()

	if err := s.storageProvider.Head(ctx, claims.ObjectKey); err != nil {
		return "", fmt.Errorf("object %s: %w", claims.ObjectKey, entities.ErrObjectNotFound)
	}

	signedURL, err := s.storageProvider.PresignPut(ctx, claims.ObjectKey, "application/octet-stream", s.presignTTLSecond)
	if err != nil {
		return "", fmt.Errorf("presign object key %s: %w", claims.ObjectKey, err)
	}

	return signedURL, nil
}

func (s *DownloadTokenService) RevokeToken(tokenID string) error {
	return nil
}
