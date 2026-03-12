package dtos

// V1PipelineStatusResponseDTO mirrors services.PipelineResponse for OpenAPI documentation
type V1PipelineStatusResponseDTO struct {
	TaskID      string `json:"task_id" example:"pipeline_task_123"`
	Status      string `json:"status" example:"completed"`
	Transcript  string `json:"transcript,omitempty" example:"Hello world"`
	RefinedText string `json:"refined_text,omitempty" example:"Hello, world."`
	Translated  string `json:"translated,omitempty" example:"Halo dunia"`
	Summary     string `json:"summary,omitempty" example:"A greeting to the world."`
	Error       string `json:"error,omitempty"`
}
