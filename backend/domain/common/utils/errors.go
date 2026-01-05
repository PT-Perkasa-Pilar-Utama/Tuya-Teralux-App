package utils

import "fmt"

// ValidationErrorDetail represents a single field validation error
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError represents an error containing multiple validation details
type ValidationError struct {
	Message string                  `json:"message"`
	Details []ValidationErrorDetail `json:"details"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Details)
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string, details []ValidationErrorDetail) *ValidationError {
	return &ValidationError{
		Message: message,
		Details: details,
	}
}
