package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
)

type WhisperLocalService struct {
	modelPath string
}

func NewWhisperLocalService(cfg *utils.Config) *WhisperLocalService {
	return &WhisperLocalService{
		modelPath: cfg.WhisperLocalModel,
	}
}

// Transcribe implements the usecases.WhisperClient interface
func (s *WhisperLocalService) Transcribe(audioPath string, language string) (*dtos.WhisperResult, error) {
	text, err := s.transcribeFull(audioPath, s.modelPath, language)
	if err != nil {
		return nil, err
	}

	return &dtos.WhisperResult{
		Transcription:    text,
		DetectedLanguage: language,
		Source:           "Local Whisper (whisper.cpp)",
	}, nil
}

func (s *WhisperLocalService) transcribeFull(wavPath string, modelPath string, lang string) (string, error) {
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

	return s.transcribeViaCLI(processedWavPath, modelPath, lang, true)
}

func (s *WhisperLocalService) transcribeViaCLI(wavPath string, modelPath string, lang string, full bool) (string, error) {
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

	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()

	if err != nil && !full {
		return "", fmt.Errorf("whisper-cli failed: %w - output: %s", err, string(out))
	}

	// Read produced .txt file
	b, err := os.ReadFile(txtPath)
	if err != nil {
		// Fallback to stdout if file not found
		outputStr := strings.TrimSpace(string(out))
		if outputStr == "" {
			return "", fmt.Errorf("no output file and empty stdout: %w", err)
		}
		// Basic stdout parsing (simplified from repo as fallback)
		return outputStr, nil
	}

	// Tidy up the text file content
	outputStr := strings.TrimSpace(string(b))
	if full {
		lines := strings.Split(outputStr, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line != "" {
				result = append(result, line)
			}
		}
		return strings.Join(result, " "), nil
	}

	return outputStr, nil
}
