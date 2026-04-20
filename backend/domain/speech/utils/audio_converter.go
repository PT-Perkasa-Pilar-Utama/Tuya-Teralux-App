package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sensio/domain/common/utils"
)

// ConvertToWav converts an input audio file (e.g., MP3) to a 16kHz mono WAV file
// as required by Whisper.
func ConvertToWav(inputPath, outputPath string) error {
	// ffmpeg -i input.mp3 -ar 16000 -ac 1 -c:a pcm_s16le output.wav
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", outputPath)
	// Log the ffmpeg command for easier debugging
	utils.LogDebug("[ffmpeg] running: %v", cmd.Args)
	output, err := cmd.CombinedOutput()
	utils.LogDebug("[ffmpeg] output: %s", string(output))
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}
	return nil
}

// NormalizeToWavPCM16k ensures the audio file is in the target format (WAV PCM 16k Mono).
// If the input satisfies the requirements, it returns the original path.
// Otherwise, it creates a temporary normalized WAV file.
// Returns: (processingPath, cleanupFn, error)
func NormalizeToWavPCM16k(inputPath string) (string, func(), error) {
	probe, err := ProbeAudio(inputPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to probe audio: %w", err)
	}

	// 1. Validation: Fail early if no audio stream
	if !probe.HasAudio {
		return "", nil, fmt.Errorf("invalid file: no audio stream detected")
	}

	// 2. Logging: Highlight extension mismatch
	ext := strings.ToLower(filepath.Ext(inputPath))
	formatMatch := strings.Contains(strings.ToLower(probe.FormatName), strings.TrimPrefix(ext, "."))
	if ext != "" && !formatMatch {
		utils.LogWarn("[audio] Extension mismatch warning: filename has %s but ffprobe detected %s for %s", ext, probe.FormatName, inputPath)
	}

	// 3. Check if already normalized
	if probe.IsWavPCM16kMono() {
		return inputPath, func() {}, nil
	}

	// 4. Create unique temp normalized file to avoid races
	utils.LogInfo("[audio] Normalizing to WAV 16k mono: %s", inputPath)
	uniqueSuffix := time.Now().UnixNano()
	outputPath := fmt.Sprintf("%s.%d.normalize.wav", inputPath, uniqueSuffix)

	if err := ConvertToWav(inputPath, outputPath); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.Remove(outputPath)
	}

	return outputPath, cleanup, nil
}
