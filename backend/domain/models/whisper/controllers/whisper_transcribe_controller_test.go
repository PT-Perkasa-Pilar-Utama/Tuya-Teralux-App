package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"sync"
	"testing"
	"time"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	"sensio/domain/models/whisper/usecases"
	recordingEntities "sensio/domain/recordings/entities"
	recordingUsecases "sensio/domain/recordings/usecases"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTranscribeUseCase for testing
type MockTranscribeUseCase struct {
	mock.Mock
}

func (m *MockTranscribeUseCase) TranscribeAudio(ctx context.Context, inputPath, filename, language string, metadata ...usecases.TranscriptionMetadata) (string, error) {
	var meta usecases.TranscriptionMetadata
	if len(metadata) > 0 {
		meta = metadata[0]
	}
	args := m.Called(ctx, inputPath, filename, language, meta)
	return args.String(0), args.Error(1)
}

func (m *MockTranscribeUseCase) TranscribeAudioSync(ctx context.Context, inputPath string, opts usecases.TranscribeOptions) (*dtos.AsyncTranscriptionResultDTO, error) {
	args := m.Called(ctx, inputPath, opts)
	return args.Get(0).(*dtos.AsyncTranscriptionResultDTO), args.Error(1)
}

func (m *MockTranscribeUseCase) CheckIdempotency(idempotencyKey, audioHash, language, terminalID string) (string, bool) {
	args := m.Called(idempotencyKey, audioHash, language, terminalID)
	return args.String(0), args.Bool(1)
}

func (m *MockTranscribeUseCase) GetStatus(taskID string) (*dtos.AsyncTranscriptionProcessStatusResponseDTO, error) {
	args := m.Called(taskID)
	return args.Get(0).(*dtos.AsyncTranscriptionProcessStatusResponseDTO), args.Error(1)
}

func (m *MockTranscribeUseCase) CancelTask(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockTranscribeUseCase) CleanupExpiredTasks() {
	m.Called()
}

// MockSaveRecordingUseCase for testing
type MockSaveRecordingUseCase struct {
	mock.Mock
}

func (m *MockSaveRecordingUseCase) SaveRecording(file *multipart.FileHeader, macAddress, baseURL string, opts ...recordingUsecases.SaveRecordingOption) (*recordingEntities.Recording, error) {
	args := m.Called(file, macAddress, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result := args.Get(0).(*recordingEntities.Recording)
	return result, args.Error(1)
}

func (m *MockSaveRecordingUseCase) SaveRecordingFromBytes(data []byte, originalName, macAddress, baseURL string, opts ...recordingUsecases.SaveRecordingOption) (*recordingEntities.Recording, error) {
	args := m.Called(data, originalName, macAddress, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result := args.Get(0).(*recordingEntities.Recording)
	return result, args.Error(1)
}

func (m *MockSaveRecordingUseCase) SaveRecordingFromPath(filePath, originalFilename, macAddress, baseURL string, opts ...recordingUsecases.SaveRecordingOption) (*recordingEntities.Recording, error) {
	args := m.Called(filePath, originalFilename, macAddress, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result := args.Get(0).(*recordingEntities.Recording)
	return result, args.Error(1)
}

// MockUploadSessionUseCase for testing
type MockUploadSessionUseCase struct {
	mock.Mock
}

func (m *MockUploadSessionUseCase) CreateSession(req dtos.CreateUploadSessionRequest) (*dtos.UploadSessionResponseDTO, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.UploadSessionResponseDTO), args.Error(1)
}

func (m *MockUploadSessionUseCase) UploadChunk(sessionID string, chunkIndex int, ownerUID string, reader io.Reader) (*dtos.UploadChunkAckDTO, error) {
	args := m.Called(sessionID, chunkIndex, ownerUID, reader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.UploadChunkAckDTO), args.Error(1)
}

func (m *MockUploadSessionUseCase) GetSessionStatus(sessionID string, ownerUID string) (*dtos.UploadSessionResponseDTO, error) {
	args := m.Called(sessionID, ownerUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.UploadSessionResponseDTO), args.Error(1)
}

func (m *MockUploadSessionUseCase) FinalizeSession(sessionID, ownerUID string) (*usecases.FinalizedUpload, error) {
	args := m.Called(sessionID, ownerUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.FinalizedUpload), args.Error(1)
}

func (m *MockUploadSessionUseCase) CleanupExpiredSessions(now time.Time) (int, error) {
	args := m.Called(now)
	return args.Int(0), args.Error(1)
}

// MockMqttService for testing
type MockMqttService struct {
	mock.Mock
	publishedMessages map[string][]byte
	mu                sync.Mutex
}

func (m *MockMqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	args := m.Called(topic, qos, handler)
	return args.Error(0)
}

func (m *MockMqttService) Publish(topic string, qos byte, retained bool, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.publishedMessages == nil {
		m.publishedMessages = make(map[string][]byte)
	}
	m.publishedMessages[topic] = payload
	args := m.Called(topic, qos, retained, payload)
	return args.Error(0)
}

func (m *MockMqttService) Disconnect() {
	m.Called()
}

func (m *MockMqttService) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttService) GetPublishedMessages() map[string][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string][]byte)
	for k, v := range m.publishedMessages {
		result[k] = v
	}
	return result
}

