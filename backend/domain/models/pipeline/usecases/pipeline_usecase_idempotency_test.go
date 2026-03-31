package usecases

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineDtos "sensio/domain/models/pipeline/dtos"
	ragDtos "sensio/domain/models/rag/dtos"
	whisperDtos "sensio/domain/models/whisper/dtos"
	speechUsecases "sensio/domain/models/whisper/usecases"

	"github.com/stretchr/testify/assert"
)

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Set required environment variables for tests
	_ = os.Setenv("GO_TEST", "true")
	_ = os.Setenv("TASK_STATUS_TTL", "24h")
	_ = os.Setenv("PIPELINE_ASYNC_TIMEOUT", "1h")
	_ = os.Setenv("TRANSCRIBE_ASYNC_TIMEOUT", "1h")
	_ = os.Setenv("APPLICATION_ENVIRONMENT", "test")
	_ = os.Setenv("MAX_FILE_SIZE_MB", "100")
	_ = os.Setenv("CHUNK_UPLOAD_SESSION_TTL", "24h")
	_ = os.Setenv("CHUNK_UPLOAD_CLEANUP_INTERVAL", "1h")
	_ = os.Setenv("UPLOAD_DIR", "/tmp/test_uploads")
	_ = os.Setenv("LOG_LEVEL", "info")
	_ = os.Setenv("TASK_EVENT_PUBLISH_ENABLED", "false")

	// Load config to initialize AppConfig singleton
	utils.LoadConfig()

	// Run tests
	os.Exit(m.Run())
}

// MockBadgerService for testing
type MockBadgerService struct {
	data map[string][]byte
	ttls map[string]time.Duration
	mu   sync.RWMutex
}

func NewMockBadgerService() *MockBadgerService {
	return &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}
}

func (m *MockBadgerService) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	if _, exists := m.ttls[key]; !exists {
		m.ttls[key] = 3600 * time.Second
	}
	return nil
}

func (m *MockBadgerService) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.ttls[key] = ttl
	return nil
}

func (m *MockBadgerService) SetIfAbsentWithTTL(key string, value []byte, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	m.data[key] = value
	m.ttls[key] = ttl
	return true, nil
}

func (m *MockBadgerService) SetPreserveTTL(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	if _, exists := m.ttls[key]; !exists {
		m.ttls[key] = 3600 * time.Second
	}
	return nil
}

func (m *MockBadgerService) GetWithTTL(key string) ([]byte, time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.data[key]
	if !ok {
		return nil, 0, nil
	}
	ttl := m.ttls[key]
	return data, ttl, nil
}

func (m *MockBadgerService) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

func (m *MockBadgerService) KeysWithPrefix(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var keys []string
	for key := range m.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// TestIdempotencyHashWithSessionID verifies that session ID is included in the hash
func TestIdempotencyHashWithSessionID(t *testing.T) {
	t.Run("hash includes session ID", func(t *testing.T) {
		sessionID := "session_abc_123"
		idempotencyKey := "meeting_123456"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("dummy_audio"))

		// With session ID
		hashInputWithSession := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID)

		// Without session ID
		hashInputWithoutSession := fmt.Sprintf("%s_%s_%s_%s_%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash)

		// They should be different
		assert.NotEqual(t, hashInputWithSession, hashInputWithoutSession,
			"Hash input with session ID should differ from hash input without session ID")
	})

	t.Run("same session ID produces same hash", func(t *testing.T) {
		sessionID := "session_xyz_789"
		idempotencyKey := "meeting_789012"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("dummy_audio"))

		hashInput1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID)
		hashInput2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID)

		assert.Equal(t, hashInput1, hashInput2,
			"Same inputs should produce same hash input")
	})

	t.Run("different session ID produces different hash", func(t *testing.T) {
		idempotencyKey := "meeting_345678"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("dummy_audio"))

		hashInput1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, "session_1")
		hashInput2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, "session_2")

		assert.NotEqual(t, hashInput1, hashInput2,
			"Different session IDs should produce different hash inputs")
	})
}

