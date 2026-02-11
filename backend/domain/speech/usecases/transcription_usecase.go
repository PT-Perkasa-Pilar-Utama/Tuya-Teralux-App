package usecases

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	ragdtos "teralux_app/domain/rag/dtos"
	ragUsecases "teralux_app/domain/rag/usecases"
	speechdtos "teralux_app/domain/speech/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
	"time"
)

// WhisperRepositoryInterface defines the methods required from the Whisper repository
type WhisperRepositoryInterface interface {
	Transcribe(wavPath string, modelPath string, lang string) (string, error)
	TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
}

type TranscriptionUsecase struct {
	whisperRepo         WhisperRepositoryInterface
	ragUsecase          *ragUsecases.RAGUsecase
	authUseCase         *tuyaUsecases.TuyaAuthUseCase
	whisperProxyUsecase *WhisperProxyUsecase
	config              *utils.Config
	badger              *infrastructure.BadgerService
	// Async transcription task tracking
	tasksMutex          sync.RWMutex
	transcribeTasks     map[string]*speechdtos.AsyncTranscriptionStatusDTO
	transcribeLongTasks map[string]*speechdtos.AsyncTranscriptionLongStatusDTO
}

func NewTranscriptionUsecase(
	whisperRepo WhisperRepositoryInterface,
	cfg *utils.Config,
	ragUsecase *ragUsecases.RAGUsecase,
	authUseCase *tuyaUsecases.TuyaAuthUseCase,
	whisperProxyUsecase *WhisperProxyUsecase,
	badger *infrastructure.BadgerService,
) *TranscriptionUsecase {
	return &TranscriptionUsecase{
		whisperRepo:         whisperRepo,
		ragUsecase:          ragUsecase,
		authUseCase:         authUseCase,
		whisperProxyUsecase: whisperProxyUsecase,
		config:              cfg,
		badger:              badger,
		transcribeTasks:     make(map[string]*speechdtos.AsyncTranscriptionStatusDTO),
		transcribeLongTasks: make(map[string]*speechdtos.AsyncTranscriptionLongStatusDTO),
	}
}



func (u *TranscriptionUsecase) HandleCommand(text string) {
	// Filter out common non-speech results
	cleanText := strings.TrimSpace(text)
	if cleanText == "" || cleanText == "[BLANK_AUDIO]" {
		utils.LogDebug("Speech: Ignoring blank or empty command")
		return
	}

	if u.ragUsecase == nil || u.authUseCase == nil {
		utils.LogWarn("Speech: RAG or Auth Usecase not initialized, skipping processing")
		return
	}

	// 1. Get Auth Token
	auth, err := u.authUseCase.Authenticate()
	if err != nil {
		utils.LogError("Speech: Failed to authenticate for RAG: %v", err)
		return
	}

	// 2. Process via RAG
	utils.LogInfo("Speech: Processing command via RAG: %q", text)
	taskID, err := u.ragUsecase.Control(text, auth.AccessToken, func(taskID string, status *ragdtos.RAGStatusDTO) {
		// Log final result
		utils.LogInfo("Speech: RAG processing completed for task %s with status %s", taskID, status.Status)

		if status.Status == "done" {
			utils.LogInfo("Speech: Result: %s", status.Result)
		} else if status.Status == "error" {
			utils.LogError("Speech: RAG Error: %s", status.Result)
		}
	})
	if err != nil {
		utils.LogError("Speech: Failed to trigger RAG processing: %v", err)
		return
	}

	utils.LogInfo("Speech: RAG processing triggered (TaskID: %s)", taskID)
}

