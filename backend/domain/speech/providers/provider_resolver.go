package providers

import (
	"context"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	whisperdtos "sensio/domain/models/whisper/dtos"
	"sensio/domain/speech/services"
	"strings"
	"time"
)

// User-selectable AI providers (excludes 'local' which is fallback-only)
var SupportedProviders = map[string]bool{
	"gemini": true,
	"openai": true,
	"groq":   true,
	"orion":  true,
}

// SupportedEngineProfiles defines product-facing profile names, separate from SupportedProviders.
var SupportedEngineProfiles = map[string]bool{
	"premium":  true,
	"standard": true,
}

// premiumProfileCandidates are the bounded provider set for the 'premium' engine profile.
var premiumProfileCandidates = []string{"openai", "groq"}

// standardProfileCandidates are the bounded provider set for the 'standard' engine profile.
var standardProfileCandidates = []string{"orion"}

// IsValidProvider checks if a provider name is supported as a user-facing provider
func IsValidProvider(provider string) bool {
	if provider == "" {
		return false
	}
	return SupportedProviders[strings.ToLower(provider)]
}

// NormalizeProvider normalizes a provider name to lowercase
func NormalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// NormalizeEngineProfile normalizes an engine profile name to lowercase.
func NormalizeEngineProfile(profile string) string {
	return strings.ToLower(strings.TrimSpace(profile))
}

// IsValidEngineProfile checks if a profile name is a known engine profile vocabulary entry.
func IsValidEngineProfile(profile string) bool {
	if profile == "" {
		return false
	}
	return SupportedEngineProfiles[NormalizeEngineProfile(profile)]
}

// ResolvedProviderSet holds the resolved LLM and Whisper clients for a specific provider
type ResolvedProviderSet struct {
	LLM           skills.LLMClient
	WhisperClient WhisperProvider
	ProviderName  string
	IsExplicit    bool   // True if provider was explicitly selected by user, false if using default/fallback
	SelectionMode string // "profile_premium" | "profile_standard" | "provider_explicit" | "default"
}

// WhisperProvider is the interface for whisper transcription services
type WhisperProvider interface {
	Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*whisperdtos.WhisperResult, error)
}

// ProviderResolver resolves the appropriate AI provider based on terminal preference
type ProviderResolver interface {
	// ResolveByTerminalID resolves provider based on terminal ID
	ResolveByTerminalID(terminalID string) (*ResolvedProviderSet, error)

	// ResolveByMacAddress resolves provider based on MAC address
	ResolveByMacAddress(macAddress string) (*ResolvedProviderSet, error)

	// ResolveDefault returns the default/global provider configuration
	ResolveDefault() *ResolvedProviderSet

	// ResolveProvider resolves a specific provider by name
	ResolveProvider(provider string) *ResolvedProviderSet

	// GetHealthAwareResolver returns the health-aware resolver for candidate-based selection
	// Returns nil if health-aware resolution is not enabled
	GetHealthAwareResolver() HealthAwareResolver

	// ExecuteWithFallback executes an operation with health-aware remote fallback
	ExecuteWithFallback(executable func(resolvedSet *ResolvedProviderSet) error, skipProviders ...string) error

	// ExecuteWithCandidateFallback executes with a bounded candidate list, using health-aware scoring within it.
	ExecuteWithCandidateFallback(candidates []string, executable func(resolvedSet *ResolvedProviderSet) error) error

	// ExecuteWithFallbackByTerminal executes with terminal-specific provider preference, then health-aware fallback
	ExecuteWithFallbackByTerminal(terminalID string, executable func(resolvedSet *ResolvedProviderSet) error) error

	// ExecuteWithFallbackByMac executes with terminal-specific provider preference (by MAC), then health-aware fallback
	ExecuteWithFallbackByMac(macAddress string, executable func(resolvedSet *ResolvedProviderSet) error) error
}

type providerResolverImpl struct {
	config *utils.Config

	// Pre-initialized provider services
	geminiService *services.GeminiService
	openaiService *services.OpenAIService
	groqService   *services.GroqService
	orionService  *services.OrionService

	// Terminal repository for looking up terminal preferences
	terminalRepo TerminalRepository

	// Health-aware resolver for candidate-based selection
	healthAwareResolver HealthAwareResolver
}

// TerminalRepository defines the minimal interface needed for terminal lookups
type TerminalRepository interface {
	GetByID(id string) (*Terminal, error)
	GetByMacAddress(macAddress string) (*Terminal, error)
}