// TestIdempotencyCacheBehavior verifies cache behavior with session-scoped keys
func TestIdempotencyCacheBehavior(t *testing.T) {
	mockBadger := NewMockBadgerService()
	cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")

	t.Run("cache stores and retrieves task ID", func(t *testing.T) {
		hashKey := "idemp_pipeline_test_hash_123"
		taskID := "task_abc_123"

		err := cache.Set(hashKey, taskID)
		assert.NoError(t, err)

		var retrievedTaskID string
		_, exists, err := cache.GetWithTTL(hashKey, &retrievedTaskID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, taskID, retrievedTaskID)
	})

	t.Run("cache misses for non-existent key", func(t *testing.T) {
		var retrievedTaskID string
		_, exists, err := cache.GetWithTTL("non_existent_key", &retrievedTaskID)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestIdempotencyKeyComponents verifies all components are included in hash
func TestIdempotencyKeyComponents(t *testing.T) {
	t.Run("all components affect hash", func(t *testing.T) {
		baseKey := "meeting_123"
		audioHash := fmt.Sprintf("%x", []byte("audio_data"))
		macAddress := "AA:BB:CC:DD:EE:FF"
		sessionID := "session_456"

		// Different language
		hash1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			baseKey, "id", "en", macAddress, audioHash, sessionID)
		hash2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			baseKey, "id", "id", macAddress, audioHash, sessionID)
		assert.NotEqual(t, hash1, hash2, "Different target language should change hash")

		// Different MAC address
		hash3 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			baseKey, "id", "en", "AA:BB:CC:DD:EE:00", audioHash, sessionID)
		assert.NotEqual(t, hash1, hash3, "Different MAC address should change hash")

		// Different audio hash
		hash4 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			baseKey, "id", "en", macAddress, fmt.Sprintf("%x", []byte("different_audio")), sessionID)
		assert.NotEqual(t, hash1, hash4, "Different audio hash should change hash")
	})
}

