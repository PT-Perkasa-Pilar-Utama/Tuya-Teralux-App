package usecases

import "teralux_app/domain/speech/dtos"

// WhisperClient is the unified interface for all whisper transcription services
type WhisperClient interface {
	Transcribe(audioPath string, language string) (*dtos.WhisperResult, error)
}
