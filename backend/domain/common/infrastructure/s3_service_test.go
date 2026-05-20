package infrastructure

import (
	"testing"

	"sensio/domain/common/utils"
)

func TestBuildObjectKey(t *testing.T) {
	s := &S3Service{prefix: "Sensio", bucket: "test-bucket", ttlSeconds: 300}

	tests := []struct {
		name     string
		category string
		filename string
		want     string
	}{
		{
			name:     "basic audio file",
			category: "audio",
			filename: "test.wav",
			want:     "Sensio/audio/test.wav",
		},
		{
			name:     "recordings category",
			category: "recordings",
			filename: "uuid-123.mp3",
			want:     "Sensio/recordings/uuid-123.mp3",
		},
		{
			name:     "file in subcategory",
			category: "audio",
			filename: "file.wav",
			want:     "Sensio/audio/file.wav",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.BuildObjectKey(tt.category, tt.filename)
			if got != tt.want {
				t.Errorf("BuildObjectKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildObjectKey_NilReceiver(t *testing.T) {
	var s *S3Service
	got := s.BuildObjectKey("audio", "test.wav")
	if got != "" {
		t.Errorf("BuildObjectKey() on nil = %v, want empty string", got)
	}
}

func TestGeneratePresignedPutURL_NilReceiver(t *testing.T) {
	var s *S3Service
	url, err := s.GeneratePresignedPutURL("audio/test.wav", "audio/wav")
	if url != "" || err != nil {
		t.Errorf("GeneratePresignedPutURL() on nil = %v, %v; want empty, nil", url, err)
	}
}

func TestGeneratePresignedGetURL_NilReceiver(t *testing.T) {
	var s *S3Service
	url, err := s.GeneratePresignedGetURL("audio/test.wav")
	if url != "" || err != nil {
		t.Errorf("GeneratePresignedGetURL() on nil = %v, %v; want empty, nil", url, err)
	}
}

func TestDeleteObject_NilReceiver(t *testing.T) {
	var s *S3Service
	err := s.DeleteObject("audio/test.wav")
	if err != nil {
		t.Errorf("DeleteObject() on nil = %v; want nil", err)
	}
}

func TestNewS3ServiceFromConfig_Disabled(t *testing.T) {
	cfg := utils.Config{
		S3Enabled: false,
	}
	svc, err := NewS3ServiceFromConfig(cfg)
	if svc != nil || err != nil {
		t.Errorf("NewS3ServiceFromConfig(S3Enabled=false) = %v, %v; want nil, nil", svc, err)
	}
}