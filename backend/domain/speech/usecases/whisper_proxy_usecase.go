package usecases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	speechdtos "teralux_app/domain/speech/dtos"

	"github.com/google/uuid"
)

type WhisperProxyUsecase interface {
	ProxyTranscribe(filePath string, fileName string, language string) (string, error)
	GetStatus(taskID string) (*speechdtos.WhisperProxyStatusDTO, error)
	HealthCheck() error
	FetchToOutsystems(filePath string, fileName string, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error)
}

type whisperProxyUsecase struct {
	cache      *tasks.BadgerTaskCache
	config     *utils.Config
	mu         sync.RWMutex
	taskStatus map[string]*speechdtos.WhisperProxyStatusDTO
}

// NewWhisperProxyUsecase creates a new whisper proxy usecase instance
func NewWhisperProxyUsecase(cache *tasks.BadgerTaskCache, cfg *utils.Config) WhisperProxyUsecase {
	return &whisperProxyUsecase{
		cache:      cache,
		config:     cfg,
		taskStatus: make(map[string]*speechdtos.WhisperProxyStatusDTO),
	}
}

// ProxyTranscribe accepts audio file and queues async transcription to external Outsystems server
func (u *whisperProxyUsecase) ProxyTranscribe(filePath string, fileName string, language string) (string, error) {
	utils.LogDebug("Whisper Proxy: Starting external transcription via Outsystems (PPU)...")
	// Generate UUID task id
	taskID := uuid.New().String()

	// Initially mark pending
	u.mu.Lock()
	u.taskStatus[taskID] = &speechdtos.WhisperProxyStatusDTO{
		Status: "pending",
	}
	pending := u.taskStatus[taskID]
	u.mu.Unlock()

	// Persist pending to cache (with TTL) if available
	if err := u.cache.Set(taskID, pending); err != nil {
		utils.LogError("Whisper Task %s: failed to cache pending task: %v", taskID, err)
	} else {
		utils.LogDebug("Whisper Task %s: pending cached with TTL", taskID)
	}

	// Run processing asynchronously
	go func(taskID, filePath, fileName, language string) {
		utils.LogInfo("Whisper Task %s: Started processing audio file: %s", taskID, fileName)

		defer func() {
			if r := recover(); r != nil {
				utils.LogError("Whisper Task %s: Panic recovered: %v", taskID, r)
				u.mu.Lock()
				u.taskStatus[taskID] = &speechdtos.WhisperProxyStatusDTO{
					Status: "error",
				}
				u.mu.Unlock()
			}
		}()

		// Step 1: Health check to verify server is online
		utils.LogDebug("Whisper Task %s: Checking server health...", taskID)
		if err := u.HealthCheck(); err != nil {
			utils.LogError("Whisper Task %s: Server health check failed: %v", taskID, err)
			statusDTO := &speechdtos.WhisperProxyStatusDTO{
				Status: "error",
			}
			u.mu.Lock()
			u.taskStatus[taskID] = statusDTO
			u.mu.Unlock()
			if err := u.cache.SetPreserveTTL(taskID, statusDTO); err != nil {
				utils.LogWarn("Whisper Task %s: failed to update persistent cache: %v", taskID, err)
			}
			return
		}

		// Step 2: Fetch to external Outsystems server
		result, err := u.FetchToOutsystems(filePath, fileName, language)

		statusDTO := &speechdtos.WhisperProxyStatusDTO{
			Status: "completed",
			Result: result,
		}

		if err != nil {
			utils.LogError("Whisper Task %s: Transcription failed: %v", taskID, err)
			statusDTO.Status = "failed"
		} else {
			utils.LogInfo("Whisper Task %s: Transcription completed successfully", taskID)
		}

		u.mu.Lock()
		u.taskStatus[taskID] = statusDTO
		u.mu.Unlock()

		// Persist final result by updating existing cache entry while preserving TTL
		if err := u.cache.SetPreserveTTL(taskID, statusDTO); err != nil {
			utils.LogError("Whisper Task %s: failed to update cached final result: %v", taskID, err)
		} else {
			utils.LogDebug("Whisper Task %s: final result cached (TTL preserved)", taskID)
		}

		// Broadcast removed
	}(taskID, filePath, fileName, language)

	return taskID, nil
}

