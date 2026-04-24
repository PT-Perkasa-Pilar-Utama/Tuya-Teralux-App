package speech

import (
	"context"

	"sensio/domain/common/interfaces"
	"sensio/domain/common/utils"

	services "sensio/domain/speech/services"
	usecases "sensio/domain/speech/usecases"
)

type SpeechModule struct {
	ProviderResolver usecases.ProviderResolver

	GeminiService *services.GeminiService
	OpenAIService *services.OpenAIService
	GroqService   *services.GroqService
	OrionService  *services.OrionService
	LlamaService  *services.LlamaLocalService
}

type terminalRepoAdapter struct {
	repo interfaces.ITerminalRepository
}

func (a *terminalRepoAdapter) GetByID(id string) (*usecases.Terminal, error) {
	term, err := a.repo.GetByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &usecases.Terminal{
		AiProvider:      &term.AiProvider,
		AiEngineProfile: &term.AiEngineProfile,
	}, nil
}

func (a *terminalRepoAdapter) GetByMacAddress(macAddress string) (*usecases.Terminal, error) {
	term, err := a.repo.GetByMacAddress(context.Background(), macAddress)
	if err != nil {
		return nil, err
	}
	return &usecases.Terminal{
		AiProvider:      &term.AiProvider,
		AiEngineProfile: &term.AiEngineProfile,
	}, nil
}

func NewSpeechModule(
	terminalRepo interfaces.ITerminalRepository,
) *SpeechModule {
	cfg := utils.GetConfig()

	geminiService := services.NewGeminiService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)
	llamaService := services.NewLlamaLocalService(cfg)

	terminalRepoAdapter := &terminalRepoAdapter{terminalRepo}

	providerResolver := usecases.NewProviderResolver(
		cfg,
		geminiService,
		openaiService,
		groqService,
		orionService,
		terminalRepoAdapter,
	)

	return &SpeechModule{
		ProviderResolver: providerResolver,
		GeminiService:    geminiService,
		OpenAIService:    openaiService,
		GroqService:      groqService,
		OrionService:     orionService,
		LlamaService:     llamaService,
	}
}