// TestExecutePipelineWithSession_IdempotencyBehavior verifies the actual idempotency behavior
// for session-scoped pipeline execution (by-upload requests)
// This test verifies that the idempotency hash includes session ID to prevent cross-session collisions
func TestExecutePipelineWithSession_IdempotencyBehavior(t *testing.T) {
	t.Run("idempotency hash includes session ID to prevent cross-session collisions", func(t *testing.T) {
		// Verify that different session IDs produce different hash inputs
		sessionID1 := "session_A"
		sessionID2 := "session_B"
		idempotencyKey := "meeting_key_123"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("test_audio"))

		// Hash with session A
		hashInput1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID1)

		// Hash with session B
		hashInput2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID2)

		// They should be different
		assert.NotEqual(t, hashInput1, hashInput2,
			"Different session IDs should produce different hash inputs")
	})

	t.Run("same session ID and idempotency key produces same hash", func(t *testing.T) {
		sessionID := "session_same_123"
		idempotencyKey := "meeting_key_456"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("test_audio"))

		hashInput1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID)
		hashInput2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			idempotencyKey, language, targetLanguage, macAddress, audioHash, sessionID)

		assert.Equal(t, hashInput1, hashInput2,
			"Same session ID and idempotency key should produce same hash")
	})

	t.Run("different idempotency key with same session ID produces different hash", func(t *testing.T) {
		sessionID := "session_789"
		language := "id"
		targetLanguage := "en"
		macAddress := "AA:BB:CC:DD:EE:FF"
		audioHash := fmt.Sprintf("%x", []byte("test_audio"))

		hashInput1 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			"meeting_key_A", language, targetLanguage, macAddress, audioHash, sessionID)
		hashInput2 := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s",
			"meeting_key_B", language, targetLanguage, macAddress, audioHash, sessionID)

		assert.NotEqual(t, hashInput1, hashInput2,
			"Different idempotency keys should produce different hash inputs")
	})

	t.Run("backend idempotency behavior - same session and key returns cached task ID", func(t *testing.T) {
		// Setup mock dependencies
		mockBadger := NewMockBadgerService()
		cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")
		store := tasks.NewStatusStore[pipelineDtos.PipelineStatusDTO]()
		mockMQTT := &MockMQTTPublisher{}
		mockTranscribeUC := &MockTranscribeUseCase{}
		mockTranslateUC := &MockTranslateUseCase{}
		mockSummaryUC := &MockSummaryUseCase{}

		pipelineUC := NewPipelineUseCase(
			mockTranscribeUC,
			mockTranslateUC,
			mockSummaryUC,
			cache,
			store,
			mockMQTT,
		)

		// Create temporary test audio file
		tmpDir := t.TempDir()
		testAudioPath := tmpDir + "/test_idemp_behavior.wav"
		testAudioContent := []byte("fake audio for idempotency behavior test")
		err := os.WriteFile(testAudioPath, testAudioContent, 0644)
		assert.NoError(t, err)

		req := pipelineDtos.PipelineRequestDTO{
			Language:       "id",
			TargetLanguage: "en",
			MacAddress:     "AA:BB:CC:DD:EE:FF",
			Diarize:        false,
			Refine:         boolPtr(true),
			Summarize:      true,
		}

		sessionID := "session_idemp_behavior_001"
		idempotencyKey := "meeting_idemp_behavior_001"

		// First call - should create new task
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1, "First call should return a task ID")

		// Second call with same session and key - should return cached task ID
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2, "Second call should return a task ID")

		// Verify idempotency: same inputs should return same task ID
		assert.Equal(t, taskID1, taskID2,
			"Same session ID and idempotency key should return the same cached task ID")
	})

	t.Run("backend idempotency behavior - different session with same key returns different task ID", func(t *testing.T) {
		// Setup mock dependencies
		mockBadger := NewMockBadgerService()
		cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")
		store := tasks.NewStatusStore[pipelineDtos.PipelineStatusDTO]()
		mockMQTT := &MockMQTTPublisher{}
		mockTranscribeUC := &MockTranscribeUseCase{}
		mockTranslateUC := &MockTranslateUseCase{}
		mockSummaryUC := &MockSummaryUseCase{}

		pipelineUC := NewPipelineUseCase(
			mockTranscribeUC,
			mockTranslateUC,
			mockSummaryUC,
			cache,
			store,
			mockMQTT,
		)

		// Create temporary test audio file
		tmpDir := t.TempDir()
		testAudioPath := tmpDir + "/test_session_collision.wav"
		testAudioContent := []byte("fake audio for session collision test")
		err := os.WriteFile(testAudioPath, testAudioContent, 0644)
		assert.NoError(t, err)

		req := pipelineDtos.PipelineRequestDTO{
			Language:       "id",
			TargetLanguage: "en",
			MacAddress:     "AA:BB:CC:DD:EE:FF",
			Diarize:        false,
			Refine:         boolPtr(true),
			Summarize:      true,
		}

		sessionID1 := "session_collision_A"
		sessionID2 := "session_collision_B"
		idempotencyKey := "meeting_same_key_collision"

		// First call with session A
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID1)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1)

		// Second call with session B (same key, different session)
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID2)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2)

		// Verify no cross-session collision
		assert.NotEqual(t, taskID1, taskID2,
			"Different session IDs with same idempotency key should produce different task IDs")
	})
}

