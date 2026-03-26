package providers

import (
	"sensio/domain/common/utils"
	"sync"
	"time"
)

// ProviderStats tracks health metrics for a single provider
type ProviderStats struct {
	mu sync.RWMutex

	// EWMA latency tracking (in milliseconds)
	ewmaLatencyMs float64
	successCount  int
	totalRequests int

	// Failure tracking
	failureStreak   int
	lastFailureTime time.Time
}

// HealthAwareResolver wraps ProviderResolver with health-aware candidate selection
type HealthAwareResolver interface {
	// GetRemoteCandidates returns ordered list of healthy remote providers
	// Excludes local fallback, orders by health score (latency + failure rate)
	GetRemoteCandidates() []string

	// RecordSuccess records a successful provider call with its duration
	RecordSuccess(provider string, durationMs int64)

	// RecordFailure records a failed provider call
	RecordFailure(provider string)

	// IsProviderHealthy checks if a provider is not in cooldown
	IsProviderHealthy(provider string) bool

	// GetProviderStats returns stats for a specific provider (for debugging)
	GetProviderStats(provider string) *ProviderStats
}

type healthAwareResolverImpl struct {
	config *utils.Config

	// Map of provider -> stats
	stats map[string]*ProviderStats
	mu    sync.RWMutex

	// Cooldown configuration
	cooldownDuration time.Duration
	ewmaAlpha        float64 // Smoothing factor for EWMA (0.3 = 30% new, 70% old)
}

// Default configuration constants
const (
	DefaultCooldownDuration = 2 * time.Minute
	DefaultEWMASmoothing    = 0.3
)

// NewHealthAwareResolver creates a health-aware resolver wrapper
func NewHealthAwareResolver(cfg *utils.Config) HealthAwareResolver {
	return &healthAwareResolverImpl{
		config:           cfg,
		stats:            make(map[string]*ProviderStats),
		cooldownDuration: DefaultCooldownDuration,
		ewmaAlpha:        DefaultEWMASmoothing,
	}
}

// GetRemoteCandidates returns an ordered list of remote providers based on health
func (r *healthAwareResolverImpl) GetRemoteCandidates() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build candidate pool from configured providers
	candidates := make([]string, 0, 4)
	priorityOrder := []string{"openai", "gemini", "groq", "orion"}

	for _, provider := range priorityOrder {
		if r.isProviderConfigured(provider) && r.isProviderHealthyInternal(provider) {
			candidates = append(candidates, provider)
		}
	}

	// If no providers configured, check LLM_PROVIDER
	if len(candidates) == 0 && r.config.LLMProvider != "" {
		provider := NormalizeProvider(r.config.LLMProvider)
		if IsValidProvider(provider) && r.isProviderHealthyInternal(provider) {
			candidates = append(candidates, provider)
		}
	}

	// Sort by health score
	r.sortByHealthScore(candidates)

	utils.LogDebug("HealthAwareResolver: GetRemoteCandidates | candidates=%v", candidates)
	return candidates
}

// RecordSuccess updates provider stats after a successful call
func (r *healthAwareResolverImpl) RecordSuccess(provider string, durationMs int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats := r.getOrCreateStats(provider)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.totalRequests++
	stats.successCount++
	stats.failureStreak = 0 // Reset failure streak on success

	// Update EWMA latency
	if stats.totalRequests == 1 {
		stats.ewmaLatencyMs = float64(durationMs)
	} else {
		stats.ewmaLatencyMs = r.ewmaAlpha*float64(durationMs) + (1-r.ewmaAlpha)*stats.ewmaLatencyMs
	}

	utils.LogDebug("HealthAwareResolver: RecordSuccess | provider=%s | duration_ms=%d | ewma_latency_ms=%.2f | success_rate=%.2f",
		provider, durationMs, stats.ewmaLatencyMs, float64(stats.successCount)/float64(stats.totalRequests))
}

// RecordFailure updates provider stats after a failed call
func (r *healthAwareResolverImpl) RecordFailure(provider string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats := r.getOrCreateStats(provider)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.totalRequests++
	stats.failureStreak++
	stats.lastFailureTime = time.Now()

	utils.LogWarn("HealthAwareResolver: RecordFailure | provider=%s | failure_streak=%d | total_requests=%d",
		provider, stats.failureStreak, stats.totalRequests)
}

