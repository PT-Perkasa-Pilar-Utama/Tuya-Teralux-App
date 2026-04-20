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

// DeviceInfoExternalService handles communication with third-party Big services
type DeviceInfoExternalService struct {
	client *http.Client
}

// NewDeviceInfoExternalService creates a new instance of DeviceInfoExternalService
func NewDeviceInfoExternalService() *DeviceInfoExternalService {
	return &DeviceInfoExternalService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDeviceInfoByMac fetches device and booking info by MAC address
func (s *DeviceInfoExternalService) GetDeviceInfoByMac(macAddress string) (map[string]interface{}, error) {
	// API endpoint
	url := utils.GetConfig().BIGAPIBaseURL + "/IOTANSJavaDasboard/rest/ProcGetDeviceByMacAddressCurrentpied"

	// Payload structure
	payload := map[string]interface{}{
		"host":       utils.GetConfig().BIGAPIBaseURL,
		"port":       "",
		"baseUrl":    "SmartMeetingRoomJavaMySQL/rest",
		"secure":     "1",
		"MacAddress": macAddress,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	utils.LogDebug("DeviceInfoExternalService: Calling API %s for MAC %s", url, macAddress)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("DeviceInfoExternalService: API request failed for MAC %s: %v", macAddress, err)
		return nil, fmt.Errorf("external API request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read and log the raw response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("DeviceInfoExternalService: Failed to read response body for MAC %s: %v", macAddress, err)
		return nil, fmt.Errorf("failed to read external API response body: %w", err)
	}

	utils.LogDebug("DeviceInfoExternalService: Raw API Response for MAC %s: %s", macAddress, string(bodyBytes))

	// Restore the response body for further processing
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		utils.LogError("DeviceInfoExternalService: API returned non-200 status %d for MAC %s", resp.StatusCode, macAddress)
		return nil, fmt.Errorf("external API returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// API returns structure: {"GetRoomByMacAddressCurrent": [...]}
	items, ok := result["GetRoomByMacAddressCurrent"].([]interface{})
	if !ok || len(items) == 0 {
		return nil, utils.NewAPIError(http.StatusNotFound, "Device information not found for given MAC address")
	}

	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected item format in API response")
	}

	return firstItem, nil
}