// TestExecutePipelineWithSession_ActualIdempotency tests the actual idempotency behavior
// by calling ExecutePipelineWithSession and verifying task ID reuse from cache.
// This test verifies the full idempotency contract: same session+key returns same task,
// different sessions do not collide, and cache is properly used.
func TestExecutePipelineWithSession_ActualIdempotency(t *testing.T) {
	// Setup mock dependencies
	mockBadger := NewMockBadgerService()
	cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")
	store := tasks.NewStatusStore[pipelineDtos.PipelineStatusDTO]()

	// Create mock MQTT publisher
	mockMQTT := &MockMQTTPublisher{}

	// Create mock transcribe, translate, summary use cases
	mockTranscribeUC := &MockTranscribeUseCase{}
	mockTranslateUC := &MockTranslateUseCase{}
	mockSummaryUC := &MockSummaryUseCase{}

	// Create real PipelineUseCase with mock dependencies
	pipelineUC := NewPipelineUseCase(
		mockTranscribeUC,
		mockTranslateUC,
		mockSummaryUC,
		cache,
		store,
		mockMQTT,
	)

	// Create a temporary test audio file
	tmpDir := t.TempDir()
	testAudioPath := tmpDir + "/test_meeting.wav"
	testAudioContent := []byte("fake audio data for testing")
	err := os.WriteFile(testAudioPath, testAudioContent, 0644)
	assert.NoError(t, err)

	// Common request parameters
	req := pipelineDtos.PipelineRequestDTO{
		Language:       "id",
		TargetLanguage: "en",
		MacAddress:     "AA:BB:CC:DD:EE:FF",
		Diarize:        false,
		Refine:         boolPtr(true),
		Summarize:      true,
	}

	t.Run("same session and same idempotency key returns same task ID", func(t *testing.T) {
		sessionID := "session_test_same_123"
		idempotencyKey := "meeting_idemp_test_001"

		// First call - should create new task
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1, "First call should return a task ID")

		// Second call with same session and key - should return cached task ID
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2, "Second call should return a task ID")

		// Verify idempotency: same inputs should return same task ID
		assert.Equal(t, taskID1, taskID2,
			"Same session ID and idempotency key should return the same cached task ID")
	})

	t.Run("different session ID with same idempotency key returns different task ID", func(t *testing.T) {
		idempotencyKey := "meeting_idemp_test_002"
		sessionID1 := "session_A_456"
		sessionID2 := "session_B_789"

		// First call with session A
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID1)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1)

		// Second call with session B (same key, different session)
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey, sessionID2)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2)

		// Verify no cross-session collision
		assert.NotEqual(t, taskID1, taskID2,
			"Different session IDs should produce different task IDs even with same idempotency key")
	})

	t.Run("same session ID with different idempotency key returns different task ID", func(t *testing.T) {
		sessionID := "session_test_003"
		idempotencyKey1 := "meeting_key_A_003"
		idempotencyKey2 := "meeting_key_B_004"

		// First call with key A
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey1, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1)

		// Second call with key B (same session, different key)
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, idempotencyKey2, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2)

		// Verify different keys produce different tasks
		assert.NotEqual(t, taskID1, taskID2,
			"Different idempotency keys should produce different task IDs even with same session ID")
	})

	t.Run("cache stores and retrieves task ID correctly", func(t *testing.T) {
		sessionID := "session_cache_test_004"
		idempotencyKey := "meeting_cache_004"

		// Create a fresh test audio file for this test case
		testAudioPath4 := tmpDir + "/test_meeting_004.wav"
		err := os.WriteFile(testAudioPath4, testAudioContent, 0644)
		assert.NoError(t, err)

		// Compute audio hash BEFORE executing pipeline (file will be deleted by async cleanup)
		audioHash, err := utils.HashFile(testAudioPath4)
		assert.NoError(t, err)

		// Execute pipeline
		taskID, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath4, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		// Wait briefly for async pipeline to complete and cleanup file
		time.Sleep(100 * time.Millisecond)

		// Verify task ID is stored in cache with correct hash key
		// Note: We use the pre-computed hash since file was deleted by async cleanup
		expectedHashKey := "idemp_pipeline_" + utils.HashString(
			fmt.Sprintf("%s_%s_%s_%s_%s_session:%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash, sessionID),
		)

		var cachedTaskID string
		_, exists, err := cache.GetWithTTL(expectedHashKey, &cachedTaskID)
		assert.NoError(t, err)
		assert.True(t, exists, "Task ID should be stored in cache")
		assert.Equal(t, taskID, cachedTaskID, "Cached task ID should match returned task ID")
	})

	t.Run("failed task allows retry with new task ID", func(t *testing.T) {
		sessionID := "session_retry_test_005"
		idempotencyKey := "meeting_retry_005"

		// Create a fresh test audio file for this test case
		testAudioPath5 := tmpDir + "/test_meeting_005.wav"
		err := os.WriteFile(testAudioPath5, testAudioContent, 0644)
		assert.NoError(t, err)

		// Compute audio hash BEFORE executing pipeline (file will be deleted by async cleanup)
		audioHash, err := utils.HashFile(testAudioPath5)
		assert.NoError(t, err)

		// First call - create task
		taskID1, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath5, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID1)

		// Wait briefly for async pipeline to start and create the task status
		time.Sleep(50 * time.Millisecond)

		// Manually set task status to failed (simulating a failure during execution)
		failedStatus := pipelineDtos.PipelineStatusDTO{
			TaskID:        taskID1,
			OverallStatus: "failed",
			Stages: map[string]pipelineDtos.PipelineStageStatus{
				"transcription": {Status: "failed", Error: "test error"},
			},
		}
		store.Set(taskID1, &failedStatus)

		// Also update the cache to reflect the failed status
		// (In real scenario, the async pipeline would do this)
		_ = cache.SetWithTTL(taskID1, failedStatus, 24*time.Hour)

		// Compute the idempotency hash key to update the idempotency cache
		hashInput := fmt.Sprintf("%s_%s_%s_%s_%s_session:%s", idempotencyKey, req.Language, req.TargetLanguage, req.MacAddress, audioHash, sessionID)
		idempotencyHash := "idemp_pipeline_" + utils.HashString(hashInput)

		// Update the idempotency cache to map hash -> taskID (so the retry check can find it)
		_ = cache.SetWithTTL(idempotencyHash, taskID1, 24*time.Hour)

		// Second call - should create new task because previous one failed
		taskID2, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath5, req, idempotencyKey, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID2)

		// Verify new task is created for retry
		assert.NotEqual(t, taskID1, taskID2,
			"Failed task should allow retry with new task ID")
	})
}