// IsProviderHealthy checks if a provider is not in cooldown
func (r *healthAwareResolverImpl) IsProviderHealthy(provider string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isProviderHealthyInternal(provider)
}

func (r *healthAwareResolverImpl) isProviderHealthyInternal(provider string) bool {
	stats, exists := r.stats[provider]
	if !exists {
		return true // No history = healthy
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	// Check if in cooldown
	if stats.failureStreak > 0 {
		timeSinceFailure := time.Since(stats.lastFailureTime)
		if timeSinceFailure < r.cooldownDuration {
			utils.LogDebug("HealthAwareResolver: Provider %s in cooldown | failures=%d | time_since_failure=%v | cooldown=%v",
				provider, stats.failureStreak, timeSinceFailure, r.cooldownDuration)
			return false
		}
		// Reset failure streak after cooldown expires
		stats.failureStreak = 0
	}

	return true
}

// GetProviderStats returns stats for a specific provider
func (r *healthAwareResolverImpl) GetProviderStats(provider string) *ProviderStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, exists := r.stats[provider]
	if !exists {
		return nil
	}

	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &ProviderStats{
		ewmaLatencyMs:   stats.ewmaLatencyMs,
		successCount:    stats.successCount,
		totalRequests:   stats.totalRequests,
		failureStreak:   stats.failureStreak,
		lastFailureTime: stats.lastFailureTime,
	}
}

// sortByHealthScore sorts candidates by health score (higher is better)
// Score = (1000 / ewma_latency) - (failure_streak * 100)
// Tie-break: LLM_PROVIDER preference, then alphabetical
func (r *healthAwareResolverImpl) sortByHealthScore(candidates []string) {
	preferredProvider := NormalizeProvider(r.config.LLMProvider)

	// Simple bubble sort (small list, performance not critical)
	for i := 0; i < len(candidates)-1; i++ {
		for j := 0; j < len(candidates)-i-1; j++ {
			providerA := candidates[j]
			providerB := candidates[j+1]

			scoreA := r.calculateScore(providerA, preferredProvider)
			scoreB := r.calculateScore(providerB, preferredProvider)

			if scoreB > scoreA {
				// Swap
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}
}

// calculateScore computes a health score for a provider
func (r *healthAwareResolverImpl) calculateScore(provider, preferredProvider string) float64 {
	stats, exists := r.stats[provider]
	if !exists {
		// No history: base score with preference bonus
		baseScore := 100.0
		if provider == preferredProvider {
			baseScore += 50.0
		}
		return baseScore
	}

	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// Base score from latency (lower latency = higher score)
	latencyScore := 0.0
	if stats.ewmaLatencyMs > 0 {
		latencyScore = 1000.0 / stats.ewmaLatencyMs
	} else {
		// No latency data yet: moderate base score
		latencyScore = 100.0
	}

	// Penalty for failures
	failurePenalty := float64(stats.failureStreak) * 100.0

	// Success rate bonus (optional, for tie-breaking)
	successRateBonus := 0.0
	if stats.totalRequests > 0 {
		successRate := float64(stats.successCount) / float64(stats.totalRequests)
		successRateBonus = successRate * 20.0 // Max 20 points bonus
	}

	// Preferred provider bonus
	preferredBonus := 0.0
	if provider == preferredProvider {
		preferredBonus = 50.0
	}

	score := latencyScore - failurePenalty + successRateBonus + preferredBonus

	utils.LogDebug("HealthAwareResolver: calculateScore | provider=%s | latency_score=%.2f | failure_penalty=%.2f | success_bonus=%.2f | preferred_bonus=%.2f | total=%.2f",
		provider, latencyScore, failurePenalty, successRateBonus, preferredBonus, score)

	return score
}

// isProviderConfigured checks if a provider has valid API configuration
func (r *healthAwareResolverImpl) isProviderConfigured(provider string) bool {
	switch provider {
	case "gemini":
		return r.config.GeminiApiKey != ""
	case "openai":
		return r.config.OpenAIApiKey != ""
	case "groq":
		return r.config.GroqApiKey != ""
	case "orion":
		return r.config.OrionApiKey != ""
	default:
		return false
	}
}

// getOrCreateStats gets or creates stats for a provider
func (r *healthAwareResolverImpl) getOrCreateStats(provider string) *ProviderStats {
	stats, exists := r.stats[provider]
	if !exists {
		stats = &ProviderStats{}
		r.stats[provider] = stats
	}
	return stats
}
