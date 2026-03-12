package dtos

// StandardResponse represents the standardized API response structure
type StandardResponse struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	// Details is only populated for 400 (Bad Request) and 422 (Unprocessable Entity) errors.
	// For all other status codes, including 500, this field is nil/omitted.
	// @Schema(oneOf=[[]ValidationErrorDetailDTO, string, object])
	Details interface{} `json:"details,omitempty"`
}

// ValidationErrorDetailDTO represents a single field validation error for OpenAPI documentation
type ValidationErrorDetailDTO struct {
	Field   string `json:"field" example:"username"`
	Message string `json:"message" example:"is required"`
}

// TaskIDResponseDTO represents a response containing only a task ID
type TaskIDResponseDTO struct {
	TaskID string `json:"task_id" example:"task-123"`
}
