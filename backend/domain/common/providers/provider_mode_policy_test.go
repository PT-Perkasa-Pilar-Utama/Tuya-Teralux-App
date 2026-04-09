package providers

import (
	"errors"
	"testing"
)

func TestProviderModePolicy_String(t *testing.T) {
	if DefaultMode.String() != "default" {
		t.Errorf("expected 'default', got %q", DefaultMode.String())
	}
	if ExplicitMode.String() != "explicit" {
		t.Errorf("expected 'explicit', got %q", ExplicitMode.String())
	}
}

func TestNewDefaultModePolicy_AllowsFallback(t *testing.T) {
	policy := NewDefaultModePolicy()

	if policy.Mode != DefaultMode {
		t.Errorf("expected mode DefaultMode, got %v", policy.Mode)
	}
	if policy.ProviderName != "" {
		t.Errorf("expected empty provider name for default mode, got %q", policy.ProviderName)
	}
	if !policy.ShouldFallback() {
		t.Error("expected default mode to allow fallback")
	}
}

func TestNewExplicitModePolicy_NoFallback(t *testing.T) {
	policy := NewExplicitModePolicy("gemini")

	if policy.Mode != ExplicitMode {
		t.Errorf("expected mode ExplicitMode, got %v", policy.Mode)
	}
	if policy.ProviderName != "gemini" {
		t.Errorf("expected provider name 'gemini', got %q", policy.ProviderName)
	}
	if policy.ShouldFallback() {
		t.Error("expected explicit mode to disallow fallback")
	}
}

func TestProviderModePolicy_FormatError_ExplicitMode(t *testing.T) {
	policy := NewExplicitModePolicy("gemini")
	err := errors.New("connection timeout")

	msg := policy.FormatError(err, []string{"gemini", "openai"})

	// Should mention the provider name and explicit mode
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	if !containsStr(msg, "Explicit Mode") {
		t.Errorf("expected 'Explicit Mode' in message, got %q", msg)
	}
	if !containsStr(msg, "gemini") {
		t.Errorf("expected provider name 'gemini' in message, got %q", msg)
	}
	if !containsStr(msg, "No fallback") {
		t.Errorf("expected 'No fallback' in message, got %q", msg)
	}
}

func TestProviderModePolicy_FormatError_DefaultMode(t *testing.T) {
	policy := NewDefaultModePolicy()
	err := errors.New("all providers down")

	msg := policy.FormatError(err, []string{"gemini", "openai"})

	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	if !containsStr(msg, "Default Mode") {
		t.Errorf("expected 'Default Mode' in message, got %q", msg)
	}
	if !containsStr(msg, "2 provider") {
		t.Errorf("expected provider count '2' in message, got %q", msg)
	}
}

func TestProviderModePolicy_FormatError_DefaultMode_EmptyAttempted(t *testing.T) {
	policy := NewDefaultModePolicy()
	err := errors.New("no candidates available")

	msg := policy.FormatError(err, nil)

	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	if !containsStr(msg, "Default Mode") {
		t.Errorf("expected 'Default Mode' in message, got %q", msg)
	}
	if !containsStr(msg, "no candidates") {
		t.Errorf("expected original error message, got %q", msg)
	}
}

func TestResolvedProviderSet_IsExplicit(t *testing.T) {
	// Test that IsExplicit field exists and works
	explicitSet := &ResolvedProviderSet{
		ProviderName: "gemini",
		IsExplicit:   true,
	}
	if !explicitSet.IsExplicit {
		t.Error("expected IsExplicit to be true")
	}

	defaultSet := &ResolvedProviderSet{
		ProviderName: "openai",
		IsExplicit:   false,
	}
	if defaultSet.IsExplicit {
		t.Error("expected IsExplicit to be false")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
