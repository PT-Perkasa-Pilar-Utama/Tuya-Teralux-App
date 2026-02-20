package skills

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
)

// LLMClient represents the external LLM client used by RAG.
type LLMClient interface {
	CallModel(prompt string, model string) (string, error)
}

// Healthcheckable is an internal interface for LLM clients that support a health check.
type Healthcheckable interface {
	HealthCheck() bool
}

// SkillContext holds the shared services and state needed by skills during execution.
type SkillContext struct {
	UID       string
	TeraluxID string
	Prompt    string
	Language  string
	History   []string
	LLM       LLMClient
	Config    *utils.Config
	Vector    *infrastructure.VectorService
	Badger    *infrastructure.BadgerService
}

// SkillResult represents the output of a skill execution.
type SkillResult struct {
	Message        string
	Data           interface{}
	IsControl      bool
	HTTPStatusCode int
}

// Skill is the interface that all modular Sensio AI capabilities must implement.
type Skill interface {
	Name() string
	Description() string // Used by the Orchestrator for natural language routing.
	Execute(ctx *SkillContext) (*SkillResult, error)
}

// TranslateService defines the translation capability required by the Orchestrator.
// This interface decouples the skills package from the usecases package to avoid circular dependencies.
type TranslateService interface {
	TranslateText(text, targetLang string) (string, error)
	TranslateTextSync(text, targetLang string) (string, error)
}
