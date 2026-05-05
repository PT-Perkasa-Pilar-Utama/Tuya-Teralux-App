package usecases

import (
	"testing"
)

func TestMergeWithDedup(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected string
	}{
		{
			name:     "No overlap",
			text1:    "Hello world.",
			text2:    "How are you?",
			expected: "Hello world. How are you?",
		},
		{
			name:     "One word overlap",
			text1:    "This is a test",
			text2:    "test of the system",
			expected: "This is a test of the system",
		},
		{
			name:     "Multiple words overlap",
			text1:    "We are going to the market",
			text2:    "to the market right now",
			expected: "We are going to the market right now",
		},
		{
			name:     "Exact match",
			text1:    "Exact match segment",
			text2:    "Exact match segment",
			expected: "Exact match segment",
		},
		{
			name:     "Case insensitive overlap",
			text1:    "Case Insensitive Overlap",
			text2:    "insensitive OVERLAP here",
			expected: "Case Insensitive Overlap here",
		},
		{
			name:     "Punctuation handling",
			text1:    "Wait, what?",
			text2:    "what? Yes.",
			expected: "Wait, what? Yes.",
		},
		{
			name:     "Overlap shorter than text",
			text1:    "Hello world",
			text2:    "world world",
			expected: "Hello world world",
		},
		{
			name:     "Empty second segment",
			text1:    "Hello world",
			text2:    "",
			expected: "Hello world",
		},
	}

	uc := &transcribeUseCase{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.mergeWithDedup(tt.text1, tt.text2)
			if result != tt.expected {
				t.Errorf("mergeWithDedup() = %q, want %q", result, tt.expected)
			}
		})
	}
}
