package providers

import (
	"context"
	"fmt"
	"sensio/domain/common/services"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	whisperdtos "sensio/domain/models/whisper/dtos"
	"strings"
)

// User-selectable AI providers (excludes 'local' which is fallback-only)
var SupportedProviders = map[string]bool{
	"gemini":  true,
	"openai":  true,
	"groq":    true,
	"orion":   true,
}

// IsValidProvider checks if a provider name is supported as a user-facing provider
func IsValidProvider(provider string) bool {
	if provider == "" {
		return false
	}
	return SupportedProviders[strings.ToLower(provider)]
}

// IsFallbackOnlyProvider checks if a provider is fallback-only (not user-selectable)
func IsFallbackOnlyProvider(provider string) bool {
	return strings.ToLower(provider) == "local"
}

// IsValidProviderOrLegacyLocal checks if a provider is valid or is the legacy 'local' value
// Used for backward compatibility when reading existing terminal records
func IsValidProviderOrLegacyLocal(provider string) bool {
	if provider == "" {
		return false
	}
	normalized := strings.ToLower(provider)
	return SupportedProviders[normalized] || IsFallbackOnlyProvider(normalized)
}

// NormalizeProvider normalizes a provider name to lowercase
func NormalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// ResolvedProviderSet holds the resolved LLM and Whisper clients for a specific provider
type ResolvedProviderSet struct {
	LLM             skills.LLMClient
	FallbackLLM     skills.LLMClient
	WhisperClient   WhisperProvider
	FallbackWhisper WhisperProvider
	ProviderName    string
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
}

type providerResolverImpl struct {
	config *utils.Config

	// Pre-initialized provider services
	geminiService  *services.GeminiService
	openaiService  *services.OpenAIService
	groqService    *services.GroqService
	orionService   *services.OrionService
	llamaService   *services.LlamaLocalService
	whisperService *services.WhisperLocalService

	// Terminal repository for looking up terminal preferences
	terminalRepo TerminalRepository
}

// TerminalRepository defines the minimal interface needed for terminal lookups
type TerminalRepository interface {
	GetByID(id string) (*Terminal, error)
	GetByMacAddress(macAddress string) (*Terminal, error)
}

// Terminal is a minimal terminal data structure for provider resolution
type Terminal struct {
	AiProvider *string
}

// NewProviderResolver creates a new ProviderResolver instance
func NewProviderResolver(
	cfg *utils.Config,
	geminiService *services.GeminiService,
	openaiService *services.OpenAIService,
	groqService *services.GroqService,
	orionService *services.OrionService,
	llamaService *services.LlamaLocalService,
	whisperService *services.WhisperLocalService,
	terminalRepo TerminalRepository,
) ProviderResolver {
	return &providerResolverImpl{
		config:         cfg,
		geminiService:  geminiService,
		openaiService:  openaiService,
		groqService:    groqService,
		orionService:   orionService,
		llamaService:   llamaService,
		whisperService: whisperService,
		terminalRepo:   terminalRepo,
	}
}

func (r *providerResolverImpl) ResolveByTerminalID(terminalID string) (*ResolvedProviderSet, error) {
	if terminalID == "" {
		return r.ResolveDefault(), nil
	}

	terminal, err := r.terminalRepo.GetByID(terminalID)
	if err != nil {
		utils.LogWarn("ProviderResolver: Failed to get terminal %s: %v, using default provider", terminalID, err)
		return r.ResolveDefault(), nil
	}

	return r.resolveFromTerminal(terminal)
}

func (r *providerResolverImpl) ResolveByMacAddress(macAddress string) (*ResolvedProviderSet, error) {
	if macAddress == "" {
		return r.ResolveDefault(), nil
	}

	terminal, err := r.terminalRepo.GetByMacAddress(macAddress)
	if err != nil {
		utils.LogWarn("ProviderResolver: Failed to get terminal for MAC %s: %v, using default provider", macAddress, err)
		return r.ResolveDefault(), nil
	}

	return r.resolveFromTerminal(terminal)
}

