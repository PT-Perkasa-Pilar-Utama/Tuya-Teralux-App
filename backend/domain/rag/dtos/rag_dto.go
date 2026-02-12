package dtos

type RAGRequestDTO struct {
	Text     string `json:"text" binding:"required"`
	Language string `json:"language,omitempty"`
}

type RAGSummaryRequestDTO struct {
	Text     string `json:"text" binding:"required" example:"This is a long transcript of a technical meeting..."`
	Language string `json:"language,omitempty" example:"id"` // "id" or "en"
	Context  string `json:"context,omitempty" example:"technical meeting"`
	Style    string `json:"style,omitempty" example:"professional"`
}

type RAGStatusDTO struct {
	Status          string            `json:"status"`
	Result          string            `json:"result,omitempty"` // raw LLM response when not structured
	Endpoint        string            `json:"endpoint,omitempty"`
	Method          string            `json:"method,omitempty"`
	Body            interface{}       `json:"body,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	ExecutionResult interface{}       `json:"execution_result,omitempty"` // holds the response from the fetched endpoint
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

type RAGSummaryResponseDTO struct {
	Summary string `json:"summary"`
	PDFUrl  string `json:"pdf_url,omitempty"`
}