// Terminal is a minimal terminal data structure for provider resolution
type Terminal struct {
	AiProvider      *string
	AiEngineProfile *string // Product-facing engine profile ("fast", "standard")
}

// NewProviderResolver creates a new ProviderResolver instance
func NewProviderResolver(
	cfg *utils.Config,
	geminiService *services.GeminiService,
	openaiService *services.OpenAIService,
	groqService *services.GroqService,
	orionService *services.OrionService,
	terminalRepo TerminalRepository,
) ProviderResolver {
	healthAwareResolver := NewHealthAwareResolver(cfg)

	return &providerResolverImpl{
		config:              cfg,
		geminiService:       geminiService,
		openaiService:       openaiService,
		groqService:         groqService,
		orionService:        orionService,
		terminalRepo:        terminalRepo,
		healthAwareResolver: healthAwareResolver,
	}
}

func (r *providerResolverImpl) ResolveByTerminalID(terminalID string) (*ResolvedProviderSet, error) {
	start := time.Now()

	if terminalID == "" {
		utils.LogDebug("ProviderResolver: ResolveByTerminalID called with empty terminalID, using default | duration_ms=%d", time.Since(start).Milliseconds())
		return r.ResolveDefault(), nil
	}

	repoStart := time.Now()
	terminal, err := r.terminalRepo.GetByID(terminalID)
	repoDuration := time.Since(repoStart)

	if err != nil {
		utils.LogWarn("ProviderResolver: Failed to get terminal %s | repo_duration_ms=%d | error=%v, using default provider", terminalID, repoDuration.Milliseconds(), err)
		totalDuration := time.Since(start)
		utils.LogDebug("ProviderResolver: ResolveByTerminalID completed | terminalID=%s | status=not_found | total_duration_ms=%d", terminalID, totalDuration.Milliseconds())
		return r.ResolveDefault(), nil
	}

	utils.LogDebug("ProviderResolver: Terminal %s retrieved | repo_duration_ms=%d", terminalID, repoDuration.Milliseconds())

	result, _ := r.resolveFromTerminal(terminal) // ignore error, always returns default on error
	totalDuration := time.Since(start)
	utils.LogDebug("ProviderResolver: ResolveByTerminalID completed | terminalID=%s | provider=%s | total_duration_ms=%d", terminalID, result.ProviderName, totalDuration.Milliseconds())
	return result, nil
}

func (r *providerResolverImpl) ResolveByMacAddress(macAddress string) (*ResolvedProviderSet, error) {
	start := time.Now()

	if macAddress == "" {
		utils.LogDebug("ProviderResolver: ResolveByMacAddress called with empty MAC, using default | duration_ms=%d", time.Since(start).Milliseconds())
		return r.ResolveDefault(), nil
	}

	repoStart := time.Now()
	terminal, err := r.terminalRepo.GetByMacAddress(macAddress)
	repoDuration := time.Since(repoStart)

	if err != nil {
		utils.LogWarn("ProviderResolver: Failed to get terminal for MAC %s | repo_duration_ms=%d | error=%v, using default provider", macAddress, repoDuration.Milliseconds(), err)
		totalDuration := time.Since(start)
		utils.LogDebug("ProviderResolver: ResolveByMacAddress completed | MAC=%s | status=not_found | total_duration_ms=%d", macAddress, totalDuration.Milliseconds())
		return r.ResolveDefault(), nil
	}

	utils.LogDebug("ProviderResolver: Terminal %s retrieved | repo_duration_ms=%d", macAddress, repoDuration.Milliseconds())

	result, _ := r.resolveFromTerminal(terminal) // ignore error, always returns default on error
	totalDuration := time.Since(start)
	utils.LogDebug("ProviderResolver: ResolveByMacAddress completed | MAC=%s | provider=%s | total_duration_ms=%d", macAddress, result.ProviderName, totalDuration.Milliseconds())
	return result, nil
}

