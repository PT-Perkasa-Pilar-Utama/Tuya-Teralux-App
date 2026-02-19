package repositories

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
)

type WhisperOrionRepository struct {
	config *utils.Config
}

func NewWhisperOrionRepository(cfg *utils.Config) *WhisperOrionRepository {
	return &WhisperOrionRepository{
		config: cfg,
	}
}

type WhisperServerResponse struct {
	Text string `json:"text"`
}

// HealthCheck verifies if the Orion Whisper server is reachable
func (r *WhisperOrionRepository) HealthCheck() bool {
	if r.config.WhisperServerURL == "" {
		return false
	}

	url := fmt.Sprintf("%s/health", r.config.WhisperServerURL)
	resp, err := http.Get(url)
	if err != nil {
		utils.LogWarn("Orion Whisper HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		utils.LogWarn("Orion Whisper HealthCheck failed: status %d", resp.StatusCode)
		return false
	}
	return true
}

func (r *WhisperOrionRepository) Transcribe(audioPath string, lang string) (string, error) {
	url := fmt.Sprintf("%s/inference", r.config.WhisperServerURL)

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer func() { _ = file.Close() }()

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add fields
	_ = writer.WriteField("response_format", "json")
	if lang != "" && lang != "auto" {
		_ = writer.WriteField("language", lang)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Send Request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to whisper server failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse Response
	var result WhisperServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode json response: %w", err)
	}

	return strings.TrimSpace(result.Text), nil
}
