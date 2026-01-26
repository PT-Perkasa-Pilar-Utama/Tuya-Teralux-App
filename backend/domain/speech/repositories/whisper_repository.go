package repositories

import "fmt"

type WhisperRepository struct {
}

func NewWhisperRepository() *WhisperRepository {
	return &WhisperRepository{}
}

func (r *WhisperRepository) Transcribe(wavPath string, modelPath string) (string, error) {
	// Placeholder: call to whisper binary / library
	return "transcribed text", nil
}

func (r *WhisperRepository) Convert(wavPath string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
