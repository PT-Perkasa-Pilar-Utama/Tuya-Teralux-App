package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const maxSegments = 1000

// AudioSegment represents a split portion of an audio file.
type AudioSegment struct {
	Index int
	Path  string
}

// SplitAudioSegments splits an audio file into chunks of segmentSec duration with overlapSec overlap.
// It encodes segments to PCM 16k mono WAV to ensure compatibility with Whisper.
func SplitAudioSegments(inputPath string, segmentSec int, overlapSec int) ([]AudioSegment, error) {
	baseDir := filepath.Dir(inputPath)
	ext := ".wav" // Standardize output segments to WAV
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outDir := filepath.Join(baseDir, baseName+"_segments")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create segment directory: %w", err)
	}

	// 1. Get total duration
	probe, err := ProbeAudio(inputPath)
	if err != nil {
		return nil, fmt.Errorf("probe failed: %v", err)
	}
	totalDuration := probe.Duration

	var segments []AudioSegment
	index := 0
	for {
		outPath := filepath.Join(outDir, fmt.Sprintf("seg_%03d%s", index, ext))

		if index >= maxSegments {
			log.Printf("warning: segment limit reached at index %d, stopping", index)
			break
		}

		// Calculate start and duration based on segmentSec and overlapSec
		start := float64(index * segmentSec)
		if index > 0 {
			start -= float64(overlapSec)
		}
		if start < 0 {
			start = 0
		}

		if start >= totalDuration {
			break
		}

		duration := float64(segmentSec)
		if start+duration < totalDuration {
			duration += float64(overlapSec)
		}

		// Encode segment: ffmpeg -ss {start} -t {duration} -i {input} -ar 16000 -ac 1 -c:a pcm_s16le {output}
		cmdArgs := []string{
			"-y",
			"-ss", fmt.Sprintf("%.3f", start),
			"-t", fmt.Sprintf("%.3f", duration),
			"-i", inputPath,
			"-ar", "16000",
			"-ac", "1",
			"-c:a", "pcm_s16le",
			outPath,
		}

		cmd := exec.Command("ffmpeg", cmdArgs...)
		if output, err := cmd.CombinedOutput(); err != nil {
			LogDebug("[ffmpeg split] error segment %d: %s", index, string(output))
			_ = os.RemoveAll(outDir)
			return nil, fmt.Errorf("ffmpeg split error at segment %d: %v", index, err)
		}

		segments = append(segments, AudioSegment{
			Index: index,
			Path:  outPath,
		})

		index++
	}

	return segments, nil
}

// CleanupSegments removes the directory containing the segment files.
func CleanupSegments(segments []AudioSegment) {
	if len(segments) > 0 {
		dir := filepath.Dir(segments[0].Path)
		_ = os.RemoveAll(dir)
		LogDebug("Cleaned up segment directory: %s", dir)
	}
}
