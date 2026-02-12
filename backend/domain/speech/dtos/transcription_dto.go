package dtos

type TranscriptionResponseDTO struct {
	Transcription  string `json:"transcription" example:"Halo dunia"`
	RefinedText    string `json:"refined_text,omitempty" example:"Hello world"`
}

type TranscriptionLongResponseDTO struct {
	Transcription    string `json:"transcription" example:"Ini adalah transkripsi yang sangat panjang..."`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`
}

// OutsystemsTranscriptionResultDTO represents the structured response from Outsystems
type OutsystemsTranscriptionResultDTO struct {
	Filename         string `json:"filename"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
	AudioUrl         string `json:"audio_url"` // Public URL to access the audio file
	Transcription    string `json:"transcription" example:"Halo dunia"`
	DetectedLanguage string `json:"detected_language" example:"id"`
}

type WhisperProxyStatusDTO struct {
	Status          string                            `json:"status" example:"completed"`
	Result          *OutsystemsTranscriptionResultDTO `json:"result,omitempty"`
	ExpiresAt       string                            `json:"expires_at,omitempty"`
	ExpiresInSecond int64                             `json:"expires_in_seconds,omitempty"`
}

type TranscriptionTaskResponseDTO struct {
	TaskID      string `json:"task_id" example:"abc-123"`
	TaskStatus  string `json:"task_status" example:"pending"`
	RecordingID string `json:"recording_id,omitempty" example:"uuid-v4"`
}

type WhisperProxyProcessResponseDTO = TranscriptionTaskResponseDTO

type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Details interface{} `json:"details,omitempty"`
}

type MqttPublishRequest struct {
	Message string `json:"message" binding:"required"`
}

// Async Transcription DTOs (Consolidated)
type AsyncTranscriptionResultDTO struct {
	Transcription    string `json:"transcription" example:"Halo dunia"`
	RefinedText      string `json:"refined_text,omitempty" example:"Hello world"`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`
}

type AsyncTranscriptionStatusDTO struct {
	Status          string                       `json:"status" example:"completed"`
	Result          *AsyncTranscriptionResultDTO `json:"result,omitempty"`
	ExpiresAt       string                       `json:"expires_at,omitempty"`
	ExpiresInSecond int64                        `json:"expires_in_seconds,omitempty"`
}

type AsyncTranscriptionProcessStatusResponseDTO struct {
	TaskID     string                       `json:"task_id"`
	TaskStatus *AsyncTranscriptionStatusDTO `json:"task_status,omitempty"`
}

type WhisperProxyProcessStatusResponseDTO struct {
	TaskID     string                 `json:"task_id"`
	TaskStatus *WhisperProxyStatusDTO `json:"task_status,omitempty"`
}

type AsyncTranscriptionProcessResponseDTO = TranscriptionTaskResponseDTO

// Keeping these for backward compatibility if used in code, but they are now aliases/identical
type AsyncTranscriptionLongResultDTO = AsyncTranscriptionResultDTO
type AsyncTranscriptionLongStatusDTO = AsyncTranscriptionStatusDTO
type AsyncTranscriptionLongProcessResponseDTO = AsyncTranscriptionProcessResponseDTO

// StatusUpdateMessage is used for real-time notifications via MQTT
type StatusUpdateMessage struct {
	Type   string      `json:"type"`
	TaskID string      `json:"task_id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}
