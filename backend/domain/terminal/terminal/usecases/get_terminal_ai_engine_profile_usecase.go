package usecases

import (
	"errors"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
)

// GetTerminalAIEngineProfileUseCase handles reading the engine profile for a terminal by MAC address
type GetTerminalAIEngineProfileUseCase struct {
	repository repositories.ITerminalRepository
}

// NewGetTerminalAIEngineProfileUseCase creates a new instance
func NewGetTerminalAIEngineProfileUseCase(repository repositories.ITerminalRepository) *GetTerminalAIEngineProfileUseCase {
	return &GetTerminalAIEngineProfileUseCase{repository: repository}
}

// GetByMac returns the engine profile for the terminal identified by MAC address
func (uc *GetTerminalAIEngineProfileUseCase) GetByMac(macAddress string) (*dtos.TerminalAIEngineProfileResponseDTO, error) {
	term, err := uc.repository.GetByMacAddress(macAddress)
	if err != nil {
		return nil, errors.New("Terminal not found")
	}

	var source string
	var effectiveProvider *string
	var effectiveMode string
	var profile *string

	if term.AiEngineProfile != nil && *term.AiEngineProfile != "" {
		source = "engine_profile"
		effectiveMode = *term.AiEngineProfile
		effectiveProvider = nil
		profile = term.AiEngineProfile
	} else if term.AiProvider != nil && *term.AiProvider != "" {
		source = "legacy_provider"
		effectiveMode = "legacy"
		effectiveProvider = term.AiProvider
		profile = nil
	} else {
		source = "default"
		effectiveMode = "default"
		effectiveProvider = nil
		profile = nil
	}

	return &dtos.TerminalAIEngineProfileResponseDTO{
		TerminalID:        term.ID,
		Profile:           profile,
		Source:            source,
		EffectiveProvider: effectiveProvider,
		EffectiveMode:     effectiveMode,
	}, nil
}
