package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"time"
)

type GroqService struct {
	config *utils.Config
}

func NewGroqService(cfg *utils.Config) *GroqService {
	return &GroqService{
		config: cfg,
	}
}

// LLM Implementation

func (s *GroqService) HealthCheck() bool {
	if s.config.GroqApiKey == "" {
		return false
	}

	url := "https://api.groq.com/openai/v1/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+s.config.GroqApiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogWarn("Groq HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (s *GroqService) CallModel(prompt string, model string) (string, error) {
	if s.config.GroqApiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY is not configured")
	}

	actualModel := model
	switch {
	case model == "high":
		actualModel = s.config.GroqModelHigh
	case model == "low":
		actualModel = s.config.GroqModelLow
	case model == "default" || model == "":
		actualModel = s.config.GroqModelLow
	}

	if actualModel == "" {
		actualModel = "llama3-8b-8192" // Safe default for Groq
	}

	url := "https://api.groq.com/openai/v1/chat/completions"
	reqBody := map[string]interface{}{
		"model": actualModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal groq request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("failed to create groq request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.GroqApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call groq api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read groq response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", utils.NewAPIError(resp.StatusCode, fmt.Sprintf("groq api returned status %d: %s", resp.StatusCode, string(body)))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal groq response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("groq api returned no choices")
	}

	result := response.Choices[0].Message.Content
	utils.LogDebug("Groq: Response received: %s", result)
	return result, nil
}

// Whisper Implementation

func (s *GroqService) Transcribe(audioPath string, language string) (*dtos.WhisperResult, error) {
	if s.config.GroqApiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is not configured")
	}

	url := "https://api.groq.com/openai/v1/audio/transcriptions"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer func() { _ = file.Close() }()

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	model := s.config.GroqModelWhisper
	if model == "" {
		model = "whisper-large-v3"
	}

	_ = writer.WriteField("model", model)
	if language != "" && language != "auto" {
		_ = writer.WriteField("language", language)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.config.GroqApiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to groq whisper failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewAPIError(resp.StatusCode, fmt.Sprintf("groq whisper returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode json response: %w", err)
	}

	return &dtos.WhisperResult{
		Transcription:    strings.TrimSpace(result.Text),
		DetectedLanguage: language,
		Source:           "Groq Whisper",
	}, nil
}
