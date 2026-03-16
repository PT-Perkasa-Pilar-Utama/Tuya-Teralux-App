package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sensio/domain/common/utils"
	"time"
)

// MacRegistrationExternalService handles communication with third-party Terminal services
type MacRegistrationExternalService struct {
	client *http.Client
}

// NewMacRegistrationExternalService creates a new instance of MacRegistrationExternalService
func NewMacRegistrationExternalService() *MacRegistrationExternalService {
	return &MacRegistrationExternalService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProcInsertMacAddress calls the external API to register the MAC address
func (s *MacRegistrationExternalService) ProcInsertMacAddress(roomID int, macAddress string, deviceTypeID int) error {
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

	utils.LogDebug("MacRegistrationExternalService: Calling API %s for MAC %s with payload %s", url, macAddress, string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.LogError("MacRegistrationExternalService: Failed to create request for MAC %s: %v", macAddress, err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("MacRegistrationExternalService: API request failed for MAC %s: %v", macAddress, err)
		return fmt.Errorf("external API request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	utils.LogDebug("MacRegistrationExternalService: API returned status %d for MAC %s", resp.StatusCode, macAddress)

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error.Message != "" {
			utils.LogError("MacRegistrationExternalService: API returned error message: %s for MAC %s", errResp.Error.Message, macAddress)
			return utils.NewAPIError(resp.StatusCode, errResp.Error.Message)
		}

		utils.LogError("MacRegistrationExternalService: API returned non-200 status %d for MAC %s", resp.StatusCode, macAddress)
		return fmt.Errorf("external API returned non-200 status: %d", resp.StatusCode)
	}

	var successResp struct {
		Pattern string `json:"Pattern"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err == nil && successResp.Pattern != "" {
		utils.LogDebug("MacRegistrationExternalService: API success pattern: %s for MAC %s", successResp.Pattern, macAddress)
	}

	return nil
}
