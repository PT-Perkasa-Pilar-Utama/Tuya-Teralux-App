package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"stt-service/domain/common/config"
	"stt-service/domain/common/utils"
	"stt-service/domain/speech/repositories"
)

type TranscriptionUsecase struct {
	whisperRepo *repositories.WhisperRepository
	ollamaRepo  *repositories.OllamaRepository
	config      *config.Config
}

func NewTranscriptionUsecase(whisperRepo *repositories.WhisperRepository, ollamaRepo *repositories.OllamaRepository, cfg *config.Config) *TranscriptionUsecase {
	return &TranscriptionUsecase{
		whisperRepo: whisperRepo,
		ollamaRepo:  ollamaRepo,
		config:      cfg,
	}
}

func (u *TranscriptionUsecase) TranscribeAudio(inputPath string) (string, error) {
	// Create temp directory for conversion if not exists
	tempDir := filepath.Dir(inputPath)

	// Convert to WAV if needed (Whisper needs 16kHz mono WAV)
	wavPath := filepath.Join(tempDir, "processed.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	// Use model path from config
	modelPath := u.config.WhisperModelPath

	// Transcribe
	text, err := u.whisperRepo.Transcribe(wavPath, modelPath)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, nil
}
