package dtos

type TranscriptionResponseDTO struct {
	Text           string `json:"text"`
	TranslatedText string `json:"translated_text,omitempty"`
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
