package usecases

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/services"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSyncDeviceStatusUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	// Mock valid Tuya Response: Flat list with Hub and Child (IR)
	// Hub: wnykq
	// Child: infrared_ac, gateway_id = hub_id
	mockDevs := `{
		"success": true,
		"result": [
			{
				"id": "hub-1",
				"name": "Smart Hub",
				"category": "wnykq",
				"online": true,
				"create_time": 1000
			},
			{
				"id": "child-1",
				"name": "AC Remote",
				"category": "infrared_ac",
				"gateway_id": "hub-1",
				"online": true,
				"create_time": 3000
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle Device List
		if r.URL.Path == "/v1.0/users/uid/devices" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockDevs))
			return
		}
		// Handle Specs (return success with empty functions to avoid unmarshal error)
		if strings.Contains(r.URL.Path, "specification") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "result": {"functions": []}}`))
			return
		}
		// Handle Status Batch
		if strings.Contains(r.URL.Path, "status") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "result": []}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	utils.AppConfig = &utils.Config{
		TuyaBaseURL:  server.URL,
		TuyaClientID: "id",
	}

	devService := services.NewTuyaDeviceService()
	// Pass nil for DeviceStateUseCase
	getAllUC := NewTuyaGetAllDevicesUseCase(devService, nil)
	syncUC := NewSyncDeviceStatusUseCase(getAllUC)

	// Execute
	devices, err := syncUC.Execute("token", "uid")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify flattening logic:
	// 1. hub-1
	// 2. child-1 (flattened from hub-1 collection)

	if len(devices) != 2 {
		t.Errorf("Expected 2 synced devices, got %d", len(devices))
	}

	// Check content
	foundChild := false
	for _, d := range devices {
		if d.ID == "child-1" {
			foundChild = true
			if !d.Online {
				t.Error("Child should be online (inherited/default)")
			}
			if d.CreateTime != 3000 {
				t.Error("Child create_time mismatch")
			}
		}
	}
	if !foundChild {
		t.Error("Child device from collection not flattened")
	}
}
