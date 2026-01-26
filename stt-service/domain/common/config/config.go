package config

import (
	"fmt"
	"os"
)

type Config struct {
	OllamaURL        string
	LLMModel         string
	WhisperModelPath string
	MaxFileSize      int64
	Port             string
}

func LoadConfig() (*Config, error) {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		return nil, fmt.Errorf("OLLAMA_URL is required")
	}

	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		return nil, fmt.Errorf("LLM_MODEL is required")
	}

	whisperModelPath := os.Getenv("WHISPER_MODEL_PATH")
	if whisperModelPath == "" {
		return nil, fmt.Errorf("WHISPER_MODEL_PATH is required")
	}

	maxFileSizeStr := os.Getenv("MAX_FILE_SIZE_MB")
	var maxFileSize int64 = 25 // Default 25MB if not specified? No, user said "hindari pake default"
	if maxFileSizeStr == "" {
		return nil, fmt.Errorf("MAX_FILE_SIZE_MB is required")
	}
	fmt.Sscanf(maxFileSizeStr, "%d", &maxFileSize)

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT is required")
	}

	return &Config{
		OllamaURL:        ollamaURL,
		LLMModel:         llmModel,
		WhisperModelPath: whisperModelPath,
		MaxFileSize:      maxFileSize * 1024 * 1024,
		Port:             port,
	}, nil
}