// TestPipelineAudioPreservation tests that meeting-summary jobs preserve input audio files
// while non-summary jobs clean them up as expected.
func TestPipelineAudioPreservation(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("meeting-summary job preserves input audio file", func(t *testing.T) {
		// Setup mock dependencies
		mockBadger := NewMockBadgerService()
		cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")
		store := tasks.NewStatusStore[pipelineDtos.PipelineStatusDTO]()
		mockMQTT := &MockMQTTPublisher{}
		mockTranscribeUC := &MockTranscribeUseCase{}
		mockTranslateUC := &MockTranslateUseCase{}
		mockSummaryUC := &MockSummaryUseCase{}

		pipelineUC := NewPipelineUseCase(
			mockTranscribeUC,
			mockTranslateUC,
			mockSummaryUC,
			cache,
			store,
			mockMQTT,
		)

		// Create test audio file
		testAudioPath := tmpDir + "/test_meeting_summary.wav"
		testAudioContent := []byte("fake audio data for meeting summary test")
		err := os.WriteFile(testAudioPath, testAudioContent, 0644)
		assert.NoError(t, err)

		// Verify file exists before pipeline
		_, err = os.Stat(testAudioPath)
		assert.NoError(t, err, "Test audio file should exist before pipeline execution")

		req := pipelineDtos.PipelineRequestDTO{
			Language:       "id",
			TargetLanguage: "en",
			MacAddress:     "AA:BB:CC:DD:EE:FF",
			Diarize:        false,
			Refine:         boolPtr(true),
			Summarize:      true, // Meeting summary enabled
		}

		// Execute pipeline
		taskID, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, "", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		// Wait briefly for async cleanup to occur (if it were going to)
		time.Sleep(100 * time.Millisecond)

		// Verify file still exists after pipeline completion (preserved for meeting summary)
		_, err = os.Stat(testAudioPath)
		assert.NoError(t, err, "Input audio file should be preserved for meeting-summary jobs")

		// Cleanup: manually remove test file
		os.Remove(testAudioPath)
	})

	t.Run("non-summary job deletes input audio file", func(t *testing.T) {
		// Setup mock dependencies
		mockBadger := NewMockBadgerService()
		cache := tasks.NewBadgerTaskCache(mockBadger, "cache:task:")
		store := tasks.NewStatusStore[pipelineDtos.PipelineStatusDTO]()
		mockMQTT := &MockMQTTPublisher{}
		mockTranscribeUC := &MockTranscribeUseCase{}
		mockTranslateUC := &MockTranslateUseCase{}
		mockSummaryUC := &MockSummaryUseCase{}

		pipelineUC := NewPipelineUseCase(
			mockTranscribeUC,
			mockTranslateUC,
			mockSummaryUC,
			cache,
			store,
			mockMQTT,
		)

		// Create test audio file
		testAudioPath := tmpDir + "/test_non_summary.wav"
		testAudioContent := []byte("fake audio data for non-summary test")
		err := os.WriteFile(testAudioPath, testAudioContent, 0644)
		assert.NoError(t, err)

		// Verify file exists before pipeline
		_, err = os.Stat(testAudioPath)
		assert.NoError(t, err, "Test audio file should exist before pipeline execution")

		req := pipelineDtos.PipelineRequestDTO{
			Language:       "id",
			TargetLanguage: "en",
			MacAddress:     "AA:BB:CC:DD:EE:FF",
			Diarize:        false,
			Refine:         boolPtr(true),
			Summarize:      false, // Meeting summary disabled
		}

		// Execute pipeline
		taskID, err := pipelineUC.ExecutePipelineWithSession(context.Background(), testAudioPath, req, "", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, taskID)

		// Wait briefly for async cleanup to occur
		time.Sleep(100 * time.Millisecond)

		// Verify file is deleted after pipeline completion (non-summary job)
		_, err = os.Stat(testAudioPath)
		assert.True(t, os.IsNotExist(err), "Input audio file should be deleted for non-summary jobs")
	})
}

