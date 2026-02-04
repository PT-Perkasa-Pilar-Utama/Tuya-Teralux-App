package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"teralux_app/domain/common/utils"
)

type AntigravityRepository struct {
	baseURL string
	apiKey  string
}

func NewAntigravityRepository() *AntigravityRepository {
	cfg := utils.GetConfig()
	return &AntigravityRepository{
		baseURL: strings.TrimRight(cfg.LLMBaseURL, "/"),
		apiKey:  cfg.LLMApiKey,
	}
}

// OpenAIMessage represents a message in the chat completion format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the payload for chat completions
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

// OpenAIResponse represents the response from the API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (r *AntigravityRepository) CallModel(prompt string, model string) (string, error) {
	if r.baseURL == "" {
		return "", fmt.Errorf("LLM_BASE_URL is not configured")
	}

	url := fmt.Sprintf("%s/chat/completions", r.baseURL)
	// Build the request body
	reqBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	}

	utils.LogDebug("Antigravity: Calling %s with model %s", url, model)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("API returned no choices")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
