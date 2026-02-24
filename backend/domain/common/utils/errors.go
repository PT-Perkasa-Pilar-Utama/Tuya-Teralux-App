package utils

import (
	"errors"
	"fmt"
	"strings"
)

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
// APIError represents an error from an external API or internal service with a status code
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// GetErrorStatusCode extracts status code from error if it's an APIError, defaults to 500
func GetErrorStatusCode(err error) int {
	if err == nil {
		return 200
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}

	// Fallback for wrapped errors or string matches
	msg := err.Error()
	if strings.Contains(msg, "status 503") || strings.Contains(msg, "503 Service Unavailable") {
		return 503
	}
	if strings.Contains(msg, "status 401") || strings.Contains(msg, "401 Unauthorized") {
		return 401
	}
	if strings.Contains(msg, "status 429") || strings.Contains(msg, "429 Too Many Requests") {
		return 429
	}
	if strings.Contains(msg, "status 400") || strings.Contains(msg, "400 Bad Request") {
		return 400
	}

	return 500
}
