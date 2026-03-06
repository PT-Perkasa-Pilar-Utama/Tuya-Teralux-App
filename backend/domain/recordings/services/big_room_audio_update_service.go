package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sensio/domain/common/utils"
)

type BIGRoomAudioUpdateService interface {
	UpdateRoomOccupiedAudio(macAddress, audioPath string) error
}

type bigRoomAudioUpdateService struct {
	client *http.Client
}

func NewBIGRoomAudioUpdateService() BIGRoomAudioUpdateService {
	return &bigRoomAudioUpdateService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type updateRoomOccupiedAudioRequest struct {
	MacAddress string `json:"MacAddress"`
	AudioPath  string `json:"audioPath"`
}

type updateRoomOccupiedAudioResponse struct {
	ResponseJSON string `json:"responsejson"`
}

func (s *bigRoomAudioUpdateService) UpdateRoomOccupiedAudio(macAddress, audioPath string) error {
	url := "https://aplikasi-big.com/IOTAN5JavaDasboard/rest/ProcUpdateRoomOccupiedAudio"

	reqBody := updateRoomOccupiedAudioRequest{
		MacAddress: macAddress,
		AudioPath:  audioPath,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	var lastErr error
	maxRetries := 3
	backoff := 500 * time.Millisecond

	attemptsMade := 0
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptsMade = attempt
		utils.LogInfo("BIGRoomAudioUpdateService: Attempt %d/%d - Sending request to %s with MacAddress: %s", attempt, maxRetries, url, macAddress)

		err, isRetryable := s.doUpdateWithRetryInfo(url, jsonData)
		if err == nil {
			utils.LogInfo("BIGRoomAudioUpdateService: Successfully updated external API for MacAddress: %s (attempt %d)", macAddress, attempt)
			return nil
		}

		lastErr = err
		if !isRetryable {
			utils.LogWarn("BIGRoomAudioUpdateService: Non-retryable error on attempt %d: %v. Failing fast.", attempt, err)
			break
		}

		utils.LogWarn("BIGRoomAudioUpdateService: Retryable error on attempt %d: %v", attempt, err)

		if attempt < maxRetries {
			utils.LogInfo("BIGRoomAudioUpdateService: Retrying in %v...", backoff)
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("failed after %d attempt(s): %w", attemptsMade, lastErr)
}

func (s *bigRoomAudioUpdateService) doUpdateWithRetryInfo(url string, jsonData []byte) (error, bool) {
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Network errors are generally retryable
		return fmt.Errorf("failed to send request: %v", err), true
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("external api returned server error: %s", resp.Status), true
	}

	if resp.StatusCode >= 400 {
		// 4xx errors are generally client errors and NOT retryable
		return fmt.Errorf("external api returned client error: %s", resp.Status), false
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("external api returned unexpected status: %s", resp.Status), false
	}

	var result updateRoomOccupiedAudioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err), false
	}

	if result.ResponseJSON != "success" {
		return fmt.Errorf("external api returned business failure: %s", result.ResponseJSON), false
	}

	return nil, false
}
