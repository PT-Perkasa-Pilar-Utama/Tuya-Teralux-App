package repositories

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"teralux_app/domain/common/utils"
)

type WhisperCppRepository struct {
	config *utils.Config
}

func NewWhisperCppRepository(cfg *utils.Config) *WhisperCppRepository {
	return &WhisperCppRepository{
		config: cfg,
	}
}

func (r *WhisperCppRepository) Transcribe(wavPath string, modelPath string, lang string) (string, error) {
	return r.transcribeViaCLI(wavPath, modelPath, lang, false)
}

func (r *WhisperCppRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
	return r.transcribeViaCLI(wavPath, modelPath, lang, true)
}

func (r *WhisperCppRepository) transcribeViaCLI(wavPath string, modelPath string, lang string, full bool) (string, error) {
	// Find whisper-cli: try local bin first, then PATH
	bin := "./bin/whisper-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		binInPath, err := exec.LookPath("whisper-cli")
		if err != nil {
			return "", fmt.Errorf("whisper-cli not found in ./bin or PATH. Run './setup.sh' to build and prepare the binary: %w", err)
		}
		bin = binInPath
	}

	// instructed whisper-cli to write a .txt file (it formats it nicely without weights/timings)
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
		s := strings.TrimSpace(string(out))
		if s == "" {
			return "", fmt.Errorf("no output file and empty stdout: %w", err)
		}
		// Basic stdout parsing
		lines := strings.Split(s, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line == "" {
				continue
			}
			low := strings.ToLower(line)
			if strings.Contains(low, "whisper_") || strings.Contains(low, "total time") || strings.Contains(low, "error") {
				continue
			}
			result = append(result, line)
		}
		if full {
			return strings.Join(result, " "), nil
		}
		if len(result) > 0 {
			return result[len(result)-1], nil
		}
		return "", nil
	}

	// Tidy up the text file content
	s := strings.TrimSpace(string(b))
	if full {
		lines := strings.Split(s, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line != "" {
				result = append(result, line)
			}
		}
		return strings.Join(result, " "), nil
	}

	// For original Transcribe behavior (last line)
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l != "" {
			return l, nil
		}
	}
	return s, nil
}
