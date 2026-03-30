package providers

import (
	"sensio/domain/common/utils"
	"testing"
	"time"
)

func TestHealthAwareResolver_GetRemoteCandidates(t *testing.T) {
	tests := []struct {
		name          string
		config        *utils.Config
		setupStats    func(*healthAwareResolverImpl)
		expectedLen   int
		expectedFirst string
		validateOrder func([]string) bool
	}{
		{
			name: "no providers configured",
			config: &utils.Config{
				LLMProvider: "",
			},
			expectedLen: 0,
		},
		{
			name: "single provider configured",
			config: &utils.Config{
				LLMProvider:  "gemini",
				GeminiApiKey: "test-key",
			},
			expectedLen:   1,
			expectedFirst: "gemini",
		},
		{
			name: "multiple providers configured",
			config: &utils.Config{
				LLMProvider:  "gemini",
				GeminiApiKey: "test-key",
				OpenAIApiKey: "test-key",
				GroqApiKey:   "test-key",
				OrionApiKey:  "test-key",
			},
			expectedLen:   4,
			expectedFirst: "gemini", // LLM_PROVIDER preference
		},
		{
			name: "provider in cooldown should be filtered out",
			config: &utils.Config{
				LLMProvider:  "gemini",
				GeminiApiKey: "test-key",
				OpenAIApiKey: "test-key",
			},
			setupStats: func(r *healthAwareResolverImpl) {
				// Put gemini in cooldown
				r.stats["gemini"] = &ProviderStats{
					failureStreak:   1,
					lastFailureTime: time.Now(),
				}
				// Initialize openai stats too
				r.stats["openai"] = &ProviderStats{}
			},
			expectedLen:   1, // Gemini should be filtered out
			expectedFirst: "openai",
		},
		{
			name: "latency-based ordering",
			config: &utils.Config{
				LLMProvider:  "openai",
				GeminiApiKey: "test-key",
				OpenAIApiKey: "test-key",
			},
			setupStats: func(r *healthAwareResolverImpl) {
				// Gemini has lower latency
				r.stats["gemini"] = &ProviderStats{
					ewmaLatencyMs: 100,
					totalRequests: 10,
					successCount:  10,
				}
				// OpenAI has higher latency
				r.stats["openai"] = &ProviderStats{
					ewmaLatencyMs: 500,
					totalRequests: 10,
					successCount:  10,
				}
			},
			expectedLen:   2,
			expectedFirst: "openai", // Preferred provider wins despite higher latency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewHealthAwareResolver(tt.config).(*healthAwareResolverImpl)

			if tt.setupStats != nil {
				tt.setupStats(resolver)
			}

			candidates := resolver.GetRemoteCandidates()

			if len(candidates) != tt.expectedLen {
				t.Errorf("expected %d candidates, got %d: %v", tt.expectedLen, len(candidates), candidates)
			}

			if tt.expectedLen > 0 && len(candidates) > 0 && candidates[0] != tt.expectedFirst {
				t.Errorf("expected first candidate to be %s, got %s", tt.expectedFirst, candidates[0])
			}

			if tt.validateOrder != nil && !tt.validateOrder(candidates) {
				t.Errorf("candidate order validation failed: %v", candidates)
			}
		})
	}
}

func TestHealthAwareResolver_RecordSuccess(t *testing.T) {
	cfg := &utils.Config{
		GeminiApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Record multiple successes
	resolver.RecordSuccess("gemini", 100)
	resolver.RecordSuccess("gemini", 150)
	resolver.RecordSuccess("gemini", 120)

	stats := resolver.GetProviderStats("gemini")
	if stats == nil {
		t.Fatal("stats should not be nil")
	}

	if stats.totalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", stats.totalRequests)
	}

	if stats.successCount != 3 {
		t.Errorf("expected 3 successes, got %d", stats.successCount)
	}

	if stats.failureStreak != 0 {
		t.Errorf("expected 0 failure streak, got %d", stats.failureStreak)
	}

	// EWMA should be between min and max
	if stats.ewmaLatencyMs < 100 || stats.ewmaLatencyMs > 150 {
		t.Errorf("EWMA latency %.2f outside expected range", stats.ewmaLatencyMs)
	}
}

