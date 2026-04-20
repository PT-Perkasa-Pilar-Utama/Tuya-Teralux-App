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

type OpenAIService struct {
	config *commonUtils.Config
}

func NewOpenAIService(cfg *commonUtils.Config) *OpenAIService {
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
		commonUtils.LogWarn("OpenAI HealthCheck failed: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

func (s *OpenAIService) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	if s.config.OpenAIApiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	promptChars := len(prompt)
	approxTokens := (promptChars + 3) / 4

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
	startTime := time.Now()
	ctxDeadline := "none"
	if deadline, ok := ctx.Deadline(); ok {
		ctxDeadline = time.Until(deadline).String()
	}
	commonUtils.LogDebug(
		"OpenAI CallModel: model=%s prompt_chars=%d approx_tokens=%d client_timeout=%s ctx_deadline=%s",
		actualModel,
		promptChars,
		approxTokens,
		(60 * time.Second).String(),
		ctxDeadline,
	)
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
		commonUtils.LogWarn(
			"OpenAI CallModel failed: model=%s duration=%s prompt_chars=%d approx_tokens=%d err=%v",
			actualModel,
			time.Since(startTime),
			promptChars,
			approxTokens,
			err,
		)
		return "", fmt.Errorf("failed to call openai api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		commonUtils.LogWarn(
			"OpenAI CallModel read body failed: model=%s duration=%s prompt_chars=%d approx_tokens=%d err=%v",
			actualModel,
			time.Since(startTime),
			promptChars,
			approxTokens,
			err,
		)
		return "", fmt.Errorf("failed to read openai response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		commonUtils.LogWarn(
			"OpenAI CallModel non-200: model=%s status=%d duration=%s resp_bytes=%d",
			actualModel,
			resp.StatusCode,
			time.Since(startTime),
			len(body),
		)
		return "", commonUtils.NewAPIError(resp.StatusCode, fmt.Sprintf("openai api returned status %d: %s", resp.StatusCode, string(body)))
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("openai api returned no choices")
	}

	result := openaiResp.Choices[0].Message.Content
	commonUtils.LogDebug(
		"OpenAI CallModel success: model=%s status=%d duration=%s resp_bytes=%d",
		actualModel,
		resp.StatusCode,
		time.Since(startTime),
		len(body),
	)
	commonUtils.LogDebug("OpenAI: Response received: %s", result)
	return result, nil
}

// Whisper Implementation

// OpenAIDirectUploadLimitBytes is the maximum file size for direct OpenAI Whisper uploads.
// OpenAI Whisper has a strict 25MB limit for direct uploads.
const OpenAIDirectUploadLimitBytes = 25 * 1024 * 1024

func (s *OpenAIService) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	if s.config.OpenAIApiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	// OpenAI Whisper has a 25MB strict limit for direct uploads.
	fileInfo, err := os.Stat(audioPath)
	if err == nil && fileInfo.Size() > OpenAIDirectUploadLimitBytes {
		return nil, fmt.Errorf("file size (%d bytes) exceeds OpenAI direct upload limit (%d bytes); use segmented transcription path", fileInfo.Size(), OpenAIDirectUploadLimitBytes)
	}

	url := "https://api.openai.com/v1/audio/transcriptions"
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

		// 1. Write metadata fields first (best practice for multipart)
		model := s.config.OpenAIModelWhisper
		if model == "" {
			model = "whisper-1"
		}

		if err := writer.WriteField("model", model); err != nil {
			commonUtils.LogError("OpenAI Transcribe: failed to write model field: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if language != "" && language != "auto" {
			if err := writer.WriteField("language", language); err != nil {
				commonUtils.LogError("OpenAI Transcribe: failed to write language field: %v", err)
				_ = pw.CloseWithError(err)
				return
			}
		}

		// 2. Write the file field last
		file, err := os.Open(audioPath)
		if err != nil {
			commonUtils.LogError("OpenAI Transcribe: failed to open file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
		if err != nil {
			commonUtils.LogError("OpenAI Transcribe: failed to create form file: %v", err)
			_ = pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			commonUtils.LogError("OpenAI Transcribe: failed to copy file: %v", err)
			_ = pw.CloseWithError(err)
			return
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
		return nil, commonUtils.NewAPIError(resp.StatusCode, fmt.Sprintf("openai whisper returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode json response: %w", err)
	}

	transcription := strings.TrimSpace(result.Text)

	// OpenAI Whisper doesn't provide diarization structure, but we can
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
		Source:            "OpenAI Whisper",
		Utterances:        utterances,
		TranscriptFormat:  transcriptFormat,
		ConfidenceSummary: speechUtils.BuildConfidenceSummary(utterances, 1),
	}, nil
}
