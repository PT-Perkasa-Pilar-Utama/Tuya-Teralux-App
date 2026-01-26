package dtos

type RAGRequest struct {
	Text string `json:"text" binding:"required"`
}

type RAGStatus string

const (
	RAGStatusEnqueue  RAGStatus = "enqueue"
	RAGStatusPending  RAGStatus = "pending"
	RAGStatusFinished RAGStatus = "finished"
)

type RAGResponse struct {
	TaskID string    `json:"task_id"`
	Status RAGStatus `json:"status"`
	Result string    `json:"result,omitempty"`
}
