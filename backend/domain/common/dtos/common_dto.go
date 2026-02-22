package dtos

// StandardResponse represents the standardized API response structure
type StandardResponse struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	// Details is only populated for 400 (Bad Request) and 422 (Unprocessable Entity) errors.
	// For all other status codes, including 500, this field is nil/omitted.
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponseDTO is a simple DTO for operations returning a success boolean
type SuccessResponseDTO struct {
	Success bool `json:"success"`
}
