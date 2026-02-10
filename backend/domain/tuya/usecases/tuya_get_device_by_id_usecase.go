package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/services"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
)

// TuyaGetDeviceByIDUseCase retrieves detailed information for a specific device.
type TuyaGetDeviceByIDUseCase struct {
	service *services.TuyaDeviceService
	cache   *infrastructure.BadgerService
}

// NewTuyaGetDeviceByIDUseCase initializes a new TuyaGetDeviceByIDUseCase.
//
// param service The TuyaDeviceService used regarding API requests.
// param cache The BadgerService used for caching device details.
// return *TuyaGetDeviceByIDUseCase A pointer to the initialized usecase.
func NewTuyaGetDeviceByIDUseCase(service *services.TuyaDeviceService, cache *infrastructure.BadgerService) *TuyaGetDeviceByIDUseCase {
	return &TuyaGetDeviceByIDUseCase{
		service: service,
		cache:   cache,
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
// return *dtos.TuyaDeviceDTO The detailed device information object.
// return error An error if the request fails.
// @throws error If the API returns a failure response.
func (uc *TuyaGetDeviceByIDUseCase) GetDeviceByID(accessToken, deviceID string) (*dtos.TuyaDeviceDTO, error) {
	// 1. Try Cache First
	cacheKey := fmt.Sprintf("tuya_device_%s", deviceID)
	cachedData, err := uc.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		var cachedDTO dtos.TuyaDeviceDTO
		if err := json.Unmarshal(cachedData, &cachedDTO); err == nil {
			utils.LogDebug("GetDeviceByID: Cache HIT for device %s", deviceID)
			return &cachedDTO, nil
		}
		utils.LogError("GetDeviceByID: failed to unmarshal cached value: %v", err)
	} else {
		utils.LogDebug("GetDeviceByID: Cache MISS for device %s (err: %v)", deviceID, err)
	}

	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path - using /v1.0/devices/{device_id} endpoint
	urlPath := fmt.Sprintf("/v1.0/devices/%s", deviceID)
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)

	utils.LogDebug("GetDeviceByID: generating signature for device=%s", deviceID)

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
	deviceResponse, err := uc.service.FetchDeviceByID(fullURL, headers)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !deviceResponse.Success {
		return nil, fmt.Errorf("tuya API failed to fetch device: %s (code: %d)", deviceResponse.Msg, deviceResponse.Code)
	}

	// Transform status
	statusDTOs := make([]dtos.TuyaDeviceStatusDTO, len(deviceResponse.Result.Status))
	for i, status := range deviceResponse.Result.Status {
		statusDTOs[i] = dtos.TuyaDeviceStatusDTO{
			Code:  status.Code,
			Value: status.Value,
		}
	}

	// Transform entity to DTO
	dto := &dtos.TuyaDeviceDTO{
		ID:          deviceResponse.Result.ID,
		Name:        deviceResponse.Result.Name,
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

	// 2. Save to Cache
	if jsonData, err := json.Marshal(dto); err == nil {
		if err := uc.cache.Set(cacheKey, jsonData); err != nil {
			utils.LogWarn("GetDeviceByID: failed to save device %s to cache: %v", deviceID, err)
		} else {
			utils.LogDebug("GetDeviceByID: Saved device %s to cache", deviceID)
		}
	} else {
		utils.LogError("GetDeviceByID: Failed to marshal device for cache: %v", err)
	}

	return dto, nil
}
