package utils

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Backup original env vars relative to this test
	originalClientID := os.Getenv("TUYA_CLIENT_ID")
	defer os.Setenv("TUYA_CLIENT_ID", originalClientID)

	// Set test env var
	testID := "test_client_id"
	os.Setenv("TUYA_CLIENT_ID", testID)

	// Force reload
	AppConfig = nil // clear global singleton if possible, or just call LoadConfig which overwrites it.
	// Note: AppConfig is exported, so we can nil it to test GetConfig lazy load
	AppConfig = nil
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
	keys := []string{"OLLAMA_URL", "LLM_MODEL", "WHISPER_MODEL_PATH", "MAX_FILE_SIZE_MB", "PORT", "GET_ALL_DEVICES_RESPONSE", "CACHE_TTL"}
	for _, k := range keys {
		backup[k] = os.Getenv(k)
	}
	defer func() {
		for k, v := range backup {
			os.Setenv(k, v)
		}
		AppConfig = nil
	}()

	os.Setenv("OLLAMA_URL", "http://test-ollama")
	os.Setenv("LLM_MODEL", "gemma-test")
	os.Setenv("WHISPER_MODEL_PATH", "/tmp/whisper.bin")
	os.Setenv("MAX_FILE_SIZE_MB", "10")
	os.Setenv("PORT", "9090")
	os.Setenv("GET_ALL_DEVICES_RESPONSE", "1")
	os.Setenv("CACHE_TTL", "30m")

	AppConfig = nil
	LoadConfig()
	cfg := GetConfig()

	if cfg.OllamaURL != "http://test-ollama" {
		t.Fatalf("expected OllamaURL to be set, got %s", cfg.OllamaURL)
	}
	if cfg.LLMModel != "gemma-test" {
		t.Fatalf("expected LLMModel to be set, got %s", cfg.LLMModel)
	}
	if cfg.WhisperModelPath != "/tmp/whisper.bin" {
		t.Fatalf("expected WhisperModelPath to be set, got %s", cfg.WhisperModelPath)
	}
	if cfg.MaxFileSize != 10*1024*1024 {
		t.Fatalf("expected MaxFileSize to be 10MB in bytes, got %d", cfg.MaxFileSize)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected Port to be 9090, got %s", cfg.Port)
	}
	if cfg.GetAllDevicesResponseType != "1" {
		t.Fatalf("expected GetAllDevicesResponseType to be '1', got %s", cfg.GetAllDevicesResponseType)
	}
}

func TestLoadConfig_InvalidMaxFileSize(t *testing.T) {
	backup := os.Getenv("MAX_FILE_SIZE_MB")
	defer os.Setenv("MAX_FILE_SIZE_MB", backup)
	os.Setenv("MAX_FILE_SIZE_MB", "notanumber")
	AppConfig = nil
	LoadConfig()
	cfg := GetConfig()
	if cfg.MaxFileSize != 0 {
		t.Fatalf("expected MaxFileSize to be 0 on invalid input, got %d", cfg.MaxFileSize)
	}
}
