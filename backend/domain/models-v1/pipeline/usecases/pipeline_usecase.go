package usecases

import (
	"context"
	"fmt"
	"sensio/domain/models-v1/pipeline/services"
)

// PipelineUseCase handles pipeline business logic.
type PipelineUseCase struct {
	pythonService *services.PythonPipelineService
}

// NewPipelineUseCase creates a new PipelineUseCase.
func NewPipelineUseCase(pythonService *services.PythonPipelineService) *PipelineUseCase {
	return &PipelineUseCase{
		pythonService: pythonService,
	}
}

// ExecutePipeline sends audio for pipeline processing to Python service.
func (uc *PipelineUseCase) ExecutePipeline(ctx context.Context, req services.PipelineRequest) (string, error) {
	resp, err := uc.pythonService.ExecuteJob(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute pipeline: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("pipeline error: %s", resp.Error)
	}

	return resp.TaskID, nil
}

// GetPipelineStatus gets the status of a pipeline task.
func (uc *PipelineUseCase) GetPipelineStatus(ctx context.Context, taskID string) (*services.PipelineResponse, error) {
	resp, err := uc.pythonService.GetStatus(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline status: %w", err)
	}

	return resp, nil
}
