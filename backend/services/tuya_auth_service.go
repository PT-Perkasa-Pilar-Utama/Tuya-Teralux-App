package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"teralux_app/entities"
	"time"
)

// TuyaAuthService is a simple HTTP client wrapper for Tuya API
type TuyaAuthService struct {
	client *http.Client
}

// NewTuyaAuthService creates a new TuyaAuthService instance
func NewTuyaAuthService() *TuyaAuthService {
	return &TuyaAuthService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchToken makes HTTP request to Tuya API with provided URL and headers
func (s *TuyaAuthService) FetchToken(url string, headers map[string]string) (*entities.TuyaAuthResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResponse entities.TuyaAuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResponse, nil
}
