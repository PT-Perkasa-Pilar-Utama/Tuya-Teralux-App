package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/services"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
)

type TuyaAuthUseCase interface {
	Authenticate() (*dtos.TuyaAuthResponseDTO, error)
	GetTuyaAccessToken() (string, error)
}

// tuyaAuthUseCase handles the core business logic for Tuya API authentication.
// It orchestrates signature generation, timestamp creation, and service interaction.
type tuyaAuthUseCase struct {
	service         *services.TuyaAuthService
	tokenCache      string
	tokenExpiry     time.Time
	tokenCacheMutex sync.RWMutex
}

// NewTuyaAuthUseCase creates a new instance of TuyaAuthUseCase.
//
// param service The TuyaAuthService used to perform the actual HTTP requests.
// return TuyaAuthUseCase The initialized usecase interface.
func NewTuyaAuthUseCase(service *services.TuyaAuthService) TuyaAuthUseCase {
	return &tuyaAuthUseCase{
		service: service,
	}
}

// Authenticate performs the full authentication flow to retrieve an access token.
// It handles signature generation (HMAC-SHA256), timestamp creation, and header preparation.
//
// Tuya API Documentation (Get Token):
// URL: https://openapi.tuyacn.com/v1.0/token?grant_type=1
// Method: GET
//
// StringToSign Format:
//
//	GET\n{content_hash}\n\n{url}
//	(content_hash is SHA256 of empty string for GET)
//
// return *dtos.TuyaAuthResponseDTO The data transfer object containing the access token, refresh token, and expiration time.
// return error An error if configuration is missing, signature generation fails, or the API call returns an error.
// @throws error if the API returns a non-success status code (e.g., invalid client ID).
func (uc *tuyaAuthUseCase) Authenticate() (*dtos.TuyaAuthResponseDTO, error) {
	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path
	urlPath := "/v1.0/token?grant_type=1"
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)

	utils.LogDebug("Authenticate: generating signature for clientId=%s", config.TuyaClientID)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, "", timestamp, stringToSign)

	// Prepare headers
	headers := map[string]string{
		"client_id":   config.TuyaClientID,
		"sign":        signature,
		"t":           timestamp,
		"sign_method": signMethod,
	}

	// Call service to fetch token
	authResponse, err := uc.service.FetchToken(fullURL, headers)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !authResponse.Success {
		return nil, fmt.Errorf("Gateway authentication failed: %s (code: %d)", authResponse.Msg, authResponse.Code)
	}

	// Transform entity to DTO (Original Tuya Response)
	// We use the UID to generate our own token
	uid := authResponse.Result.UID
	if config.TuyaUserID != "" {
		uid = config.TuyaUserID
	}

	// Cache the Tuya token
	uc.tokenCacheMutex.Lock()
	uc.tokenCache = authResponse.Result.AccessToken
	// Subtract 60 seconds from expiry to be safe
	uc.tokenExpiry = time.Now().Add(time.Duration(authResponse.Result.ExpireTime-60) * time.Second)
	uc.tokenCacheMutex.Unlock()

	// Generate BE JWT
	beToken, err := utils.GenerateToken(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to generate BE token: %w", err)
	}

	dto := &dtos.TuyaAuthResponseDTO{
		AccessToken:  beToken,
		ExpireTime:   -1, // No expiration
		RefreshToken: "none",
		UID:          uid,
	}

	return dto, nil
}

// GetTuyaAccessToken returns a valid Tuya access token, using cache or fetching a new one if needed.
func (uc *tuyaAuthUseCase) GetTuyaAccessToken() (string, error) {
	uc.tokenCacheMutex.RLock()
	if uc.tokenCache != "" && time.Now().Before(uc.tokenExpiry) {
		token := uc.tokenCache
		uc.tokenCacheMutex.RUnlock()
		return token, nil
	}
	uc.tokenCacheMutex.RUnlock()

	// Need to fetch new token
	utils.LogDebug("GetTuyaAccessToken: cache miss or expired, fetching new token")
	_, err := uc.Authenticate()
	if err != nil {
		return "", err
	}

	uc.tokenCacheMutex.RLock()
	defer uc.tokenCacheMutex.RUnlock()
	return uc.tokenCache, nil
}
