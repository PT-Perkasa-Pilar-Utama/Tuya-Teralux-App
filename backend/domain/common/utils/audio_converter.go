package utils

import (
	"fmt"
	"os/exec"
)

// ConvertToWav converts an input audio file (e.g., MP3) to a 16kHz mono WAV file
// as required by Whisper.
func ConvertToWav(inputPath, outputPath string) error {
	// ffmpeg -i input.mp3 -ar 16000 -ac 1 -c:a pcm_s16le output.wav
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", outputPath)
	// Log the ffmpeg command for easier debugging
	LogDebug("[ffmpeg] running: %v", cmd.Args)
	output, err := cmd.CombinedOutput()
	LogDebug("[ffmpeg] output: %s", string(output))
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}
	return nil
}
