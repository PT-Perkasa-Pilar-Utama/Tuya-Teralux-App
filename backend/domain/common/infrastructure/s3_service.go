package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sensio/domain/common/utils"
)

var DefaultS3Service *S3Service

type S3Service struct {
	s3Client   *s3.Client
	bucket     string
	region     string
	prefix     string
	ttlSeconds int
}

func NewS3ServiceFromConfig(cfg utils.Config) (*S3Service, error) {
	if !cfg.S3Enabled {
		return nil, nil
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS_ACCESS_KEY_ID,
			cfg.AWS_SECRET_ACCESS_KEY,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Service{
		s3Client:   client,
		bucket:     cfg.S3Bucket,
		region:     cfg.S3Region,
		prefix:     strings.TrimSuffix(cfg.S3Prefix, "/"),
		ttlSeconds: cfg.S3SignedURLTTLSeconds,
	}, nil
}

func (s *S3Service) BuildObjectKey(category string, filename string) string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s.prefix, category, filename)
}

func (s *S3Service) GeneratePresignedPutURL(objectKey string, contentType string) (string, error) {
	if s == nil || s.s3Client == nil {
		return "", nil
	}
	presignClient := s3.NewPresignClient(s.s3Client)
	req, err := presignClient.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(PresignExpiresFromTTL(s.ttlSeconds)))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return req.URL, nil
}

func (s *S3Service) GeneratePresignedGetURL(objectKey string) (string, error) {
	if s == nil || s.s3Client == nil {
		return "", nil
	}
	presignClient := s3.NewPresignClient(s.s3Client)
	req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(PresignExpiresFromTTL(s.ttlSeconds)))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return req.URL, nil
}

func (s *S3Service) DeleteObject(objectKey string) error {
	if s == nil || s.s3Client == nil {
		return nil
	}
	_, err := s.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (s *S3Service) UploadFile(srcPath, objectKey, contentType string) (string, error) {
	if s == nil || s.s3Client == nil {
		return "", nil
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", srcPath, err)
	}
	_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload %s to S3: %w", objectKey, err)
	}
	utils.LogDebug("S3Service: Uploaded %s -> s3://%s/%s", srcPath, s.bucket, objectKey)
	return objectKey, nil
}

func PresignExpiresFromTTL(ttlSeconds int) time.Duration {
	return time.Duration(ttlSeconds) * time.Second
}