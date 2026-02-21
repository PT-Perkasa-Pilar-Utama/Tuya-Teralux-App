package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
)

type GeminiService struct {
	apiKey string
	config *utils.Config
}

func NewGeminiService(cfg *utils.Config) *GeminiService {
	return &GeminiService{
		apiKey: cfg.GeminiApiKey,
		config: cfg,
	}
}

// LLM Implementation

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (s *GeminiService) HealthCheck() bool {
	if s.apiKey == "" {
		return false
	}

	// Quick test with models list endpoint
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", s.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		utils.LogWarn("Gemini HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		utils.LogWarn("Gemini HealthCheck failed: status %d", resp.StatusCode)
		return false
	}
	return true
}

func (s *GeminiService) CallModel(prompt string, model string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	// Map abstract model names to actual models from config
	actualModel := model
	switch {
	case model == "high":
		actualModel = s.config.GeminiModelHigh
	case model == "low":
		actualModel = s.config.GeminiModelLow
	case model == "default" || model == "":
		actualModel = s.config.GeminiModelLow
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", actualModel, s.apiKey)
	utils.LogDebug("Gemini: Calling URL: https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", actualModel)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("failed to call gemini api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", utils.NewAPIError(resp.StatusCode, fmt.Sprintf("gemini api returned status %d: %s", resp.StatusCode, string(body)))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini api returned no candidates")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	utils.LogDebug("Gemini: Response received: %s", responseText)
	return responseText, nil
}

// Whisper Implementation

func (s *GeminiService) Transcribe(audioPath string, language string) (*dtos.WhisperResult, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	// Detect mime type based on extension
	ext := filepath.Ext(audioPath)
	mimeType := "audio/mp3" // default
	switch ext {
	case ".wav":
		mimeType = "audio/wav"
	case ".ogg":
		mimeType = "audio/ogg"
	case ".m4a":
		mimeType = "audio/m4a"
	}

	// Build prompt
	promptText := "Transcribe this audio file exactly as spoken."
	if language != "" {
		promptText += fmt.Sprintf(" The language is %s.", language)
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": promptText},
					{
						"inline_data": map[string]string{
							"mime_type": mimeType,
							"data":      base64.StdEncoding.EncodeToString(audioData),
						},
					},
				},
			},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use configured Whisper model
	model := s.config.GeminiModelWhisper
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, s.apiKey)

	client := &http.Client{Timeout: 0}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewAPIError(resp.StatusCode, fmt.Sprintf("gemini api returned status %d: %s", resp.StatusCode, string(body)))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini api returned no candidates")
	}

	transcription := geminiResp.Candidates[0].Content.Parts[0].Text

	return &dtos.WhisperResult{
		Transcription:    transcription,
		DetectedLanguage: language,
		Source:           "Gemini Whisper",
	}, nil
}
