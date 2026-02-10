package dtos

type TranscriptionResponseDTO struct {
	Text           string `json:"text" example:"Halo dunia"`
	TranslatedText string `json:"translated_text,omitempty" example:"Hello world"`
}

type TranscriptionLongResponseDTO struct {
	Text             string `json:"text" example:"Ini adalah transkripsi yang sangat panjang..."`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`
}

// OutsystemsTranscriptionResultDTO represents the structured response from Outsystems
type OutsystemsTranscriptionResultDTO struct {
	Filename         string `json:"filename" example:"audio.mp3"`
	Transcription    string `json:"transcription" example:"Halo dunia"`
	DetectedLanguage string `json:"detected_language" example:"id"`
}

type WhisperProxyStatusDTO struct {
	Status          string                            `json:"status" example:"completed"`
	Result          *OutsystemsTranscriptionResultDTO `json:"result,omitempty"`
	ExpiresAt       string                            `json:"expires_at,omitempty"`
	ExpiresInSecond int64                             `json:"expires_in_seconds,omitempty"`
}

type WhisperProxyProcessResponseDTO struct {
	TaskID string                 `json:"task_id"`
	Status *WhisperProxyStatusDTO `json:"status,omitempty"`
}

type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details string      `json:"details,omitempty"`
}

type MqttPublishRequest struct {
	Message string `json:"message" binding:"required"`
}

// Async Transcription DTOs
type AsyncTranscriptionResultDTO struct {
	Text           string `json:"text" example:"Halo dunia"`
	TranslatedText string `json:"translated_text,omitempty" example:"Hello world"`
}

type AsyncTranscriptionLongResultDTO struct {
	Text             string `json:"text" example:"Ini adalah transkripsi yang sangat panjang..."`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`
}

type AsyncTranscriptionStatusDTO struct {
	Status          string                       `json:"status" example:"completed"`
	Result          *AsyncTranscriptionResultDTO `json:"result,omitempty"`
	ExpiresAt       string                       `json:"expires_at,omitempty"`
	ExpiresInSecond int64                        `json:"expires_in_seconds,omitempty"`
}

type AsyncTranscriptionLongStatusDTO struct {
	Status          string                           `json:"status" example:"completed"`
	Result          *AsyncTranscriptionLongResultDTO `json:"result,omitempty"`
	ExpiresAt       string                           `json:"expires_at,omitempty"`
	ExpiresInSecond int64                            `json:"expires_in_seconds,omitempty"`
}

type AsyncTranscriptionProcessResponseDTO struct {
	TaskID string                       `json:"task_id"`
	Status *AsyncTranscriptionStatusDTO `json:"status,omitempty"`
}

type AsyncTranscriptionLongProcessResponseDTO struct {
	TaskID string                           `json:"task_id"`
	Status *AsyncTranscriptionLongStatusDTO `json:"status,omitempty"`
}

// StatusUpdateMessage is used for real-time notifications via MQTT
type StatusUpdateMessage struct {
	Type   string      `json:"type"`
	TaskID string      `json:"task_id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}
