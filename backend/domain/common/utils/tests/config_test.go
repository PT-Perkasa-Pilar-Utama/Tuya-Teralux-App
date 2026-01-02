package tests

import (
	"os"
	"testing"
	"teralux_app/domain/common/utils"
)

func TestLoadConfig(t *testing.T) {
	// Backup original env vars relative to this test
	originalClientID := os.Getenv("TUYA_CLIENT_ID")
	defer os.Setenv("TUYA_CLIENT_ID", originalClientID)

	// Set test env var
	testID := "test_client_id"
	os.Setenv("TUYA_CLIENT_ID", testID)

	// Force reload
	utils.AppConfig = nil // clear global singleton if possible, or just call LoadConfig which overwrites it.
	// Note: utils.AppConfig is exported, so we can nil it to test GetConfig lazy load
	utils.AppConfig = nil
	cfg := utils.GetConfig()

	if cfg.TuyaClientID != testID {
		t.Errorf("GetConfig().TuyaClientID = %q; want %q", cfg.TuyaClientID, testID)
	}

	// Verify other fields are loaded (even if empty, structure should exist)
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
}
