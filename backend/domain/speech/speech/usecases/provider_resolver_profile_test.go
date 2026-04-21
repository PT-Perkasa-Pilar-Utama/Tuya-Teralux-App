package usecases

import (
	"fmt"
	"sensio/domain/common/utils"
	"testing"
)

// TestNormalizeEngineProfile verifies case and whitespace normalization
func TestNormalizeEngineProfile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Premium", "premium"},
		{"  STANDARD  ", "standard"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeEngineProfile(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeEngineProfile(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

// TestIsValidEngineProfile verifies the vocabulary lookup
func TestIsValidEngineProfile(t *testing.T) {
	valid := []string{"premium", "standard"}
	for _, p := range valid {
		if !IsValidEngineProfile(p) {
			t.Errorf("IsValidEngineProfile(%q) should be true", p)
		}
	}
	invalid := []string{"openai", "gemini", "", "turbo", "ultra", "plaud", "fast"}
	for _, p := range invalid {
		if IsValidEngineProfile(p) {
			t.Errorf("IsValidEngineProfile(%q) should be false", p)
		}
	}
}

// newTerminalWithProfile is a helper to build a Terminal with both fields
func newTerminalWithProfile(profile, provider *string) *Terminal {
	return &Terminal{
		AiEngineProfile: profile,
		AiProvider:      provider,
	}
}

func strPtr(s string) *string { return &s }

// TestResolveFromTerminal_PremiumProfile verifies that premium profile returns profile_premium selection mode
func TestResolveFromTerminal_PremiumProfile(t *testing.T) {
	cfg := makeMinimalConfig()
	r := newMinimalResolver(cfg)

	terminal := newTerminalWithProfile(strPtr("premium"), nil)
	result, err := r.resolveFromTerminal(terminal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SelectionMode != "profile_premium" {
		t.Errorf("expected SelectionMode=profile_premium, got %s", result.SelectionMode)
	}
	if !result.IsExplicit {
		t.Error("expected IsExplicit=true for premium profile")
	}
}

// TestResolveFromTerminal_StandardProfile verifies that standard profile returns profile_standard selection mode
func TestResolveFromTerminal_StandardProfile(t *testing.T) {
	cfg := makeMinimalConfig()
	r := newMinimalResolver(cfg)

	terminal := newTerminalWithProfile(strPtr("standard"), nil)
	result, err := r.resolveFromTerminal(terminal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SelectionMode != "profile_standard" {
		t.Errorf("expected SelectionMode=profile_standard, got %s", result.SelectionMode)
	}
	if !result.IsExplicit {
		t.Error("expected IsExplicit=true for standard profile")
	}
}

// TestResolveFromTerminal_ProfileWinsOverProvider verifies profile takes precedence over ai_provider
func TestResolveFromTerminal_ProfileWinsOverProvider(t *testing.T) {
	cfg := makeMinimalConfig()
	r := newMinimalResolver(cfg)

	// premium profile + ai_provider=orion → premium must win
	terminal := newTerminalWithProfile(strPtr("premium"), strPtr("orion"))
	result, err := r.resolveFromTerminal(terminal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SelectionMode != "profile_premium" {
		t.Errorf("expected profile_premium to win over ai_provider; got SelectionMode=%s", result.SelectionMode)
	}
}

// TestResolveFromTerminal_LegacyProviderPreserved verifies that explicit ai_provider still works when no profile is set
func TestResolveFromTerminal_LegacyProviderPreserved(t *testing.T) {
	cfg := makeMinimalConfig()
	r := newMinimalResolver(cfg)

	terminal := newTerminalWithProfile(nil, strPtr("openai"))
	result, err := r.resolveFromTerminal(terminal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SelectionMode != "provider_explicit" {
		t.Errorf("expected provider_explicit for legacy ai_provider; got SelectionMode=%s", result.SelectionMode)
	}
	if result.ProviderName != "openai" {
		t.Errorf("expected provider openai, got %s", result.ProviderName)
	}
}

// TestResolveFromTerminal_BothNullUsesDefault verifies default path when both fields are nil
func TestResolveFromTerminal_BothNullUsesDefault(t *testing.T) {
	cfg := makeMinimalConfig()
	r := newMinimalResolver(cfg)

	terminal := &Terminal{}
	result, err := r.resolveFromTerminal(terminal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SelectionMode != "default" {
		t.Errorf("expected default SelectionMode, got %s", result.SelectionMode)
	}
}

// makeMinimalConfig builds a Config with openai so the default resolver can function
func makeMinimalConfig() *utils.Config {
	return &utils.Config{
		LLMProvider:  "openai",
		OpenAIApiKey: "test-key",
	}
}

// newMinimalResolver creates a resolver with just an OpenAI service for testing resolver logic
func newMinimalResolver(cfg *utils.Config) *providerResolverImpl {
	geminiSvc, openaiSvc, groqSvc, orionSvc := GetProviderServices(cfg)
	repo := &stubTerminalRepo{}
	r := NewProviderResolver(cfg, geminiSvc, openaiSvc, groqSvc, orionSvc, repo).(*providerResolverImpl)
	return r
}

// stubTerminalRepo is a no-op terminal repository for resolver unit tests
type stubTerminalRepo struct{}

func (s *stubTerminalRepo) GetByID(id string) (*Terminal, error) {
	return nil, fmt.Errorf("not found")
}

func (s *stubTerminalRepo) GetByMacAddress(macAddress string) (*Terminal, error) {
	return nil, fmt.Errorf("not found")
}

// TestExecuteWithCandidateFallback_FiltersUnconfigured verifies that only configured providers are tried
func TestExecuteWithCandidateFallback_FiltersUnconfigured(t *testing.T) {
	cfg := &utils.Config{
		OpenAIApiKey: "present",
		// GroqApiKey is missing
	}
	r := newMinimalResolver(cfg)

	tried := make(map[string]bool)
	executable := func(s *ResolvedProviderSet) error {
		tried[s.ProviderName] = true
		return nil
	}

	err := r.ExecuteWithCandidateFallback([]string{"openai", "groq"}, executable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tried["openai"] {
		t.Error("expected openai to be tried")
	}
	if tried["groq"] {
		t.Error("expected groq to be skipped (unconfigured)")
	}
}

// TestExecuteWithCandidateFallback_FailsIfNoneConfigured verifies it fails if no candidates have keys
func TestExecuteWithCandidateFallback_FailsIfNoneConfigured(t *testing.T) {
	cfg := &utils.Config{} // no keys
	r := newMinimalResolver(cfg)

	err := r.ExecuteWithCandidateFallback([]string{"openai", "groq"}, func(s *ResolvedProviderSet) error { return nil })
	if err == nil {
		t.Error("expected error when no candidates are configured")
	}
}

// TestExecuteWithFallbackByTerminal_PremiumProfile verifies that premium profile uses premium candidates
func TestExecuteWithFallbackByTerminal_PremiumProfile(t *testing.T) {
	cfg := &utils.Config{OpenAIApiKey: "present"}
	r := newMinimalResolver(cfg)

	// Mock repo that returns a terminal with premium profile
	profile := "premium"
	repo := &stubTerminalRepoWithData{terminal: &Terminal{AiEngineProfile: &profile}}
	r.terminalRepo = repo

	tried := make(map[string]bool)
	executable := func(s *ResolvedProviderSet) error {
		tried[s.ProviderName] = true
		return nil
	}

	err := r.ExecuteWithFallbackByTerminal("t1", executable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tried["openai"] {
		t.Error("expected openai (premium candidate) to be tried")
	}
}

type stubTerminalRepoWithData struct {
	terminal *Terminal
}

func (s *stubTerminalRepoWithData) GetByID(id string) (*Terminal, error) {
	return s.terminal, nil
}
func (s *stubTerminalRepoWithData) GetByMacAddress(macAddress string) (*Terminal, error) {
	return s.terminal, nil
}
