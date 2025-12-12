package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"teralux_app/controllers"
	"teralux_app/dtos"
	"teralux_app/routes"
	"teralux_app/services"
	test_utils "teralux_app/test/utils"
	"teralux_app/usecases"
	"teralux_app/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupRouter initializes the Gin router with all routes for testing
func setupRouter() *gin.Engine {
	// Load config
	utils.LoadConfig()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize router
	router := gin.Default()

	// Initialize dependency chain
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)
	tuyaAuthController := controllers.NewTuyaAuthController(tuyaAuthUseCase)

	// Register routes
	routes.SetupTuyaAuthRoutes(router, tuyaAuthController)

	return router
}

// TestAuth_LoginWithValidCredentials tests login with valid credentials
func TestAuth_LoginWithValidCredentials(t *testing.T) {
	// Restore config after test
	originalConfig := *utils.GetConfig()
	defer func() {
		utils.AppConfig = &originalConfig
	}()

	router := setupRouter()

	test_utils.PrintTestHeader(t, "Auth Login", "Valid Credentials")

	// Create request
	req, _ := http.NewRequest("POST", "/api/tuya/auth", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Parse response
	var response dtos.TuyaAuthResponseDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	// Log Result
	if w.Code == http.StatusOK && response.AccessToken != "" {
		test_utils.LogResult(t, "200 OK with Access Token", fmt.Sprintf("%d OK with Token", w.Code), true)
	} else {
		test_utils.LogResult(t, "200 OK with Access Token", fmt.Sprintf("Status %d, Token: %s", w.Code, response.AccessToken), false)
	}
}

// TestAuth_LoginWithInvalidCredentials tests login with invalid credentials
func TestAuth_LoginWithInvalidCredentials(t *testing.T) {
	// Backup and restore config/env
	originalClientID := os.Getenv("TUYA_CLIENT_ID")
	originalClientSecret := os.Getenv("TUYA_ACCESS_SECRET")
	defer func() {
		os.Setenv("TUYA_CLIENT_ID", originalClientID)
		os.Setenv("TUYA_ACCESS_SECRET", originalClientSecret)
		utils.LoadConfig() // Reload original config
	}()

	// Set invalid credentials
	os.Setenv("TUYA_CLIENT_ID", "invalid_id")
	os.Setenv("TUYA_ACCESS_SECRET", "invalid_secret")
	utils.LoadConfig() // Reload config with invalid values

	router := setupRouter()

	test_utils.PrintTestHeader(t, "Auth Login", "Invalid Credentials")

	// Create request
	req, _ := http.NewRequest("POST", "/api/tuya/auth", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Log Result
	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Response", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Response", fmt.Sprintf("Status %d", w.Code), false)
	}
}

// TestAuth_LoginWithoutCredentials tests login without credentials (empty env)
func TestAuth_LoginWithoutCredentials(t *testing.T) {
	// Backup and restore config/env
	originalClientID := os.Getenv("TUYA_CLIENT_ID")
	originalClientSecret := os.Getenv("TUYA_ACCESS_SECRET")
	defer func() {
		os.Setenv("TUYA_CLIENT_ID", originalClientID)
		os.Setenv("TUYA_ACCESS_SECRET", originalClientSecret)
		utils.LoadConfig() // Reload original config
	}()

	// Set empty credentials
	os.Setenv("TUYA_CLIENT_ID", "")
	os.Setenv("TUYA_ACCESS_SECRET", "")
	utils.LoadConfig() // Reload config

	router := setupRouter()

	test_utils.PrintTestHeader(t, "Auth Login", "Missing Credentials")

	// Create request
	req, _ := http.NewRequest("POST", "/api/tuya/auth", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Log Result
	if w.Code != http.StatusOK {
		test_utils.LogResult(t, "Error / Failure Response", fmt.Sprintf("Status %d", w.Code), true)
	} else {
		test_utils.LogResult(t, "Error / Failure Response", fmt.Sprintf("Status %d", w.Code), false)
	}
}
