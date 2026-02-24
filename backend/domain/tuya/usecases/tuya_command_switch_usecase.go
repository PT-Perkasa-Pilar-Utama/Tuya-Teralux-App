package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/entities"
	"teralux_app/domain/tuya/services"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
)

// TuyaCommandSwitchUseCase defines the interface for sending switch commands to Tuya devices.
type TuyaCommandSwitchUseCase interface {
	SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error)
}

type tuyaCommandSwitchUseCase struct {
	service       *services.TuyaDeviceService
	deviceStateUC DeviceStateUseCase
}

// NewTuyaCommandSwitchUseCase initializes a new tuyaCommandSwitchUseCase.
func NewTuyaCommandSwitchUseCase(service *services.TuyaDeviceService, deviceStateUC DeviceStateUseCase) TuyaCommandSwitchUseCase {
	return &tuyaCommandSwitchUseCase{
		service:       service,
		deviceStateUC: deviceStateUC,
	}
}

// SendSwitchCommand sends switch commands to a specific device.
func (uc *tuyaCommandSwitchUseCase) SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error) {
	config := utils.GetConfig()
	urlPath := fmt.Sprintf("/v1.0/iot-03/devices/%s/commands", deviceID)
	fullURL := config.TuyaBaseURL + urlPath
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Map DTO to Entity
	entityCommands := make([]entities.TuyaCommand, len(commands))
	for i, cmd := range commands {
		entityCommands[i] = entities.TuyaCommand{
			Code:  cmd.Code,
			Value: cmd.Value,
		}
	}

	reqBody := entities.TuyaCommandRequest{Commands: entityCommands}
	jsonBody, _ := json.Marshal(reqBody)

	// Calculate content hash
	h := sha256.New()
	h.Write(jsonBody)
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("POST", contentHash, "", urlPath)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	// Prepare headers
	headers := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         signature,
		"t":            timestamp,
		"sign_method":  signMethod,
		"access_token": accessToken,
	}

	// Call service
	utils.LogDebug("SendCommand: Sending Switch command")
	utils.LogDebug("SendCommand: DeviceID=%s, URL=%s", deviceID, fullURL)
	utils.LogDebug("SendCommand: Headers: client_id=%s, t=%s, sign_method=%s, access_token=%s...",
		headers["client_id"], headers["t"], headers["sign_method"], headers["access_token"][:10])
	utils.LogDebug("SendCommand: Body: %s", string(jsonBody))

	resp, err := uc.service.SendCommand(fullURL, headers, entityCommands)
	if err != nil {
		utils.LogError("SendCommand: Network error calling Tuya: %v", err)
		return false, err
	}

	utils.LogDebug("SendCommand: Tuya response received: success=%v, code=%d, msg=%s, result=%v",
		resp.Success, resp.Code, resp.Msg, resp.Result)

	if !resp.Success {
		utils.LogError("Tuya API Command Failed. Code: %d, Msg: %s", resp.Code, resp.Msg)

		// RETRY LOGIC for "switch_" mismatch (switch_1 -> switch1)
		if resp.Code == 2008 {
			var retryCommands []entities.TuyaCommand
			shouldRetry := false

			for _, cmd := range entityCommands {
				newCode := cmd.Code
				if strings.HasPrefix(cmd.Code, "switch_") {
					newCode = strings.Replace(cmd.Code, "_", "", 1)
					if newCode != cmd.Code {
						shouldRetry = true
					}
				}
				retryCommands = append(retryCommands, entities.TuyaCommand{Code: newCode, Value: cmd.Value})
			}

			if shouldRetry {
				utils.LogDebug("Retrying with corrected commands: %+v", retryCommands)

				// Use LEGACY endpoint for DP instructions (v1.0/devices/{id}/commands) instead of iot-03
				retryUrlPath := fmt.Sprintf("/v1.0/devices/%s/commands", deviceID)
				retryFullURL := config.TuyaBaseURL + retryUrlPath

				// Re-create request body
				retryReqBody := entities.TuyaCommandRequest{Commands: retryCommands}
				retryJsonBody, _ := json.Marshal(retryReqBody)

				// Re-calculate content hash
				hRetry := sha256.New()
				hRetry.Write(retryJsonBody)
				retryContentHash := hex.EncodeToString(hRetry.Sum(nil))

				// Re-sign
				retryStringToSign := tuya_utils.GenerateTuyaStringToSign("POST", retryContentHash, "", retryUrlPath)
				retrySignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, retryStringToSign)

				// Re-prepare headers
				retryHeaders := map[string]string{
					"client_id":    config.TuyaClientID,
					"sign":         retrySignature,
					"t":            timestamp,
					"sign_method":  signMethod,
					"access_token": accessToken,
				}

				// Retry call
				retryResp, retryErr := uc.service.SendCommand(retryFullURL, retryHeaders, retryCommands)
				switch {
				case retryErr == nil && retryResp.Success:
					utils.LogInfo("Retry success with corrected commands!")
					return retryResp.Result, nil
				case retryErr != nil:
					utils.LogError("Retry failed: %v", retryErr)
				default:
					utils.LogError("Retry API failed: %d %s", retryResp.Code, retryResp.Msg)
				}
			}
		}

		return false, fmt.Errorf("Gateway API failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	// Save state after successful command
	if uc.deviceStateUC != nil {
		stateCommands := make([]dtos.DeviceStateCommandDTO, len(commands))
		for i, cmd := range commands {
			stateCommands[i] = dtos.DeviceStateCommandDTO(cmd)
		}
		if err := uc.deviceStateUC.SaveDeviceState(deviceID, stateCommands); err != nil {
			utils.LogWarn("Failed to save device state for %s: %v", deviceID, err)
		}
	}

	return resp.Result, nil
}
