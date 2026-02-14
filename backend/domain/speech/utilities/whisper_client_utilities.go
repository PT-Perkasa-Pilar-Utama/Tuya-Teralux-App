package utilities

import (
	"fmt"
	"os"
	"path/filepath"

	"teralux_app/domain/common/utils"
	speechdtos "teralux_app/domain/speech/dtos"
)

// WhisperResult represents the result of a transcription
type WhisperResult struct {
	Transcription    string
	DetectedLanguage string
	Source           string // Which service was used: "PPU", "Orion", "Local"
}

// WhisperClient is the unified interface for all whisper transcription services
type WhisperClient interface {
	Transcribe(audioPath string, language string) (*WhisperResult, error)
}

// Healthcheckable is an optional interface for clients that support health checking
type Healthcheckable interface {
	HealthCheck() bool
}

// WhisperClientWithFallback implements automatic failover between multiple whisper providers
type WhisperClientWithFallback struct {
	primary   WhisperClient
	secondary WhisperClient
	tertiary  WhisperClient
}

// NewWhisperClientWithFallback creates a new whisper client with automatic failover.
// Tries providers in order: primary -> secondary -> tertiary until one succeeds.
func NewWhisperClientWithFallback(primary WhisperClient, secondary WhisperClient, tertiary WhisperClient) WhisperClient {
	return &WhisperClientWithFallback{
		primary:   primary,
		secondary: secondary,
		tertiary:  tertiary,
	}
}

// Transcribe attempts transcription with automatic failover
func (c *WhisperClientWithFallback) Transcribe(audioPath string, language string) (*WhisperResult, error) {
	// Try Primary (PPU)
	if c.primary != nil {
		if hp, ok := c.primary.(Healthcheckable); ok {
			utils.LogDebug("WhisperClientFallback: Checking primary (PPU) health...")
			if hp.HealthCheck() {
				utils.LogDebug("WhisperClientFallback: Primary (PPU) is healthy, proceeding.")
				result, err := c.primary.Transcribe(audioPath, language)
				if err == nil {
					return result, nil
				}
				utils.LogWarn("WhisperClientFallback: Primary (PPU) call failed: %v. Falling back to secondary.", err)
			} else {
				utils.LogWarn("WhisperClientFallback: Primary (PPU) is UNHEALTHY. Falling back to secondary.")
			}
		} else {
			result, err := c.primary.Transcribe(audioPath, language)
			if err == nil {
				return result, nil
			}
			utils.LogWarn("WhisperClientFallback: Primary call failed: %v. Falling back to secondary.", err)
		}
	}

	// Try Secondary (Orion)
	if c.secondary != nil {
		if hs, ok := c.secondary.(Healthcheckable); ok {
			utils.LogDebug("WhisperClientFallback: Checking secondary (Orion) health...")
			if hs.HealthCheck() {
				utils.LogDebug("WhisperClientFallback: Secondary (Orion) is healthy, proceeding.")
				result, err := c.secondary.Transcribe(audioPath, language)
				if err == nil {
					return result, nil
				}
				utils.LogWarn("WhisperClientFallback: Secondary (Orion) call failed: %v. Falling back to tertiary.", err)
			} else {
				utils.LogWarn("WhisperClientFallback: Secondary (Orion) is UNHEALTHY. Falling back to tertiary.")
			}
		} else {
			result, err := c.secondary.Transcribe(audioPath, language)
			if err == nil {
				return result, nil
			}
			utils.LogWarn("WhisperClientFallback: Secondary call failed: %v. Falling back to tertiary.", err)
		}
	}

	// Fallback to Tertiary (Local)
	utils.LogInfo("WhisperClientFallback: Using tertiary (local) whisper client.")
	if c.tertiary != nil {
		return c.tertiary.Transcribe(audioPath, language)
	}

	return nil, fmt.Errorf("all whisper clients failed or unavailable")
}

// ============= ADAPTERS =============

// PPUWhisperClient adapts WhisperProxyUsecase to WhisperClient interface
type PPUWhisperClient struct {
	proxy interface {
		FetchToOutsystems(filePath string, fileName string, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error)
		HealthCheck() error
	}
}

func NewPPUWhisperClient(proxy interface {
	FetchToOutsystems(filePath string, fileName string, language string) (*speechdtos.OutsystemsTranscriptionResultDTO, error)
	HealthCheck() error
}) WhisperClient {
	return &PPUWhisperClient{proxy: proxy}
}

func (c *PPUWhisperClient) Transcribe(audioPath string, language string) (*WhisperResult, error) {
	fileName := filepath.Base(audioPath)
	result, err := c.proxy.FetchToOutsystems(audioPath, fileName, language)
	if err != nil {
		return nil, err
	}

	lang := result.DetectedLanguage
	if lang == "" {
		lang = language
	}

	return &WhisperResult{
		Transcription:    result.Transcription,
		DetectedLanguage: lang,
		Source:           "PPU (Outsystems)",
	}, nil
}

func (c *PPUWhisperClient) HealthCheck() bool {
	return c.proxy.HealthCheck() == nil
}

// OrionWhisperClient adapts WhisperOrionRepository to WhisperClient interface
type OrionWhisperClient struct {
	repo interface {
		Transcribe(audioPath string, lang string) (string, error)
		HealthCheck() bool
	}
}

func NewOrionWhisperClient(repo interface {
	Transcribe(audioPath string, lang string) (string, error)
	HealthCheck() bool
}) WhisperClient {
	return &OrionWhisperClient{repo: repo}
}

func (c *OrionWhisperClient) Transcribe(audioPath string, language string) (*WhisperResult, error) {
	text, err := c.repo.Transcribe(audioPath, language)
	if err != nil {
		return nil, err
	}

	return &WhisperResult{
		Transcription:    text,
		DetectedLanguage: language,
		Source:           "Orion Whisper",
	}, nil
}

func (c *OrionWhisperClient) HealthCheck() bool {
	return c.repo.HealthCheck()
}

// LocalWhisperClient adapts WhisperCppRepository to WhisperClient interface
type LocalWhisperClient struct {
	repo interface {
		TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
	}
	modelPath string
}

func NewLocalWhisperClient(repo interface {
	TranscribeFull(wavPath string, modelPath string, lang string) (string, error)
}, modelPath string) WhisperClient {
	return &LocalWhisperClient{
		repo:      repo,
		modelPath: modelPath,
	}
}

func (c *LocalWhisperClient) Transcribe(audioPath string, language string) (*WhisperResult, error) {
	tempDir := filepath.Dir(audioPath)
	wavPath := filepath.Join(tempDir, "processed.wav")

	if err := utils.ConvertToWav(audioPath, wavPath); err != nil {
		return nil, fmt.Errorf("failed to convert audio: %w", err)
	}
	defer os.Remove(wavPath)

	text, err := c.repo.TranscribeFull(wavPath, c.modelPath, language)
	if err != nil {
		return nil, fmt.Errorf("transcription failed: %w", err)
	}

	return &WhisperResult{
		Transcription:    text,
		DetectedLanguage: language,
		Source:           "Local Whisper (whisper.cpp)",
	}, nil
}