func (r *providerResolverImpl) resolveFromTerminal(terminal *Terminal) (*ResolvedProviderSet, error) {
	start := time.Now()

	// Precedence 1: AiEngineProfile (product abstraction layer)
	if terminal.AiEngineProfile != nil && *terminal.AiEngineProfile != "" {
		profile := NormalizeEngineProfile(*terminal.AiEngineProfile)

		switch profile {
		case "premium":
			utils.LogInfo("ProviderResolver: resolveFromTerminal | engine_profile=premium | candidates=%v | selection_mode=profile_premium | duration_ms=%d",
				premiumProfileCandidates, time.Since(start).Milliseconds())
			return &ResolvedProviderSet{
				LLM:           nil,
				WhisperClient: nil,
				ProviderName:  "profile_premium",
				IsExplicit:    true,
				SelectionMode: "profile_premium",
			}, nil

		case "standard":
			utils.LogInfo("ProviderResolver: resolveFromTerminal | engine_profile=standard | candidates=%v | selection_mode=profile_standard | duration_ms=%d",
				standardProfileCandidates, time.Since(start).Milliseconds())
			return &ResolvedProviderSet{
				LLM:           nil,
				WhisperClient: nil,
				ProviderName:  "profile_standard",
				IsExplicit:    true,
				SelectionMode: "profile_standard",
			}, nil

		default:
			utils.LogWarn("ProviderResolver: unknown engine_profile '%s'; falling back to legacy behavior | duration_ms=%d",
				*terminal.AiEngineProfile, time.Since(start).Milliseconds())
		}
	}

	// Precedence 2: AiProvider (legacy explicit provider)
	if terminal.AiProvider != nil && *terminal.AiProvider != "" {
		provider := NormalizeProvider(*terminal.AiProvider)

		if IsValidProvider(provider) {
			utils.LogDebug("ProviderResolver: Using terminal provider '%s' | selection_mode=provider_explicit | duration_ms=%d", provider, time.Since(start).Milliseconds())
			result := r.ResolveProvider(provider)
			result.IsExplicit = true
			result.SelectionMode = "provider_explicit"
			utils.LogDebug("ProviderResolver: resolveFromTerminal completed | provider=%s | selection_mode=provider_explicit | duration_ms=%d", provider, time.Since(start).Milliseconds())
			return result, nil
		}
		utils.LogWarn("ProviderResolver: Invalid provider '%s' in terminal, using default | duration_ms=%d", *terminal.AiProvider, time.Since(start).Milliseconds())
	}

	// Precedence 3: Default
	result := r.ResolveDefault()
	result.IsExplicit = false
	result.SelectionMode = "default"
	utils.LogDebug("ProviderResolver: Using default provider '%s' | selection_mode=default | duration_ms=%d", result.ProviderName, time.Since(start).Milliseconds())
	return result, nil
}

func (r *providerResolverImpl) ResolveDefault() *ResolvedProviderSet {
	provider := NormalizeProvider(r.config.LLMProvider)

	// CRITICAL: Never allow invalid providers to be the primary provider through the global default path
	// Select remote default if LLM_PROVIDER is empty or invalid
	if provider == "" || !IsValidProvider(provider) {
		if provider != "" {
			utils.LogWarn("ProviderResolver: Invalid LLM_PROVIDER '%s', selecting remote default", provider)
		} else {
			utils.LogDebug("ProviderResolver: Global default provider is empty, selecting remote default")
		}
		provider = r.selectRemoteDefault()
		if provider == "" {
			// No remote providers configured - return error state
			utils.LogError("ProviderResolver: No remote providers configured and LLM_PROVIDER is invalid/empty")
			return &ResolvedProviderSet{
				LLM:           nil,
				WhisperClient: nil,
				ProviderName:  "",
				IsExplicit:    false,
			}
		}
	}

	return r.ResolveProvider(provider)
}

// selectRemoteDefault selects a remote provider in deterministic order:
// openai -> gemini -> groq -> orion
// Returns empty string if no remote providers are configured
func (r *providerResolverImpl) selectRemoteDefault() string {
	// Check in fixed priority order
	if r.config.OpenAIApiKey != "" {
		utils.LogInfo("ProviderResolver: Selecting OpenAI as remote default provider")
		return "openai"
	}
	if r.config.GeminiApiKey != "" {
		utils.LogInfo("ProviderResolver: Selecting Gemini as remote default provider")
		return "gemini"
	}
	if r.config.GroqApiKey != "" {
		utils.LogInfo("ProviderResolver: Selecting Groq as remote default provider")
		return "groq"
	}
	if r.config.OrionApiKey != "" {
		utils.LogInfo("ProviderResolver: Selecting Orion as remote default provider")
		return "orion"
	}

	// No remote providers configured
	return ""
}

