package dtos

type TranscriptionResponseDTO struct {
	Text string `json:"text"`
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
