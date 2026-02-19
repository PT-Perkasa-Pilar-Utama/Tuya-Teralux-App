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

type OrionService struct {
	config *utils.Config
}

func NewOrionService(cfg *utils.Config) *OrionService {
	return &OrionService{
		config: cfg,
	}
}

// LLM Implementation

type orionResponse struct {
	ID     string        `json:"id"`
	Output []orionOutput `json:"output"`
	Status string        `json:"status"`
}

type orionOutput struct {
	Type    string         `json:"type"`
	Content []orionContent `json:"content"`
}

type orionContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *OrionService) HealthCheck() bool {
	if s.config.OrionBaseURL == "" || s.config.OrionApiKey == "" {
		return false
	}

	// Remove trailing slash from baseURL if present
	baseURL := s.config.OrionBaseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	url := fmt.Sprintf("%s/v1/models", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+s.config.OrionApiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogWarn("Orion HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		utils.LogWarn("Orion HealthCheck failed: status %d", resp.StatusCode)
		return false
	}
	return true
}

func (s *OrionService) CallModel(prompt string, model string) (string, error) {
	if s.config.OrionApiKey == "" {
		return "", fmt.Errorf("ORION_API_KEY is not configured")
	}

	// Remove trailing slash from baseURL if present
	baseURL := s.config.OrionBaseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	url := fmt.Sprintf("%s/v1/responses", baseURL)

	// Use configured model regardless of input (strict Orion behavior as requested)
	reqBody := map[string]interface{}{
		"model": s.config.OrionModel,
		"input": prompt,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal orion request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("failed to create orion request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.OrionApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call orion api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read orion response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("orion api returned status %d: %s", resp.StatusCode, string(body))
	}

	var orionResp orionResponse
	if err := json.Unmarshal(body, &orionResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal orion response: %w", err)
	}

	// Extract text
	for _, output := range orionResp.Output {
		if output.Type == "message" {
			for _, content := range output.Content {
				if content.Type == "output_text" && content.Text != "" {
					utils.LogDebug("Orion: Response received: %s", content.Text)
					return content.Text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("orion api returned no text content")
}

// Whisper Implementation (PPU)

type OrionWhisperResponse struct {
	Text string `json:"text"`
}

func (s *OrionService) WhisperHealthCheck() bool {
	if s.config.OrionWhisperBaseURL == "" {
		return false
	}
	url := fmt.Sprintf("%s/health", s.config.OrionWhisperBaseURL)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

// Transcribe implements the usecases.WhisperClient interface
func (s *OrionService) Transcribe(audioPath string, lang string) (*dtos.WhisperResult, error) {
	url := fmt.Sprintf("%s/inference", s.config.OrionWhisperBaseURL)

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

	_ = writer.WriteField("response_format", "json")
	if lang != "" && lang != "auto" {
		_ = writer.WriteField("language", lang)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to whisper server failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("whisper server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result OrionWhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode json response: %w", err)
	}

	return &dtos.WhisperResult{
		Transcription:    strings.TrimSpace(result.Text),
		DetectedLanguage: lang,
		Source:           "Orion Whisper (PPU)",
	}, nil
}
