package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"teralux_app/domain/common/utils"
	"time"
)

// MailExternalService handles communication with third-party Teralux services for mail
type MailExternalService struct {
	client *http.Client
}

// NewMailExternalService creates a new instance of MailExternalService
func NewMailExternalService() *MailExternalService {
	return &MailExternalService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDeviceInfoByMac fetches device and booking info by MAC address
func (s *MailExternalService) GetDeviceInfoByMac(macAddress string) (map[string]interface{}, error) {
	url := "https://aplikasi-big.com/SmartMeetingRoomJavaMySQL/rest/ProcGetDeviceTeralux"

	payload := map[string]interface{}{
		"MacAddress": macAddress,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	utils.LogDebug("MailExternalService: Calling API %s for MAC %s", url, macAddress)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("MailExternalService: API request failed for MAC %s: %v", macAddress, err)
		return nil, fmt.Errorf("external API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.LogError("MailExternalService: API returned non-200 status %d for MAC %s", resp.StatusCode, macAddress)
		return nil, fmt.Errorf("external API returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Based on investigation, results are in "SDTGetRoomTeralux" array
	items, ok := result["SDTGetRoomTeralux"].([]interface{})
	if !ok || len(items) == 0 {
		return nil, utils.NewAPIError(http.StatusNotFound, "Device information not found for given MAC address")
	}

	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected item format in API response")
	}

	return firstItem, nil
}
