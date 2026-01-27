package repositories

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"teralux_app/domain/common/utils"
)

type WhisperRepository struct {
}

func NewWhisperRepository() *WhisperRepository {
	return &WhisperRepository{}
}

// Transcribe executes the local whisper-cli binary (from whisper.cpp) to transcribe the provided WAV file.
// It prefers to instruct whisper-cli to write a .txt output file (using -otxt -of <base>) and read that file,
// falling back to stdout parsing only if the file is not produced.
func (r *WhisperRepository) Transcribe(wavPath string, modelPath string) (string, error) {
	// Find whisper-cli in PATH
	bin, err := exec.LookPath("whisper-cli")
	if err != nil {
		return "", fmt.Errorf("whisper-cli not found in PATH. Run ./setup.sh to build whisper.cpp and ensure whisper-cli is available: %w", err)
	}

	// First try: ask whisper-cli to write a .txt file next to the wav
	base := strings.TrimSuffix(wavPath, ".wav")
	txtPath := base + ".txt"
	_ = os.Remove(txtPath)

	cmd := exec.Command(bin, "-m", modelPath, "-f", wavPath, "--no-timestamps", "-otxt", "-of", base)
	utils.LogDebug("[whisper] running (file output): %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	utils.LogDebug("[whisper] output: %s", string(out))
	if err == nil {
		// If CLI succeeded, try to read produced .txt file
		b, err2 := os.ReadFile(txtPath)
		if err2 == nil {
			s := strings.TrimSpace(string(b))
			if s != "" {
				lines := strings.Split(s, "\n")
				for i := len(lines) - 1; i >= 0; i-- {
					l := strings.TrimSpace(lines[i])
					if l != "" {
						return l, nil
					}
				}
			}
		}
	}

	// Fallback: capture stdout and filter out timing / progress lines
	utils.LogDebug("[whisper] falling back to stdout parse")
	cmd2 := exec.Command(bin, "-m", modelPath, "-f", wavPath, "--no-timestamps")
	utils.LogDebug("[whisper] running: %v", cmd2.Args)
	out2, err2 := cmd2.CombinedOutput()
	utils.LogDebug("[whisper] output: %s", string(out2))
	if err2 != nil {
		return "", fmt.Errorf("whisper-cli failed: %w - output: %s", err2, string(out2))
	}

	s2 := strings.TrimSpace(string(out2))
	lines2 := strings.Split(s2, "\n")
	for i := len(lines2) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines2[i])
		if l == "" {
			continue
		}
		low := strings.ToLower(l)
		// ignore known timing/progress lines
		if strings.Contains(low, "whisper_print_timings") || strings.Contains(low, "total time") || strings.Contains(low, "processing") || strings.Contains(low, "error") {
			continue
		}
		return l, nil
	}

	return s2, nil
} 
func (r *WhisperRepository) Convert(wavPath string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
