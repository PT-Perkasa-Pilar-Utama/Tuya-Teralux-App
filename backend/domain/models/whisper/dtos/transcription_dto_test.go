package dtos

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAsyncTranscriptionResultDTO_SerializationValidFalse ensures that
// transcript_valid=false is explicitly present in JSON output (not omitted).
// This is critical for frontend to reliably detect rejection states.
func TestAsyncTranscriptionResultDTO_SerializationValidFalse(t *testing.T) {
	result := AsyncTranscriptionResultDTO{
		Transcription:             "Test transcript",
		DetectedLanguage:          "id",
		AudioClass:                "silent",
		TranscriptValid:           false, // Explicitly false
		TranscriptRejectionReason: "audio_silent",
		ProviderSkipped:           true,
		ProviderName:              "Orion",
	}

	jsonBytes, err := json.Marshal(result)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)

	// CRITICAL: transcript_valid must be present even when false
	assert.Contains(t, jsonStr, `"transcript_valid":false`, "transcript_valid must be explicitly false in JSON")

	// CRITICAL: provider_skipped must be present even when true
	assert.Contains(t, jsonStr, `"provider_skipped":true`, "provider_skipped must be present in JSON")

	// CRITICAL: audio_class must be present
	assert.Contains(t, jsonStr, `"audio_class":"silent"`, "audio_class must be present in JSON")

	// CRITICAL: transcript_rejection_reason must be present when rejection occurred
	assert.Contains(t, jsonStr, `"transcript_rejection_reason":"audio_silent"`, "transcript_rejection_reason must be present in JSON")
}

// TestAsyncTranscriptionResultDTO_SerializationValidTrue ensures that
// transcript_valid=true is explicitly present in JSON output.
func TestAsyncTranscriptionResultDTO_SerializationValidTrue(t *testing.T) {
	result := AsyncTranscriptionResultDTO{
		Transcription:             "Hello world",
		RefinedText:               "Hello world",
		DetectedLanguage:          "en",
		AudioClass:                "active",
		TranscriptValid:           true,
		TranscriptRejectionReason: "", // Empty when valid
		ProviderSkipped:           false,
		ProviderName:              "Gemini",
	}

	jsonBytes, err := json.Marshal(result)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)

	// CRITICAL: transcript_valid must be present even when true
	assert.Contains(t, jsonStr, `"transcript_valid":true`, "transcript_valid must be explicitly true in JSON")

	// CRITICAL: provider_skipped must be present even when false
	assert.Contains(t, jsonStr, `"provider_skipped":false`, "provider_skipped must be present in JSON")

	// CRITICAL: audio_class must be present
	assert.Contains(t, jsonStr, `"audio_class":"active"`, "audio_class must be present in JSON")

	// Rejection reason should be empty string (present but empty)
	assert.Contains(t, jsonStr, `"transcript_rejection_reason":""`, "transcript_rejection_reason must be present (empty) in JSON")
}

// TestAsyncTranscriptionResultDTO_RejectionContract ensures the complete
// rejection contract is properly serialized for frontend consumption.
func TestAsyncTranscriptionResultDTO_RejectionContract(t *testing.T) {
	testCases := []struct {
		name              string
		audioClass        string
		transcriptValid   bool
		rejectionReason   string
		providerSkipped   bool
		expectedInJSON    []string
		notExpectedInJSON []string
	}{
		{
			name:            "Silent audio rejection",
			audioClass:      "silent",
			transcriptValid: false,
			rejectionReason: "audio_silent",
			providerSkipped: true,
			expectedInJSON: []string{
				`"transcript_valid":false`,
				`"provider_skipped":true`,
				`"audio_class":"silent"`,
				`"transcript_rejection_reason":"audio_silent"`,
			},
		},
		{
			name:            "Near-silent audio with provider call",
			audioClass:      "near_silent",
			transcriptValid: false,
			rejectionReason: "hallucination_detected",
			providerSkipped: false,
			expectedInJSON: []string{
				`"transcript_valid":false`,
				`"provider_skipped":false`,
				`"audio_class":"near_silent"`,
				`"transcript_rejection_reason":"hallucination_detected"`,
			},
		},
		{
			name:            "Valid transcript",
			audioClass:      "active",
			transcriptValid: true,
			rejectionReason: "",
			providerSkipped: false,
			expectedInJSON: []string{
				`"transcript_valid":true`,
				`"provider_skipped":false`,
				`"audio_class":"active"`,
				`"transcript_rejection_reason":""`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AsyncTranscriptionResultDTO{
				Transcription:             "Test",
				DetectedLanguage:          "id",
				AudioClass:                tc.audioClass,
				TranscriptValid:           tc.transcriptValid,
				TranscriptRejectionReason: tc.rejectionReason,
				ProviderSkipped:           tc.providerSkipped,
				ProviderName:              "TestProvider",
			}

			jsonBytes, err := json.Marshal(result)
			assert.NoError(t, err)

			jsonStr := string(jsonBytes)

			// Verify all expected fields are present
			for _, expected := range tc.expectedInJSON {
				assert.Contains(t, jsonStr, expected, "Expected field missing: %s", expected)
			}

			// Verify no unexpected omissions
			for _, notExpected := range tc.notExpectedInJSON {
				assert.NotContains(t, jsonStr, notExpected, "Unexpected field present: %s", notExpected)
			}
		})
	}
}

