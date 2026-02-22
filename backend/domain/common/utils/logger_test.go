package utils

import (
	"os"
	"testing"
)

func TestUpdateLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"DEBUG Level", "DEBUG", LevelDebug},
		{"INFO Level", "INFO", LevelInfo},
		{"WARN Level", "WARN", LevelWarn},
		{"ERROR Level", "ERROR", LevelError},
		{"Lowercase debug", "debug", LevelDebug},
		{"Invalid Value", "INVALID", LevelInfo}, // Defaults to INFO
		{"Empty Value", "", LevelInfo},          // Defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			original := os.Getenv("LOG_LEVEL")
			defer func() { _ = os.Setenv("LOG_LEVEL", original) }()

			// Save and clear global AppConfig to avoid interference
			originalConfig := AppConfig
			AppConfig = nil
			defer func() { AppConfig = originalConfig }()

			// Set test value
			_ = os.Setenv("LOG_LEVEL", tt.envValue)
			UpdateLogLevel()

			if currentLogLevel != tt.expected {
				t.Errorf("Expected log level %d, got %d", tt.expected, currentLogLevel)
			}
		})
	}
}

func TestShouldLog(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		messageLevel int
		expected     bool
	}{
		{"Debug message with Debug level", LevelDebug, LevelDebug, true},
		{"Info message with Debug level", LevelDebug, LevelInfo, true},
		{"Debug message with Info level", LevelInfo, LevelDebug, false},
		{"Error message with Info level", LevelInfo, LevelError, true},
		{"Warn message with Error level", LevelError, LevelWarn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily set current log level
			oldLevel := currentLogLevel
			currentLogLevel = tt.currentLevel
			defer func() { currentLogLevel = oldLevel }()

			result := shouldLog(tt.messageLevel)
			if result != tt.expected {
				t.Errorf("Expected shouldLog to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLogFunctions(t *testing.T) {
	// Set to DEBUG to capture all logs
	oldLevel := currentLogLevel
	currentLogLevel = LevelDebug
	defer func() { currentLogLevel = oldLevel }()

	// These tests just ensure the functions don't panic
	t.Run("LogDebug", func(t *testing.T) {
		LogDebug("Test debug message: %s", "value")
	})

	t.Run("LogInfo", func(t *testing.T) {
		LogInfo("Test info message: %d", 123)
	})

	t.Run("LogWarn", func(t *testing.T) {
		LogWarn("Test warn message")
	})

	t.Run("LogError", func(t *testing.T) {
		LogError("Test error message: %v", "error")
	})
}

func TestLogMessage(t *testing.T) {
	oldLevel := currentLogLevel
	currentLogLevel = LevelInfo
	defer func() { currentLogLevel = oldLevel }()

	t.Run("Message Below Threshold Not Logged", func(t *testing.T) {
		// Debug message should not be logged when level is INFO
		// This test just ensures no panic occurs
		logMessage(LevelDebug, "This should not appear")
	})

	t.Run("Message Above Threshold Logged", func(t *testing.T) {
		// Error message should be logged when level is INFO
		// This test just ensures no panic occurs
		logMessage(LevelError, "This should appear: %s", "test")
	})
}

func TestLevelNames(t *testing.T) {
	expectedNames := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	if len(levelNames) != len(expectedNames) {
		t.Errorf("Expected %d level names, got %d", len(expectedNames), len(levelNames))
	}

	for i, name := range expectedNames {
		if levelNames[i] != name {
			t.Errorf("Expected level name %s at index %d, got %s", name, i, levelNames[i])
		}
	}
}