func (r *providerResolverImpl) resolveFromTerminal(terminal *Terminal) (*ResolvedProviderSet, error) {
	// If terminal has a provider preference, use it
	if terminal.AiProvider != nil && *terminal.AiProvider != "" {
		provider := NormalizeProvider(*terminal.AiProvider)

		// Handle legacy 'local' value: treat as unset/default for primary provider resolution
		// Local remains available as internal fallback
		if IsFallbackOnlyProvider(provider) {
			utils.LogDebug("ProviderResolver: Legacy 'local' provider detected in terminal, using default primary provider with local fallback")
			return r.ResolveDefault(), nil
		}

		if IsValidProvider(provider) {
			return r.resolveProvider(provider), nil
		}
		utils.LogWarn("ProviderResolver: Invalid provider '%s' in terminal, using default", *terminal.AiProvider)
	}

	// Fall back to default
	return r.ResolveDefault(), nil
}

func (r *providerResolverImpl) ResolveDefault() *ResolvedProviderSet {
	provider := NormalizeProvider(r.config.LLMProvider)
	
	// CRITICAL: Never allow 'local' to be the primary provider through the global default path
	// Also handle invalid provider values (e.g., LLM_PROVIDER=foo) by selecting remote default
	if provider == "" || IsFallbackOnlyProvider(provider) || !IsValidProvider(provider) {
		if provider != "" && !IsFallbackOnlyProvider(provider) {
			utils.LogWarn("ProviderResolver: Invalid LLM_PROVIDER '%s', selecting remote default", provider)
		} else {
			utils.LogDebug("ProviderResolver: Global default provider is '%s', selecting remote default", provider)
		}
		provider = r.selectRemoteDefault()
		if provider == "" {
			// No remote providers configured - return error state rather than promoting local
			utils.LogError("ProviderResolver: No remote providers configured and LLM_PROVIDER is invalid/empty")
			// Return a minimal set with clear error indication - callers should handle this
			return &ResolvedProviderSet{
				LLM:             nil,
				FallbackLLM:     r.llamaService,
				WhisperClient:   nil,
				FallbackWhisper: r.whisperService,
				ProviderName:    "",
			}
		}
	}
	
	return r.resolveProvider(provider)
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

func (r *providerResolverImpl) resolveProvider(provider string) *ResolvedProviderSet {
	var llm skills.LLMClient
	var whisper WhisperProvider

	switch provider {
	case "gemini":
		utils.LogDebug("ProviderResolver: Using Gemini provider")
		llm = r.geminiService
		whisper = r.geminiService
	case "openai":
		utils.LogDebug("ProviderResolver: Using OpenAI provider")
		llm = r.openaiService
		whisper = r.openaiService
	case "groq":
		utils.LogDebug("ProviderResolver: Using Groq provider")
		llm = r.groqService
		whisper = r.groqService
	case "orion":
		utils.LogDebug("ProviderResolver: Using Orion provider")
		llm = r.orionService
		whisper = r.orionService
	default:
		// Default to local if invalid (should not happen for validated providers)
		utils.LogError("ProviderResolver: Invalid provider '%s', using local fallback", provider)
		llm = r.llamaService
		whisper = r.whisperService
	}

	// Fallback is always local
	fallbackLLM := r.llamaService
	fallbackWhisper := r.whisperService

	return &ResolvedProviderSet{
		LLM:             llm,
		FallbackLLM:     fallbackLLM,
		WhisperClient:   whisper,
		FallbackWhisper: fallbackWhisper,
		ProviderName:    provider,
	}
}

// GetProviderServices returns all available provider services for initialization
func GetProviderServices(cfg *utils.Config) (
	*services.GeminiService,
	*services.OpenAIService,
	*services.GroqService,
	*services.OrionService,
	*services.LlamaLocalService,
	*services.WhisperLocalService,
) {
	geminiService := services.NewGeminiService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)
	llamaService := services.NewLlamaLocalService(cfg)
	whisperService := services.NewWhisperLocalService(cfg)

	return geminiService, openaiService, groqService, orionService, llamaService, whisperService
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
