package usecases

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
)

// fake Ollama client for testing
type fakeOllama struct {
	resp string
}

func (f *fakeOllama) CallModel(prompt string, model string) (string, error) {
	return f.resp, nil
}

func TestControlUseCase_Execute(t *testing.T) {
	utils.LoadConfig()
	vectorSvc := infrastructure.NewVectorService("")
	// seed a device doc containing 'lamp'
	deviceDoc := map[string]interface{}{"id": "lamp123", "name": "Living Room Lamp", "category": "switch"}
	b, _ := json.Marshal(deviceDoc)
	_ = vectorSvc.Upsert("tuya:device:lamp123", string(b), nil)

	// Prepare fake LLM response
	llmResp := map[string]interface{}{
		"endpoint":  "/api/tuya/devices/lamp123/commands/switch",
		"method":    "POST",
		"device_id": "lamp123",
		"body":      map[string]interface{}{"commands": []map[string]interface{}{{"code": "switch", "value": 1}}},
	}
	rb, _ := json.Marshal(llmResp)
	fake := &fakeOllama{resp: string(rb)}
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

	u := NewControlUseCase(vectorSvc, fake, utils.GetConfig(), nil, nil, store)
	statusUC := tasks.NewGenericStatusUseCase[ragdtos.RAGStatusDTO](nil, store)

	task, err := u.ControlFromText("turn on the lamp", "mock-token", nil)
	if err != nil {
		t.Fatalf("expected no error from Execute, got %v", err)
	}
	if task == "" {
		t.Fatalf("expected non-empty task id")
	}

	status, err := statusUC.GetTaskStatus(task)
	if err != nil {
		t.Fatalf("expected no error from StatusUseCase.Execute, got %v", err)
	}
	if status == nil || status.Status == "" {
		t.Fatalf("expected valid status result, got %+v", status)
	}
	if status.Status != "pending" {
		t.Fatalf("expected pending status right after Execute, got %s", status.Status)
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
			status, err = statusUC.GetTaskStatus(task)
			if err != nil {
				t.Fatalf("expected no error from StatusUseCase.Execute during polling, got %v", err)
			}
			if status.Status == "completed" {
				done = true
			} else if status.Status == "failed" {
				t.Fatalf("task failed: %v", status.Result)
			}
		}
	}

	// verify structured result contains chosen endpoint, method and body
	if status.Endpoint != "/api/tuya/devices/lamp123/commands/switch" {
		t.Fatalf("expected endpoint /api/tuya/devices/lamp123/commands/switch, got %s", status.Endpoint)
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
	vectorSvc := infrastructure.NewVectorService("")
	// seed device doc
	deviceDoc := map[string]interface{}{"id": "lamp123", "name": "Living Room Lamp", "category": "switch"}
	b, _ := json.Marshal(deviceDoc)
	_ = vectorSvc.Upsert("tuya:device:lamp123", string(b), nil)

	// fake LLM as before
	llmResp := map[string]interface{}{
		"endpoint":  "/api/tuya/devices/lamp123/commands/switch",
		"method":    "POST",
		"device_id": "lamp123",
		"body":      map[string]interface{}{"commands": []map[string]interface{}{{"code": "switch", "value": 1}}},
	}
	rb, _ := json.Marshal(llmResp)
	fake := &fakeOllama{resp: string(rb)}
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

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
	cache := tasks.NewBadgerTaskCache(badgerSvc, "rag:task:")

	u := NewControlUseCase(vectorSvc, fake, utils.GetConfig(), cache, nil, store)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)

	task, err := u.ControlFromText("turn on the lamp", "mock-token", nil)
	if err != nil {
		t.Fatalf("expected no error from Execute, got %v", err)
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

	status, err := statusUC.GetTaskStatus(task)
	if err != nil {
		t.Fatalf("expected no error from StatusUseCase.Execute after completion, got %v", err)
	}
	if status.Status != "completed" {
		t.Fatalf("expected status completed after completion, got %s", status.Status)
	}

	// Now GetStatus should retrieve the persisted copy from Badger even if memory is empty
	newStore := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()
	statusUC2 := tasks.NewGenericStatusUseCase(cache, newStore)
	status2, err := statusUC2.GetTaskStatus(task)
	if err != nil {
		t.Fatalf("expected persisted task to be found after restart, got error: %v", err)
	}
	if status2.Status != "completed" {
		t.Fatalf("expected persisted status completed, got %s", status2.Status)
	}
}

func TestPendingCachedWithTTLAndPreservedOnFinalize(t *testing.T) {
	utils.LoadConfig()
	vectorSvc := infrastructure.NewVectorService("")
	// seed device doc
	deviceDoc := map[string]interface{}{"id": "lamp123", "name": "Living Room Lamp", "category": "switch"}
	b, _ := json.Marshal(deviceDoc)
	_ = vectorSvc.Upsert("tuya:device:lamp123", string(b), nil)

	// fake LLM as before
	llmResp := map[string]interface{}{
		"endpoint":  "/api/tuya/devices/lamp123/commands/switch",
		"method":    "POST",
		"device_id": "lamp123",
		"body":      map[string]interface{}{"commands": []map[string]interface{}{{"code": "switch", "value": 1}}},
	}
	rb, _ := json.Marshal(llmResp)
	fake := &fakeOllama{resp: string(rb)}
	store := tasks.NewStatusStore[ragdtos.RAGStatusDTO]()

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
	cache := tasks.NewBadgerTaskCache(badgerSvc, "rag:task:")

	u := NewControlUseCase(vectorSvc, fake, utils.GetConfig(), cache, nil, store)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)

	task, err := u.ControlFromText("turn on the lamp", "mock-token", nil)
	if err != nil {
		t.Fatalf("expected no error from Execute, got %v", err)
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

	status, err := statusUC.GetTaskStatus(task)
	if err != nil {
		t.Fatalf("expected no error from StatusUseCase.Execute after completion, got %v", err)
	}
	if status.Status != "completed" {
		t.Fatalf("expected status completed after completion, got %s", status.Status)
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
