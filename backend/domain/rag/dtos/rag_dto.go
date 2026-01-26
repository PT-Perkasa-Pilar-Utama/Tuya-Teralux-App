package dtos

type RAGRequestDTO struct {
	Text string `json:"text"`
}

type RAGStatusDTO struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details string      `json:"details,omitempty"`
}
