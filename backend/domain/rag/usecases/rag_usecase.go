package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
	badger     *infrastructure.BadgerService
	mu         sync.RWMutex
	taskStatus map[string]*ragdtos.RAGStatusDTO
}

// extractLastJSONObject scans a string that may contain multiple JSON objects
// (for example, streaming fragments concatenated together) and returns the last
// successfully parsed JSON object. This helps recover structured output from
// tokenized/streaming LLM responses.
func extractLastJSONObject(s string) (map[string]interface{}, error) {
	var last map[string]interface{}
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		depth := 0
		for j := i; j < len(s); j++ {
			if s[j] == '{' {
				depth++
			} else if s[j] == '}' {
				depth--
				if depth == 0 {
					candidate := s[i : j+1]
					var parsed map[string]interface{}
					if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
						last = parsed
					}
					break
				}
			}
		}
	}
	if last == nil {
		return nil, fmt.Errorf("no valid JSON object found")
	}
	return last, nil
}

// tryParseOnce attempts to parse the provided string as JSON. It first
// unmarshals the whole string and, if that fails, attempts to extract
// the last valid JSON object from streaming/fractured output.
func tryParseOnce(s string) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err == nil && parsed != nil {
		return parsed, nil
	}
	return extractLastJSONObject(s)
}

func NewRAGUsecase(vectorSvc *infrastructure.VectorService, ollama OllamaClient, cfg *utils.Config, badgerSvc *infrastructure.BadgerService) *RAGUsecase {
	return &RAGUsecase{vectorSvc: vectorSvc, ollama: ollama, config: cfg, badger: badgerSvc, taskStatus: make(map[string]*ragdtos.RAGStatusDTO)}
}

// Process accepts user text, queues work to query the vector store and LLM, and stores the result under a task ID
// which can later be fetched via GetStatus. Processing is done asynchronously and this method returns immediately.
func (u *RAGUsecase) Process(text string) (string, error) {
	// Generate UUID task id
	taskID := uuid.New().String()
	// Initially mark pending
	u.mu.Lock()
	u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "pending", Result: ""}
	pending := u.taskStatus[taskID]
	u.mu.Unlock()
	// persist pending to cache (with TTL) if available
	if u.badger != nil {
		b, _ := json.Marshal(pending)
		if err := u.badger.Set("rag:task:"+taskID, b); err != nil {
			utils.LogError("RAG Task %s: failed to cache pending task: %v", taskID, err)
		} else {
			utils.LogDebug("RAG Task %s: pending cached with TTL", taskID)
		}
	}

	// Run processing asynchronously
	go func(taskID, text string) {
		defer func() {
			if r := recover(); r != nil {
				u.mu.Lock()
				u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: fmt.Sprintf("panic: %v", r)}
				u.mu.Unlock()
			}
		}()

		// 1) Search vector DB for a device match
		candidates, err := u.vectorSvc.Search(text)
		if err != nil {
			u.mu.Lock()
			u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: err.Error()}
			u.mu.Unlock()
			return
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
			return
		}

		// Log prompt and raw response for debugging
		utils.LogDebug("RAG Task %s prompt: %s", taskID, prompt)
		utils.LogDebug("RAG Task %s raw LLM response: %s", taskID, resp)

		// 4) Try to parse LLM response as JSON (flattened flow to avoid nested ifs)
		var parsed map[string]interface{}
		var perr error
		parsed, perr = tryParseOnce(resp)
		if perr != nil || parsed == nil {
			utils.LogDebug("RAG Task %s: initial parse/extract failed: %v", taskID, perr)
			// Retry once with a stricter instruction
			utils.LogDebug("RAG Task %s: retrying with stricter instruction", taskID)
			retryPrompt := prompt + "\nIMPORTANT: Your output must be valid JSON ONLY, with keys: endpoint, method, device_id, body. Do not add any extra text."
			resp2, err2 := u.ollama.CallModel(retryPrompt, model)
			if err2 != nil {
				utils.LogDebug("RAG Task %s: LLM retry error: %v", taskID, err2)
			} else {
				parsed, perr = tryParseOnce(resp2)
				if perr != nil || parsed == nil {
					utils.LogDebug("RAG Task %s: retry also failed to produce JSON: %v", taskID, perr)
				} else {
					utils.LogDebug("RAG Task %s: extracted JSON from retry", taskID)
				}
			}
		}

		if parsed == nil {
			// If still not JSON, store raw response and mark done
			u.mu.Lock()
			statusDTO := &ragdtos.RAGStatusDTO{Status: "done", Result: resp}
			u.taskStatus[taskID] = statusDTO
			u.mu.Unlock()
			utils.LogDebug("RAG Task %s: LLM returned non-JSON after retry: %s", taskID, resp)
			// persist final result by updating existing cache entry while preserving TTL
			if u.badger != nil {
				b, _ := json.Marshal(statusDTO)
				_ = u.badger.SetPreserveTTL("rag:task:"+taskID, b)
			}
			return
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

		// Store structured result (do NOT perform external fetch)
		statusDTO := &ragdtos.RAGStatusDTO{Status: "done", Endpoint: endpoint, Method: method, Body: bodyObj}
		u.mu.Lock()
		u.taskStatus[taskID] = statusDTO
		u.mu.Unlock()
		// persist final result by updating existing cache entry while preserving TTL
		if u.badger != nil {
			b, _ := json.Marshal(statusDTO)
			if err := u.badger.SetPreserveTTL("rag:task:"+taskID, b); err != nil {
				utils.LogError("RAG Task %s: failed to update cached final result: %v", taskID, err)
			} else {
				utils.LogDebug("RAG Task %s: final result cached (TTL preserved)", taskID)
			}
		}
		return
	}(taskID, text)

	return taskID, nil
}

func (u *RAGUsecase) GetStatus(taskID string) (*ragdtos.RAGStatusDTO, error) {
	// First try in-memory map with read lock
	u.mu.RLock()
	if s, ok := u.taskStatus[taskID]; ok {
		u.mu.RUnlock()
		// augment with TTL info if available
		if u.badger != nil {
			key := "rag:task:" + taskID
			_, ttl, err := u.badger.GetWithTTL(key)
			if err == nil && ttl > 0 {
				s.ExpiresInSecond = int64(ttl.Seconds())
				s.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
		}
		return s, nil
	}
	u.mu.RUnlock()

	// If not found in-memory, try persistent store (Badger) if configured
	if u.badger != nil {
		key := "rag:task:" + taskID
		b, ttl, err := u.badger.GetWithTTL(key)
		if err != nil {
			return nil, fmt.Errorf("failed to read persistent task: %w", err)
		}
		if b != nil {
			var status ragdtos.RAGStatusDTO
			if err := json.Unmarshal(b, &status); err == nil {
				// Cache into memory for faster subsequent reads
				u.mu.Lock()
				u.taskStatus[taskID] = &status
				u.mu.Unlock()
				if ttl > 0 {
					status.ExpiresInSecond = int64(ttl.Seconds())
					status.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
				}
				utils.LogDebug("RAG Task %s: retrieved from badger, ttl=%v", taskID, ttl)
				return &status, nil
			}
		}
		// Not found in badger either
		utils.LogDebug("RAG Task %s: not found in cache", taskID)
	}

	return nil, fmt.Errorf("task not found")
}
