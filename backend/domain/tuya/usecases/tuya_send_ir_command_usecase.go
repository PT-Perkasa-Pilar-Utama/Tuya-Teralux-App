package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
)

// TuyaSendIRCommandUseCase defines the interface for sending IR commands to Tuya devices.
type TuyaSendIRCommandUseCase interface {
	SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error)
}

type tuyaSendIRCommandUseCase struct {
	service *services.TuyaDeviceService
}

// NewTuyaSendIRCommandUseCase initializes a new tuyaSendIRCommandUseCase.
func NewTuyaSendIRCommandUseCase(service *services.TuyaDeviceService) TuyaSendIRCommandUseCase {
	return &tuyaSendIRCommandUseCase{
		service: service,
	}
}

// SendIRACCommand sends specific commands to an Infrared (IR) controlled Air Conditioner.
func (uc *tuyaSendIRCommandUseCase) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	config := utils.GetConfig()
	var gatewayID string

	// 1. Fetch Device Details to get correct GatewayID
	deviceUrlPath := fmt.Sprintf("/v1.0/iot-03/devices/%s", remoteID)
	deviceFullURL := config.TuyaBaseURL + deviceUrlPath
	deviceTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	hEmpty := sha256.New()
	hEmpty.Write([]byte(""))
	deviceContentHash := hex.EncodeToString(hEmpty.Sum(nil))

	deviceStringToSign := tuya_utils.GenerateTuyaStringToSign("GET", deviceContentHash, "", deviceUrlPath)
	deviceSignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, deviceTimestamp, deviceStringToSign)

	deviceHeaders := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         deviceSignature,
		"t":            deviceTimestamp,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken,
	}

	utils.LogDebug("SendIRACCommand: Fetching device details for RemoteID=%s", remoteID)
	deviceResp, err := uc.service.FetchDeviceByID(deviceFullURL, deviceHeaders)
	if err == nil && deviceResp.Success {
		if deviceResp.Result.GatewayID != "" {
			utils.LogDebug("SendIRACCommand: Found GatewayID=%s. Using it as InfraredID.", deviceResp.Result.GatewayID)
			gatewayID = deviceResp.Result.GatewayID
			infraredID = gatewayID
		}
	}

	// 2. Prepare IR Command
	irUrlPath := fmt.Sprintf("/v1.0/infrareds/%s/air-conditioners/%s/scenes/command", infraredID, remoteID)
	irFullURL := config.TuyaBaseURL + irUrlPath
	irTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Tuya IR AC API expects direct parameters
	irBody := make(map[string]interface{})

	// Map parameters
	for k, v := range params {
		switch k {
		case "power":
			irBody["power"] = v // 0 = off, 1 = on
		case "temp", "temperature":
			irBody["temp"] = v // 16-30 degrees
			if _, exists := irBody["power"]; !exists {
				irBody["power"] = 1 // Ensure ON
			}
		case "mode":
			irBody["mode"] = v // 0=cool, 1=heat, 2=auto, 3=wind, 4=dry
			if _, exists := irBody["power"]; !exists {
				irBody["power"] = 1
			}
		case "wind", "fan":
			irBody["wind"] = v // 0=auto, 1=low, 2=med, 3=high
			if _, exists := irBody["power"]; !exists {
				irBody["power"] = 1
			}
		default:
			// For unknown codes, try sending as-is
			irBody[k] = v
		}
	}

	irJsonBody, _ := json.Marshal(irBody)

	hIR := sha256.New()
	hIR.Write(irJsonBody)
	irContentHash := hex.EncodeToString(hIR.Sum(nil))

	irStringToSign := tuya_utils.GenerateTuyaStringToSign("POST", irContentHash, "", irUrlPath)
	irSignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, irTimestamp, irStringToSign)

	irHeaders := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         irSignature,
		"t":            irTimestamp,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken,
		"Content-Type": "application/json",
	}

	utils.LogDebug("SendIRACCommand: Sending IR command")
	utils.LogDebug("SendIRACCommand: URL: %s", irFullURL)
	utils.LogDebug("SendIRACCommand: Headers: client_id=%s, t=%s, sign_method=%s, access_token=%s...",
		irHeaders["client_id"], irHeaders["t"], irHeaders["sign_method"], irHeaders["access_token"][:10])
	utils.LogDebug("SendIRACCommand: Body: %s", string(irJsonBody))

	resp, err := uc.service.SendIRCommand(irFullURL, irHeaders, irJsonBody)
	if err != nil {
		utils.LogError("SendIRACCommand: Network error calling Tuya: %v", err)
		return false, err
	}

	utils.LogDebug("SendIRACCommand: Tuya response received: success=%v, code=%d, msg=%s, result=%v",
		resp.Success, resp.Code, resp.Msg, resp.Result)

	if !resp.Success {
		return false, fmt.Errorf("Gateway IR API failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	return resp.Result, nil
}
