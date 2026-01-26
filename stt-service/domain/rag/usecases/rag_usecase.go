package usecases

import (
	"os"
	"stt-service/domain/rag/dtos"
	"stt-service/domain/rag/repositories"
	speechRepos "stt-service/domain/speech/repositories"
)

type RAGUsecase interface {
	ProcessText(text string) (dtos.RAGResponse, error)
	GetStatus(taskID string) (dtos.RAGResponse, error)
}

type ragUsecase struct {
	vectorRepo repositories.VectorRepository
	ollamaRepo *speechRepos.OllamaRepository
}

func NewRAGUsecase(vectorRepo repositories.VectorRepository, ollamaRepo *speechRepos.OllamaRepository) RAGUsecase {
	return &ragUsecase{
		vectorRepo: vectorRepo,
		ollamaRepo: ollamaRepo,
	}
}

func (u *ragUsecase) ProcessText(text string) (dtos.RAGResponse, error) {
	// For now, we skip the RAG process (retrieval) and focus on LLM connection
	// We send the input as-is to the local Ollama instance

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://host.docker.internal:11434"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gemma" // Default model as per Makefile
	}

	res, err := u.ollamaRepo.ProcessPrompt(ollamaURL, model, text)
	if err != nil {
		return dtos.RAGResponse{
			TaskID: "task-error",
			Status: dtos.RAGStatusFinished,
			Result: "Error from Ollama: " + err.Error(),
		}, nil // Returning nil error so controller can return the error result to user
	}

	return dtos.RAGResponse{
		TaskID: "task-12345",
		Status: dtos.RAGStatusFinished,
		Result: res,
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