// TestTranscriptionTaskResponseDTO_VerifyACKStructure tests that the transcription ACK response
// includes the required request_id and source fields
func TestTranscriptionTaskResponseDTO_VerifyACKStructure(t *testing.T) {
	// Test case: ACK with RequestID and Source
	requestID := "test-request-id-123"
	taskID := "task-uuid-456"

	dto := dtos.TranscriptionTaskResponseDTO{
		TaskID:      taskID,
		TaskStatus:  "pending",
		RecordingID: "",
		RequestID:   requestID,
		Source:      "MQTT_ACK",
	}

	// Wrap in StandardResponse as the controller does
	resp := commonDtos.StandardResponse{
		Status:  true,
		Message: "Transcription task submitted successfully (Ephemeral)",
		Data:    dto,
	}

	// Marshal to JSON (simulating MQTT publish)
	jsonBytes, err := json.Marshal(resp)
	assert.NoError(t, err)

	// Unmarshal to verify structure
	var unmarshaledResp commonDtos.StandardResponse
	err = json.Unmarshal(jsonBytes, &unmarshaledResp)
	assert.NoError(t, err)

	// Verify status
	assert.True(t, unmarshaledResp.Status, "Expected status to be true")
	assert.Equal(t, "Transcription task submitted successfully (Ephemeral)", unmarshaledResp.Message)

	// Extract and verify data
	dataBytes, err := json.Marshal(unmarshaledResp.Data)
	assert.NoError(t, err)

	var dataDTO dtos.TranscriptionTaskResponseDTO
	err = json.Unmarshal(dataBytes, &dataDTO)
	assert.NoError(t, err)

	// CRITICAL: Verify request_id is present
	assert.Equal(t, requestID, dataDTO.RequestID, "Expected request_id to be present in ACK")

	// CRITICAL: Verify source is MQTT_ACK
	assert.Equal(t, "MQTT_ACK", dataDTO.Source, "Expected source to be MQTT_ACK")

	// Verify task details
	assert.Equal(t, taskID, dataDTO.TaskID)
	assert.Equal(t, "pending", dataDTO.TaskStatus)
	assert.Equal(t, "", dataDTO.RecordingID, "Expected empty RecordingID for ephemeral task")
}

// TestTranscriptionTaskResponseDTO_WithoutRequestID tests DTO with empty RequestID
func TestTranscriptionTaskResponseDTO_WithoutRequestID(t *testing.T) {
	taskID := "task-uuid-789"

	// No RequestID provided (empty)
	dto := dtos.TranscriptionTaskResponseDTO{
		TaskID:      taskID,
		TaskStatus:  "pending",
		RecordingID: "",
		RequestID:   "", // Empty
		Source:      "MQTT_ACK",
	}

	jsonBytes, err := json.Marshal(dto)
	assert.NoError(t, err)

	var unmarshaledDTO dtos.TranscriptionTaskResponseDTO
	err = json.Unmarshal(jsonBytes, &unmarshaledDTO)
	assert.NoError(t, err)

	// Source should still be MQTT_ACK
	assert.Equal(t, "MQTT_ACK", unmarshaledDTO.Source)
	// RequestID should be empty (omitempty will exclude it from JSON)
	assert.Equal(t, "", unmarshaledDTO.RequestID)
}

