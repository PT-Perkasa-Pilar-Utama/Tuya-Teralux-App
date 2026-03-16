package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/tuya/dtos"
	"sensio/domain/tuya/services"
	tuya_utils "sensio/domain/tuya/utils"
	"strconv"
	"time"
)

// TuyaGetDeviceByIDUseCase retrieves detailed information for a specific device.
type TuyaGetDeviceByIDUseCase struct {
	service       *services.TuyaDeviceService
	deviceStateUC DeviceStateUseCase
}

// NewTuyaGetDeviceByIDUseCase initializes a new TuyaGetDeviceByIDUseCase.
//
// param service The TuyaDeviceService used regarding API requests.
// param deviceStateUC The DeviceStateUseCase used for populating infrared_ac status.
// return *TuyaGetDeviceByIDUseCase A pointer to the initialized usecase.
func NewTuyaGetDeviceByIDUseCase(service *services.TuyaDeviceService, deviceStateUC DeviceStateUseCase) *TuyaGetDeviceByIDUseCase {
	return &TuyaGetDeviceByIDUseCase{
		service:       service,
		deviceStateUC: deviceStateUC,
	}
}

// GetDeviceByID fetches the details of a single device from the Tuya API.
//
// Tuya API Documentation (Get Device):
// URL: https://openapi.tuyacn.com/v1.0/devices/{device_id}
// Method: GET
//
// param accessToken The valid OAuth 2.0 access token.
// param deviceID The unique ID of the device to fetch.
// param remoteID (optional) The unique ID of the remote sub-device.
// return *dtos.TuyaDeviceDTO The detailed device information object.
// return error An error if the request fails.
// @throws error If the API returns a failure response.
func (uc *TuyaGetDeviceByIDUseCase) GetDeviceByID(accessToken, deviceID, remoteID string) (*dtos.TuyaDeviceDTO, error) {
	// 1. Determine Target ID (Use remoteID if provided, otherwise gateway deviceID)
	targetID := deviceID
	if remoteID != "" {
		targetID = remoteID
	}

	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path - using /v1.0/devices/{targetID} endpoint
	urlPath := fmt.Sprintf("/v1.0/devices/%s", targetID)
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)

	utils.LogDebug("GetDeviceByID: generating signature for target=%s (hub=%s)", targetID, deviceID)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	// Prepare headers with access token
	headers := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         signature,
		"t":            timestamp,
		"sign_method":  signMethod,
		"access_token": accessToken,
	}

	// Call service to fetch device
	apiStart := time.Now()
	deviceResponse, err := uc.service.FetchDeviceByID(fullURL, headers)
	apiDuration := time.Since(apiStart)
	utils.LogDebug("GetDeviceByID: Tuya API call completed | target=%s | duration_ms=%d | success=%v | code=%d", targetID, apiDuration.Milliseconds(), deviceResponse.Success, deviceResponse.Code)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !deviceResponse.Success {
		return nil, fmt.Errorf("Gateway API failed to fetch device details: %s (code: %d)", deviceResponse.Msg, deviceResponse.Code)
	}

	// Transform status
	statusDTOs := make([]dtos.TuyaDeviceStatusDTO, len(deviceResponse.Result.Status))
	for i, status := range deviceResponse.Result.Status {
		statusDTOs[i] = dtos.TuyaDeviceStatusDTO{
			Code:  status.Code,
			Value: status.Value,
		}
	}

	// For infrared_ac devices, fetch specialized status from Tuya V2 API
	if deviceResponse.Result.Category == "infrared_ac" {
		utils.LogDebug("GetDeviceByID: Fetching specialized status for infrared_ac %s (hub=%s)", targetID, deviceID)

		irUrlPath := fmt.Sprintf("/v2.0/infrareds/%s/remotes/%s/ac/status", deviceID, targetID)
		irFullURL := config.TuyaBaseURL + irUrlPath

		// Generate signature for IR status request
		irTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		irStringToSign := tuya_utils.GenerateTuyaStringToSign("GET", contentHash, "", irUrlPath)
		irSignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, irTimestamp, irStringToSign)

		irHeaders := map[string]string{
			"client_id":    config.TuyaClientID,
			"sign":         irSignature,
			"t":            irTimestamp,
			"sign_method":  signMethod,
			"access_token": accessToken,
		}

		irApiStart := time.Now()
		irResp, err := uc.service.FetchIRACStatus(irFullURL, irHeaders)
		irApiDuration := time.Since(irApiStart)
		if err == nil && irResp.Success {
			utils.LogDebug("GetDeviceByID: Successfully fetched real IR status for %s | duration_ms=%d", targetID, irApiDuration.Milliseconds())
			statusDTOs = make([]dtos.TuyaDeviceStatusDTO, 0, len(irResp.Result))
			for code, val := range irResp.Result {
				// Convert string values to appropriate types if needed (Tuya returns strings for IR status)
				var typedVal interface{} = val
				if intVal, err := strconv.Atoi(val); err == nil {
					typedVal = intVal
				}

				statusDTOs = append(statusDTOs, dtos.TuyaDeviceStatusDTO{
					Code:  code,
					Value: typedVal,
				})
			}
		} else {
			if err != nil {
				utils.LogWarn("GetDeviceByID: Failed to fetch IR status from API: %v", err)
			} else {
				utils.LogWarn("GetDeviceByID: Tuya IR API returned failure: %s", irResp.Msg)
			}

			// Fallback to saved state
			if uc.deviceStateUC != nil {
				stateStart := time.Now()
				savedState, err := uc.deviceStateUC.GetDeviceState(targetID)
				stateDuration := time.Since(stateStart)
				if err == nil && savedState != nil && len(savedState.LastCommands) > 0 {
					utils.LogDebug("GetDeviceByID: Falling back to saved state for %s | duration_ms=%d", targetID, stateDuration.Milliseconds())
					statusDTOs = make([]dtos.TuyaDeviceStatusDTO, len(savedState.LastCommands))
					for i, cmd := range savedState.LastCommands {
						statusDTOs[i] = dtos.TuyaDeviceStatusDTO(cmd)
					}
				} else {
					utils.LogDebug("GetDeviceByID: Using default status for %s (no API status and no saved state)", targetID)
					statusDTOs = []dtos.TuyaDeviceStatusDTO{
						{Code: "power", Value: 0},
						{Code: "temp", Value: 24},
						{Code: "mode", Value: 0},
						{Code: "wind", Value: 0},
					}
				}
			}
		}
	}

	// Determine display name (Use RemoteName if available)
	displayName := deviceResponse.Result.Name
	if deviceResponse.Result.RemoteName != "" {
		displayName = deviceResponse.Result.RemoteName
	}

	// Transform entity to DTO
	dto := &dtos.TuyaDeviceDTO{
		ID:          deviceID, // Still use the path ID as the main ID for consistency with client expectations
		RemoteID:    remoteID,
		Name:        displayName,
		Category:    deviceResponse.Result.Category,
		ProductName: deviceResponse.Result.ProductName,
		Online:      deviceResponse.Result.Online,
		Icon:        deviceResponse.Result.Icon,
		Status:      statusDTOs,
		CustomName:  deviceResponse.Result.CustomName,
		Model:       deviceResponse.Result.Model,
		IP:          deviceResponse.Result.IP,
		LocalKey:    deviceResponse.Result.LocalKey,
		CreateTime:  deviceResponse.Result.CreateTime,
		UpdateTime:  deviceResponse.Result.UpdateTime,
	}

	return dto, nil
}
