package usecases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"

	"github.com/google/uuid"
)

// LLMClient represents the external LLM client used by RAG.
// This is an interface to allow testing with fakes.
type LLMClient interface {
	CallModel(prompt string, model string) (string, error)
}

type RAGUsecase struct {
	vectorSvc  *infrastructure.VectorService
	llm        LLMClient
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
	// First, try to clean markdown fences if present
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			// Find first line with { and last line with }
			var start, end int = -1, -1
			for i, line := range lines {
				if strings.Contains(line, "{") && start == -1 {
					start = i
				}
				if strings.Contains(line, "}") {
					end = i
				}
			}
			if start != -1 && end != -1 && end >= start {
				candidate := strings.Join(lines[start:end+1], "\n")
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
					return parsed, nil
				}
			}
		}
	}

	// Fallback to brace counting logic
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
					// Don't break, keep looking for potentially better/longer/later JSON
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

func NewRAGUsecase(vectorSvc *infrastructure.VectorService, llm LLMClient, cfg *utils.Config, badgerSvc *infrastructure.BadgerService) *RAGUsecase {
	return &RAGUsecase{vectorSvc: vectorSvc, llm: llm, config: cfg, badger: badgerSvc, taskStatus: make(map[string]*ragdtos.RAGStatusDTO)}
}

