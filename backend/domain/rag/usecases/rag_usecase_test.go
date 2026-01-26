package usecases

import (
	"encoding/json"
	"strings"
	"testing"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
)

// fake Ollama client for testing
type fakeOllama struct {
	resp string
}

func (f *fakeOllama) CallModel(prompt string, model string) (string, error) {
	return f.resp, nil
}

func TestRAGUsecase_ProcessAndGetStatus(t *testing.T) {
	utils.LoadConfig()
	vectorSvc := infrastructure.NewVectorService()
	// seed a device doc containing 'lamp'
	deviceDoc := map[string]interface{}{"id": "lamp123", "name": "Living Room Lamp", "category": "switch"}
	b, _ := json.Marshal(deviceDoc)
	vectorSvc.Upsert("tuya:device:lamp123", string(b), nil)

	// Prepare fake LLM response
	llmResp := map[string]interface{}{
		"endpoint":  "/api/tuya/devices/{id}/commands/switch",
		"method":    "POST",
		"device_id": "lamp123",
		"body":      map[string]interface{}{"commands": []map[string]interface{}{{"code": "switch", "value": 1}}},
	}
	rb, _ := json.Marshal(llmResp)
	fake := &fakeOllama{resp: string(rb)}

	u := NewRAGUsecase(vectorSvc, fake, utils.GetConfig())

	task, err := u.Process("turn on the lamp")
	if err != nil {
		t.Fatalf("expected no error from Process, got %v", err)
	}
	if task == "" {
		t.Fatalf("expected non-empty task id")
	}

	status, err := u.GetStatus(task)
	if err != nil {
		t.Fatalf("expected no error from GetStatus, got %v", err)
	}
	if status == nil || status.Status == "" {
		t.Fatalf("expected valid status result, got %+v", status)
	}

	// verify result contains chosen endpoint and device
	if !strings.Contains(status.Result, "endpoint=") || !strings.Contains(status.Result, "device_id=lamp123") {
		t.Fatalf("unexpected result: %s", status.Result)
	}
}
