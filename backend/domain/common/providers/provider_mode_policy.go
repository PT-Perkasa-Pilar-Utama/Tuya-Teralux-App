package providers

import "fmt"

// ProviderMode defines the runtime policy for provider selection.
// Two modes exist:
// - DefaultMode: Health-aware fallback across all available providers. May switch providers for resilience.
// - ExplicitMode: Strict use of the user-selected provider only. No cross-provider fallback.
type ProviderMode int

const (
	// DefaultMode uses the health-aware fallback chain. If the primary provider fails,
	// the system attempts other available providers in health-score order.
	DefaultMode ProviderMode = iota

	// ExplicitMode uses only the user-selected provider. If that provider fails,
	// the error is returned to the user — no silent switch to another provider.
	ExplicitMode
)

// String returns a human-readable name for the provider mode.
func (m ProviderMode) String() string {
	switch m {
	case DefaultMode:
		return "default"
	case ExplicitMode:
		return "explicit"
	default:
		return "unknown"
	}
}

// ProviderModePolicy encapsulates the runtime behavior rules for each provider mode.
// This type exists to make the policy explicit, testable, and documented.
type ProviderModePolicy struct {
	Mode            ProviderMode
	ProviderName    string // Empty for DefaultMode, set for ExplicitMode
	FallbackAllowed bool
}

// NewDefaultModePolicy creates a policy for default mode (health-aware fallback).
func NewDefaultModePolicy() *ProviderModePolicy {
	return &ProviderModePolicy{
		Mode:            DefaultMode,
		FallbackAllowed: true,
	}
}

// NewExplicitModePolicy creates a policy for explicit mode (strict provider, no fallback).
func NewExplicitModePolicy(providerName string) *ProviderModePolicy {
	return &ProviderModePolicy{
		Mode:            ExplicitMode,
		ProviderName:    providerName,
		FallbackAllowed: false,
	}
}

// FormatError formats an error message according to the provider mode policy.
// For explicit mode: includes the provider name so the user knows which provider failed.
// For default mode: lists the attempted providers.
func (p *ProviderModePolicy) FormatError(err error, attemptedProviders []string) string {
	if p.Mode == ExplicitMode {
		return fmt.Sprintf("[Explicit Mode] Provider %q failed: %v. No fallback to other providers — a specific provider was selected.", p.ProviderName, err)
	}

	if len(attemptedProviders) > 0 {
		return fmt.Sprintf("[Default Mode] All %d provider(s) failed (%v): %v. Please try again later or select a specific provider.",
			len(attemptedProviders), attemptedProviders, err)
	}

	return fmt.Sprintf("[Default Mode] Summary generation failed: %v", err)
}

// ShouldFallback returns whether fallback to another provider is allowed.
func (p *ProviderModePolicy) ShouldFallback() bool {
	return p.FallbackAllowed
}