// ResolveProvider resolves a specific provider by name
func (r *providerResolverImpl) ResolveProvider(provider string) *ResolvedProviderSet {
	start := time.Now()

	var llm skills.LLMClient
	var whisper WhisperProvider

	switch provider {
	case "gemini":
		utils.LogDebug("ProviderResolver: Using Gemini provider | duration_ms=%d", time.Since(start).Milliseconds())
		llm = r.geminiService
		whisper = r.geminiService
	case "openai":
		utils.LogDebug("ProviderResolver: Using OpenAI provider | duration_ms=%d", time.Since(start).Milliseconds())
		llm = r.openaiService
		whisper = r.openaiService
	case "groq":
		utils.LogDebug("ProviderResolver: Using Groq provider | duration_ms=%d", time.Since(start).Milliseconds())
		llm = r.groqService
		whisper = r.groqService
	case "orion":
		utils.LogDebug("ProviderResolver: Using Orion provider | duration_ms=%d", time.Since(start).Milliseconds())
		llm = r.orionService
		whisper = r.orionService
	default:
		// Invalid provider - return nil
		utils.LogError("ProviderResolver: Invalid provider '%s' | duration_ms=%d", provider, time.Since(start).Milliseconds())
		return &ResolvedProviderSet{
			LLM:           nil,
			WhisperClient: nil,
			ProviderName:  "",
			IsExplicit:    false,
		}
	}

	utils.LogDebug("ProviderResolver: resolveProvider completed | provider=%s | duration_ms=%d", provider, time.Since(start).Milliseconds())
	return &ResolvedProviderSet{
		LLM:           llm,
		WhisperClient: whisper,
		ProviderName:  provider,
		IsExplicit:    false,
	}
}

// ExecuteWithFallback executes an operation with health-aware remote fallback
func (r *providerResolverImpl) ExecuteWithFallback(executable func(resolvedSet *ResolvedProviderSet) error, skipProviders ...string) error {
	skipMap := make(map[string]bool)
	for _, p := range skipProviders {
		skipMap[p] = true
	}

	healthResolver := r.healthAwareResolver
	if healthResolver == nil {
		// Fallback to default primary provider
		defaultSet := r.ResolveDefault()
		if defaultSet == nil || defaultSet.LLM == nil {
			return fmt.Errorf("no default provider available")
		}
		return executable(defaultSet)
	}

	candidates := healthResolver.GetRemoteCandidates()
	if len(candidates) == 0 {
		utils.LogWarn("ProviderResolver: No remote candidates available, using default provider")
		defaultSet := r.ResolveDefault()
		if defaultSet == nil || defaultSet.LLM == nil {
			return fmt.Errorf("no default provider available")
		}
		return executable(defaultSet)
	}

	var lastErr error
	attemptedProviders := make([]string, 0, len(candidates))

	for _, provider := range candidates {
		attemptedProviders = append(attemptedProviders, provider)

		if skipMap[provider] {
			utils.LogDebug("ProviderResolver: Skipping already attempted provider %s from fallback chain", provider)
			continue
		}

		providerSet := r.ResolveProvider(provider)
		if providerSet == nil || providerSet.LLM == nil {
			utils.LogWarn("ProviderResolver: No client available for provider %s, skipping", provider)
			continue
		}

		attemptStart := time.Now()
		err := executable(providerSet)
		attemptDuration := time.Since(attemptStart)

		if err == nil {
			healthResolver.RecordSuccess(provider, attemptDuration.Milliseconds())
			utils.LogInfo("ProviderResolver: Execution succeeded | provider=%s | duration_ms=%d | attempts=%d | providers_tried=%v",
				provider, attemptDuration.Milliseconds(), len(attemptedProviders), attemptedProviders)
			return nil
		}

		healthResolver.RecordFailure(provider)
		lastErr = err
		utils.LogWarn("ProviderResolver: Provider %s failed (attempt %d/%d): %v", provider, len(attemptedProviders), len(candidates), err)
	}

	utils.LogError("ProviderResolver: All providers failed | providers_tried=%v | last_error=%v", attemptedProviders, lastErr)
	if lastErr != nil {
		return fmt.Errorf("all providers failed, last error: %w", lastErr)
	}
	return fmt.Errorf("all providers failed")
}

