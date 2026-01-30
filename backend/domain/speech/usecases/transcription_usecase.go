package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/repositories"
)

type TranscriptionUsecase struct {
	whisperRepo *repositories.WhisperRepository
	ollamaRepo  *repositories.OllamaRepository
	mqttRepo    *repositories.MqttRepository
	config      *utils.Config
}

func NewTranscriptionUsecase(whisperRepo *repositories.WhisperRepository, ollamaRepo *repositories.OllamaRepository, mqttRepo *repositories.MqttRepository, cfg *utils.Config) *TranscriptionUsecase {
	return &TranscriptionUsecase{
		whisperRepo: whisperRepo,
		ollamaRepo:  ollamaRepo,
		mqttRepo:    mqttRepo,
		config:      cfg,
	}
}

func (u *TranscriptionUsecase) PublishToWhisper(message string) error {
	return u.mqttRepo.Publish(message)
}

func (u *TranscriptionUsecase) StartListening() {
	err := u.mqttRepo.Subscribe(func(payload []byte) {
		utils.LogDebug("🔊 Whisper received audio payload: %d bytes", len(payload))

		// If payload is very small, it might look like text command
		if len(payload) < 256 {
			msg := string(payload)
			// Simple check if it's likely text (printable ascii)
			isText := true
			for _, b := range msg {
				if b < 32 || b > 126 {
					isText = false
					break
				}
			}
			if isText {
				utils.LogInfo("🔊 Received text command: %s", msg)
				return
			}
		}

		// Assume it's audio. Save to temp file.
		// We don't know the format yet. Android MediaRecorder usually uses AAC/M4A or 3GP.
		// Converting from ANY recognizable format to WAV is handled by ffmpeg in ConvertToWav.
		// Let's name it .bin or try .m4a as default container hint for ffmpeg.
		tempDir := "./tmp"
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			os.Mkdir(tempDir, 0755)
		}

		filePath := filepath.Join(tempDir, fmt.Sprintf("mqtt_audio_%s.m4a", utils.GenerateUUID()))
		if err := os.WriteFile(filePath, payload, 0644); err != nil {
			utils.LogError("Failed to write temp audio file: %v", err)
			return
		}
		defer os.Remove(filePath)

		utils.LogDebug("Saved audio to %s, processing...", filePath)

		text, err := u.TranscribeAudio(filePath)
		if err != nil {
			utils.LogError("Transcription failed: %v", err)
			// Publish failure?
			return
		}

		utils.LogInfo("🎯 Transcription Result: %s", text)

		// Publish result back
		if err := u.PublishToWhisper("Result: " + text); err != nil {
			utils.LogError("Failed to publish result: %v", err)
		}
	})
	if err != nil {
		utils.LogError("Failed to register MQTT callback: %v", err)
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
