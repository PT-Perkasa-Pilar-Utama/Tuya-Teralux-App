package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/skills"
	"time"
)

type QueryOpenAIModelUseCase interface {
	Query(prompt string, trigger string) (*dtos.RAGRawPromptResponseDTO, error)
}

type queryOpenAIModelUseCase struct {
	llm skills.LLMClient
}

func NewQueryOpenAIModelUseCase(llm skills.LLMClient) QueryOpenAIModelUseCase {
	return &queryOpenAIModelUseCase{
		llm: llm,
	}
}

func (u *queryOpenAIModelUseCase) Query(prompt string, trigger string) (*dtos.RAGRawPromptResponseDTO, error) {
	startTime := time.Now()

	response := &dtos.RAGRawPromptResponseDTO{
		Status:    "pending",
		Trigger:   trigger,
		StartedAt: startTime.Format(time.RFC3339),
	}

	result, err := u.llm.CallModel(prompt, "low")
	
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
