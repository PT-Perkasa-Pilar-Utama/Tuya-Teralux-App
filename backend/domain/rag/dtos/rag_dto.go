package dtos

type RAGRequestDTO struct {
	Text string `json:"text"`
}

type RAGStatusDTO struct {
	Status          string            `json:"status"`
	Result          string            `json:"result,omitempty"` // raw LLM response when not structured
	Endpoint        string            `json:"-"`
	Method          string            `json:"-"`
	Body            interface{}       `json:"-"`
	Headers         map[string]string `json:"-"`
	ExecutionResult interface{}       `json:"-"` // holds the response from the fetched endpoint
	ExpiresAt       string            `json:"expires_at,omitempty"`
	ExpiresInSecond int64             `json:"expires_in_seconds,omitempty"`
}

// RAGProcessResponseDTO is the payload returned by POST /api/rag
// It contains the generated task id and optionally a `status` DTO when available.
type RAGProcessResponseDTO struct {
	TaskID     string        `json:"task_id"`
	TaskStatus *RAGStatusDTO `json:"task_status,omitempty"`
}

type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details string      `json:"details,omitempty"`
}
