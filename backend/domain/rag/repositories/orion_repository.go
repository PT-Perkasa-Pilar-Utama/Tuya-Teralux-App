package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"teralux_app/domain/common/utils"
	"time"
)

type OrionRepository struct {
	baseURL string
	apiKey  string
	model   string
}

func NewOrionRepository() *OrionRepository {
	cfg := utils.GetConfig()
	return &OrionRepository{
		baseURL: cfg.OrionBaseURL,
		apiKey:  cfg.OrionApiKey,
		model:   cfg.OrionModel,
	}
}

func (r *OrionRepository) HealthCheck() bool {
	if r.baseURL == "" || r.apiKey == "" {
		return false
	}

	// Remove trailing slash from baseURL if present
	baseURL := r.baseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	url := fmt.Sprintf("%s/v1/models", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogWarn("Orion HealthCheck failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.LogWarn("Orion HealthCheck failed: status %d", resp.StatusCode)
		return false
	}
	return true
}

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

func (r *OrionRepository) CallModel(prompt string, model string) (string, error) {
	if r.apiKey == "" {
		return "", fmt.Errorf("ORION_API_KEY is not configured")
	}

	// Always use Orion's configured model, ignore parameter
	model = r.model

	// Remove trailing slash from baseURL if present
	baseURL := r.baseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	url := fmt.Sprintf("%s/v1/responses", baseURL)

	reqBody := map[string]interface{}{
		"model": model,
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

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0} // No timeout, handled by async task system
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call orion api: %w", err)
	}
	defer resp.Body.Close()

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

	// Extract text from message output (skip reasoning output)
	for _, output := range orionResp.Output {
		if output.Type == "message" {
			for _, content := range output.Content {
				if content.Type == "output_text" && content.Text != "" {
					return content.Text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("orion api returned no text content")
}
