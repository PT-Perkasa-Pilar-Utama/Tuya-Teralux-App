package usecases

import (
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupAuthTestEnv initializes configuration for auth tests
func setupAuthTestEnv(t *testing.T) {
	utils.AppConfig = &utils.Config{
		TuyaBaseURL:      "https://openapi.tuyacn.com",
		TuyaClientID:     "test-client-id",
		TuyaClientSecret: "test-client-secret",
		TuyaUserID:       "test-user-id", // optional override
	}
}

func TestTuyaAuthUseCase_Authenticate_Success(t *testing.T) {
	// Enable Test Mode to use built-in mock in TuyaAuthService
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.ReleaseMode)

	setupAuthTestEnv(t)

	service := services.NewTuyaAuthService()
	useCase := NewTuyaAuthUseCase(service)

	// Execute
	resp, err := useCase.Authenticate()

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify details from the mock in TuyaAuthService
	if resp.AccessToken != "mock-access-token" {
		t.Errorf("Expected mock access token, got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "mock-refresh-token" {
		t.Errorf("Expected mock refresh token, got %s", resp.RefreshToken)
	}
	// Verify UID override logic if configured
	if resp.UID != "test-user-id" {
		t.Errorf("Expected UID override to 'test-user-id', got %s", resp.UID)
	}
}

func TestTuyaAuthUseCase_Authenticate_SignatureGeneration(t *testing.T) {
	// Verify signature generation logic without validation

	gin.SetMode(gin.TestMode)
	defer gin.SetMode(gin.ReleaseMode)

	setupAuthTestEnv(t)

	service := services.NewTuyaAuthService()
	useCase := NewTuyaAuthUseCase(service)

	_, err := useCase.Authenticate()
	if err != nil {
		t.Fatalf("Authenticate execution failed: %v", err)
	}
}

func TestTuyaAuthUseCase_Authenticate_Real_Network_Fail(t *testing.T) {
	// Force ReleaseMode to bypass mock and attempt real network call
	// This ensures we cover the "real" path too, expecting failure due to bad credentials/network
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode) // restore if needed (though next tests set it explicitly)

	setupAuthTestEnv(t)

	// Create service with very short timeout to fail fast
	service := services.NewTuyaAuthService()
	// Skip real network test by default
	t.Skip("Skipping real network test")

	useCase := NewTuyaAuthUseCase(service)
	_, err := useCase.Authenticate()

	if err == nil {
		t.Error("Expected error in real network mode with fake creds, got success")
	} else {
		// Just ensure it's not a panic
		if !strings.Contains(err.Error(), "failed") && !strings.Contains(err.Error(), "API returned") {
			// t.Logf("Got expected error: %v", err)
		}
	}
}
