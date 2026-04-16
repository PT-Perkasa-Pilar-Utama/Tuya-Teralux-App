package utils

import (
	"testing"
)

func TestMapOrionErrorToCode_ModelNotLoaded(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		errMsg     string
		wantCode   OrionErrorCode
	}{
		{
			name:       "500 with model load error in body",
			statusCode: 500,
			body:       `{"error": "failed to load model"}`,
			errMsg:     "Orion server returned error status 500: {\"error\": \"failed to load model\"}",
			wantCode:   OrionErrorCodeModelNotLoaded,
		},
		{
			name:       "500 with model load error in message",
			statusCode: 500,
			body:       `{"error": "something else"}`,
			errMsg:     "Orion server returned error status 500: model load failed",
			wantCode:   OrionErrorCodeModelNotLoaded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOrionErrorToCode(tt.statusCode, tt.body, tt.errMsg)
			if result == nil {
				t.Fatal("MapOrionErrorToCode returned nil")
			}
			if result.ErrorCode != string(tt.wantCode) {
				t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, tt.wantCode)
			}
		})
	}
}

func TestMapOrionErrorToCode_GPUOOM(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		errMsg     string
	}{
		{
			name:       "500 with memory error",
			statusCode: 500,
			body:       `{"error": "out of memory"}`,
			errMsg:     "Orion server returned error status 500",
		},
		{
			name:       "500 with oom error",
			statusCode: 500,
			body:       `{"error": "gpu oom"}`,
			errMsg:     "Orion server returned error status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOrionErrorToCode(tt.statusCode, tt.body, tt.errMsg)
			if result == nil {
				t.Fatal("MapOrionErrorToCode returned nil")
			}
			if result.ErrorCode != string(OrionErrorCodeGPUOOM) {
				t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, OrionErrorCodeGPUOOM)
			}
		})
	}
}

func TestMapOrionErrorToCode_UnsupportedAudioFormat(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		errMsg     string
	}{
		{
			name:       "500 with format error",
			statusCode: 500,
			body:       `{"error": "unsupported audio format"}`,
			errMsg:     "Orion server returned error status 500",
		},
		{
			name:       "500 with codec error",
			statusCode: 500,
			body:       `{"error": "codec not supported"}`,
			errMsg:     "Orion server returned error status 500",
		},
		{
			name:       "400 client error",
			statusCode: 400,
			body:       `{"error": "bad request"}`,
			errMsg:     "Orion server returned error status 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOrionErrorToCode(tt.statusCode, tt.body, tt.errMsg)
			if result == nil {
				t.Fatal("MapOrionErrorToCode returned nil")
			}
			if result.ErrorCode != string(OrionErrorCodeUnsupportedAudioFmt) {
				t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, OrionErrorCodeUnsupportedAudioFmt)
			}
		})
	}
}

func TestMapOrionErrorToCode_Upstream500(t *testing.T) {
	result := MapOrionErrorToCode(500, `{"error": "internal server error"}`, "Orion server returned error status 500")
	if result == nil {
		t.Fatal("MapOrionErrorToCode returned nil")
	}
	if result.ErrorCode != string(OrionErrorCodeUpstream500) {
		t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, OrionErrorCodeUpstream500)
	}
}

func TestMapOrionErrorToCode_503(t *testing.T) {
	result := MapOrionErrorToCode(503, `{"error": "service unavailable"}`, "Orion server returned error status 503")
	if result == nil {
		t.Fatal("MapOrionErrorToCode returned nil")
	}
	if result.ErrorCode != string(OrionErrorCodeModelNotLoaded) {
		t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, OrionErrorCodeModelNotLoaded)
	}
}

func TestMapOrionErrorToCode_Unknown(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   OrionErrorCode
	}{
		{name: "404", statusCode: 404, wantCode: OrionErrorCodeUnsupportedAudioFmt},
		{name: "401", statusCode: 401, wantCode: OrionErrorCodeUnsupportedAudioFmt},
		{name: "418", statusCode: 418, wantCode: OrionErrorCodeUnsupportedAudioFmt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOrionErrorToCode(tt.statusCode, `{}`, "some error")
			if result == nil {
				t.Fatal("MapOrionErrorToCode returned nil")
			}
			if result.ErrorCode != string(tt.wantCode) {
				t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, tt.wantCode)
			}
		})
	}
}

func TestNewOrionTranscribeError(t *testing.T) {
	structuredErr := NewStructuredError(OrionErrorCodeUpstream500, "Transcription failed", "status=500 body=error")
	err := NewOrionTranscribeError(500, structuredErr)

	if err.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", err.StatusCode)
	}
	if err.Message != "Transcription failed" {
		t.Errorf("Message = %v, want 'Transcription failed'", err.Message)
	}
	if err.StructuredError == nil {
		t.Fatal("StructuredError is nil")
	}
	if err.StructuredError.ErrorCode != string(OrionErrorCodeUpstream500) {
		t.Errorf("StructuredError.ErrorCode = %v, want %v", err.StructuredError.ErrorCode, OrionErrorCodeUpstream500)
	}
}

func TestGetOrionStructuredError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		result := GetOrionStructuredError(nil)
		if result != nil {
			t.Errorf("GetOrionStructuredError(nil) = %v, want nil", result)
		}
	})

	t.Run("non-Orion error", func(t *testing.T) {
		err := NewAPIError(500, "some error")
		result := GetOrionStructuredError(err)
		if result != nil {
			t.Errorf("GetOrionStructuredError(APIError) = %v, want nil", result)
		}
	})

	t.Run("Orion error", func(t *testing.T) {
		structuredErr := NewStructuredError(OrionErrorCodeGPUOOM, "Out of memory", "details")
		err := NewOrionTranscribeError(500, structuredErr)
		result := GetOrionStructuredError(err)
		if result == nil {
			t.Fatal("GetOrionStructuredError(OrionTranscribeError) = nil, want non-nil")
		}
		if result.ErrorCode != string(OrionErrorCodeGPUOOM) {
			t.Errorf("ErrorCode = %v, want %v", result.ErrorCode, OrionErrorCodeGPUOOM)
		}
	})
}

func TestNewStructuredError(t *testing.T) {
	err := NewStructuredError(OrionErrorCodeModelNotLoaded, "Service unavailable", "technical details")

	if err.ErrorCode != string(OrionErrorCodeModelNotLoaded) {
		t.Errorf("ErrorCode = %v, want %v", err.ErrorCode, OrionErrorCodeModelNotLoaded)
	}
	if err.Message != "Service unavailable" {
		t.Errorf("Message = %v, want 'Service unavailable'", err.Message)
	}
	if err.Details != "technical details" {
		t.Errorf("Details = %v, want 'technical details'", err.Details)
	}
}

func TestStructuredError_UserSafeMessage(t *testing.T) {
	structuredErr := NewStructuredError(
		OrionErrorCodeUpstream500,
		"Transcription service encountered an error. Please try again.",
		"status=500 body=Internal Server Error at Orion transcription layer",
	)

	if structuredErr.Message == "" {
		t.Error("Message should not be empty for user display")
	}

	if len(structuredErr.Details) > 0 {
		t.Log("Details field properly populated for logging")
	}
}