// ExecuteWithFallbackByTerminal executes with terminal-specific provider preference, then health-aware fallback
func (r *providerResolverImpl) ExecuteWithFallbackByTerminal(terminalID string, executable func(resolvedSet *ResolvedProviderSet) error) error {
	// First, try to resolve provider from terminal preference
	resolved, err := r.ResolveByTerminalID(terminalID)
	if err == nil && resolved != nil && resolved.ProviderName != "" {
		// Handle engine profile selection modes
		if resolved.SelectionMode == "profile_premium" {
			utils.LogDebug("ProviderResolver: Executing with premium profile candidates for terminal %s", terminalID)
			return r.ExecuteWithCandidateFallback(premiumProfileCandidates, executable)
		}
		if resolved.SelectionMode == "profile_standard" {
			utils.LogDebug("ProviderResolver: Executing with standard profile candidates for terminal %s", terminalID)
			return r.ExecuteWithCandidateFallback(standardProfileCandidates, executable)
		}

		// Legacy explicit provider path
		if resolved.IsExplicit && resolved.LLM != nil {
			// Explicit provider: execute and return immediately, NO fallback
			attemptStart := time.Now()
			err := executable(resolved)
			attemptDuration := time.Since(attemptStart)

			if err == nil {
				if r.healthAwareResolver != nil {
					r.healthAwareResolver.RecordSuccess(resolved.ProviderName, attemptDuration.Milliseconds())
				}
				utils.LogInfo("ProviderResolver: Explicit provider execution succeeded | terminalID=%s | provider=%s | selection_mode=%s | duration_ms=%d",
					terminalID, resolved.ProviderName, resolved.SelectionMode, attemptDuration.Milliseconds())
			} else {
				if r.healthAwareResolver != nil {
					r.healthAwareResolver.RecordFailure(resolved.ProviderName)
				}
				utils.LogError("ProviderResolver: Explicit provider %s failed for terminal %s: %v (no fallback per explicit choice policy)",
					resolved.ProviderName, terminalID, err)
			}
			return err
		}
	}

	// No valid provider from terminal, use health-aware fallback
	return r.ExecuteWithFallback(executable)
}

// ExecuteWithFallbackByMac executes with terminal-specific provider preference (by MAC), then health-aware fallback
func (r *providerResolverImpl) ExecuteWithFallbackByMac(macAddress string, executable func(resolvedSet *ResolvedProviderSet) error) error {
	// First, try to resolve provider from terminal preference
	resolved, err := r.ResolveByMacAddress(macAddress)
	if err == nil && resolved != nil && resolved.ProviderName != "" {
		// Handle engine profile selection modes
		if resolved.SelectionMode == "profile_premium" {
			utils.LogDebug("ProviderResolver: Executing with premium profile candidates for MAC %s", macAddress)
			return r.ExecuteWithCandidateFallback(premiumProfileCandidates, executable)
		}
		if resolved.SelectionMode == "profile_standard" {
			utils.LogDebug("ProviderResolver: Executing with standard profile candidates for MAC %s", macAddress)
			return r.ExecuteWithCandidateFallback(standardProfileCandidates, executable)
		}

		// Legacy explicit provider path
		if resolved.IsExplicit && resolved.LLM != nil {
			// Explicit provider: execute and return immediately, NO fallback
			attemptStart := time.Now()
			err := executable(resolved)
			attemptDuration := time.Since(attemptStart)

			if err == nil {
				if r.healthAwareResolver != nil {
					r.healthAwareResolver.RecordSuccess(resolved.ProviderName, attemptDuration.Milliseconds())
				}
				utils.LogInfo("ProviderResolver: Explicit provider execution succeeded | macAddress=%s | provider=%s | selection_mode=%s | duration_ms=%d",
					macAddress, resolved.ProviderName, resolved.SelectionMode, attemptDuration.Milliseconds())
			} else {
				if r.healthAwareResolver != nil {
					r.healthAwareResolver.RecordFailure(resolved.ProviderName)
				}
				utils.LogError("ProviderResolver: Explicit provider %s failed for MAC %s: %v (no fallback per explicit choice policy)",
					resolved.ProviderName, macAddress, err)
			}
			return err
		}
	}

	// No valid provider from terminal, use health-aware fallback
	return r.ExecuteWithFallback(executable)
}

// GetHealthAwareResolver returns the health-aware resolver for candidate-based selection
func (r *providerResolverImpl) GetHealthAwareResolver() HealthAwareResolver {
	return r.healthAwareResolver
}

