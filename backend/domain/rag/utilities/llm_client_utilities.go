package utilities

import (

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

