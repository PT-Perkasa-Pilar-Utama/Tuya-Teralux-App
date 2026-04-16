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

// OrionErrorCode represents structured error codes for Orion transcription failures
type OrionErrorCode string

const (
	OrionErrorCodeModelNotLoaded      OrionErrorCode = "model-not-loaded"
	OrionErrorCodeGPUOOM              OrionErrorCode = "gpu-oom"
	OrionErrorCodeUnsupportedAudioFmt OrionErrorCode = "unsupported-audio-format"
	OrionErrorCodeUpstream500         OrionErrorCode = "upstream-500"
	OrionErrorCodeUnknown             OrionErrorCode = "unknown"
)

// StructuredError represents a terminal failure with structured error information.
// This enables reliable error handling on Android without exposing technical details.
type StructuredError struct {
	ErrorCode string `json:"error_code"`        // Machine-readable code for programmatic handling
	Message   string `json:"message"`           // User-safe message for display
	Details   string `json:"details,omitempty"` // Technical details for logging (not sent to client)
}

// NewStructuredError creates a new StructuredError with the given code, user message, and technical details.
func NewStructuredError(code OrionErrorCode, userMessage string, technicalDetails string) *StructuredError {
	return &StructuredError{
		ErrorCode: string(code),
		Message:   userMessage,
		Details:   technicalDetails,
	}
}

// OrionTranscribeError wraps APIError with structured error information for Orion transcription failures.
// This allows upstream handlers to extract machine-readable error codes without parsing error messages.
type OrionTranscribeError struct {
	APIError
	StructuredError *StructuredError
}

// NewOrionTranscribeError creates a new OrionTranscribeError with the given structured error info.
func NewOrionTranscribeError(statusCode int, structuredErr *StructuredError) *OrionTranscribeError {
	return &OrionTranscribeError{
		APIError: APIError{
			StatusCode: statusCode,
			Message:    structuredErr.Message,
		},
		StructuredError: structuredErr,
	}
}

// GetOrionStructuredError extracts Orion structured error from an error if it is an OrionTranscribeError.
func GetOrionStructuredError(err error) *StructuredError {
	if err == nil {
		return nil
	}
	var orionErr *OrionTranscribeError
	if errors.As(err, &orionErr) {
		return orionErr.StructuredError
	}
	return nil
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

// MapOrionErrorToCode maps an Orion/Outsystems error response to a structured error code.
// It examines the error message and response details to determine the appropriate code.
func MapOrionErrorToCode(statusCode int, responseBody string, errMsg string) *StructuredError {
	technicalDetails := fmt.Sprintf("status=%d body=%s original=%s", statusCode, responseBody, errMsg)

	// Map specific status codes and error patterns to codes
	switch {
	case statusCode == 500:
		// Check for specific 500 error patterns in the response body
		lowerBody := strings.ToLower(responseBody)
		lowerMsg := strings.ToLower(errMsg)

		switch {
		case strings.Contains(lowerBody, "model") && strings.Contains(lowerBody, "load"):
			return NewStructuredError(OrionErrorCodeModelNotLoaded,
				"Transcription service is temporarily unavailable. Please try again.",
				technicalDetails)
		case strings.Contains(lowerBody, "memory") || strings.Contains(lowerBody, "oom") || strings.Contains(lowerBody, "out of memory"):
			return NewStructuredError(OrionErrorCodeGPUOOM,
				"Audio file is too complex for real-time processing. Please try a shorter recording.",
				technicalDetails)
		case strings.Contains(lowerBody, "format") || strings.Contains(lowerBody, "unsupported") || strings.Contains(lowerBody, "codec"):
			return NewStructuredError(OrionErrorCodeUnsupportedAudioFmt,
				"Audio format not supported. Please ensure you're recording in a standard format.",
				technicalDetails)
		case strings.Contains(lowerMsg, "model") && strings.Contains(lowerMsg, "load"):
			return NewStructuredError(OrionErrorCodeModelNotLoaded,
				"Transcription service is temporarily unavailable. Please try again.",
				technicalDetails)
		default:
			return NewStructuredError(OrionErrorCodeUpstream500,
				"Transcription service encountered an error. Please try again.",
				technicalDetails)
		}
	case statusCode == 503:
		return NewStructuredError(OrionErrorCodeModelNotLoaded,
			"Transcription service is starting up. Please try again in a moment.",
			technicalDetails)
	case statusCode >= 400 && statusCode < 500:
		return NewStructuredError(OrionErrorCodeUnsupportedAudioFmt,
			"Invalid audio file. Please check the recording and try again.",
			technicalDetails)
	default:
		return NewStructuredError(OrionErrorCodeUnknown,
			"An unexpected error occurred. Please try again.",
			technicalDetails)
	}
}
