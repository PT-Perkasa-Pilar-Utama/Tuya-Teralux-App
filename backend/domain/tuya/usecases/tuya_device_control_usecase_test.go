package usecases

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTuyaDeviceControlUseCase_SendCommand_Success(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if !strings.Contains(r.URL.Path, "/v1.0/iot-03/devices/dev-1/commands") {
			t.Errorf("Unexpected URL: %s", r.URL.Path)
		}
		// Verify Headers
		if r.Header.Get("access_token") != "test-token" {
			t.Errorf("Missing access token")
		}
		// Return Success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "result": true, "code": 200}`))
	}))
	defer server.Close()

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:      server.URL,
		TuyaClientID:     "client-id",
		TuyaClientSecret: "client-secret",
	}

	deviceService := services.NewTuyaDeviceService()
	// deviceStateUC is nil for this test to isolate control logic
	useCase := NewTuyaDeviceControlUseCase(deviceService, nil)

	commands := []dtos.TuyaCommandDTO{
		{Code: "switch_1", Value: true},
	}

	success, err := useCase.SendCommand("test-token", "dev-1", commands)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !success {
		t.Error("Expected success=true")
	}
}

func TestTuyaDeviceControlUseCase_SendCommand_RetryLogic(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	// Mock Server with 2 calls:
	// 1. Initial call -> fails with 2008
	// 2. Retry call -> succeeds
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call: iot-03 endpoint, return 2008 error
			if !strings.Contains(r.URL.Path, "/v1.0/iot-03/devices/dev-1/commands") {
				t.Errorf("Call 1: Unexpected URL: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": false, "code": 2008, "msg": "command not found"}`))
			return
		}
		if callCount == 2 {
			// Second call: legacy endpoint, return success
			if !strings.Contains(r.URL.Path, "/v1.0/devices/dev-1/commands") {
				t.Errorf("Call 2: Unexpected URL for retry: %s", r.URL.Path)
			}
			// Verify body has corrected code (switch_1 -> switch1)
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			cmds := body["commands"].([]interface{})
			firstCmd := cmds[0].(map[string]interface{})
			if firstCmd["code"] != "switch1" {
				t.Errorf("Expected corrected code 'switch1', got %v", firstCmd["code"])
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "result": true}`))
			return
		}
	}))
	defer server.Close()

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:      server.URL,
		TuyaClientID:     "client-id",
		TuyaClientSecret: "secret",
	}

	deviceService := services.NewTuyaDeviceService()
	useCase := NewTuyaDeviceControlUseCase(deviceService, nil)

	commands := []dtos.TuyaCommandDTO{
		{Code: "switch_1", Value: true},
	}

	success, err := useCase.SendCommand("token", "dev-1", commands)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !success {
		t.Error("Expected success after retry")
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestTuyaDeviceControlUseCase_SendIRACCommand_Standard(t *testing.T) {
	// Test normal IR command flow
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Expect call to fetch device details first
		if strings.Contains(r.URL.Path, "/v1.0/iot-03/devices/remote-1") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "result": {"id": "remote-1", "gateway_id": "gw-1"}}`))
			return
		}
		// Expect call to IR command
		if strings.Contains(r.URL.Path, "/v2.0/infrareds/gw-1/air-conditioners/remote-1/command") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "result": true}`))
			return
		}
		t.Logf("Unexpected URL: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:      server.URL,
		TuyaClientID:     "id",
		TuyaClientSecret: "sec",
	}

	deviceService := services.NewTuyaDeviceService()
	useCase := NewTuyaDeviceControlUseCase(deviceService, nil)

	ok, err := useCase.SendIRACCommand("token", "ir-1", "remote-1", "power", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !ok {
		t.Error("Expected success")
	}
}