// Process accepts user text, queues work to query the vector store and LLM, and stores the result under a task ID
// which can later be fetched via GetStatus. Processing is done asynchronously and this method returns immediately.
func (u *RAGUsecase) Process(text string, authToken string) (string, error) {
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
	go func(taskID, text, authToken string) {
		utils.LogInfo("RAG Task %s: Started processing user request: %q", taskID, text)
		utils.LogInfo("--- Available Tuya Endpoints ---")
		utils.LogInfo("1) Send IR AC Command")
		utils.LogInfo("   URL:      /api/tuya/devices/{id}/commands/ir")
		utils.LogInfo("   Method:   POST")
		utils.LogInfo("   Body:     { \"remote_id\": \"string\", \"code\": \"string\", \"value\": integer }")
		utils.LogInfo("   Headers:  { \"Authorization\": \"Bearer <token>\", \"Content-Type\": \"application/json\" }")
		utils.LogInfo("2) Send Command to Device (Switch/Light/etc)")
		utils.LogInfo("   URL:      /api/tuya/devices/{id}/commands/switch")
		utils.LogInfo("   Method:   POST")
		utils.LogInfo("   Body:     { \"code\": \"string\", \"value\": boolean|integer|string }")
		utils.LogInfo("   Headers:  { \"Authorization\": \"Bearer <token>\", \"Content-Type\": \"application/json\" }")
		utils.LogInfo("3) Get Sensor Data")
		utils.LogInfo("   URL:      /api/tuya/devices/{id}/sensor")
		utils.LogInfo("   Method:   GET")
		utils.LogInfo("   Body:     null")
		utils.LogInfo("   Headers:  { \"Authorization\": \"Bearer <token>\" }")
		utils.LogInfo("--------------------------------")

		defer func() {
			if r := recover(); r != nil {
				u.mu.Lock()
				u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: fmt.Sprintf("panic: %v", r)}
				u.mu.Unlock()
			}
		}()

		// 0) Step 0: Grammar Correction (Fix pronunciation or transcription artifacts)
		correctedText := text
		grammarPrompt := fmt.Sprintf("You are a smart home transcription assistant. Fix any pronunciation errors, 'word salads', or grammar mistakes in the following user request while keeping its core intent and language. Return ONLY the corrected text.\n\nRequest: %q", text)

		utils.LogInfo("RAG Task %s: Correcting grammar/pronunciation...", taskID)
		if gResp, gErr := u.llm.CallModel(grammarPrompt, u.config.LLMModel); gErr == nil {
			gResp = strings.TrimSpace(gResp)
			// Remove surrounding quotes if model added them
			gResp = strings.Trim(gResp, `"'`)
			if gResp != "" {
				utils.LogInfo("RAG Task %s: Corrected text: %q -> %q", taskID, text, gResp)
				correctedText = gResp
			}
		} else {
			utils.LogWarn("RAG Task %s: Grammar correction failed, falling back to original: %v", taskID, gErr)
		}

		// 1) Search vector DB for all potential device matches using CORRECTED text
		docCount := u.vectorSvc.Count()
		utils.LogInfo("RAG Task %s: Searching vector DB (%d documents total) with text: %q", taskID, docCount, correctedText)
		candidates, err := u.vectorSvc.Search(correctedText)
		if err != nil {
			u.mu.Lock()
			u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: err.Error()}
			u.mu.Unlock()
			return
		}

		// Collect all device documents found
		var deviceContexts []string
		for _, id := range candidates {
			if strings.HasPrefix(id, "tuya:device:") {
				if doc, ok := u.vectorSvc.Get(id); ok {
					deviceContexts = append(deviceContexts, doc)
				}
			}
		}

		if len(deviceContexts) > 0 {
			utils.LogInfo("RAG Task %s: Found %d potential device matches in Vector DB", taskID, len(deviceContexts))
			// Extract and log names for easier debugging
			for i, ctx := range deviceContexts {
				var d map[string]interface{}
				if err := json.Unmarshal([]byte(ctx), &d); err == nil {
					name, _ := d["name"].(string)
					utils.LogDebug("Candidate %d: %s (ID: %v)", i+1, name, d["id"])
				}
			}
		} else {
			utils.LogInfo("RAG Task %s: No specific device context found in Vector DB", taskID)
		}

		// 2) Build prompt for LLM
		prompt := "You are a smart home assistant. Compare the user's request with the available devices and endpoints to decide the best action.\n\n"
		prompt += "--- Available Endpoints ---\n"
		prompt += "1) Send IR AC Command (for AC units controlled via a Smart IR Hub)\n"
		prompt += "   URL:      /api/tuya/devices/{hub_id}/commands/ir\n"
		prompt += "   Method:   POST\n"
		prompt += "   Body:     { \"remote_id\": \"string\", \"code\": \"string\", \"value\": integer }\n"
		prompt += "   CRITICAL for Endpoint 1:\n"
		prompt += "     - {hub_id} in URL MUST be the 'id' field of the Smart Hub.\n"
		prompt += "     - 'remote_id' in Body MUST be the 'remote_id' field of the same device context.\n"
		prompt += "     - Codes: 'temp' (16-30), 'power' (1=ON, 0=OFF), 'mode' (0-4), 'wind' (0-3).\n"
		prompt += "   Headers:  { \"Content-Type\": \"application/json\" }\n\n"
		prompt += "2) Send Command to Device (e.g., Switch/Light/etc)\n"
		prompt += "   URL:      /api/tuya/devices/{id}/commands/switch\n"
		prompt += "   Method:   POST\n"
		prompt += "   Body:     { \"code\": \"string\", \"value\": boolean|integer|string }\n"
		prompt += "   Headers:  { \"Content-Type\": \"application/json\" }\n\n"
		prompt += "3) Get Sensor Data\n"
		prompt += "   URL:      /api/tuya/devices/{id}/sensor\n"
		prompt += "   Method:   GET\n"
		prompt += "   Body:     null\n"
		prompt += "   Headers:  {}\n\n"

		prompt += fmt.Sprintf("User request: %q\n\n", correctedText)

		if len(deviceContexts) > 0 {
			prompt += "--- Matching Devices from Vector DB ---\n"
			for i, doc := range deviceContexts {
				prompt += fmt.Sprintf("Candidate %d:\n%s\n\n", i+1, doc)
			}
			prompt += "Instructions:\n"
			prompt += "- MATCH the user's requested device (e.g., brand name like 'Daikin' or room name) with the 'name' or 'product_name' fields in the Candidates.\n"
			prompt += "- If the request mentions a specific brand (like 'Daikin'), you MUST choose the Candidate that contains that brand name in its 'name' or 'product_name'.\n"
			prompt += "- If the request is for an AC, pick the Candidate with 'remote_category': 'infrared_ac' OR 'category': 'infrared_ac'.\n"
			prompt += "- For AC: Use its 'id' for the URL's {hub_id} and its 'remote_id' for the body's 'remote_id'. Do NOT swap them.\n"
			prompt += "- If multiple Candidates match, pick the most relevant one based on the brand or name.\n"
			prompt += "CRITICAL: You MUST use ONLY the codes listed in the 'status' list of the selected Candidate. Do NOT invent new codes or values.\n"
			prompt += "- The 'status' list defines the valid commands (e.g., 'switch_1', 'led_switch').\n"
			prompt += "- If the requested function is NOT in the 'status' list, return an error. DO NOT hallunicate a code.\n"
			prompt += "- If the device name matches (e.g. 'Smart Switch') but lacks the capability (e.g. 'alarm'), FAIL instead of picking a different device (e.g. 'Receptionist').\n"
		} else {
			prompt += "IMPORTANT: No matching devices found. Try to infer or use {id} as placeholder.\n"
		}

		prompt += "\nReturn ONLY a JSON object with keys: endpoint_index (1-3), endpoint (string, replace placeholders with IDs), method (POST/GET), device_id (string), body (object|null), headers ({\"Content-Type\": \"application/json\"}).\n"

		// 3) Call the LLM
		model := u.config.LLMModel
		if model == "" {
			model = "default"
		}

		utils.LogInfo("RAG Task %s: Generating decision using LLM (model: %s)...", taskID, model)
		resp, err := u.llm.CallModel(prompt, model)
		if err != nil {
			u.mu.Lock()
			u.taskStatus[taskID] = &ragdtos.RAGStatusDTO{Status: "error", Result: err.Error()}
			u.mu.Unlock()
			return
		}

		// Log raw response for debugging
		utils.LogDebug("RAG Task %s: Raw LLM Response: %s", taskID, resp)

		// 4) Try to parse LLM response as JSON (flattened flow to avoid nested ifs)
		var parsed map[string]interface{}
		var perr error
		parsed, perr = tryParseOnce(resp)
		if perr != nil || parsed == nil {
			// Retry once with a stricter instruction if needed (logging only failure)
			retryPrompt := prompt + "\nIMPORTANT: Your output must be valid JSON ONLY, with keys: endpoint, method, device_id, body, headers. Do not add any extra text."
			resp2, err2 := u.llm.CallModel(retryPrompt, model)
			if err2 == nil {
				parsed, _ = tryParseOnce(resp2)
			}
		}

		if parsed == nil {
			// If still not JSON, log the raw response to see what's wrong
			utils.LogError("RAG Task %s: LLM failed to return valid JSON. Raw Response: %s", taskID, resp)

			u.mu.Lock()
			statusDTO := &ragdtos.RAGStatusDTO{Status: "error", Result: "invalid llm response format"}
			u.taskStatus[taskID] = statusDTO
			u.mu.Unlock()
			return
		}

		// Extract fields
		endpoint, _ := parsed["endpoint"].(string)
		method, _ := parsed["method"].(string)
		deviceID, _ := parsed["device_id"].(string)
		bodyObj := parsed["body"]

		var headers map[string]string
		if hObj, ok := parsed["headers"].(map[string]interface{}); ok {
			headers = make(map[string]string)
			for k, v := range hObj {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}

		// Re-extract deviceID for logging if needed
		logDeviceID := deviceID
		if logDeviceID == "" {
			// Try to extract from URL if possible
			parts := strings.Split(endpoint, "/")
			for i, p := range parts {
				if p == "devices" && i+1 < len(parts) {
					logDeviceID = parts[i+1]
					break
				}
			}
		}

		// Determine Endpoint Description for logging
		endpointIdxRaw := parsed["endpoint_index"]
		var endpointIdx int
		switch v := endpointIdxRaw.(type) {
		case float64:
			endpointIdx = int(v)
		case string:
			endpointIdx, _ = strconv.Atoi(v)
		}

		endpointDesc := "Unknown"
		switch endpointIdx {
		case 1:
			endpointDesc = "1) Send IR AC Command"
		case 2:
			endpointDesc = "2) Send Command to Device (Switch/Light/etc)"
		case 3:
			endpointDesc = "3) Get Sensor Data"
		default:
			// Fallback: Infer from endpoint string if index is missing
			if strings.Contains(endpoint, "/commands/ir") {
				endpointDesc = "1) Send IR AC Command"
			} else if strings.Contains(endpoint, "/commands/switch") {
				endpointDesc = "2) Send Command to Device (Switch/Light/etc)"
			} else if strings.Contains(endpoint, "/sensor") {
				endpointDesc = "3) Get Sensor Data"
			}
		}

		// Format body for logging
		bodyStr := "null"
		if bodyObj != nil {
			if b, err := json.Marshal(bodyObj); err == nil {
				bodyStr = string(b)
			}
		}

		// Use a single LogInfo for the entire decision to prevent splitting and ensure all data is shown
		utils.LogInfo("\n--- RAG Decision ---\nUser Request: %q\naccessed %s\nDetail: [%s] %s\nDeviceID: %s\nBody: %s\nHeaders: {\"Authorization\": %q}\n--------------------",
			text, endpointDesc, method, endpoint, logDeviceID, bodyStr, authToken)

		// Store structured result (do NOT perform external fetch)
		statusDTO := &ragdtos.RAGStatusDTO{Status: "done", Endpoint: endpoint, Method: method, Body: bodyObj, Headers: headers}

		// EXECUTE the decision by fetching the internal endpoint
		if endpoint != "" && authToken != "" {
			utils.LogInfo("RAG Task %s: Executing action via internal fetch...", taskID)
			execRes := u.executeAction(method, endpoint, bodyObj, authToken)
			statusDTO.ExecutionResult = execRes
			utils.LogInfo("RAG Task %s: Action execution completed", taskID)
		}

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
	}(taskID, text, authToken)

	return taskID, nil
}

// executeAction performs the actual HTTP request to the internal API
func (u *RAGUsecase) executeAction(method, path string, body interface{}, token string) interface{} {
	// Base URL for internal calls (use loopback)
	baseURL := "http://localhost:" + u.config.Port
	fullURL := baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	}

	// Clean and format token
	token = strings.TrimSpace(token)
	if token == "" {
		return map[string]string{"error": "empty authorization token"}
	}
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return map[string]string{"error": "failed to create request: " + err.Error()}
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Host = "localhost"

	client := &http.Client{Timeout: 15 * time.Second}
	utils.LogDebug("RAG: Internal fetch to %s (method=%s) with token: %s", fullURL, method, token)
	resp, err := client.Do(req)
	if err != nil {
		return map[string]string{"error": "internal fetch failed: " + err.Error()}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var resData interface{}
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return string(bodyBytes)
	}

	return resData
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
