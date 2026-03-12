package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MQTTCredentials holds the plaintext MQTT credentials returned by the Rust Auth Service
type MQTTCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type mqttCreateRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	IsSuperuser bool   `json:"is_superuser"`
}

type mqttAuthResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *MQTTCredentials `json:"data"`
}

// MqttAuthClient is an HTTP client for the EMQX Auth Service (Rust)
type MqttAuthClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewMqttAuthClient creates a new MQTT Auth HTTP client
func NewMqttAuthClient(baseURL, apiKey string) *MqttAuthClient {
	return &MqttAuthClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// CreateMQTTUser calls POST /mqtt/create in the Rust Auth Service
// Returns: alreadyExists (bool), error
func (c *MqttAuthClient) CreateMQTTUser(username, password string) (alreadyExists bool, err error) {
	body, _ := json.Marshal(mqttCreateRequest{
		Username:    username,
		Password:    password,
		IsSuperuser: false,
	})

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/mqtt/create", bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Errorf("failed to build mqtt create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("mqtt create request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return true, nil // Scenario C: already exists
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("mqtt create failed (status %d): %s", resp.StatusCode, string(b))
	}

	return false, nil
}

// GetMQTTCredentials calls GET /mqtt/credentials/{username} in the Rust Auth Service
// Returns: decrypted credentials, error
func (c *MqttAuthClient) GetMQTTCredentials(username string) (*MQTTCredentials, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/mqtt/credentials/%s", c.baseURL, username), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build mqtt credentials request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mqtt credentials request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // MQTT user not found
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mqtt credentials failed (status %d): %s", resp.StatusCode, string(b))
	}

	var result mqttAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mqtt credentials response: %w", err)
	}

	return result.Data, nil
}
