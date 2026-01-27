package usecases

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

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

	u := NewRAGUsecase(vectorSvc, fake, utils.GetConfig(), nil)

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
	if status.Status != "pending" {
		t.Fatalf("expected pending status right after Process, got %s", status.Status)
	}

	// wait for completion
	timeout := time.After(2 * time.Second)
	tick := time.Tick(50 * time.Millisecond)
	done := false
	for !done {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for task to complete, last status: %+v", status)
		case <-tick:
			status, err = u.GetStatus(task)
			if err != nil {
				t.Fatalf("expected no error from GetStatus during polling, got %v", err)
			}
			if status.Status == "done" {
				done = true
			} else if status.Status == "error" {
				t.Fatalf("task failed: %s", status.Result)
			}
		}
	}

	// verify result contains chosen endpoint and device
	if !strings.Contains(status.Result, "endpoint=") || !strings.Contains(status.Result, "device_id=lamp123") {
		t.Fatalf("unexpected result: %s", status.Result)
	}
}

func TestPersistentStorageAfterCompletion(t *testing.T) {
	utils.LoadConfig()
	vectorSvc := infrastructure.NewVectorService()
	// seed device doc
	deviceDoc := map[string]interface{}{"id": "lamp123", "name": "Living Room Lamp", "category": "switch"}
	b, _ := json.Marshal(deviceDoc)
	vectorSvc.Upsert("tuya:device:lamp123", string(b), nil)

	// fake LLM as before
	llmResp := map[string]interface{}{
		"endpoint":  "/api/tuya/devices/{id}/commands/switch",
		"method":    "POST",
		"device_id": "lamp123",
		"body":      map[string]interface{}{"commands": []map[string]interface{}{{"code": "switch", "value": 1}}},
	}
	rb, _ := json.Marshal(llmResp)
	fake := &fakeOllama{resp: string(rb)}

	// Prepare a temporary badger DB dir
	dbDir := "./tmp/badger-test-rag"
	_ = os.RemoveAll(dbDir)
	badgerSvc, err := infrastructure.NewBadgerService(dbDir)
	if err != nil {
		t.Fatalf("failed to create badger service: %v", err)
	}
	defer func() {
		badgerSvc.Close()
		_ = os.RemoveAll(dbDir)
	}()

	u := NewRAGUsecase(vectorSvc, fake, utils.GetConfig(), badgerSvc)

	task, err := u.Process("turn on the lamp")
	if err != nil {
		t.Fatalf("expected no error from Process, got %v", err)
	}

	// wait for completion
	time.Sleep(200 * time.Millisecond)

	status, err := u.GetStatus(task)
	if err != nil {
		t.Fatalf("expected no error from GetStatus after completion, got %v", err)
	}
	if status.Status != "done" {
		t.Fatalf("expected status done after completion, got %s", status.Status)
	}

	// Remove in-memory entry to simulate restart/eviction
	u.mu.Lock()
	delete(u.taskStatus, task)
	u.mu.Unlock()

	// Now GetStatus should retrieve the persisted copy from Badger
	status2, err := u.GetStatus(task)
	if err != nil {
		t.Fatalf("expected persisted task to be found, got error: %v", err)
	}
	if status2.Status != "done" {
		t.Fatalf("expected persisted status done, got %s", status2.Status)
	}
}
