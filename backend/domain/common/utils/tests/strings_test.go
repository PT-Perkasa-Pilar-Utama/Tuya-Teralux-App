package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"teralux_app/domain/common/utils"
)

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		sep      string
		expected string
	}{
		{
			name:     "Join valid strings",
			input:    []string{"hello", "world"},
			sep:      ", ",
			expected: "hello, world",
		},
		{
			name:     "Join empty slice",
			input:    []string{},
			sep:      ", ",
			expected: "",
		},
		{
			name:     "Join single element",
			input:    []string{"foo"},
			sep:      "-",
			expected: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.JoinStrings(tt.input, tt.sep)
			if result != tt.expected {
				t.Errorf("JoinStrings(%v, %q) = %q; want %q", tt.input, tt.sep, result, tt.expected)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	input := "test_string"

	// manually calculate hash
	h := sha256.New()
	h.Write([]byte(input))
	expected := hex.EncodeToString(h.Sum(nil))

	result := utils.HashString(input)

	if result != expected {
		t.Errorf("HashString(%q) = %q; want %q", input, result, expected)
	}

	// Verify consistentcy
	result2 := utils.HashString(input)
	if result != result2 {
		t.Error("HashString is not deterministic")
	}
}

// Additional strings tests can be added here
func TestToSnakeCase(t *testing.T) {
	// Assuming ToSnakeCase exists in utils/strings.go ?
	// Checking file content from Step 35: Only JoinStrings and HashString were shown.
	// So I won't add TestToSnakeCase unless I see it.
}
