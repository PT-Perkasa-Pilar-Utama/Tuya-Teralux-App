package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
)

// OllamaClient represents the external LLM client used by RAG.
// This is an interface to allow testing with fakes.
type OllamaClient interface {
	CallModel(prompt string, model string) (string, error)
}

type RAGUsecase struct {
	vectorSvc  *infrastructure.VectorService
	ollama     OllamaClient
	config     *utils.Config
	mu         sync.RWMutex
	taskStatus map[string]*ragdtos.RAGStatusDTO
}

func NewRAGUsecase(vectorSvc *infrastructure.VectorService, ollama OllamaClient, cfg *utils.Config) *RAGUsecase {
	return &RAGUsecase{vectorSvc: vectorSvc, ollama: ollama, config: cfg, taskStatus: make(map[string]*ragdtos.RAGStatusDTO)}
}

// Process accepts user text, queries vector store for the best device match, asks the LLM to choose endpoint and body,
// and stores the result under a task ID which can later be fetched via GetStatus.
func (u *RAGUsecase) Process(text string) (string, error) {
	// Generate a simple task id
	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
	// Initially mark pending
	u.mu.Lock()
	u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "pending", Result: ""}
	u.mu.Unlock()

	// 1) Search vector DB for a device match
	candidates, err := u.vectorSvc.Search(text)
	if err != nil {
		return taskID, err
	}

	// Pick first device id that looks like device doc
	var chosenDeviceID string
	var deviceDoc string
	for _, id := range candidates {
		if strings.HasPrefix(id, "tuya:device:") {
			chosenDeviceID = strings.TrimPrefix(id, "tuya:device:")
			if doc, ok := u.vectorSvc.Get(id); ok {
				deviceDoc = doc
			}
			break
		}
	}

	// 2) Build prompt for LLM
	// Provide available endpoints and the chosen device doc (if any)
	prompt := "You are an assistant that maps user intent to one of these endpoints:\n"
	prompt += "1) POST /api/tuya/devices/{id}/commands/ir - Send IR AC Command\n"
	prompt += "2) POST /api/tuya/devices/{id}/commands/switch - Send Command to Device\n"
	prompt += "3) GET /api/tuya/devices/{id}/sensor - Get Sensor Data\n\n"
	prompt += fmt.Sprintf("User request: %s\n\n", text)
	if deviceDoc != "" {
		prompt += fmt.Sprintf("Candidate device example:\n%s\n\n", deviceDoc)
	}
	prompt += "Return a JSON object with keys: endpoint (string), method (GET/POST), device_id (string), body (object|null). Only output the JSON.\n"

	// 3) Call the LLM
	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}
	resp, err := u.ollama.CallModel(prompt, model)
	if err != nil {
		u.mu.Lock()
		u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: err.Error()}
		u.mu.Unlock()
		return taskID, err
	}

	// 4) Try to parse LLM response as JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		// If not JSON, store raw response
		u.mu.Lock()
		u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "done", Result: resp}
		u.mu.Unlock()
		utils.LogDebug("RAG Task %s: LLM returned non-JSON: %s", taskID, resp)
		return taskID, nil
	}

	// Extract fields
	endpoint, _ := parsed["endpoint"].(string)
	method, _ := parsed["method"].(string)
	// if device_id is provided, prefer it; else use chosenDeviceID
	deviceID, _ := parsed["device_id"].(string)
	if deviceID == "" {
		deviceID = chosenDeviceID
	}
	bodyObj := parsed["body"]
	bodyB, _ := json.Marshal(bodyObj)

	// Log the decision
	utils.LogDebug("RAG Task %s decided endpoint=%s method=%s device_id=%s body=%s", taskID, endpoint, method, deviceID, string(bodyB))

	// Store result
	resStr := fmt.Sprintf("endpoint=%s method=%s device_id=%s body=%s", endpoint, method, deviceID, string(bodyB))
	u.mu.Lock()
	u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "done", Result: resStr}
	u.mu.Unlock()

	return taskID, nil
}

func (u *RAGUsecase) GetStatus(taskID string) (*ragdtos.RAGStatusDTO, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if s, ok := u.taskStatus[taskID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("task not found")
}
