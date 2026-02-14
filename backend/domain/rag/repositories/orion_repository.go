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

	url := fmt.Sprintf("%s/v1/models", r.baseURL)
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

func (r *OrionRepository) CallModel(prompt string, model string) (string, error) {
	if r.apiKey == "" {
		return "", fmt.Errorf("ORION_API_KEY is not configured")
	}

	if model == "" || model == "default" {
		model = r.model
	}

	url := fmt.Sprintf("%s/v1/responses", r.baseURL)
	
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

	client := &http.Client{Timeout: 60 * time.Second}
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

	var result string
	if err := json.Unmarshal(body, &result); err == nil {
		return result, nil
	}

	return string(body), nil
}
