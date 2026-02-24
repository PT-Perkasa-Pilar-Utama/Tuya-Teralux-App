package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"teralux_app/domain/common/utils"
	"time"
)

// TeraluxExternalService handles communication with third-party Teralux services
type TeraluxExternalService struct {
	client *http.Client
}

// NewTeraluxExternalService creates a new instance of TeraluxExternalService
func NewTeraluxExternalService() *TeraluxExternalService {
	return &TeraluxExternalService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProcInsertMacAddress calls the external API to register the MAC address
func (s *TeraluxExternalService) ProcInsertMacAddress(roomID int, macAddress string, deviceTypeID int) error {
	url := "https://aplikasi-big.com/IOTAN5JavaDasboard/rest/ProcInsertMacAddress"

	payload := map[string]interface{}{
		"roomid":       roomID,
		"macAddress":   macAddress,
		"DeviceTypeId": deviceTypeID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	utils.LogDebug("TeraluxExternalService: Calling API %s for MAC %s with payload %s", url, macAddress, string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.LogError("TeraluxExternalService: Failed to create request for MAC %s: %v", macAddress, err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("TeraluxExternalService: API request failed for MAC %s: %v", macAddress, err)
		return fmt.Errorf("external API request failed: %w", err)
	}
	defer resp.Body.Close()

	utils.LogDebug("TeraluxExternalService: API returned status %d for MAC %s", resp.StatusCode, macAddress)

	if resp.StatusCode != http.StatusOK {
		utils.LogError("TeraluxExternalService: API returned non-200 status %d for MAC %s", resp.StatusCode, macAddress)
		return fmt.Errorf("external API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
