package usecases

import (
	"fmt"
	"stt-service/domain/common/config"
	"stt-service/domain/rag/dtos"
	"stt-service/domain/rag/repositories"
	speechRepos "stt-service/domain/speech/repositories"
	"sync"

	"github.com/google/uuid"
)

type RAGUsecase interface {
	ProcessText(text string) (dtos.RAGResponse, error)
	GetStatus(taskID string) (dtos.RAGResponse, error)
}

type ragUsecase struct {
	vectorRepo repositories.VectorRepository
	ollamaRepo *speechRepos.OllamaRepository
	config     *config.Config
	tasks      sync.Map
}

func NewRAGUsecase(vectorRepo repositories.VectorRepository, ollamaRepo *speechRepos.OllamaRepository, cfg *config.Config) RAGUsecase {
	return &ragUsecase{
		vectorRepo: vectorRepo,
		ollamaRepo: ollamaRepo,
		config:     cfg,
		tasks:      sync.Map{},
	}
}

func (u *ragUsecase) ProcessText(text string) (dtos.RAGResponse, error) {
	taskID := uuid.NewString()

	// Initialize task as pending
	initialResponse := dtos.RAGResponse{
		TaskID: taskID,
		Status: dtos.RAGStatusPending,
	}
	u.tasks.Store(taskID, initialResponse)

	// Start asynchronous processing
	go func(id string, input string) {
		// Simulate some processing delay if needed, but Ollama usually takes time anyway
		res, err := u.ollamaRepo.ProcessPrompt(u.config.OllamaURL, u.config.LLMModel, input)

		finalResponse := dtos.RAGResponse{
			TaskID: id,
			Status: dtos.RAGStatusFinished,
		}

		if err != nil {
			finalResponse.Result = "Error from Ollama: " + err.Error()
		} else {
			finalResponse.Result = res
		}

		// Update task storage with final result
		u.tasks.Store(id, finalResponse)
	}(taskID, text)

	return initialResponse, nil
}

func (u *ragUsecase) GetStatus(taskID string) (dtos.RAGResponse, error) {
	val, ok := u.tasks.Load(taskID)
	if !ok {
		return dtos.RAGResponse{}, fmt.Errorf("task not found: %s", taskID)
	}

	return val.(dtos.RAGResponse), nil
}
