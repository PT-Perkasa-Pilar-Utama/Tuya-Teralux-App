package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"
	ragUsecases "teralux_app/domain/rag/usecases"
	"teralux_app/domain/speech/repositories"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
)

// WhisperRepositoryInterface defines the methods required from the Whisper repository
type WhisperRepositoryInterface interface {
	Transcribe(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
}

type TranscriptionUsecase struct {
	whisperRepo WhisperRepositoryInterface
	ollamaRepo  *repositories.OllamaRepository
	geminiRepo  *repositories.GeminiRepository
	mqttRepo    *repositories.MqttRepository
	ragUsecase  *ragUsecases.RAGUsecase
	authUseCase *tuyaUsecases.TuyaAuthUseCase
	config      *utils.Config
}

func NewTranscriptionUsecase(
	whisperRepo WhisperRepositoryInterface,
	ollamaRepo *repositories.OllamaRepository,
	geminiRepo *repositories.GeminiRepository,
	mqttRepo *repositories.MqttRepository,
	cfg *utils.Config,
	ragUsecase *ragUsecases.RAGUsecase,
	authUseCase *tuyaUsecases.TuyaAuthUseCase,
) *TranscriptionUsecase {
	return &TranscriptionUsecase{
		whisperRepo: whisperRepo,
		ollamaRepo:  ollamaRepo,
		geminiRepo:  geminiRepo,
		mqttRepo:    mqttRepo,
		ragUsecase:  ragUsecase,
		authUseCase: authUseCase,
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

			// Loop prevention: ignore our own results
			if strings.HasPrefix(msg, "Result: ") {
				return
			}

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
				u.HandleCommand(msg)
				return
			}
		}

		// Assume it's audio. Save to temp file.
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
			return
		}

		utils.LogInfo("🎯 Transcription Result: %s", text)

		// Publish result back
		if err := u.PublishToWhisper("Result: " + text); err != nil {
			utils.LogError("Failed to publish result: %v", err)
		}

		// 1. Translate if needed (e.g. from Indonesian to English)
		translatedText, err := u.TranslateToEnglish(text)
		if err != nil {
			utils.LogWarn("Translation failed, using original text: %v", err)
			translatedText = text
		} else {
			utils.LogInfo("🌐 Translated Result: %s", translatedText)
			// Publish translated result back
			if err := u.PublishToWhisper("Translated: " + translatedText); err != nil {
				utils.LogError("Failed to publish translated result: %v", err)
			}
		}

		// 2. Process via RAG (using English text)
		u.HandleCommand(translatedText)
	})
	if err != nil {
		utils.LogError("Failed to register MQTT callback: %v", err)
	}
}

func (u *TranscriptionUsecase) TranslateToEnglish(text string) (string, error) {
	// Simple check: if text is basic ascii and seems like english, maybe skip?
	// But it's safer to always ask LLM to translate or refine.
	prompt := fmt.Sprintf(`You are a translator. Translate the following Indonesian smart home command to a clear, concise English command. 
If it is already in English, just return it as is.
Only return the translated text without any explanation or quotes.

Indonesian: "%s"
English:`, text)

	var translated string
	var err error

	if u.config.LLMProvider == "ollama" {
		if u.ollamaRepo == nil {
			return text, fmt.Errorf("ollama repo not initialized")
		}
		translated, err = u.ollamaRepo.CallModel(prompt, u.config.LLMModel)
	} else {
		// Default to gemini if provider is gemini or empty
		if u.geminiRepo == nil {
			return text, fmt.Errorf("gemini repo not initialized")
		}
		translated, err = u.geminiRepo.CallModel(prompt, u.config.LLMModel)
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(translated), nil
}

func (u *TranscriptionUsecase) HandleCommand(text string) {
	// Filter out common non-speech results
	cleanText := strings.TrimSpace(text)
	if cleanText == "" || cleanText == "[BLANK_AUDIO]" {
		utils.LogDebug("Speech: Ignoring blank or empty command")
		return
	}

	if u.ragUsecase == nil || u.authUseCase == nil {
		utils.LogWarn("Speech: RAG or Auth Usecase not initialized, skipping processing")
		return
	}

	// 1. Get Auth Token
	auth, err := u.authUseCase.Authenticate()
	if err != nil {
		utils.LogError("Speech: Failed to authenticate for RAG: %v", err)
		return
	}

	// 2. Process via RAG
	utils.LogInfo("Speech: Processing command via RAG: %q", text)
	taskID, err := u.ragUsecase.Process(text, auth.AccessToken)
	if err != nil {
		utils.LogError("Speech: Failed to trigger RAG processing: %v", err)
		return
	}

	utils.LogInfo("Speech: RAG processing triggered (TaskID: %s)", taskID)
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

	// Transcribe with "id" for Indonesian support as requested
	// You can also use "auto" for auto detection
	text, err := u.whisperRepo.Transcribe(wavPath, modelPath, "id")
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, nil
}

func (u *TranscriptionUsecase) TranscribeLongAudio(inputPath string, lang string) (string, error) {
	// Create temp directory for conversion if not exists
	tempDir := filepath.Dir(inputPath)

	// Convert to WAV if needed (Whisper needs 16kHz mono WAV)
	wavPath := filepath.Join(tempDir, "processed_long.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	// Use model path from config
	modelPath := u.config.WhisperModelPath

	// Use TranscribeFull to get all text
	text, err := u.whisperRepo.TranscribeFull(wavPath, modelPath, lang)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, nil
}
