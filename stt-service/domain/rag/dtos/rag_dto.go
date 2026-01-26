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

type RAGTask struct {
	ID     string    `json:"id"`
	Status RAGStatus `json:"status"`
	Result string    `json:"result,omitempty"`
}

type RAGResponse struct {
	Tasks RAGTask `json:"tasks"`
}
