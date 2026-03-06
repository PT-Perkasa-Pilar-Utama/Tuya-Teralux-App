package utils

import (
	"os"
	"testing"
)

func TestValidateEnvDuration(t *testing.T) {
	t.Run("Valid Duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION_VALID", "8h")
		defer os.Unsetenv("TEST_DURATION_VALID")

		val, err := validateEnvDuration("TEST_DURATION_VALID")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if val != "8h" {
			t.Errorf("expected 8h, got %s", val)
		}
	})

	t.Run("Missing Duration", func(t *testing.T) {
		os.Unsetenv("TEST_DURATION_MISSING")

		_, err := validateEnvDuration("TEST_DURATION_MISSING")
		if err == nil {
			t.Error("expected error for missing duration, got nil")
		}
	})

	t.Run("Invalid Duration format", func(t *testing.T) {
		os.Setenv("TEST_DURATION_INVALID", "3600") // No unit
		defer os.Unsetenv("TEST_DURATION_INVALID")

		_, err := validateEnvDuration("TEST_DURATION_INVALID")
		if err == nil {
			t.Error("expected error for invalid duration format, got nil")
		}
	})
}
