package usecases

import (
	"teralux_app/domain/common/tasks"
	ragdtos "teralux_app/domain/rag/dtos"
)

// RAGStatusUseCase is a type alias for the generic status usecase from common/tasks.
// This provides status retrieval functionality without duplicating code.
type RAGStatusUseCase = tasks.GenericStatusUseCase[ragdtos.RAGStatusDTO]

// NewRAGStatusUseCase creates a new RAG status usecase using the generic implementation.
// This is a convenience wrapper around tasks.NewGenericStatusUseCase.
//
// Parameters:
//   - cache: BadgerTaskCache for persistent storage (can be nil)
//   - store: StatusStore for in-memory storage (required)
//
// Returns:
//   - RAGStatusUseCase: A status usecase instance
func NewRAGStatusUseCase(cache *tasks.BadgerTaskCache, store *tasks.StatusStore[ragdtos.RAGStatusDTO]) RAGStatusUseCase {
	return tasks.NewGenericStatusUseCase(cache, store)
}