// HealthCheck performs a health check to verify the Outsystems server is online
func (u *whisperProxyUsecase) HealthCheck() error {
	outsystemsURL := u.config.OutsystemsTranscribeURL
	if outsystemsURL == "" {
		utils.LogError("Whisper: OUTSYSTEMS_TRANSCRIBE_URL not configured")
		return fmt.Errorf("outsystems URL not configured")
	}

	// Extract base URL (remove /whisper/transcribe)
	baseURL := outsystemsURL[:len(outsystemsURL)-len("/whisper/transcribe")]

	healthCheckURL := baseURL + "/"

	// Create HTTP request
	req, err := http.NewRequest("GET", healthCheckURL, nil)
	if err != nil {
		utils.LogError("Whisper: Failed to create health check request: %v", err)
		return fmt.Errorf("health check request creation failed")
	}

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	utils.LogDebug("Whisper: Health check to %s", healthCheckURL)

	resp, err := client.Do(req)
	if err != nil {
		utils.LogError("Whisper: Health check request failed - Outsystems server is unreachable: %v", err)
		return fmt.Errorf("outsystems server unreachable")
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		utils.LogError("Whisper: Health check failed - server returned status %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("outsystems server returned error status %d", resp.StatusCode)
	}

	utils.LogDebug("Whisper: Health check passed, server is online")
	return nil
}

// FetchToOutsystems sends the audio file to the external Outsystems server and returns parsed result
func (u *whisperProxyUsecase) FetchToOutsystems(filePath string, fileName string, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error) {
	outsystemsURL := u.config.OutsystemsTranscribeURL
	if outsystemsURL == "" {
		utils.LogError("Whisper: OUTSYSTEMS_TRANSCRIBE_URL not configured")
		return nil, fmt.Errorf("outsystems URL not configured")
	}

	// Create multipart form data
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	// Read file first
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		utils.LogError("Whisper: Failed to read audio file from %s: %v", filePath, err)
		return nil, fmt.Errorf("audio file read failed")
	}

	// Add file to multipart form
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		utils.LogError("Whisper: Failed to create form file: %v", err)
		return nil, fmt.Errorf("form file creation failed")
	}

	if _, err := fileWriter.Write(fileData); err != nil {
		utils.LogError("Whisper: Failed to write file to form: %v", err)
		return nil, fmt.Errorf("file write to form failed")
	}

	// Add language field AFTER file is written
	if err := writer.WriteField("language", language); err != nil {
		utils.LogError("Whisper: Failed to write language field: %v", err)
		return nil, fmt.Errorf("language field write failed")
	}

	if err := writer.Close(); err != nil {
		utils.LogError("Whisper: Failed to close multipart writer: %v", err)
		return nil, fmt.Errorf("multipart writer close failed")
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", outsystemsURL, bodyBuf)
	if err != nil {
		utils.LogError("Whisper: Failed to create transcribe request: %v", err)
		return nil, fmt.Errorf("request creation failed")
	}

	// Set multipart content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	client := &http.Client{Timeout: 60 * time.Second}
	utils.LogDebug("Whisper: Fetching to %s", outsystemsURL)

	resp, err := client.Do(req)
	if err != nil {
		utils.LogError("Whisper: Transcribe request to Outsystems failed: %v", err)
		return nil, fmt.Errorf("transcribe request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("Whisper: Failed to read transcribe response: %v", err)
		return nil, fmt.Errorf("response read failed")
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		utils.LogError("Whisper: Outsystems server returned error status %d: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("outsystems error: status %d", resp.StatusCode)
	}

	// Parse response as JSON
	var result speechdtos.OutsystemsTranscriptionResultDTO
	if err := json.Unmarshal(respBody, &result); err != nil {
		utils.LogError("Whisper: Failed to parse outsystems response: %v", err)
		return nil, fmt.Errorf("response parse failed")
	}

	utils.LogInfo("Whisper: Successfully transcribed audio, filename: %s", result.Filename)
	return &result, nil
}

// GetStatus retrieves the status of a whisper transcription task
func (u *whisperProxyUsecase) GetStatus(taskID string) (*speechdtos.WhisperProxyStatusDTO, error) {
	// First try in-memory map with read lock
	u.mu.RLock()
	if s, ok := u.taskStatus[taskID]; ok {
		u.mu.RUnlock()
		// Augment with TTL info if available
		if u.cache != nil {
			var cached speechdtos.WhisperProxyStatusDTO
			if ttl, found, err := u.cache.GetWithTTL(taskID, &cached); err == nil && found && ttl > 0 {
				s.ExpiresInSecond = int64(ttl.Seconds())
				s.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
		}
		return s, nil
	}
	u.mu.RUnlock()

	// If not found in-memory, try persistent store (Badger) if configured
	if u.cache != nil {
		var status speechdtos.WhisperProxyStatusDTO
		ttl, found, err := u.cache.GetWithTTL(taskID, &status)
		if err != nil {
			utils.LogError("Whisper: Failed to read task %s from persistent cache: %v", taskID, err)
			return nil, fmt.Errorf("persistent cache read failed")
		}
		if found {
			// Cache into memory for faster subsequent reads
			u.mu.Lock()
			u.taskStatus[taskID] = &status
			u.mu.Unlock()
			if ttl > 0 {
				status.ExpiresInSecond = int64(ttl.Seconds())
				status.ExpiresAt = time.Now().Add(ttl).UTC().Format(time.RFC3339)
			}
			utils.LogDebug("Whisper Task %s: retrieved from badger, ttl=%v", taskID, ttl)
			return &status, nil
		}
		// Not found in badger either
		utils.LogDebug("Whisper Task %s: not found in cache", taskID)
	}

	utils.LogWarn("Whisper: Task %s not found in any cache", taskID)
	return nil, fmt.Errorf("task not found")
}
