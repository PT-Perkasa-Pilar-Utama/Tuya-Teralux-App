package usecases

import (
	"errors"
	"sensio/domain/common/providers"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
)

// UpdateTerminalAIEngineProfileUseCase handles writing the engine profile for a terminal
type UpdateTerminalAIEngineProfileUseCase struct {
	repository repositories.ITerminalRepository
	cfg        *utils.Config
}

// NewUpdateTerminalAIEngineProfileUseCase creates a new instance
func NewUpdateTerminalAIEngineProfileUseCase(repository repositories.ITerminalRepository, cfg *utils.Config) *UpdateTerminalAIEngineProfileUseCase {
	return &UpdateTerminalAIEngineProfileUseCase{repository: repository, cfg: cfg}
}

// Update sets the engine profile for the terminal identified by ID.
// Accepted values: "premium", "standard". Empty/nil clears the profile.
// Unknown values are rejected with a validation error.
func (uc *UpdateTerminalAIEngineProfileUseCase) Update(id string, req *dtos.UpdateTerminalAIEngineProfileRequestDTO) (*dtos.TerminalAIEngineProfileResponseDTO, error) {
	term, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, errors.New("Terminal not found")
	}

	if req.Profile == nil || *req.Profile == "" {
		// Clear the profile
		utils.LogInfo("UpdateTerminalAIEngineProfileUseCase: clearing engine profile | terminal_id=%s", id)
		term.AiEngineProfile = nil
	} else {
		normalized := providers.NormalizeEngineProfile(*req.Profile)

		switch normalized {
		case "premium":
			openAiErr := providers.ValidateProviderConfig("openai", uc.cfg)
			groqErr := providers.ValidateProviderConfig("groq", uc.cfg)
			if openAiErr != nil && groqErr != nil {
				utils.LogWarn("UpdateTerminalAIEngineProfileUseCase: premium profile update blocked by missing config | terminal_id=%s | openai_err=%v | groq_err=%v", id, openAiErr, groqErr)
				return nil, utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
					{Field: "profile", Message: "premium profile is unavailable because neither OpenAI nor Groq is configured"},
				})
			}
			term.AiEngineProfile = &normalized
			utils.LogInfo("UpdateTerminalAIEngineProfileUseCase: setting engine profile to premium | terminal_id=%s", id)
		case "standard":
			if err := providers.ValidateProviderConfig("orion", uc.cfg); err != nil {
				utils.LogWarn("UpdateTerminalAIEngineProfileUseCase: standard profile update blocked by missing config | terminal_id=%s | error=%v", id, err)
				return nil, utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
					{Field: "profile", Message: "standard profile is unavailable because Orion is not configured"},
				})
			}
			term.AiEngineProfile = &normalized
			utils.LogInfo("UpdateTerminalAIEngineProfileUseCase: setting engine profile to standard | terminal_id=%s", id)
		default:
			return nil, utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
				{Field: "profile", Message: "invalid profile; supported values: premium, standard"},
			})
		}
	}

	if err := uc.repository.Update(term); err != nil {
		return nil, err
	}

	if err := uc.repository.InvalidateCache(id); err != nil {
		utils.LogWarn("UpdateTerminalAIEngineProfileUseCase: cache invalidation failed for terminal %s: %v", id, err)
	}

	var source string
	var effectiveProvider *string
	var effectiveMode string

	if term.AiEngineProfile != nil && *term.AiEngineProfile != "" {
		source = "engine_profile"
		effectiveMode = *term.AiEngineProfile
		effectiveProvider = nil
	} else if term.AiProvider != nil && *term.AiProvider != "" {
		source = "legacy_provider"
		effectiveMode = "legacy"
		effectiveProvider = term.AiProvider
	} else {
		source = "default"
		effectiveMode = "default"
		effectiveProvider = nil
	}

	return &dtos.TerminalAIEngineProfileResponseDTO{
		TerminalID:        term.ID,
		Profile:           term.AiEngineProfile,
		Source:            source,
		EffectiveProvider: effectiveProvider,
		EffectiveMode:     effectiveMode,
	}, nil
}
