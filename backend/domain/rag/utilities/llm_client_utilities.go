package utilities

import (
	"fmt"
	"teralux_app/domain/common/utils"
)

// LLMClient represents the external LLM client used by RAG.
// This is an interface to allow testing with fakes.
type LLMClient interface {
	CallModel(prompt string, model string) (string, error)
}

// Healthcheckable is an internal interface for LLM clients that support a health check.
type Healthcheckable interface {
	HealthCheck() bool
}

// LLMClientFallback implements automatic failover between multiple LLM providers.
// It attempts to use the primary client first, then falls back to secondary and tertiary
// if the primary fails or is unhealthy.
type LLMClientFallback struct {
	primary   LLMClient
	secondary LLMClient
	tertiary  LLMClient
}

// NewLLMClientWithFallback creates a new LLM client with automatic failover support.
// The client will attempt to use providers in order: primary -> secondary -> tertiary.
//
// Parameters:
//   - primary: The preferred LLM client (e.g., Orion)
//   - secondary: The fallback LLM client (e.g., Gemini)
//   - tertiary: The final fallback LLM client (e.g., Ollama)
//
// Returns:
//   - LLMClient: An LLM client with automatic failover capability
func NewLLMClientWithFallback(primary LLMClient, secondary LLMClient, tertiary LLMClient) LLMClient {
	return &LLMClientFallback{
		primary:   primary,
		secondary: secondary,
		tertiary:  tertiary,
	}
}

// CallModel attempts to call the LLM with automatic failover.
// It tries each client in order (primary -> secondary -> tertiary) until one succeeds.
//
// Parameters:
//   - prompt: The input prompt for the LLM
//   - model: The model identifier to use
//
// Returns:
//   - string: The LLM response
//   - error: Any error encountered (only if all clients fail)
func (c *LLMClientFallback) CallModel(prompt string, model string) (string, error) {
	// Try Primary (Orion)
	if c.primary != nil {
		if hp, ok := c.primary.(Healthcheckable); ok {
			utils.LogDebug("LLMClientFallback: Checking primary (Orion) health...")
			if hp.HealthCheck() {
				utils.LogDebug("LLMClientFallback: Primary (Orion) is healthy, proceeding.")
				res, err := c.primary.CallModel(prompt, model)
				if err == nil {
					return res, nil
				}
				utils.LogWarn("LLMClientFallback: Primary (Orion) call failed: %v. Falling back to secondary.", err)
			} else {
				utils.LogWarn("LLMClientFallback: Primary (Orion) is UNHEALTHY. Falling back to secondary.")
			}
		} else {
			// If no healthcheck, try primary anyway and catch error
			res, err := c.primary.CallModel(prompt, model)
			if err == nil {
				return res, nil
			}
			utils.LogWarn("LLMClientFallback: Primary (Orion) call failed: %v. Falling back to secondary.", err)
		}
	}

	// Try Secondary (Gemini)
	if c.secondary != nil {
		if hs, ok := c.secondary.(Healthcheckable); ok {
			utils.LogDebug("LLMClientFallback: Checking secondary (Gemini) health...")
			if hs.HealthCheck() {
				utils.LogDebug("LLMClientFallback: Secondary (Gemini) is healthy, proceeding.")
				res, err := c.secondary.CallModel(prompt, model)
				if err == nil {
					return res, nil
				}
				utils.LogWarn("LLMClientFallback: Secondary (Gemini) call failed: %v. Falling back to tertiary.", err)
			} else {
				utils.LogWarn("LLMClientFallback: Secondary (Gemini) is UNHEALTHY. Falling back to tertiary.")
			}
		} else {
			res, err := c.secondary.CallModel(prompt, model)
			if err == nil {
				return res, nil
			}
			utils.LogWarn("LLMClientFallback: Secondary (Gemini) call failed: %v. Falling back to tertiary.", err)
		}
	}

	// Fallback to Tertiary (Ollama)
	if c.tertiary != nil {
		utils.LogInfo("LLMClientFallback: Using tertiary (Ollama) LLM client.")
		return c.tertiary.CallModel(prompt, model)
	}

	return "", fmt.Errorf("all LLM clients failed or unavailable")
}
