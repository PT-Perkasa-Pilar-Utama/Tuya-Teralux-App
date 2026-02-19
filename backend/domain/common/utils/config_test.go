package utils

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Backup original env vars relative to this test
	originalClientID := os.Getenv("TUYA_CLIENT_ID")
	defer func() { _ = os.Setenv("TUYA_CLIENT_ID", originalClientID) }()

	// Set test env var
	testID := "test_client_id"
	_ = os.Setenv("TUYA_CLIENT_ID", testID)

	// Force reload
	AppConfig = nil // clear global singleton if possible, or just call LoadConfig which overwrites it.
	cfg := GetConfig()

	if cfg.TuyaClientID != testID {
		t.Errorf("GetConfig().TuyaClientID = %q; want %q", cfg.TuyaClientID, testID)
	}

	// Verify other fields are loaded (even if empty, structure should exist)
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
}

func TestLoadConfig_SetsValues(t *testing.T) {
	// Backup and restore env
	backup := map[string]string{}
	keys := []string{"GEMINI_MODEL", "WHISPER_LOCAL_MODEL", "MAX_FILE_SIZE_MB", "PORT", "CACHE_TTL"}
	for _, k := range keys {
		backup[k] = os.Getenv(k)
	}
	defer func() {
		for k, v := range backup {
			_ = os.Setenv(k, v)
		}
		AppConfig = nil
	}()

	_ = os.Setenv("GEMINI_MODEL", "gemini-test")
	_ = os.Setenv("WHISPER_LOCAL_MODEL", "/tmp/whisper.bin")
	_ = os.Setenv("MAX_FILE_SIZE_MB", "10")
	_ = os.Setenv("PORT", "9090")
	_ = os.Setenv("CACHE_TTL", "30m")

	AppConfig = nil
	LoadConfig()
	cfg := GetConfig()

	if cfg.GeminiModel != "gemini-test" {
		t.Fatalf("expected GeminiModel to be set, got %s", cfg.GeminiModel)
	}
	if cfg.WhisperLocalModel != "/tmp/whisper.bin" {
		t.Fatalf("expected WhisperLocalModel to be set, got %s", cfg.WhisperLocalModel)
	}
	if cfg.MaxFileSize != 10*1024*1024 {
		t.Fatalf("expected MaxFileSize to be 10MB in bytes, got %d", cfg.MaxFileSize)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected Port to be 9090, got %s", cfg.Port)
	}
}

func TestLoadConfig_InvalidMaxFileSize(t *testing.T) {
	backup := os.Getenv("MAX_FILE_SIZE_MB")
	defer func() { _ = os.Setenv("MAX_FILE_SIZE_MB", backup) }()
	_ = os.Setenv("MAX_FILE_SIZE_MB", "notanumber")
	AppConfig = nil
	LoadConfig()
	cfg := GetConfig()
	if cfg.MaxFileSize != 0 {
		t.Fatalf("expected MaxFileSize to be 0 on invalid input, got %d", cfg.MaxFileSize)
	}
}
