package dtos

// WhisperResult represents the result of a transcription from any provider
type WhisperResult struct {
	Transcription    string
	DetectedLanguage string
	Source           string // Which service was used: "Orion", "Local", "Gemini"
}

type WhisperMqttRequestDTO struct {
	Audio     string `json:"audio" binding:"required"` // Base64 encoded audio
	Language  string `json:"language,omitempty"`
	TeraluxID string `json:"teralux_id" binding:"required"`
	UID       string `json:"uid,omitempty"`
}


// OutsystemsTranscriptionResultDTO represents the structured response from Outsystems
type OutsystemsTranscriptionResultDTO struct {
	Filename         string `json:"filename"`
	Transcription    string `json:"transcription" example:"Halo dunia"`
	DetectedLanguage string `json:"detected_language" example:"id"`
}


type TranscriptionTaskResponseDTO struct {
	TaskID      string `json:"task_id" example:"abc-123"`
	TaskStatus  string `json:"task_status" example:"pending"`
	RecordingID string `json:"recording_id,omitempty" example:"uuid-v4"`
}


type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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
	Error           string                       `json:"error,omitempty" example:"service unavailable"`
	Trigger         string                       `json:"trigger,omitempty" example:"/api/speech/models/gemini"`
	HTTPStatusCode  int                          `json:"-"`
	StartedAt       string                       `json:"started_at,omitempty" example:"2026-02-21T11:00:00Z"`
	DurationSeconds float64                      `json:"duration_seconds,omitempty" example:"1.5"`
	ExpiresAt       string                       `json:"expires_at,omitempty"`
	ExpiresInSecond int64                        `json:"expires_in_seconds,omitempty"`
}

// SetExpiry implements tasks.StatusWithExpiry interface
func (s *AsyncTranscriptionStatusDTO) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSecond = expiresInSeconds
}

type AsyncTranscriptionProcessStatusResponseDTO struct {
	TaskID     string                       `json:"task_id"`
	TaskStatus *AsyncTranscriptionStatusDTO `json:"task_status,omitempty"`
}


type AsyncTranscriptionProcessResponseDTO = TranscriptionTaskResponseDTO

// StatusUpdateMessage is used for real-time notifications via MQTT
type StatusUpdateMessage struct {
	Type   string      `json:"type"`
	TaskID string      `json:"task_id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}
