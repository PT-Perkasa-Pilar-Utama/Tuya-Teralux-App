package dtos

// StandardResponse represents the standardized API response structure
type StandardResponse struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorResponse represents a standardized error response for non-validation errors (e.g. 401, 404, 500)
type ErrorResponse struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"An error occurred"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationErrorResponse represents a standardized error response for validation errors (400, 422)
type ValidationErrorResponse struct {
	Status  bool                       `json:"status" example:"false"`
	Message string                     `json:"message" example:"Validation Error"`
	Data    interface{}                `json:"data,omitempty"`
	Details []ValidationErrorDetailDTO `json:"details"`
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
