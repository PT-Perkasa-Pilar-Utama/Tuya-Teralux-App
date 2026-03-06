package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sensio/domain/common/utils"
	"sensio/domain/speech/dtos"
	"strings"
	"time"
)

type OpenAIService struct {
	config *utils.Config
}

func NewOpenAIService(cfg *utils.Config) *OpenAIService {
	return &OpenAIService{
		config: cfg,
	}
}

// LLM Implementation

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *OpenAIService) HealthCheck() bool {
	if s.config.OpenAIApiKey == "" {
		return false
	}

	url := "https://api.openai.com/v1/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+s.config.OpenAIApiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogWarn("OpenAI HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (s *OpenAIService) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	if s.config.OpenAIApiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	actualModel := model
	switch {
	case model == "high":
		actualModel = s.config.OpenAIModelHigh
	case model == "low":
		actualModel = s.config.OpenAIModelLow
	case model == "default" || model == "":
		actualModel = s.config.OpenAIModelLow
	}

	if actualModel == "" {
		actualModel = "gpt-3.5-turbo" // Safe default
	}

	url := "https://api.openai.com/v1/chat/completions"
	reqBody := openaiRequest{
		Model: actualModel,
		Messages: []openaiMessage{
			{Role: "user", Content: prompt},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.OpenAIApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call openai api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read openai response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", utils.NewAPIError(resp.StatusCode, fmt.Sprintf("openai api returned status %d: %s", resp.StatusCode, string(body)))
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("openai api returned no choices")
	}

	result := openaiResp.Choices[0].Message.Content
	utils.LogDebug("OpenAI: Response received: %s", result)
	return result, nil
}

// Whisper Implementation

func (s *OpenAIService) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	if s.config.OpenAIApiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	url := "https://api.openai.com/v1/audio/transcriptions"
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		file, err := os.Open(audioPath)
		if err != nil {
			utils.LogError("OpenAI Transcribe: failed to open file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
		if err != nil {
			utils.LogError("OpenAI Transcribe: failed to create form file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			utils.LogError("OpenAI Transcribe: failed to copy file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		model := s.config.OpenAIModelWhisper
		if model == "" {
			model = "whisper-1"
		}

		if err := writer.WriteField("model", model); err != nil {
			utils.LogError("OpenAI Transcribe: failed to write model field: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if language != "" && language != "auto" {
			if err := writer.WriteField("language", language); err != nil {
				utils.LogError("OpenAI Transcribe: failed to write language field: %v", err)
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", url, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.config.OpenAIApiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to openai whisper failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewAPIError(resp.StatusCode, fmt.Sprintf("openai whisper returned status %d: %s", resp.StatusCode, string(respBody)))
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
		Diarized:         false,
		Source:           "OpenAI Whisper",
	}, nil
}
