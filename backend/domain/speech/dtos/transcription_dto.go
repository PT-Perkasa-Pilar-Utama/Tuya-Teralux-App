package dtos

type TranscriptionResponseDTO struct {
	Text           string `json:"text" example:"Halo dunia"`
	TranslatedText string `json:"translated_text,omitempty" example:"Hello world"`
}

type TranscriptionLongResponseDTO struct {
	Text             string `json:"text" example:"Ini adalah transkripsi yang sangat panjang..."`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`
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
