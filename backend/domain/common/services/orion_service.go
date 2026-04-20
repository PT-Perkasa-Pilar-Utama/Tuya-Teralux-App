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
	"sensio/domain/models/whisper/dtos"
	"strings"
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

func (s *OrionService) CallModel(ctx context.Context, prompt string, model string) (string, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
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
		return "", utils.NewAPIError(resp.StatusCode, fmt.Sprintf("orion api returned status %d: %s", resp.StatusCode, string(body)))
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

// Whisper Implementation (Orion)

// OrionDirectUploadLimitBytes is the maximum file size for direct Orion Whisper uploads.
// Orion's backend limit is not explicitly documented, so we use 20MB as a conservative
// safe default to match Gemini's inline limit and avoid remote timeouts/rejections.
const OrionDirectUploadLimitBytes = 20 * 1024 * 1024

const OrionWhisperTranscribePath = "/v1/audio/transcriptions"
const OrionHealthCheckPath = "/v1/models"

type OrionWhisperResponse struct {
	Text string `json:"text"`
}

func (s *OrionService) WhisperHealthCheck() bool {
	baseURL := s.config.OrionWhisperBaseURL
	if baseURL == "" {
		utils.LogError("Orion: ORION_WHISPER_BASE_URL not configured")
		return false
	}

	healthCheckURL := strings.TrimSuffix(baseURL, "/") + OrionHealthCheckPath

	req, err := http.NewRequest("GET", healthCheckURL, nil)
	if err != nil {
		utils.LogError("Orion: Failed to create health check request: %v", err)
		return false
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogWarn("Orion: Health check request failed - server unreachable: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (s *OrionService) Transcribe(ctx context.Context, audioPath string, lang string, diarize bool) (*dtos.WhisperResult, error) {
	baseURL := s.config.OrionWhisperBaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("ORION_WHISPER_BASE_URL not configured")
	}

	transcribeURL := strings.TrimSuffix(baseURL, "/") + OrionWhisperTranscribePath

	fileInfo, err := os.Stat(audioPath)
	if err == nil && fileInfo.Size() > OrionDirectUploadLimitBytes {
		return nil, fmt.Errorf("file size (%d bytes) exceeds Orion direct upload limit (%d bytes); use segmented transcription path", fileInfo.Size(), OrionDirectUploadLimitBytes)
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		file, err := os.Open(audioPath)
		if err != nil {
			utils.LogError("Orion Transcribe: failed to open file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
		if err != nil {
			utils.LogError("Orion Transcribe: failed to create form file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			utils.LogError("Orion Transcribe: failed to copy file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if err := writer.WriteField("language", lang); err != nil {
			utils.LogError("Orion Transcribe: failed to write language field: %v", err)
			_ = pw.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", transcribeURL, pr)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	// Set multipart content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	timeout, err := time.ParseDuration(s.config.OrionTranscribeTimeout)
	if err != nil {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transcribe request to Orion failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read failed: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Orion server returned error status %d: %s", resp.StatusCode, string(respBody))
		structuredErr := utils.MapOrionErrorToCode(resp.StatusCode, string(respBody), errMsg)
		return nil, utils.NewOrionTranscribeError(resp.StatusCode, structuredErr)
	}

	// Parse response as JSON
	var result dtos.OutsystemsTranscriptionResultDTO
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Orion response: %w", err)
	}

	detectedLang := result.DetectedLanguage
	if detectedLang == "" {
		detectedLang = lang
	}

	transcription := result.Transcription

	// Orion/Outsystems doesn't provide diarization structure, but we can
	// attempt to parse speaker labels if they exist in the text
	var utterances []dtos.Utterance
	var transcriptFormat dtos.TranscriptFormat

	if diarize {
		utterances = utils.ParseUtterancesFromText(transcription)
		if len(utterances) > 0 {
			transcriptFormat = dtos.TranscriptFormatUtteranceList
		}
	}

	if len(utterances) == 0 {
		transcriptFormat = dtos.TranscriptFormatPlainText
	}

	// Diarized is true ONLY if actual speaker-labeled utterances were extracted
	diarized := diarize && len(utterances) > 0

	return &dtos.WhisperResult{
		Transcription:     transcription,
		DetectedLanguage:  detectedLang,
		Diarized:          diarized,
		Source:            "Orion (Outsystems)",
		Utterances:        utterances,
		TranscriptFormat:  transcriptFormat,
		ConfidenceSummary: utils.BuildConfidenceSummary(utterances, 1),
	}, nil
}
