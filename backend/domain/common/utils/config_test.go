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

func TestGetByteConfigWithMBFallback(t *testing.T) {
	t.Run("Byte-based env var takes precedence", func(t *testing.T) {
		os.Setenv("TEST_BYTE_VAR", "2097152") // 2 MB in bytes
		os.Setenv("TEST_MB_VAR", "8")         // 8 MB
		defer os.Unsetenv("TEST_BYTE_VAR")
		defer os.Unsetenv("TEST_MB_VAR")

		result := getByteConfigWithMBFallback("TEST_BYTE_VAR", "TEST_MB_VAR", 1024*1024)
		if result != 2097152 {
			t.Errorf("expected 2097152 (2 MB in bytes), got %d", result)
		}
	})

	t.Run("MB fallback when byte var not set", func(t *testing.T) {
		os.Unsetenv("TEST_BYTE_VAR")
		os.Setenv("TEST_MB_VAR", "4") // 4 MB
		defer os.Unsetenv("TEST_MB_VAR")

		result := getByteConfigWithMBFallback("TEST_BYTE_VAR", "TEST_MB_VAR", 1024*1024)
		expected := int64(4 * 1024 * 1024)
		if result != expected {
			t.Errorf("expected %d (4 MB in bytes), got %d", expected, result)
		}
	})

	t.Run("Default when neither env var set", func(t *testing.T) {
		os.Unsetenv("TEST_BYTE_VAR")
		os.Unsetenv("TEST_MB_VAR")

		defaultValue := int64(256 * 1024) // 256 KB
		result := getByteConfigWithMBFallback("TEST_BYTE_VAR", "TEST_MB_VAR", defaultValue)
		if result != defaultValue {
			t.Errorf("expected default %d, got %d", defaultValue, result)
		}
	})

	t.Run("Invalid byte var falls back to MB", func(t *testing.T) {
		os.Setenv("TEST_BYTE_VAR", "invalid")
		os.Setenv("TEST_MB_VAR", "16") // 16 MB
		defer os.Unsetenv("TEST_BYTE_VAR")
		defer os.Unsetenv("TEST_MB_VAR")

		result := getByteConfigWithMBFallback("TEST_BYTE_VAR", "TEST_MB_VAR", 1024*1024)
		expected := int64(16 * 1024 * 1024)
		if result != expected {
			t.Errorf("expected %d (16 MB in bytes), got %d", expected, result)
		}
	})

	t.Run("Invalid MB var falls back to default", func(t *testing.T) {
		os.Setenv("TEST_BYTE_VAR", "invalid")
		os.Setenv("TEST_MB_VAR", "invalid")
		defer os.Unsetenv("TEST_BYTE_VAR")
		defer os.Unsetenv("TEST_MB_VAR")

		defaultValue := int64(512 * 1024) // 512 KB
		result := getByteConfigWithMBFallback("TEST_BYTE_VAR", "TEST_MB_VAR", defaultValue)
		if result != defaultValue {
			t.Errorf("expected default %d, got %d", defaultValue, result)
		}
	})
}
