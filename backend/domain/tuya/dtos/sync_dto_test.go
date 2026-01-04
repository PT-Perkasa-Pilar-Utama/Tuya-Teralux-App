package dtos

import (
	"encoding/json"
	"testing"
)

func TestTuyaSyncDeviceDTO_JSON(t *testing.T) {
	// Create a sample DTO
	dto := TuyaSyncDeviceDTO{
		ID:         "dev-1",
		RemoteID:   "rem-1",
		Online:     true,
		CreateTime: 1622548800,
		UpdateTime: 1622552400,
	}

	// Marshal to JSON
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("Failed to marshal TuyaSyncDeviceDTO: %v", err)
	}

	// Unmarshal back to map to verify keys (locking down json tags)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Expected keys and values
	expected := map[string]interface{}{
		"id":          "dev-1",
		"remote_id":   "rem-1",
		"online":      true,
		"create_time": float64(1622548800), // JSON numbers are floats in Go interface{}
		"update_time": float64(1622552400),
	}

	for k, v := range expected {
		if val, ok := raw[k]; !ok {
			t.Errorf("Missing JSON key %s", k)
		} else if val != v {
			t.Errorf("Key %s: expected %v, got %v", k, v, val)
		}
	}

	// Test omitempty
	dtoEmpty := TuyaSyncDeviceDTO{ID: "dev-2"}
	dataEmpty, _ := json.Marshal(dtoEmpty)
	if string(dataEmpty) == "" {
		t.Error("Empty serialization result")
	}

	// Unmarshal to map to check missing remote_id
	var rawEmpty map[string]interface{}
	json.Unmarshal(dataEmpty, &rawEmpty)

	if _, ok := rawEmpty["remote_id"]; ok {
		t.Error("remote_id should be omitted when empty")
	}
}