func (u *TranscriptionUsecase) TranscribeAudio(inputPath string) (string, string, error) {
	utils.LogDebug("Speech: Starting local transcription via whisper.cpp...")
	// Create temp directory for conversion if not exists
	tempDir := filepath.Dir(inputPath)

	// Convert to WAV if needed (Whisper needs 16kHz mono WAV)
	wavPath := filepath.Join(tempDir, "processed.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	// Use model path from config
	modelPath := u.config.WhisperModelPath

	// Use TranscribeFull to get all text
	text, err := u.whisperRepo.TranscribeFull(wavPath, modelPath, "id")
	if err != nil {
		return "", "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, "id", nil
}

func (u *TranscriptionUsecase) TranscribeLongAudio(inputPath string, lang string) (string, error) {
	utils.LogDebug("Speech: Starting local LONG transcription via whisper.cpp...")
	// Create temp directory for conversion if not exists
	tempDir := filepath.Dir(inputPath)

	// Convert to WAV if needed (Whisper needs 16kHz mono WAV)
	wavPath := filepath.Join(tempDir, "processed_long.wav")
	if err := utils.ConvertToWav(inputPath, wavPath); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	// Use model path from config
	modelPath := u.config.WhisperModelPath

	// Use TranscribeFull to get all text
	text, err := u.whisperRepo.TranscribeFull(wavPath, modelPath, lang)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return text, nil
}

// ProxyTranscribeAudio starts async transcription task for short audio
func (u *TranscriptionUsecase) ProxyTranscribeAudio(inputPath string, fileName string) (string, error) {
	// Validate file
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcription: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	// Generate task ID
	taskID := utils.GenerateUUID()

	// Store initial status in memory
	status := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	u.tasksMutex.Lock()
	if u.transcribeTasks == nil {
		u.transcribeTasks = make(map[string]*speechdtos.AsyncTranscriptionStatusDTO)
	}
	u.transcribeTasks[taskID] = status
	u.tasksMutex.Unlock()

	// Store in persistent cache if available
	if u.badger != nil {
		taskData, _ := json.Marshal(status)
		key := "transcribe:task:" + taskID
		if err := u.badger.Set(key, taskData); err != nil {
			utils.LogWarn("Transcription Task %s: failed to save to persistent cache: %v", taskID, err)
		}
	}

	utils.LogInfo("Transcription Task %s: Started processing audio file: %s", taskID, fileName)

	// Start async processing
	go u.processTranscribeAudioAsync(taskID, inputPath)

	return taskID, nil
}

// ProxyTranscribeLongAudio starts async transcription task for long audio
func (u *TranscriptionUsecase) ProxyTranscribeLongAudio(inputPath string, fileName string, lang string) (string, error) {
	// Validate file
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcription Long: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	// Generate task ID
	taskID := utils.GenerateUUID()

	// Store initial status in memory
	status := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    "pending",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	u.tasksMutex.Lock()
	if u.transcribeLongTasks == nil {
		u.transcribeLongTasks = make(map[string]*speechdtos.AsyncTranscriptionLongStatusDTO)
	}
	u.transcribeLongTasks[taskID] = status
	u.tasksMutex.Unlock()

	// Store in persistent cache if available
	if u.badger != nil {
		taskData, _ := json.Marshal(status)
		key := "transcribe_long:task:" + taskID
		if err := u.badger.Set(key, taskData); err != nil {
			utils.LogWarn("Transcription Long Task %s: failed to save to persistent cache: %v", taskID, err)
		}
	}

	utils.LogInfo("Transcription Long Task %s: Started processing audio file: %s", taskID, fileName)

	// Start async processing
	go u.processTranscribeLongAudioAsync(taskID, inputPath, lang)

	return taskID, nil
}

// processTranscribeAudioAsync processes audio transcription in background
func (u *TranscriptionUsecase) processTranscribeAudioAsync(taskID string, inputPath string) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Transcription Task %s: Panic recovered: %v", taskID, r)
			u.updateTranscribeTaskStatus(taskID, "failed", nil)
		}
	}()

	// Try PPU first if available
	var text string
	var lang string
	var err error
	var usedPath string

	if u.whisperProxyUsecase != nil {
		utils.LogDebug("Transcription Task %s: Checking PPU (Outsystems) availability...", taskID)
		if proxyErr := u.whisperProxyUsecase.HealthCheck(); proxyErr == nil {
			utils.LogInfo("Transcription Task %s: PPU is active, trying external transcription...", taskID)
			result, fetchErr := u.whisperProxyUsecase.FetchToOutsystems(inputPath, filepath.Base(inputPath))
			if fetchErr == nil && result != nil {
				text = result.Transcription
				lang = "id" // PPU currently returns Indonesian
				usedPath = "PPU (Outsystems)"
				utils.LogInfo("Transcription Task %s: Successfully transcribed via PPU", taskID)
			} else {
				utils.LogWarn("Transcription Task %s: PPU transcription failed, falling back to local: %v", taskID, fetchErr)
			}
		} else {
			utils.LogDebug("Transcription Task %s: PPU is inactive/unreachable, falling back to local", taskID)
		}
	}

	// Fallback to local if PPU was not used or failed
	if text == "" {
		utils.LogInfo("Transcription Task %s: Using local Whisper (whisper.cpp) path", taskID)
		text, lang, err = u.TranscribeAudio(inputPath)
		usedPath = "Local Whisper (whisper.cpp)"
		if err != nil {
			utils.LogError("Transcription Task %s: Local transcription failed: %v", taskID, err)
			u.updateTranscribeTaskStatus(taskID, "failed", nil)
			return
		}
	}

	utils.LogInfo("Transcription Task %s: Transcription finished using %s", taskID, usedPath)

	// Try to translate
	translated, _ := u.ragUsecase.Translate(text)

	// Build result
	result := &speechdtos.AsyncTranscriptionResultDTO{
		Transcription:    text,
		TranslatedText:   translated,
		DetectedLanguage: lang,
	}

	utils.LogInfo("Transcription Task %s: Successfully transcribed, triggering RAG processing", taskID)

	// If translated, trigger RAG processing (DISABLED as per user request)
	// if translated != "" {
	// 	go u.HandleCommand(translated)
	// }

	u.updateTranscribeTaskStatus(taskID, "completed", result)
}

