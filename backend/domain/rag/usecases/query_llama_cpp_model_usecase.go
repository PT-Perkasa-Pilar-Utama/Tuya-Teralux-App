package usecases

import (
	"context"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/skills"
	"time"
)

type QueryLlamaCppModelUseCase interface {
	Query(ctx context.Context, prompt string, trigger string) (*dtos.RAGRawPromptResponseDTO, error)
}

type queryLlamaCppModelUseCase struct {
	llm skills.LLMClient
}

func NewQueryLlamaCppModelUseCase(llm skills.LLMClient) QueryLlamaCppModelUseCase {
	return &queryLlamaCppModelUseCase{
		llm: llm,
	}
}

func (u *queryLlamaCppModelUseCase) Query(ctx context.Context, prompt string, trigger string) (*dtos.RAGRawPromptResponseDTO, error) {
	startTime := time.Now()

	response := &dtos.RAGRawPromptResponseDTO{
		Status:    "pending",
		Trigger:   trigger,
		StartedAt: startTime.Format(time.RFC3339),
	}

	result, err := u.llm.CallModel(ctx, prompt, "low")

	duration := time.Since(startTime).Seconds()
	response.DurationSeconds = duration

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
		response.HTTPStatusCode = utils.GetErrorStatusCode(err)
		return response, err
	}

	if result == "" {
		response.Status = "failed"
		errEmpty := fmt.Errorf("llm returned an empty response")
		response.Error = errEmpty.Error()
		response.HTTPStatusCode = 500
		return response, errEmpty
	}

	response.Status = "completed"
	response.Result = result
	response.HTTPStatusCode = 200

	return response, nil
}