// TestAsyncTranscriptionResultDTO_BackwardCompatibility ensures that
// existing optional fields still use omitempty for backward compatibility.
func TestAsyncTranscriptionResultDTO_BackwardCompatibility(t *testing.T) {
	result := AsyncTranscriptionResultDTO{
		Transcription:    "Test",
		DetectedLanguage: "id",
		// Leave optional fields empty/zero
		Utterances:        nil,
		Segments:          nil,
		TranscriptFormat:  "",
		ConfidenceSummary: nil,
		// Critical fields set
		AudioClass:      "active",
		TranscriptValid: true,
		ProviderSkipped: false,
	}

	jsonBytes, err := json.Marshal(result)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)

	// Optional fields should be omitted when empty
	assert.False(t, strings.Contains(jsonStr, `"utterances"`), "utterances should be omitted when nil")
	assert.False(t, strings.Contains(jsonStr, `"segments"`), "segments should be omitted when nil")
	assert.False(t, strings.Contains(jsonStr, `"transcript_format"`), "transcript_format should be omitted when empty")
	assert.False(t, strings.Contains(jsonStr, `"confidence_summary"`), "confidence_summary should be omitted when nil")
	assert.False(t, strings.Contains(jsonStr, `"normalization_applied"`), "normalization_applied should be omitted when false")

	// Critical fields must still be present
	assert.Contains(t, jsonStr, `"transcript_valid"`, "transcript_valid must be present")
	assert.Contains(t, jsonStr, `"provider_skipped"`, "provider_skipped must be present")
	assert.Contains(t, jsonStr, `"audio_class"`, "audio_class must be present")
}

// TestAsyncTranscriptionResultDTO_FullSerialization tests a complete realistic
// rejection scenario to ensure all contract fields serialize correctly.
func TestAsyncTranscriptionResultDTO_FullSerialization(t *testing.T) {
	result := AsyncTranscriptionResultDTO{
		Transcription:             "",
		RefinedText:               "",
		DetectedLanguage:          "id",
		Utterances:                nil,
		Segments:                  nil,
		TranscriptFormat:          "",
		ConfidenceSummary:         nil,
		NormalizationApplied:      false,
		AudioClass:                "silent",
		TranscriptValid:           false,
		TranscriptRejectionReason: "audio_silent",
		ProviderSkipped:           true,
		ProviderName:              "Orion",
	}

	jsonBytes, err := json.Marshal(result)
	assert.NoError(t, err)

	// Parse back to verify round-trip
	var parsed AsyncTranscriptionResultDTO
	err = json.Unmarshal(jsonBytes, &parsed)
	assert.NoError(t, err)

	// Verify critical fields survive round-trip
	assert.Equal(t, false, parsed.TranscriptValid, "transcript_valid must survive round-trip")
	assert.Equal(t, true, parsed.ProviderSkipped, "provider_skipped must survive round-trip")
	assert.Equal(t, "silent", parsed.AudioClass, "audio_class must survive round-trip")
	assert.Equal(t, "audio_silent", parsed.TranscriptRejectionReason, "transcript_rejection_reason must survive round-trip")
}
