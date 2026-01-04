package dtos

import (
	"encoding/json"
	"testing"
)

func TestTuyaDeviceDTO_JSON(t *testing.T) {
	status := []TuyaDeviceStatusDTO{
		{Code: "switch_1", Value: true},
	}

	dto := TuyaDeviceDTO{
		ID:          "dev-1",
		Name:        "Smart Light",
		Category:    "dj",
		ProductName: "Light Bulb",
		Online:      true,
		Status:      status,
		LocalKey:    "key123",
		GatewayID:   "gw1",
		CreateTime:  1000,
		UpdateTime:  2000,
		// Optional fields empty
	}

	data, _ := json.Marshal(dto)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	// Check core fields
	if raw["id"] != "dev-1" {
		t.Error("ID mismatch")
	}
	if raw["category"] != "dj" {
		t.Error("Category mismatch")
	}

	// Check Status array
	statusList, ok := raw["status"].([]interface{})
	if !ok || len(statusList) != 1 {
		t.Error("Status array invalid")
	} else {
		item := statusList[0].(map[string]interface{})
		if item["code"] != "switch_1" {
			t.Error("Status code mismatch")
		}
		if item["value"] != true {
			t.Error("Status value mismatch")
		}
	}

	// Check omitempty fields
	if _, ok := raw["remote_id"]; ok {
		t.Error("remote_id should be omitted")
	}
	if _, ok := raw["collections"]; ok {
		t.Error("collections should be omitted")
	}
}

func TestTuyaCommandsRequestDTO_JSON(t *testing.T) {
	cmd := TuyaCommandDTO{Code: "switch_1", Value: true}
	req := TuyaCommandsRequestDTO{
		Commands: []TuyaCommandDTO{cmd},
	}

	data, _ := json.Marshal(req)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	cmds, ok := raw["commands"].([]interface{})
	if !ok || len(cmds) != 1 {
		t.Fatal("Commands parsing failed")
	}

	first := cmds[0].(map[string]interface{})
	if first["code"] != "switch_1" || first["value"] != true {
		t.Error("Command content mismatch")
	}
}

func TestDeviceStateDTO_JSON(t *testing.T) {
	cmd := DeviceStateCommandDTO{Code: "temp", Value: 25}
	state := DeviceStateDTO{
		DeviceID:     "d1",
		LastCommands: []DeviceStateCommandDTO{cmd},
		UpdatedAt:    12345,
	}

	data, _ := json.Marshal(state)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["device_id"] != "d1" {
		t.Error("device_id mismatch")
	}
	if raw["updated_at"] != float64(12345) {
		t.Error("updated_at mismatch")
	}
}
