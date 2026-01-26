package repositories

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type WhisperRepository struct{}

func NewWhisperRepository() *WhisperRepository {
	return &WhisperRepository{}
}

// Transcribe transcribes a WAV file using whisper-cli
func (r *WhisperRepository) Transcribe(wavPath, modelPath string) (string, error) {
	// We'll use a CLI command for simplicity in this initial implementation
	// whisper-cli -m models/ggml-base.bin -f input.wav -otxt
	// Note: We'll need to make sure whisper-cli is available.

	basePath := strings.TrimSuffix(wavPath, ".wav")
	outputPath := basePath + ".txt"
	cmd := exec.Command("whisper-cli", "-m", modelPath, "-f", wavPath, "-otxt", "-of", basePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper-cli error: %v, output: %s", err, string(output))
	}

	// Read the resulting .txt file
	txtContent, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription file: %v", err)
	}

	return strings.TrimSpace(string(txtContent)), nil
}