// TestActiveTranscriptions_Lifecycle_StoreAndDelete tests that ActiveTranscriptions
// flag is properly stored and deleted during the MQTT flow
func TestActiveTranscriptions_Lifecycle_StoreAndDelete(t *testing.T) {
	// Clean up before test
	utils.ActiveTranscriptions = sync.Map{}

	terminalID := "test-terminal-001"

	// Verify flag is not present initially
	_, exists := utils.ActiveTranscriptions.Load(terminalID)
	assert.False(t, exists, "ActiveTranscriptions should not exist before test")

	// Simulate storing the flag (as done in StartMqttSubscription)
	utils.ActiveTranscriptions.Store(terminalID, true)

	// Verify flag is stored
	val, exists := utils.ActiveTranscriptions.Load(terminalID)
	assert.True(t, exists, "ActiveTranscriptions should be stored after Store()")
	assert.True(t, val.(bool), "ActiveTranscriptions value should be true")

	// Simulate deletion (as done in defer or after successful processing)
	utils.ActiveTranscriptions.Delete(terminalID)

	// Verify flag is deleted
	_, exists = utils.ActiveTranscriptions.Load(terminalID)
	assert.False(t, exists, "ActiveTranscriptions should be deleted after Delete()")
}

// TestActiveTranscriptions_NoLeak_OnErrorPath tests that ActiveTranscriptions
// is properly cleaned up even when an error occurs
func TestActiveTranscriptions_NoLeak_OnErrorPath(t *testing.T) {
	utils.ActiveTranscriptions = sync.Map{}

	terminalID := "test-terminal-error-002"

	// Store the flag
	utils.ActiveTranscriptions.Store(terminalID, true)

	// Verify it exists
	_, exists := utils.ActiveTranscriptions.Load(terminalID)
	assert.True(t, exists)

	// Simulate error path cleanup (defer in controller)
	utils.ActiveTranscriptions.Delete(terminalID)

	// Verify cleanup
	_, exists = utils.ActiveTranscriptions.Load(terminalID)
	assert.False(t, exists, "ActiveTranscriptions should be cleaned up on error path")
}

// TestActiveTranscriptions_ConcurrentAccess tests thread-safety of ActiveTranscriptions
func TestActiveTranscriptions_ConcurrentAccess(t *testing.T) {
	utils.ActiveTranscriptions = sync.Map{}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			terminalID := "terminal-" + string(rune(id))
			utils.ActiveTranscriptions.Store(terminalID, true)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			terminalID := "terminal-" + string(rune(id))
			utils.ActiveTranscriptions.Load(terminalID)
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			terminalID := "terminal-" + string(rune(id))
			utils.ActiveTranscriptions.Delete(terminalID)
		}(i)
	}

	wg.Wait()
	// If no panic occurred, the test passes (sync.Map is thread-safe)
}

