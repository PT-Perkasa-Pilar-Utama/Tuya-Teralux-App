package usecases

import (
	"encoding/json"
	"os"
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

	// verify structured result contains chosen endpoint, method and body
	if status.Endpoint != "/api/tuya/devices/{id}/commands/switch" {
		t.Fatalf("expected endpoint /api/tuya/devices/{id}/commands/switch, got %s", status.Endpoint)
	}
	if status.Method != "POST" {
		t.Fatalf("expected method POST, got %s", status.Method)
	}
	// verify body structure
	bodyMap, ok := status.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected body to be object, got %+v", status.Body)
	}
	if _, ok := bodyMap["commands"]; !ok {
		t.Fatalf("expected body to contain commands, got %+v", bodyMap)
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

	// verify pending cached in badger immediately
	b1, err := badgerSvc.Get("rag:task:" + task)
	if err != nil {
		t.Fatalf("failed to read from badger: %v", err)
	}
	if b1 == nil {
		t.Fatalf("expected pending task to be cached in badger")
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
	// Simulate a restart by creating a new usecase instance without in-memory cache
	u2 := NewRAGUsecase(vectorSvc, fake, utils.GetConfig(), badgerSvc)
	status2, err := u2.GetStatus(task)
	if err != nil {
		t.Fatalf("expected persisted task to be found after restart, got error: %v", err)
	}
	if status2.Status != "done" {
		t.Fatalf("expected persisted status done, got %s", status2.Status)
	}
}

func TestPendingCachedWithTTLAndPreservedOnFinalize(t *testing.T) {
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

	// Prepare a temporary badger DB dir and short TTL
	dbDir := "./tmp/badger-test-rag-ttl"
	_ = os.RemoveAll(dbDir)
	// override cache ttl to small value for test
	origConfig := utils.AppConfig
	utils.AppConfig = &utils.Config{CacheTTL: "1s"}
	defer func() { utils.AppConfig = origConfig }()

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

	// verify pending cached in Badger immediately and TTL exists
	_, ttl1, err := badgerSvc.GetWithTTL("rag:task:" + task)
	if err != nil {
		t.Fatalf("failed to read from badger: %v", err)
	}
	if ttl1 <= 0 {
		t.Fatalf("expected TTL > 0 for cached pending task, got %v", ttl1)
	}

	// wait a bit and then finalization should preserve TTL (not extend it)
	time.Sleep(300 * time.Millisecond)

	// wait for completion
	time.Sleep(400 * time.Millisecond)

	status, err := u.GetStatus(task)
	if err != nil {
		t.Fatalf("expected no error from GetStatus after completion, got %v", err)
	}
	if status.Status != "done" {
		t.Fatalf("expected status done after completion, got %s", status.Status)
	}

	// Check TTL after finalize
	_, ttl2, err := badgerSvc.GetWithTTL("rag:task:" + task)
	if err != nil {
		t.Fatalf("failed to read from badger after finalize: %v", err)
	}
	// If ttl2 is 0 it might have expired while we waited; accept that. Otherwise it must not have increased.
	if ttl2 > 0 {
		// TTL should not have increased; allow some small margin
		if ttl2 > ttl1+200*time.Millisecond {
			t.Fatalf("expected TTL not to be extended on finalize (ttl1=%v ttl2=%v)", ttl1, ttl2)
		}
	}
}