func TestHealthAwareResolver_RecordFailure(t *testing.T) {
	cfg := &utils.Config{
		GeminiApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Record failures
	resolver.RecordFailure("gemini")
	resolver.RecordFailure("gemini")

	stats := resolver.GetProviderStats("gemini")
	if stats == nil {
		t.Fatal("stats should not be nil")
	}

	if stats.failureStreak != 2 {
		t.Errorf("expected 2 failure streak, got %d", stats.failureStreak)
	}

	if stats.totalRequests != 2 {
		t.Errorf("expected 2 total requests, got %d", stats.totalRequests)
	}

	if stats.lastFailureTime.IsZero() {
		t.Error("lastFailureTime should be set")
	}
}

func TestHealthAwareResolver_IsProviderHealthy(t *testing.T) {
	cfg := &utils.Config{
		GeminiApiKey: "test-key",
		OpenAIApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Fresh provider should be healthy
	if !resolver.IsProviderHealthy("gemini") {
		t.Error("fresh provider should be healthy")
	}

	// Record failure
	resolver.RecordFailure("gemini")

	// Should be unhealthy (in cooldown)
	if resolver.IsProviderHealthy("gemini") {
		t.Error("provider should be unhealthy after failure")
	}

	// Other provider should still be healthy
	if !resolver.IsProviderHealthy("openai") {
		t.Error("other provider should still be healthy")
	}
}

func TestHealthAwareResolver_CooldownExpiry(t *testing.T) {
	cfg := &utils.Config{
		GeminiApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Record failure
	resolver.RecordFailure("gemini")

	// Should be in cooldown
	if resolver.IsProviderHealthy("gemini") {
		t.Error("provider should be in cooldown")
	}

	// Manually expire cooldown
	resolver.stats["gemini"].lastFailureTime = time.Now().Add(-3 * time.Minute)

	// Should be healthy again
	if !resolver.IsProviderHealthy("gemini") {
		t.Error("provider should be healthy after cooldown expires")
	}
}

func TestHealthAwareResolver_CalculateScore(t *testing.T) {
	cfg := &utils.Config{
		LLMProvider:  "gemini",
		GeminiApiKey: "test-key",
		OpenAIApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Setup stats
	resolver.stats["gemini"] = &ProviderStats{
		ewmaLatencyMs: 200,
		totalRequests: 10,
		successCount:  10,
		failureStreak: 0,
	}
	resolver.stats["openai"] = &ProviderStats{
		ewmaLatencyMs: 200,
		totalRequests: 10,
		successCount:  10,
		failureStreak: 0,
	}

	// Equal latency, but gemini should have higher score due to preference
	geminiScore := resolver.calculateScore("gemini", "gemini")
	openaiScore := resolver.calculateScore("openai", "gemini")

	if geminiScore <= openaiScore {
		t.Errorf("gemini score (%.2f) should be higher than openai (%.2f) due to preference", geminiScore, openaiScore)
	}
}

func TestHealthAwareResolver_SortByHealthScore(t *testing.T) {
	cfg := &utils.Config{
		LLMProvider:  "gemini",
		GeminiApiKey: "test-key",
		OpenAIApiKey: "test-key",
		GroqApiKey:   "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// Setup different latency profiles
	// Initialize all providers first
	resolver.stats["gemini"] = &ProviderStats{
		ewmaLatencyMs: 300,
	}
	resolver.stats["openai"] = &ProviderStats{
		ewmaLatencyMs: 100, // Fastest
	}
	resolver.stats["groq"] = &ProviderStats{
		ewmaLatencyMs: 500, // Slowest
	}

	candidates := []string{"gemini", "openai", "groq"}
	resolver.sortByHealthScore(candidates)

	// Gemini should be first (preferred provider), then openai (fastest), then groq
	if candidates[0] != "gemini" {
		t.Errorf("expected gemini first (preferred), got %s", candidates[0])
	}
}

func TestHealthAwareResolver_EWMACalculation(t *testing.T) {
	cfg := &utils.Config{
		GeminiApiKey: "test-key",
	}
	resolver := NewHealthAwareResolver(cfg).(*healthAwareResolverImpl)

	// First request: EWMA = value
	resolver.RecordSuccess("gemini", 200)
	stats := resolver.GetProviderStats("gemini")
	if stats.ewmaLatencyMs != 200 {
		t.Errorf("first EWMA should be 200, got %.2f", stats.ewmaLatencyMs)
	}

	// Second request: EWMA = 0.3 * 300 + 0.7 * 200 = 90 + 140 = 230
	resolver.RecordSuccess("gemini", 300)
	stats = resolver.GetProviderStats("gemini")
	expected := 0.3*300 + 0.7*200
	if stats.ewmaLatencyMs < expected-1 || stats.ewmaLatencyMs > expected+1 {
		t.Errorf("second EWMA should be ~%.2f, got %.2f", expected, stats.ewmaLatencyMs)
	}
}

func TestHealthAwareResolver_IsProviderConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   *utils.Config
		provider string
		expected bool
	}{
		{
			name: "gemini configured",
			config: &utils.Config{
				GeminiApiKey: "test-key",
			},
			provider: "gemini",
			expected: true,
		},
		{
			name: "gemini not configured",
			config: &utils.Config{
				GeminiApiKey: "",
			},
			provider: "gemini",
			expected: false,
		},
		{
			name: "openai configured",
			config: &utils.Config{
				OpenAIApiKey: "test-key",
			},
			provider: "openai",
			expected: true,
		},
		{
			name:     "invalid provider",
			config:   &utils.Config{},
			provider: "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewHealthAwareResolver(tt.config).(*healthAwareResolverImpl)
			result := resolver.isProviderConfigured(tt.provider)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