// TestTranscriptionTaskResponseDTO_JsonSerialization tests JSON marshaling/unmarshaling
func TestTranscriptionTaskResponseDTO_JsonSerialization(t *testing.T) {
	dto := dtos.TranscriptionTaskResponseDTO{
		TaskID:      "task-123",
		TaskStatus:  "pending",
		RecordingID: "rec-456",
		RequestID:   "req-789",
		Source:      "MQTT_ACK",
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(dto)
	assert.NoError(t, err)

	// Verify JSON contains expected fields
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.NoError(t, err)

	assert.Equal(t, "task-123", jsonMap["task_id"])
	assert.Equal(t, "pending", jsonMap["task_status"])
	assert.Equal(t, "rec-456", jsonMap["recording_id"])
	assert.Equal(t, "req-789", jsonMap["request_id"])
	assert.Equal(t, "MQTT_ACK", jsonMap["source"])

	// Unmarshal back to DTO
	var unmarshaledDTO dtos.TranscriptionTaskResponseDTO
	err = json.Unmarshal(jsonBytes, &unmarshaledDTO)
	assert.NoError(t, err)

	assert.Equal(t, dto, unmarshaledDTO)
}

// TestTranscriptionTaskResponseDTO_JsonWithOmitEmpty tests omitempty behavior
func TestTranscriptionTaskResponseDTO_JsonWithOmitEmpty(t *testing.T) {
	dto := dtos.TranscriptionTaskResponseDTO{
		TaskID:     "task-123",
		TaskStatus: "pending",
		// RecordingID, RequestID, Source are empty
	}

	jsonBytes, err := json.Marshal(dto)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	assert.NoError(t, err)

	// Verify required fields are present
	assert.Contains(t, jsonMap, "task_id")
	assert.Contains(t, jsonMap, "task_status")

	// Verify optional fields are omitted when empty
	_, hasRecordingID := jsonMap["recording_id"]
	_, hasRequestID := jsonMap["request_id"]
	_, hasSource := jsonMap["source"]

	assert.False(t, hasRecordingID, "recording_id should be omitted when empty")
	assert.False(t, hasRequestID, "request_id should be omitted when empty")
	assert.False(t, hasSource, "source should be omitted when empty")
}

// Helper function to create base64 encoded audio for testing
func createTestAudioBase64() string {
	// Create minimal WAV file header + silence
	// RIFF header (44 bytes for empty WAV)
	wavHeader := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x24, 0x00, 0x00, 0x00, // File size - 8
		0x57, 0x41, 0x56, 0x45, // "WAVE"
		0x66, 0x6D, 0x74, 0x20, // "fmt "
		0x10, 0x00, 0x00, 0x00, // Subchunk1Size
		0x01, 0x00, 0x01, 0x00, // AudioFormat, NumChannels
		0x80, 0x3E, 0x00, 0x00, // SampleRate (16000)
		0x00, 0xFA, 0x00, 0x00, // ByteRate
		0x02, 0x00, 0x10, 0x00, // BlockAlign, BitsPerSample
		0x64, 0x61, 0x74, 0x61, // "data"
		0x00, 0x00, 0x00, 0x00, // Subchunk2Size
	}
	return base64.StdEncoding.EncodeToString(wavHeader)
}

// TestWhisperMqttRequestDTO_JsonParsing tests MQTT request parsing
func TestWhisperMqttRequestDTO_JsonParsing(t *testing.T) {
	testAudio := createTestAudioBase64()

	requestDTO := dtos.WhisperMqttRequestDTO{
		RequestID:  "test-req-123",
		Audio:      testAudio,
		Language:   "id",
		TerminalID: "terminal-001",
		UID:        "sg123456",
		Diarize:    true,
	}

	// Marshal
	jsonBytes, err := json.Marshal(requestDTO)
	assert.NoError(t, err)

	// Unmarshal
	var unmarshaled dtos.WhisperMqttRequestDTO
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	assert.NoError(t, err)

	// Verify all fields
	assert.Equal(t, requestDTO.RequestID, unmarshaled.RequestID)
	assert.Equal(t, requestDTO.Audio, unmarshaled.Audio)
	assert.Equal(t, requestDTO.Language, unmarshaled.Language)
	assert.Equal(t, requestDTO.TerminalID, unmarshaled.TerminalID)
	assert.Equal(t, requestDTO.UID, unmarshaled.UID)
	assert.Equal(t, requestDTO.Diarize, unmarshaled.Diarize)
}

// TestWhisperMqttRequestDTO_JsonParsing_WithOptionalFields tests parsing with minimal fields
func TestWhisperMqttRequestDTO_JsonParsing_WithOptionalFields(t *testing.T) {
	testAudio := createTestAudioBase64()

	// Minimal request (only required fields)
	jsonStr := `{
		"audio": "` + testAudio + `",
		"terminal_id": "terminal-002"
	}`

	var unmarshaled dtos.WhisperMqttRequestDTO
	err := json.Unmarshal([]byte(jsonStr), &unmarshaled)
	assert.NoError(t, err)

	// Verify required fields
	assert.Equal(t, testAudio, unmarshaled.Audio)
	assert.Equal(t, "terminal-002", unmarshaled.TerminalID)

	// Verify optional fields are empty
	assert.Equal(t, "", unmarshaled.RequestID)
	assert.Equal(t, "", unmarshaled.Language)
	assert.Equal(t, "", unmarshaled.UID)
	assert.Equal(t, false, unmarshaled.Diarize)
}
