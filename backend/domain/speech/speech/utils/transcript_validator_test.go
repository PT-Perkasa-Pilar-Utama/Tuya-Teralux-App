package utils

import (
	"testing"
)

func TestTranscriptValidator_Normalize(t *testing.T) {
	validator := NewTranscriptValidator().(*transcriptValidator)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim_whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "lowercase",
			input:    "HELLO WORLD",
			expected: "hello world",
		},
		{
			name:     "collapse_whitespace",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "strip_surrounding_punctuation",
			input:    "...hello world!!!",
			expected: "hello world",
		},
		{
			name:     "mixed_case_with_punctuation",
			input:    "  TERIMA KASIH...  ",
			expected: "terima kasih",
		},
		{
			name:     "empty_after_normalization",
			input:    "   ...   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTranscriptValidator_Validate_Empty(t *testing.T) {
	validator := NewTranscriptValidator()

	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
		wantReason string
	}{
		{
			name:       "empty_string",
			transcript: "",
			audioClass: "active",
			wantValid:  false,
			wantReason: "empty_after_normalization",
		},
		{
			name:       "whitespace_only",
			transcript: "   ",
			audioClass: "active",
			wantValid:  false,
			wantReason: "empty_after_normalization",
		},
		{
			name:       "punctuation_only",
			transcript: "...!!!   ",
			audioClass: "active",
			wantValid:  false,
			wantReason: "empty_after_normalization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.transcript, tt.audioClass)
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
			if result.RejectionReason != tt.wantReason {
				t.Errorf("RejectionReason = %q, want %q", result.RejectionReason, tt.wantReason)
			}
		})
	}
}

func TestTranscriptValidator_Validate_HallucinationPhrases(t *testing.T) {
	validator := NewTranscriptValidator()

	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
		wantReason string
	}{
		// Indonesian gratitude
		{
			name:       "terima_kasih",
			transcript: "terima kasih",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		{
			name:       "terima_kasih_sudah_menonton",
			transcript: "terima kasih sudah menonton",
			audioClass: "active",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		// English gratitude
		{
			name:       "thank_you",
			transcript: "thank you",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		{
			name:       "thanks_for_watching",
			transcript: "thanks for watching",
			audioClass: "active",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		// Subscription prompts
		{
			name:       "jangan_lupa_subscribe",
			transcript: "jangan lupa subscribe",
			audioClass: "active",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		{
			name:       "like_comment_share",
			transcript: "like comment share",
			audioClass: "active",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		// Credit patterns
		{
			name:       "translated_by",
			transcript: "translated by",
			audioClass: "active",
			wantValid:  false,
			wantReason: "subtitle_credit_pattern", // Matches regex pattern
		},
		// Closing phrases
		{
			name:       "selamat_menonton",
			transcript: "selamat menonton",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
		{
			name:       "see_you_next_time",
			transcript: "see you next time",
			audioClass: "active",
			wantValid:  false,
			wantReason: "known_hallucination_phrase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.transcript, tt.audioClass)
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
			if result.RejectionReason != tt.wantReason {
				t.Errorf("RejectionReason = %q, want %q", result.RejectionReason, tt.wantReason)
			}
		})
	}
}

func TestTranscriptValidator_Validate_CommandPreservation(t *testing.T) {
	validator := NewTranscriptValidator()

	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
	}{
		{
			name:       "nyalakan_lampu",
			transcript: "nyalakan lampu",
			audioClass: "active",
			wantValid:  true,
		},
		{
			name:       "matikan_ac",
			transcript: "matikan ac",
			audioClass: "active",
			wantValid:  true,
		},
		{
			name:       "turn_on_light",
			transcript: "turn on the light",
			audioClass: "active",
			wantValid:  true,
		},
		{
			name:       "berapa_suhu",
			transcript: "berapa suhu ruangan",
			audioClass: "active",
			wantValid:  true,
		},
		{
			name:       "tolong_nyalakan",
			transcript: "tolong nyalakan lampu kamar",
			audioClass: "active",
			wantValid:  true,
		},
		// Even with hallucination phrase, command-like should be preserved
		{
			name:       "command_with_thanks",
			transcript: "terima kasih nyalakan lampu",
			audioClass: "active",
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.transcript, tt.audioClass)
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
			if !result.IsCommandLike {
				t.Errorf("IsCommandLike = false, want true for command-like transcript")
			}
		})
	}
}

func TestTranscriptValidator_Validate_CreditPatterns(t *testing.T) {
	validator := NewTranscriptValidator()

	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
		wantReason string
	}{
		{
			name:       "subscribe_count",
			transcript: "subscribe 1000 subscribers",
			audioClass: "active",
			wantValid:  false,
			wantReason: "subtitle_credit_pattern",
		},
		{
			name:       "youtube_analytics",
			transcript: "check your youtube analytics",
			audioClass: "active",
			wantValid:  false,
			wantReason: "subtitle_credit_pattern",
		},
		{
			name:       "monetization",
			transcript: "youtube studio monetization",
			audioClass: "active",
			wantValid:  false,
			wantReason: "subtitle_credit_pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.transcript, tt.audioClass)
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
			if result.RejectionReason != tt.wantReason {
				t.Errorf("RejectionReason = %q, want %q", result.RejectionReason, tt.wantReason)
			}
		})
	}
}

func TestTranscriptValidator_Validate_LowSignalGeneric(t *testing.T) {
	validator := NewTranscriptValidator()

	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
		wantReason string
	}{
		{
			name:       "near_silent_halo",
			transcript: "halo",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase", // "halo" is in hallucination list
		},
		{
			name:       "near_silent_oke",
			transcript: "oke",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase", // "oke" is in hallucination list
		},
		{
			name:       "near_silent_short_phrase",
			transcript: "ya tidak",
			audioClass: "near_silent",
			wantValid:  false,
			wantReason: "known_hallucination_phrase", // "ya" and "tidak" are in hallucination list
		},
		// Active audio should allow short phrases that aren't in hallucination list
		{
			name:       "active_short_phrase",
			transcript: "test",
			audioClass: "active",
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.transcript, tt.audioClass)
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
			if result.RejectionReason != tt.wantReason {
				t.Errorf("RejectionReason = %q, want %q", result.RejectionReason, tt.wantReason)
			}
		})
	}
}

func TestTranscriptValidator_IsValidTranscript(t *testing.T) {
	tests := []struct {
		name       string
		transcript string
		audioClass string
		wantValid  bool
	}{
		{"valid_command", "nyalakan lampu kamar", "active", true},
		{"invalid_hallucination", "terima kasih", "near_silent", false},
		{"invalid_empty", "", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTranscript(tt.transcript, tt.audioClass)
			if got != tt.wantValid {
				t.Errorf("IsValidTranscript(%q, %q) = %v, want %v", tt.transcript, tt.audioClass, got, tt.wantValid)
			}
		})
	}
}
