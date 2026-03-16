package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
)

type WhisperLocalService struct {
	modelPath string
}

func NewWhisperLocalService(cfg *utils.Config) *WhisperLocalService {
	return &WhisperLocalService{
		modelPath: cfg.WhisperLocalModel,
	}
}

// HealthCheck validates that whisper-cli is available, can resolve libwhisper.so.1,
// and the configured model exists. This should be called at startup to ensure
// local fallback is healthy.
func (s *WhisperLocalService) HealthCheck() bool {
	// Find whisper-cli: try local bin first, then PATH
	bin := "./bin/whisper-cli"
	binFound := false
	binPath := ""

	if _, err := os.Stat(bin); err == nil {
		binFound = true
		binPath = bin
	} else {
		if path, err := exec.LookPath("whisper-cli"); err == nil {
			binFound = true
			binPath = path
		}
	}

	if !binFound {
		utils.LogError("WhisperLocal HealthCheck: whisper-cli not found in ./bin or PATH")
		return false
	}

	// CRITICAL: Actually execute the binary to verify libwhisper.so.1 resolution.
	// Running with --help is lightweight and will fail if shared libraries are missing.
	cmd := exec.Command(binPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// Check for library loading errors
		if strings.Contains(outputStr, "libwhisper.so.1") ||
			strings.Contains(outputStr, "shared library") ||
			strings.Contains(outputStr, "cannot open shared object") ||
			strings.Contains(outputStr, "error while loading shared libraries") {
			utils.LogError("WhisperLocal HealthCheck: whisper-cli failed to load libwhisper.so.1 | error=%v | output=%s", err, outputStr)
			return false
		}
		// Other errors may still indicate runtime issues
		utils.LogError("WhisperLocal HealthCheck: whisper-cli --help failed | error=%v | output=%s", err, outputStr)
		return false
	}

	// Validate that the configured model file exists
	if s.modelPath == "" {
		utils.LogError("WhisperLocal HealthCheck: no model path configured")
		return false
	}

	if _, err := os.Stat(s.modelPath); os.IsNotExist(err) {
		utils.LogError("WhisperLocal HealthCheck: model file not found at %s: %v", s.modelPath, err)
		return false
	}

	return true
}

// Transcribe implements the usecases.WhisperClient interface
func (s *WhisperLocalService) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	text, err := s.transcribeFull(ctx, audioPath, s.modelPath, language)
	if err != nil {
		return nil, err
	}

	return &dtos.WhisperResult{
		Transcription:    text,
		DetectedLanguage: language,
		Diarized:         false,
		Source:           "Local Whisper (whisper.cpp)",
	}, nil
}

func (s *WhisperLocalService) transcribeFull(ctx context.Context, wavPath string, modelPath string, lang string) (string, error) {
	// Ensure input is WAV. If not, convert.
	// Note: previous repo assumed conversion was done outside or inside via a separate method?
	// Let's check `whisper_client_utilities.go`: `NewLocalWhisperClient` adapter did the conversion.
	// `whisper_cpp_repository.go` method `TranscribeFull` assumed wavPath input.
	// But `Transcribe` in `whisper_client_utilities.go` did the conversion.
	// Since this service implements `WhisperClient`, it SHOULD handle conversion.

	tempDir := filepath.Dir(wavPath)
	processedWavPath := filepath.Join(tempDir, "processed_local_service.wav")

	if err := utils.ConvertToWav(wavPath, processedWavPath); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}
	defer func() { _ = os.Remove(processedWavPath) }()

	return s.transcribeViaCLI(ctx, processedWavPath, modelPath, lang, true)
}

func (s *WhisperLocalService) transcribeViaCLI(ctx context.Context, wavPath string, modelPath string, lang string, full bool) (string, error) {
	// Find whisper-cli: try local bin first, then PATH
	bin := "./bin/whisper-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		binInPath, err := exec.LookPath("whisper-cli")
		if err != nil {
			return "", fmt.Errorf("whisper-cli not found in ./bin or PATH: %w", err)
		}
		bin = binInPath
	}

	// instructed whisper-cli to write a .txt file
	base := strings.TrimSuffix(wavPath, ".wav")
	txtPath := base + ".txt"
	_ = os.Remove(txtPath)

	args := []string{"-m", modelPath, "-f", wavPath, "--no-timestamps", "-otxt", "-of", base}
	if lang != "" && lang != "auto" {
		args = append(args, "-l", lang)
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(out))

	// Always return an error when whisper-cli exits non-zero, regardless of full mode.
	// Never use stderr/stdout as transcript when process exit failed.
	if err != nil {
		return "", fmt.Errorf("whisper-cli failed with exit error: %w - output: %s", err, outputStr)
	}

	// Read produced .txt file
	b, err := os.ReadFile(txtPath)
	if err != nil {
		// Fallback to stdout if file not found (only when exit was successful)
		if outputStr == "" {
			return "", fmt.Errorf("no output file and empty stdout: %w", err)
		}
		// Basic stdout parsing (simplified from repo as fallback)
		return outputStr, nil
	}

	// Tidy up the text file content
	text := strings.TrimSpace(string(b))
	if full {
		lines := strings.Split(text, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line != "" {
				result = append(result, line)
			}
		}
		return strings.Join(result, " "), nil
	}

	return text, nil
}
