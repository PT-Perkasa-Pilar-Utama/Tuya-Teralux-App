package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	stypes "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"sensio/domain/common/utils"
)

// StorageProvider defines generic object storage operations.
// It intentionally parallels FileService style without replacing it.
type StorageProvider interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	PresignPut(ctx context.Context, key string, contentType string, ttlSeconds int64) (string, error)
}

type s3StorageProvider struct {
	client        *s3.Client
	bucket        string
	prefix        string
	localFallback StorageProvider
}

type localStorageProvider struct {
	baseDir string
}

// NewStorageProvider returns S3 when enabled, otherwise local storage.
func NewStorageProvider(cfg *utils.Config) (StorageProvider, error) {
	if cfg == nil || !cfg.S3Enabled {
		return NewLocalStorageProvider("uploads"), nil
	}

	return NewS3StorageProvider(cfg)
}

// NewS3StorageProvider creates an S3-backed storage provider.
func NewS3StorageProvider(cfg *utils.Config) (StorageProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("s3 config is nil")
	}

	if strings.TrimSpace(cfg.S3Bucket) == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}
	if strings.TrimSpace(cfg.S3Region) == "" {
		return nil, fmt.Errorf("s3 region is required")
	}

	loadOptions := []func(*config.LoadOptions) error{config.WithRegion(cfg.S3Region)}
	if strings.TrimSpace(cfg.AWSAccessKeyID) != "" && strings.TrimSpace(cfg.AWSSecretAccessKey) != "" {
		loadOptions = append(loadOptions,
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, "")),
		)
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws s3 config: %w", err)
	}

	provider := &s3StorageProvider{
		client:        s3.NewFromConfig(awsCfg),
		bucket:        cfg.S3Bucket,
		prefix:        enforceSensioPrefix(cfg.S3Prefix),
		localFallback: NewLocalStorageProvider("uploads"),
	}

	return provider, nil
}

// NewLocalStorageProvider creates a local filesystem storage provider.
func NewLocalStorageProvider(baseDir string) StorageProvider {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		baseDir = "uploads"
	}

	return &localStorageProvider{baseDir: baseDir}
}

func (s *s3StorageProvider) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if len(data) == 0 {
		return fmt.Errorf("put object: empty payload")
	}

	s3Key := s.objectKey(key)
	if s3Key == "" {
		return fmt.Errorf("put object: key is required")
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s3Key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("put object %s: %w", s3Key, err)
	}

	return nil
}

func (s *s3StorageProvider) Get(ctx context.Context, key string) ([]byte, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("get object: key is required")
	}

	if looksLikeLegacyUploadsPath(key) {
		if data, err := s.localFallback.Get(ctx, key); err == nil {
			return data, nil
		}
	}

	s3Key := s.objectKey(key)
	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		var noSuchKey *stypes.NoSuchKey
		if errors.As(err, &noSuchKey) && looksLikeLegacyUploadsPath(key) {
			return s.localFallback.Get(ctx, key)
		}
		return nil, fmt.Errorf("get object %s: %w", s3Key, err)
	}
	defer func() { _ = res.Body.Close() }()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read object %s body: %w", s3Key, readErr)
	}

	return body, nil
}

func (s *s3StorageProvider) Delete(ctx context.Context, key string) error {
	s3Key := s.objectKey(key)
	if s3Key == "" {
		return fmt.Errorf("delete object: key is required")
	}

	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}); err != nil {
		return fmt.Errorf("delete object %s: %w", s3Key, err)
	}

	return nil
}

func (s *s3StorageProvider) PresignPut(ctx context.Context, key string, contentType string, ttlSeconds int64) (string, error) {
	s3Key := s.objectKey(key)
	if s3Key == "" {
		return "", fmt.Errorf("presign put object: key is required")
	}

	if ttlSeconds <= 0 {
		ttlSeconds = 900
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}

	presignClient := s3.NewPresignClient(s.client)
	presigned, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(ttlSeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("presign put object %s: %w", s3Key, err)
	}

	return presigned.URL, nil
}

func (l *localStorageProvider) Put(_ context.Context, key string, data []byte, _ string) error {
	if len(data) == 0 {
		return fmt.Errorf("put object: empty payload")
	}

	localPath, err := l.resolveLocalPath(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
		return fmt.Errorf("ensure local directory %s: %w", filepath.Dir(localPath), err)
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return fmt.Errorf("write local object %s: %w", localPath, err)
	}

	return nil
}

func (l *localStorageProvider) Get(_ context.Context, key string) ([]byte, error) {
	localPath, err := l.resolveLocalPath(key)
	if err != nil {
		return nil, err
	}

	body, readErr := os.ReadFile(localPath)
	if readErr != nil {
		return nil, fmt.Errorf("read local object %s: %w", localPath, readErr)
	}

	return body, nil
}

func (l *localStorageProvider) Delete(_ context.Context, key string) error {
	localPath, err := l.resolveLocalPath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(localPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete local object %s: %w", localPath, err)
	}

	return nil
}

func (l *localStorageProvider) PresignPut(_ context.Context, _ string, _ string, _ int64) (string, error) {
	return "", fmt.Errorf("presign put not supported for local storage")
}

func (s *s3StorageProvider) objectKey(key string) string {
	cleanKey := sanitizeObjectKey(key)
	if cleanKey == "" {
		return ""
	}

	if strings.HasPrefix(cleanKey, s.prefix) {
		return cleanKey
	}

	return s.prefix + cleanKey
}

func (l *localStorageProvider) resolveLocalPath(key string) (string, error) {
	cleanKey := sanitizeObjectKey(key)
	if cleanKey == "" {
		return "", fmt.Errorf("local storage key is required")
	}

	if strings.HasPrefix(cleanKey, "uploads/") {
		return cleanKey, nil
	}

	return filepath.Join(l.baseDir, cleanKey), nil
}

func sanitizeObjectKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	clean := path.Clean("/" + strings.TrimPrefix(key, "/"))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." {
		return ""
	}

	return clean
}

func enforceSensioPrefix(prefix string) string {
	prefix = sanitizeObjectKey(prefix)
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		return "Sensio/"
	}

	if !strings.HasPrefix(prefix, "Sensio") {
		prefix = path.Join("Sensio", prefix)
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return prefix
}

func looksLikeLegacyUploadsPath(key string) bool {
	clean := sanitizeObjectKey(key)
	return strings.HasPrefix(clean, "uploads/")
}
