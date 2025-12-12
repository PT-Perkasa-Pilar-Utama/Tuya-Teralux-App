package devices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"teralux_app/controllers"
	"teralux_app/dtos"
	"teralux_app/routes"
	"teralux_app/services"
	test_utils "teralux_app/test/utils"
	"teralux_app/usecases"
	"teralux_app/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupRouter initializes the Gin router with all routes for testing
func setupRouter() *gin.Engine {
	// Load config
	utils.LoadConfig()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize router
	router := gin.Default()

	// Initialize auth dependency chain
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)
	tuyaAuthController := controllers.NewTuyaAuthController(tuyaAuthUseCase)

	// Initialize device dependency chain
	tuyaDeviceService := services.NewTuyaDeviceService()
	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService)
	tuyaGetAllDevicesController := controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase)

	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService)
	tuyaGetDeviceByIDController := controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase)

	// Register routes
	routes.SetupTuyaAuthRoutes(router, tuyaAuthController)
	routes.SetupTuyaDeviceRoutes(router, tuyaGetAllDevicesController, tuyaGetDeviceByIDController)

	return router
}

// getAuthToken is a helper function to get a valid access token and its full response
func getAuthToken(t *testing.T, router *gin.Engine) map[string]interface{} {
	authReq, _ := http.NewRequest("POST", "/api/tuya/auth", nil)
	authW := httptest.NewRecorder()
	router.ServeHTTP(authW, authReq)

	var authResponse map[string]interface{}
	err := json.Unmarshal(authW.Body.Bytes(), &authResponse)
	assert.NoError(t, err, "Should get valid auth token for setup")
	return authResponse
}

// TestGetAllDevices_WithValidToken tests fetching devices with a valid access token
// Scenario: User authenticates and then fetches all devices
// Expected: 200 OK and list of devices
func TestGetAllDevices_WithValidToken(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get All Devices", "Valid Token")

	// Step 1: Authenticate to get valid token
	authBody := getAuthToken(t, router)

	// Check if auth passed
	if val, ok := authBody["access_token"]; ok {
		accessToken := val.(string)

		// Step 2: Fetch devices
		req, _ := http.NewRequest("GET", "/api/tuya/devices", nil)
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

			// Check for specific Tuya environment error (Data Center Suspended)
			if strings.Contains(errorResponse.Message, "28841107") || strings.Contains(errorResponse.Message, "No permission") {
				test_utils.LogResult(t, "200 OK (or Env Error 28841107)", "Env Error 28841107 (Allowed)", true)
				return
			}
		}

		var response dtos.TuyaDevicesResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)

		// Assert 200 OK and non-nil list
		if w.Code == http.StatusOK && err == nil && response.Devices != nil {
			test_utils.LogResult(t, "200 OK with Devices List", fmt.Sprintf("200 OK, %d devices", len(response.Devices)), true)
		} else {
			test_utils.LogResult(t, "200 OK with Devices List", fmt.Sprintf("Status %d", w.Code), false)
		}

	} else {
		test_utils.LogResult(t, "Valid Access Token", "Failed to get token", false)
		t.Fatal("Failed to get access token from setup. Check Auth tests.")
	}
}

// TestGetAllDevices_WithInvalidToken tests fetching devices with an invalid token
// Scenario: User attempts to fetch devices using a malformed or expired token
// Expected: Error (Tuya 1004/1010/500)
func TestGetAllDevices_WithInvalidToken(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get All Devices", "Invalid Token")

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/api/tuya/devices", nil)
	req.Header.Set("access_token", "invalid_token_12345")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert failure
	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), false)
	}
}

// TestGetAllDevices_WithoutToken tests fetching devices without any token
// Scenario: User attempts to fetch devices without access_token header
// Expected: Error (400 or 500/401)
func TestGetAllDevices_WithoutToken(t *testing.T) {
	router := setupRouter()

	test_utils.PrintTestHeader(t, "Get All Devices", "Without Token")

	// Create request without header
	req, _ := http.NewRequest("GET", "/api/tuya/devices", nil)
	// No header set
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert failure
	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Status", fmt.Sprintf("Status %d", w.Code), false)
	}
}