// ExecuteWithCandidateFallback executes an operation with a bounded candidate list using health-aware scoring.
// It does not fall outside the provided candidate list regardless of global fallback settings.
func (r *providerResolverImpl) ExecuteWithCandidateFallback(candidates []string, executable func(resolvedSet *ResolvedProviderSet) error) error {
	if len(candidates) == 0 {
		return fmt.Errorf("no candidates provided for bounded fallback")
	}

	// Filter candidates to those with configured API keys; sort by health score if available.
	configured := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if r.isProviderConfigured(c) {
			configured = append(configured, c)
		} else {
			utils.LogDebug("ProviderResolver: ExecuteWithCandidateFallback skipping unconfigured candidate %s", c)
		}
	}

	if len(configured) == 0 {
		return fmt.Errorf("no configured providers available for requested engine profile")
	}

	eligible := make([]string, 0, len(configured))
	healthResolver := r.healthAwareResolver
	for _, c := range configured {
		if healthResolver != nil && !healthResolver.IsProviderHealthy(c) {
			utils.LogDebug("ProviderResolver: ExecuteWithCandidateFallback skipping unhealthy candidate %s", c)
			continue
		}
		eligible = append(eligible, c)
	}
	if len(eligible) == 0 {
		// All configured candidates in cooldown — retry with full configured list rather than silently failing
		utils.LogWarn("ProviderResolver: All configured bounded candidates unhealthy, retrying with full configured list: %v", configured)
		eligible = configured
	}

	// Sort eligible candidates by health score (highest first)
	if healthResolver != nil && len(eligible) > 1 {
		healthResolver.SortByHealthScore(eligible)
		utils.LogDebug("ProviderResolver: ExecuteWithCandidateFallback sorted candidates by health: %v", eligible)
	}

	var lastErr error
	attemptedProviders := make([]string, 0, len(eligible))

	for _, provider := range eligible {
		attemptedProviders = append(attemptedProviders, provider)

		providerSet := r.ResolveProvider(provider)
		if providerSet == nil || providerSet.LLM == nil {
			utils.LogWarn("ProviderResolver: ExecuteWithCandidateFallback: no client for candidate %s, skipping", provider)
			continue
		}

		attemptStart := time.Now()
		err := executable(providerSet)
		attemptDuration := time.Since(attemptStart)

		if err == nil {
			if healthResolver != nil {
				healthResolver.RecordSuccess(provider, attemptDuration.Milliseconds())
			}
			utils.LogInfo("ProviderResolver: ExecuteWithCandidateFallback succeeded | provider=%s | duration_ms=%d | attempts=%d | providers_tried=%v",
				provider, attemptDuration.Milliseconds(), len(attemptedProviders), attemptedProviders)
			return nil
		}

		if healthResolver != nil {
			healthResolver.RecordFailure(provider)
		}
		lastErr = err
		utils.LogWarn("ProviderResolver: ExecuteWithCandidateFallback candidate %s failed (attempt %d/%d): %v",
			provider, len(attemptedProviders), len(eligible), err)
	}

	utils.LogError("ProviderResolver: ExecuteWithCandidateFallback all candidates failed | providers_tried=%v | last_error=%v", attemptedProviders, lastErr)
	if lastErr != nil {
		return fmt.Errorf("all bounded candidates failed, last error: %w", lastErr)
	}
	return fmt.Errorf("all bounded candidates failed")
}

// GetProviderServices returns all available provider services for initialization
func GetProviderServices(cfg *utils.Config) (
	*services.GeminiService,
	*services.OpenAIService,
	*services.GroqService,
	*services.OrionService,
) {
	geminiService := services.NewGeminiService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)

	return geminiService, openaiService, groqService, orionService
}

// ValidateProviderConfig checks if a provider has valid configuration
func ValidateProviderConfig(provider string, cfg *utils.Config) error {
	provider = NormalizeProvider(provider)

	switch provider {
	case "gemini":
		if cfg.GeminiApiKey == "" {
			return fmt.Errorf("gemini provider requires GEMINI_API_KEY")
		}
	case "openai":
		if cfg.OpenAIApiKey == "" {
			return fmt.Errorf("openai provider requires OPENAI_API_KEY")
		}
	case "groq":
		if cfg.GroqApiKey == "" {
			return fmt.Errorf("groq provider requires GROQ_API_KEY")
		}
	case "orion":
		if cfg.OrionApiKey == "" {
			return fmt.Errorf("orion provider requires ORION_API_KEY")
		}
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	return nil
}

func (r *providerResolverImpl) isProviderConfigured(provider string) bool {
	return ValidateProviderConfig(provider, r.config) == nil
}
