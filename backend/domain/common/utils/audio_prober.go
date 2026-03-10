package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// AudioProbeResult contains metadata about an audio file probed via ffprobe.
type AudioProbeResult struct {
	FormatName string  `json:"format_name"`
	Duration   float64 `json:"duration"`
	CodecName  string  `json:"codec_name"`
	Channels   int     `json:"channels"`
	SampleRate int     `json:"sample_rate"`
	HasAudio   bool    `json:"has_audio"`
}

// ProbeAudio uses ffprobe to extract detailed metadata from an audio file.
func ProbeAudio(inputPath string) (*AudioProbeResult, error) {
	// ffprobe -v error -show_entries format=format_name,duration -show_entries stream=codec_name,codec_type,channels,sample_rate -of json input.mp3
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=format_name,duration",
		"-show_entries", "stream=codec_name,codec_type,channels,sample_rate",
		"-of", "json",
		inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %v, output: %s", err, string(output))
	}

	var raw struct {
		Format struct {
			FormatName string `json:"format_name"`
			Duration   string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecName  string `json:"codec_name"`
			CodecType  string `json:"codec_type"`
			Channels   int    `json:"channels"`
			SampleRate string `json:"sample_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe json: %w", err)
	}

	result := &AudioProbeResult{
		FormatName: raw.Format.FormatName,
	}

	if dur, err := strconv.ParseFloat(raw.Format.Duration, 64); err == nil {
		result.Duration = dur
	}

	// Find the first audio stream
	for _, stream := range raw.Streams {
		if stream.CodecType == "audio" {
			result.HasAudio = true
			result.CodecName = stream.CodecName
			result.Channels = stream.Channels
			if sr, err := strconv.Atoi(stream.SampleRate); err == nil {
				result.SampleRate = sr
			}
			break
		}
	}

	return result, nil
}

// IsWavPCM16kMono checks if the audio file matches the requested 16kHz mono PCM format.
func (r *AudioProbeResult) IsWavPCM16kMono() bool {
	// Case-insensitive check for wav format and pcm_s16le codec
	isWav := strings.Contains(strings.ToLower(r.FormatName), "wav")
	isPCM := strings.Contains(strings.ToLower(r.CodecName), "pcm_s16le")
	return isWav && isPCM && r.Channels == 1 && r.SampleRate == 16000
}
