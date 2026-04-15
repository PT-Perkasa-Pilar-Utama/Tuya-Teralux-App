package contracts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensio/domain/models/whisper/dtos"
)

func TestAsyncTranscriptionResultDTOContract(t *testing.T) {
	dto := dtos.AsyncTranscriptionResultDTO{
		Transcription:    "Halo dunia",
		RefinedText:      "Hello world",
		DetectedLanguage: "id",
		AudioClass:       "active",
		TranscriptValid:  true,
		ProviderName:     "openai",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Halo dunia", decoded["transcription"])
	assert.Equal(t, "Hello world", decoded["refined_text"])
	assert.Equal(t, "id", decoded["detected_language"])
	assert.Equal(t, "active", decoded["audio_class"])
	assert.Equal(t, true, decoded["transcript_valid"])
	assert.Equal(t, "openai", decoded["provider_name"])
}

func TestAsyncTranscriptionStatusDTOContract(t *testing.T) {
	dto := dtos.AsyncTranscriptionStatusDTO{
		Status:          "completed",
		Trigger:         "/api/whisper/transcribe",
		MacAddress:      "AA:BB:CC:DD:EE:FF",
		TerminalID:      "term-123",
		StartedAt:       "2026-04-15T10:00:00Z",
		DurationSeconds: 5.2,
		Result: &dtos.AsyncTranscriptionResultDTO{
			Transcription:    "Test transcription",
			DetectedLanguage: "id",
			AudioClass:       "active",
			TranscriptValid:  true,
		},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "completed", decoded["status"])
	assert.Equal(t, "/api/whisper/transcribe", decoded["trigger"])
	assert.NotNil(t, decoded["result"])

	result := decoded["result"].(map[string]interface{})
	assert.Equal(t, "Test transcription", result["transcription"])
}

func TestTranscriptionTaskResponseDTOContract(t *testing.T) {
	dto := dtos.TranscriptionTaskResponseDTO{
		TaskID:      "task-123",
		TaskStatus:  "completed",
		RecordingID: "rec-456",
		RequestID:   "req-789",
		Source:      "WHISPER_WORKER",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "task-123", decoded["task_id"])
	assert.Equal(t, "completed", decoded["task_status"])
	assert.Equal(t, "rec-456", decoded["recording_id"])
	assert.Equal(t, "WHISPER_WORKER", decoded["source"])
}

func TestUploadSessionResponseDTOContract(t *testing.T) {
	dto := dtos.UploadSessionResponseDTO{
		SessionID:      "session-123",
		State:          "uploading",
		TotalChunks:    10,
		ChunkSizeBytes: 1024,
		TotalSizeBytes: 10240,
		ReceivedBytes:  5120,
		MissingRanges:  []string{"5", "6", "7"},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "session-123", decoded["session_id"])
	assert.Equal(t, "uploading", decoded["state"])
	assert.Equal(t, float64(10), decoded["total_chunks"])
	assert.Equal(t, float64(1024), decoded["chunk_size_bytes"])
	assert.NotNil(t, decoded["missing_ranges"])
}

func TestWhisperMqttRequestDTOContract(t *testing.T) {
	dto := dtos.WhisperMqttRequestDTO{
		RequestID:  "req-123",
		Audio:      "base64encodedaudio...",
		Language:   "id",
		TerminalID: "term-123",
		UID:        "uid-456",
		Diarize:    true,
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "req-123", decoded["request_id"])
	assert.Equal(t, "base64encodedaudio...", decoded["audio"])
	assert.Equal(t, "id", decoded["language"])
	assert.Equal(t, "term-123", decoded["terminal_id"])
	assert.Equal(t, true, decoded["diarize"])
}
