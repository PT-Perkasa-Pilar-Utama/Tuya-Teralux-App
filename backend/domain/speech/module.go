package speech

import (
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"

	usecases "sensio/domain/speech/speech/usecases"
	services "sensio/domain/speech/speech/services"
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
	repo terminal_repositories.ITerminalRepository
}

func (a *terminalRepoAdapter) GetByID(id string) (*usecases.Terminal, error) {
	term, err := a.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &usecases.Terminal{
		AiProvider:      term.AiProvider,
		AiEngineProfile: term.AiEngineProfile,
	}, nil
}

func (a *terminalRepoAdapter) GetByMacAddress(macAddress string) (*usecases.Terminal, error) {
	term, err := a.repo.GetByMacAddress(macAddress)
	if err != nil {
		return nil, err
	}
	return &usecases.Terminal{
		AiProvider:      term.AiProvider,
		AiEngineProfile: term.AiEngineProfile,
	}, nil
}

func NewSpeechModule(
	badger *infrastructure.BadgerService,
) *SpeechModule {
	cfg := utils.GetConfig()

	geminiService := services.NewGeminiService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	groqService := services.NewGroqService(cfg)
	orionService := services.NewOrionService(cfg)
	llamaService := services.NewLlamaLocalService(cfg)

	terminalRepo := terminal_repositories.NewTerminalRepository(badger)
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