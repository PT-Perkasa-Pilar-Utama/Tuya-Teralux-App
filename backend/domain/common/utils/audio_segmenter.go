package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// AudioSegment represents a split portion of an audio file.
type AudioSegment struct {
	Index int
	Path  string
}

// SplitAudioSegments splits an audio file into chunks of segmentSec duration with overlapSec overlap.
func SplitAudioSegments(inputPath string, segmentSec int, overlapSec int) ([]AudioSegment, error) {
	baseDir := filepath.Dir(inputPath)
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	outDir := filepath.Join(baseDir, baseName+"_segments")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create segment directory: %w", err)
	}

	// 1. Get total duration
	durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputPath)
	durationOut, err := durationCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %v", err)
	}
	totalDuration, err := strconv.ParseFloat(strings.TrimSpace(string(durationOut)), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %v", err)
	}

	var segments []AudioSegment
	index := 0
	for {
		outPath := filepath.Join(outDir, fmt.Sprintf("seg_%03d%s", index, ext))

		// Calculate start and duration based on segmentSec and overlapSec
		start := float64(index * segmentSec)
		if index > 0 {
			start -= float64(overlapSec)
		}
		if start < 0 {
			start = 0
		}

		duration := float64(segmentSec)
		if start+duration < totalDuration {
			duration += float64(overlapSec)
		}

		// ffmpeg -ss {start} -t {duration} -i {input} -c copy {output}
		cmdArgs := []string{
			"-y",
			"-ss", fmt.Sprintf("%.3f", start),
			"-t", fmt.Sprintf("%.3f", duration),
			"-i", inputPath,
			"-c", "copy",
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
		if float64(index*segmentSec) >= totalDuration {
			break
		}
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
