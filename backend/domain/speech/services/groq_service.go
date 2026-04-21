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
	commonUtils "sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	speechUtils "sensio/domain/speech/utils"
	"strings"
	"time"
)

type GroqService struct {
	config *commonUtils.Config
}

func NewGroqService(cfg *commonUtils.Config) *GroqService {
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
		commonUtils.LogWarn("Groq HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (s *GroqService) CallModel(ctx context.Context, prompt string, model string) (string, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("failed to create groq request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.GroqApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
		return "", commonUtils.NewAPIError(resp.StatusCode, fmt.Sprintf("groq api returned status %d: %s", resp.StatusCode, string(body)))
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
	commonUtils.LogDebug("Groq: Response received: %s", result)
	return result, nil
}

// Whisper Implementation

// GroqDirectUploadLimitBytes is the maximum file size for direct Groq Whisper uploads.
// Groq does not document a specific limit, so we use 25MB as a safe default
// to match OpenAI's limit and avoid remote 413 errors.
const GroqDirectUploadLimitBytes = 25 * 1024 * 1024

func (s *GroqService) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	if s.config.GroqApiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is not configured")
	}

	// Preflight size check to avoid sending oversized files to Groq
	fileInfo, err := os.Stat(audioPath)
	if err == nil && fileInfo.Size() > GroqDirectUploadLimitBytes {
		return nil, fmt.Errorf("file size (%d bytes) exceeds Groq direct upload limit (%d bytes); use segmented transcription path", fileInfo.Size(), GroqDirectUploadLimitBytes)
	}

	url := "https://api.groq.com/openai/v1/audio/transcriptions"
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Create cancellable context to signal goroutine to stop on early error
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel() // Signal main function that goroutine is done
		defer pw.Close()
		defer writer.Close()

		select {
		case <-ctx.Done():
			return // Goroutine was cancelled, don't continue
		default:
		}

		file, err := os.Open(audioPath)
		if err != nil {
			commonUtils.LogError("Groq Transcribe: failed to open file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
		if err != nil {
			commonUtils.LogError("Groq Transcribe: failed to create form file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			commonUtils.LogError("Groq Transcribe: failed to copy file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		model := s.config.GroqModelWhisper
		if model == "" {
			model = "whisper-large-v3"
		}

		if err := writer.WriteField("model", model); err != nil {
			commonUtils.LogError("Groq Transcribe: failed to write model field: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if language != "" && language != "auto" {
			if err := writer.WriteField("language", language); err != nil {
				commonUtils.LogError("Groq Transcribe: failed to write language field: %v", err)
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", url, pr)
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
		return nil, commonUtils.NewAPIError(resp.StatusCode, fmt.Sprintf("groq whisper returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode json response: %w", err)
	}

	transcription := strings.TrimSpace(result.Text)

	// Groq Whisper doesn't provide diarization structure, but we can
	// attempt to parse speaker labels if they exist in the text
	var utterances []dtos.Utterance
	var transcriptFormat dtos.TranscriptFormat

	if diarize {
		utterances = speechUtils.ParseUtterancesFromText(transcription)
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
		DetectedLanguage:  language,
		Diarized:          diarized,
		Source:            "Groq Whisper",
		Utterances:        utterances,
		TranscriptFormat:  transcriptFormat,
		ConfidenceSummary: speechUtils.BuildConfidenceSummary(utterances, 1),
	}, nil
}
