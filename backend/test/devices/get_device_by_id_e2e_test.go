package devices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"teralux_app/dtos"
	test_utils "teralux_app/test/utils"
	"testing"
)

// TestGetDeviceByID_WithValidTokenAndID tests fetching a specific device with valid credentials
// Scenario: User authenticates, gets list of devices, picks one, and fetches details
// Expected: 200 OK and device details
func TestGetDeviceByID_WithValidTokenAndID(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get Device By ID", "Valid Token & ID")

	// Step 1: Authenticate
	authBody := getAuthToken(t, router)

	if val, ok := authBody["access_token"]; ok {
		accessToken := val.(string)

		// Step 2: Get all devices to find a valid ID
		devicesReq, _ := http.NewRequest("GET", "/api/tuya/devices", nil)
		devicesReq.Header.Set("access_token", accessToken)
		devicesW := httptest.NewRecorder()
		router.ServeHTTP(devicesW, devicesReq)

		// Check Env failure tolerance
		if devicesW.Code == http.StatusInternalServerError {
			var errorResponse struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			json.Unmarshal(devicesW.Body.Bytes(), &errorResponse)
			if strings.Contains(errorResponse.Message, "28841107") || strings.Contains(errorResponse.Message, "No permission") {
				test_utils.LogResult(t, "Fetch List Success (or Env Error)", "Env Error 28841107 (Allowed)", true)
				return
			}
		}

		if devicesW.Code != http.StatusOK {
			test_utils.LogResult(t, "Fetch Devices List", fmt.Sprintf("Failed Status %d", devicesW.Code), false)
			return
		}

		var devicesResponse dtos.TuyaDevicesResponseDTO
		json.Unmarshal(devicesW.Body.Bytes(), &devicesResponse)

		if len(devicesResponse.Devices) == 0 {
			test_utils.LogResult(t, "Device Selection", "No devices found (Warning)", true)
			return
		}

		deviceID := devicesResponse.Devices[0].ID

		// Step 3: Get specific device
		req, _ := http.NewRequest("GET", "/api/tuya/devices/"+deviceID, nil)
		req.Header.Set("access_token", accessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Validate response
		if w.Code == http.StatusInternalServerError {
			var errorResponse struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			json.Unmarshal(w.Body.Bytes(), &errorResponse)
			if strings.Contains(errorResponse.Message, "28841107") || strings.Contains(errorResponse.Message, "No permission") {
				test_utils.LogResult(t, "200 OK (or Env Error)", "Env Error 28841107 (Allowed)", true)
				return
			}
		}

		var response dtos.TuyaDeviceResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)

		if w.Code == http.StatusOK && err == nil && response.Device.ID == deviceID {
			test_utils.LogResult(t, "200 OK with Correct Device ID", fmt.Sprintf("200 OK, ID: %s", response.Device.ID), true)
		} else {
			test_utils.LogResult(t, "200 OK with Correct Device ID", fmt.Sprintf("Status %d, ID Mismatch or Error", w.Code), false)
		}

	} else {
		test_utils.LogResult(t, "Valid Access Token", "Failed to get token", false)
		t.Fatal("Failed to get access token from setup")
	}
}

// TestGetDeviceByID_WithInvalidToken tests fetching device with invalid token
// Scenario: User tries to fetch device with invalid/expired token
// Expected: Error
func TestGetDeviceByID_WithInvalidToken(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get Device By ID", "Invalid Token")

	req, _ := http.NewRequest("GET", "/api/tuya/devices/any_id", nil)
	req.Header.Set("access_token", "invalid_token_123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), false)
	}
}

// TestGetDeviceByID_WithInvalidID tests fetching non-existent device
// Scenario: User authenticates but requests invalid device ID
// Expected: Error (Tuya API failure)
func TestGetDeviceByID_WithInvalidID(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get Device By ID", "Invalid Device ID")

	authBody := getAuthToken(t, router)
	if val, ok := authBody["access_token"]; ok {
		accessToken := val.(string)

		req, _ := http.NewRequest("GET", "/api/tuya/devices/invalid_device_id_99999", nil)
		req.Header.Set("access_token", accessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Assert failure (Tuya API should return error)
		if w.Code != http.StatusOK {
			test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), true)
		} else {
			test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), false)
		}
	}
}

// TestGetDeviceByID_WithoutToken tests fetching device without token
// Scenario: User requests device without auth header
// Expected: Error
func TestGetDeviceByID_WithoutToken(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get Device By ID", "Without Token")

	req, _ := http.NewRequest("GET", "/api/tuya/devices/any_id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), false)
	}
}
