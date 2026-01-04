package dtos

import (
	"encoding/json"
	"testing"
)

func TestSensorDataDTO_JSON(t *testing.T) {
	dto := SensorDataDTO{
		Temperature:       25.5,
		Humidity:          60,
		BatteryPercentage: 90,
		StatusText:        "OK",
		TempUnit:          "C",
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	expected := map[string]interface{}{
		"temperature":        25.5,
		"humidity":           float64(60),
		"battery_percentage": float64(90),
		"status_text":        "OK",
		"temp_unit":          "C",
	}

	for k, v := range expected {
		if raw[k] != v {
			t.Errorf("Key %s: got %v, want %v", k, raw[k], v)
		}
	}
}
