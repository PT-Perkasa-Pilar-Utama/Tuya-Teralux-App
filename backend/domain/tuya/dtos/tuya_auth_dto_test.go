package dtos

import (
	"encoding/json"
	"testing"
)

func TestTuyaAuthResponseDTO_JSON(t *testing.T) {
	dto := TuyaAuthResponseDTO{
		AccessToken:  "access-token-123",
		ExpireTime:   3600,
		RefreshToken: "refresh-token-456",
		UID:          "user-1",
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	expected := map[string]interface{}{
		"access_token":  "access-token-123",
		"expire_time":   float64(3600),
		"refresh_token": "refresh-token-456",
		"uid":           "user-1",
	}

	for k, v := range expected {
		val, ok := raw[k]
		if !ok {
			t.Errorf("Missing key %s", k)
		}
		if val != v {
			t.Errorf("Key %s mismatch: got %v, want %v", k, val, v)
		}
	}
}

func TestErrorResponseDTO_JSON(t *testing.T) {
	dto := ErrorResponseDTO{
		Error:   "invalid_request",
		Message: "Missing params",
	}

	data, _ := json.Marshal(dto)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["error"] != "invalid_request" {
		t.Errorf("Expected error 'invalid_request', got %v", raw["error"])
	}
	if raw["message"] != "Missing params" {
		t.Errorf("Expected message 'Missing params', got %v", raw["message"])
	}
}