// processTranscribeLongAudioAsync processes long audio transcription in background
func (u *TranscriptionUsecase) processTranscribeLongAudioAsync(taskID string, inputPath string, lang string) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Transcription Long Task %s: Panic recovered: %v", taskID, r)
			u.updateTranscribeLongTaskStatus(taskID, "failed", nil)
		}
	}()

	// Transcribe long audio
	text, err := u.TranscribeLongAudio(inputPath, lang)
	if err != nil {
		utils.LogError("Transcription Long Task %s: Transcription failed: %v", taskID, err)
		u.updateTranscribeLongTaskStatus(taskID, "failed", nil)
		return
	}

	// Build result
	result := &speechdtos.AsyncTranscriptionLongResultDTO{
		Transcription:    text,
		DetectedLanguage: lang,
	}

	utils.LogInfo("Transcription Long Task %s: Successfully transcribed", taskID)
	u.updateTranscribeLongTaskStatus(taskID, "completed", result)
}

// updateTranscribeTaskStatus updates the status of a short transcription task
func (u *TranscriptionUsecase) updateTranscribeTaskStatus(taskID string, status string, result *speechdtos.AsyncTranscriptionResultDTO) {
	newStatus := &speechdtos.AsyncTranscriptionStatusDTO{
		Status:    status,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	// Update in memory
	u.tasksMutex.Lock()
	if u.transcribeTasks == nil {
		u.transcribeTasks = make(map[string]*speechdtos.AsyncTranscriptionStatusDTO)
	}
	u.transcribeTasks[taskID] = newStatus
	u.tasksMutex.Unlock()

	// Update in persistent cache
	if u.badger != nil {
		taskData, _ := json.Marshal(newStatus)
		key := "transcribe:task:" + taskID
		if err := u.badger.SetPreserveTTL(key, taskData); err != nil {
			utils.LogWarn("Transcription Task %s: failed to update persistent cache: %v", taskID, err)
		}
	}

	utils.LogDebug("Transcription Task %s: Status updated to %s", taskID, status)
}

// updateTranscribeLongTaskStatus updates the status of a long transcription task
func (u *TranscriptionUsecase) updateTranscribeLongTaskStatus(taskID string, status string, result *speechdtos.AsyncTranscriptionLongResultDTO) {
	newStatus := &speechdtos.AsyncTranscriptionLongStatusDTO{
		Status:    status,
		Result:    result,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	// Update in memory
	u.tasksMutex.Lock()
	if u.transcribeLongTasks == nil {
		u.transcribeLongTasks = make(map[string]*speechdtos.AsyncTranscriptionLongStatusDTO)
	}
	u.transcribeLongTasks[taskID] = newStatus
	u.tasksMutex.Unlock()

	// Update in persistent cache
	if u.badger != nil {
		taskData, _ := json.Marshal(newStatus)
		key := "transcribe_long:task:" + taskID
		if err := u.badger.SetPreserveTTL(key, taskData); err != nil {
			utils.LogWarn("Transcription Long Task %s: failed to update persistent cache: %v", taskID, err)
		}
	}

	utils.LogDebug("Transcription Long Task %s: Status updated to %s", taskID, status)
}

// GetTranscriptionStatus retrieves the status of a short transcription task
func (u *TranscriptionUsecase) GetTranscriptionStatus(taskID string) (*speechdtos.AsyncTranscriptionStatusDTO, error) {
	u.tasksMutex.RLock()
	status, exists := u.transcribeTasks[taskID]
	u.tasksMutex.RUnlock()

	if exists && status != nil {
		calculateTranscriptionTTL(status)
		return status, nil
	}

	// If not found in-memory, try persistent store (Badger) if configured
	if u.badger != nil {
		key := "transcribe:task:" + taskID
		b, ttl, err := u.badger.GetWithTTL(key)
		if err != nil {
			utils.LogDebug("Transcription Task %s: not found in cache", taskID)
		}

		if len(b) > 0 {
			var status speechdtos.AsyncTranscriptionStatusDTO
			if err := json.Unmarshal(b, &status); err == nil {
				status.ExpiresInSecond = int64(ttl.Seconds())
				return &status, nil
			}
		}
	}

	utils.LogWarn("Transcription: Task %s not found in any cache", taskID)
	return nil, fmt.Errorf("task not found")
}

// GetTranscriptionLongStatus retrieves the status of a long transcription task
func (u *TranscriptionUsecase) GetTranscriptionLongStatus(taskID string) (*speechdtos.AsyncTranscriptionLongStatusDTO, error) {
	u.tasksMutex.RLock()
	status, exists := u.transcribeLongTasks[taskID]
	u.tasksMutex.RUnlock()

	if exists && status != nil {
		calculateTranscriptionLongTTL(status)
		return status, nil
	}

	// If not found in-memory, try persistent store (Badger) if configured
	if u.badger != nil {
		key := "transcribe_long:task:" + taskID
		b, ttl, err := u.badger.GetWithTTL(key)
		if err != nil {
			utils.LogDebug("Transcription Long Task %s: not found in cache", taskID)
		}

		if len(b) > 0 {
			var status speechdtos.AsyncTranscriptionLongStatusDTO
			if err := json.Unmarshal(b, &status); err == nil {
				status.ExpiresInSecond = int64(ttl.Seconds())
				return &status, nil
			}
		}
	}

	utils.LogWarn("Transcription Long: Task %s not found in any cache", taskID)
	return nil, fmt.Errorf("task not found")
}

// calculateTranscriptionTTL calculates TTL for short transcription status
func calculateTranscriptionTTL(status *speechdtos.AsyncTranscriptionStatusDTO) {
	if status.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, status.ExpiresAt); err == nil {
			status.ExpiresInSecond = int64(time.Until(t).Seconds())
		}
	}
}

// calculateTranscriptionLongTTL calculates TTL for long transcription status
func calculateTranscriptionLongTTL(status *speechdtos.AsyncTranscriptionLongStatusDTO) {
	if status.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, status.ExpiresAt); err == nil {
			status.ExpiresInSecond = int64(time.Until(t).Seconds())
		}
	}
}
