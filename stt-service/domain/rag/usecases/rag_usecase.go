package usecases

import (
	"stt-service/domain/rag/dtos"
	"stt-service/domain/rag/repositories"
)

type RAGUsecase interface {
	ProcessText(text string) (dtos.RAGResponse, error)
	GetStatus(taskID string) (dtos.RAGResponse, error)
}

type ragUsecase struct {
	vectorRepo repositories.VectorRepository
}

func NewRAGUsecase(vectorRepo repositories.VectorRepository) RAGUsecase {
	return &ragUsecase{
		vectorRepo: vectorRepo,
	}
}

func (u *ragUsecase) ProcessText(text string) (dtos.RAGResponse, error) {
	// For now, we skip the RAG process and return a mocked response
	// Simulate the flow: enqueue -> pending -> finished
	// In a real scenario, this might involve async workers or multiple steps

	return dtos.RAGResponse{
		TaskID: "task-12345",
		Status: dtos.RAGStatusFinished,
		Result: "Processed: " + text + " (Note: RAG logic is skipped)",
	}, nil
}

func (u *ragUsecase) GetStatus(taskID string) (dtos.RAGResponse, error) {
	// Mocked status retrieval
	return dtos.RAGResponse{
		TaskID: taskID,
		Status: dtos.RAGStatusFinished,
		Result: "Final processed result for " + taskID,
	}, nil
}