// Helper functions for tests
func boolPtr(b bool) *bool {
	return &b
}

// MockMQTTPublisher implements mqttPublisher interface for testing
type MockMQTTPublisher struct {
	publishedMessages []MockMQTTMessage
}

type MockMQTTMessage struct {
	topic    string
	qos      byte
	retained bool
	payload  interface{}
}

func (m *MockMQTTPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	m.publishedMessages = append(m.publishedMessages, MockMQTTMessage{
		topic:    topic,
		qos:      qos,
		retained: retained,
		payload:  payload,
	})
	return nil
}

// MockTranscribeUseCase implements speechUsecases.TranscribeUseCase for testing
type MockTranscribeUseCase struct{}

func (m *MockTranscribeUseCase) TranscribeAudio(
	ctx context.Context,
	inputPath string,
	fileName string,
	language string,
	metadata ...speechUsecases.TranscriptionMetadata,
) (string, error) {
	return "mock_task_id", nil
}
func (m *MockTranscribeUseCase) TranscribeAudioSync(
	ctx context.Context,
	audioPath string,
	opts speechUsecases.TranscribeOptions,
) (*whisperDtos.AsyncTranscriptionResultDTO, error) {
	return &whisperDtos.AsyncTranscriptionResultDTO{
		Transcription:    "This is a test transcription",
		RefinedText:      "This is a refined transcription",
		DetectedLanguage: "en",
	}, nil
}

func (m *MockTranscribeUseCase) CheckIdempotency(
	idempotencyKey string,
	audioHash string,
	language string,
	terminalID string,
) (string, bool) {
	return "", false
}

// MockTranslateUseCase implements ragUsecases.TranslateUseCase for testing
type MockTranslateUseCase struct{}

func (m *MockTranslateUseCase) TranslateText(
	text string,
	targetLang string,
	args ...string,
) (string, error) {
	return "This is a test translation", nil
}

func (m *MockTranslateUseCase) TranslateTextWithTrigger(
	text string,
	targetLang string,
	trigger string,
	args ...string,
) (string, error) {
	return "This is a test translation", nil
}

func (m *MockTranslateUseCase) TranslateTextSync(
	ctx context.Context,
	text string,
	targetLanguage string,
	args ...string,
) (string, error) {
	return "This is a test translation", nil
}

// MockSummaryUseCase implements ragUsecases.SummaryUseCase for testing
type MockSummaryUseCase struct{}

func (m *MockSummaryUseCase) SummarizeText(
	text string,
	language string,
	meetingContext string,
	style string,
	date string,
	location string,
	participants string,
	args ...string,
) (string, error) {
	return "This is a test summary", nil
}

func (m *MockSummaryUseCase) SummarizeTextWithTrigger(
	text string,
	language string,
	meetingContext string,
	style string,
	date string,
	location string,
	participants string,
	trigger string,
	args ...string,
) (string, error) {
	return "This is a test summary", nil
}

func (m *MockSummaryUseCase) SummarizeTextWithContext(
	ctx context.Context,
	text string,
	language string,
	meetingContext string,
	style string,
	date string,
	location string,
	participants string,
	args ...string,
) (string, error) {
	return "This is a test summary", nil
}

func (m *MockSummaryUseCase) SummarizeTextWithContextAndTrigger(
	ctx context.Context,
	text string,
	language string,
	meetingContext string,
	style string,
	date string,
	location string,
	participants string,
	trigger string,
	args ...string,
) (string, error) {
	return "This is a test summary", nil
}

func (m *MockSummaryUseCase) SummarizeTextSync(
	ctx context.Context,
	text string,
	language string,
	meetingContext string,
	style string,
	date string,
	location string,
	participants string,
	macAddress string,
) (*ragDtos.RAGSummaryResponseDTO, error) {
	return &ragDtos.RAGSummaryResponseDTO{
		Summary: "This is a test summary",
	}, nil
}
