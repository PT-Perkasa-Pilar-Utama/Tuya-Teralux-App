package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sensio/domain/common/utils"
	"time"
)

type WANotificationService struct {
	client  *http.Client
	baseURL string
}

func NewWANotificationService(baseURL string) *WANotificationService {
	return &WANotificationService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

type WASendRequest struct {
	Number  string `json:"number"`
	Content string `json:"content"`
}

type WASendResponse struct {
	Status  interface{} `json:"status"`
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
}

func (s *WANotificationService) SendMessage(phoneNumber string, content string) error {
	payload := WASendRequest{
		Number:  phoneNumber,
		Content: content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal WA payload: %w", err)
	}

	utils.LogDebug("WANotificationService: Sending message to %s", phoneNumber)

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("WANotificationService: Failed to send to %s: %v", phoneNumber, err)
		return fmt.Errorf("failed to send WA message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	utils.LogDebug("WANotificationService: WA response for %s: %s", phoneNumber, string(bodyBytes))

	var result WASendResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != 200 {
		utils.LogError("WANotificationService: WA API returned error for %s: %s", phoneNumber, result.Message)
		return fmt.Errorf("WA API error: %s", result.Message)
	}

	utils.LogInfo("WANotificationService: Message sent successfully to %s", phoneNumber)
	return nil
}
