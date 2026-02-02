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
func (r *WhisperRepository) Transcribe(wavPath string, modelPath string, lang string) (string, error) {
	// Find whisper-cli: try local bin first, then PATH
	bin := "./bin/whisper-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		binInPath, err := exec.LookPath("whisper-cli")
		if err != nil {
			return "", fmt.Errorf("whisper-cli not found in ./bin or PATH. Run './setup.sh' to build and prepare the binary: %w", err)
		}
		bin = binInPath
	}

	// First try: ask whisper-cli to write a .txt file next to the wav
	base := strings.TrimSuffix(wavPath, ".wav")
	txtPath := base + ".txt"
	_ = os.Remove(txtPath)

	args := []string{"-m", modelPath, "-f", wavPath, "--no-timestamps", "-otxt", "-of", base}
	if lang != "" {
		args = append(args, "-l", lang)
	}

	cmd := exec.Command(bin, args...)
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
	argsFallback := []string{"-m", modelPath, "-f", wavPath, "--no-timestamps"}
	if lang != "" {
		argsFallback = append(argsFallback, "-l", lang)
	}
	cmd2 := exec.Command(bin, argsFallback...)
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

// TranscribeFull executes whisper-cli and returns a concatenated string of all transcription segments.
func (r *WhisperRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
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
	utils.LogDebug("[whisper-full] running: %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper-cli failed: %w - output: %s", err, string(out))
	}

	// Read produced .txt file
	b, err := os.ReadFile(txtPath)
	if err != nil {
		// If file not produced, try to parse stdout (similar to Transcribe but join all)
		s := strings.TrimSpace(string(out))
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
		return strings.Join(result, " "), nil
	}

	// Tidy up the text file content
	s := strings.TrimSpace(string(b))
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

func (r *WhisperRepository) Convert(wavPath string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
