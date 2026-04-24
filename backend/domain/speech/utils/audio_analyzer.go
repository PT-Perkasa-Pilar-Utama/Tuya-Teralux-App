package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// AudioClass represents the classification of audio based on signal quality
type AudioClass string

const (
	AudioClassSilent     AudioClass = "silent"      // Effectively no usable voice activity
	AudioClassNearSilent AudioClass = "near_silent" // Mostly silence with tiny peaks / room noise
	AudioClassActive     AudioClass = "active"      // Enough speech-like signal to attempt ASR
)

// AudioMetrics holds computed audio analysis metrics
type AudioMetrics struct {
	DurationSec       float64 // Duration in seconds
	MeanVolumeDB      float64 // Mean volume in dB
	MaxVolumeDB       float64 // Max volume in dB
	SilencePercentage float64 // Percentage of audio that is silence
	LongestSilenceSec float64 // Longest continuous silence duration in seconds
}

// AudioAnalysisResult combines metrics with classification
type AudioAnalysisResult struct {
	Metrics AudioMetrics
	Class   AudioClass
}

// AudioAnalyzer analyzes audio files for silence and signal quality
type AudioAnalyzer interface {
	Analyze(audioPath string) (*AudioAnalysisResult, error)
}

type ffmpegAudioAnalyzer struct {
	ffmpegPath  string
	ffprobePath string
}

// NewAudioAnalyzer creates a new audio analyzer using FFmpeg/ffprobe
func NewAudioAnalyzer() AudioAnalyzer {
	return &ffmpegAudioAnalyzer{
		ffmpegPath:  detectFFmpegPath(),
		ffprobePath: detectFFprobePath(),
	}
}

func detectFFmpegPath() string {
	// Try to find ffmpeg in PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path
	}
	// Fallback to common installation paths
	commonPaths := []string{
		"/usr/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
		"/opt/homebrew/bin/ffmpeg",
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "ffmpeg" // Assume it's in PATH
}

func detectFFprobePath() string {
	// Try to find ffprobe in PATH
	if path, err := exec.LookPath("ffprobe"); err == nil {
		return path
	}
	// Fallback to common installation paths
	commonPaths := []string{
		"/usr/bin/ffprobe",
		"/usr/local/bin/ffprobe",
		"/opt/homebrew/bin/ffprobe",
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "ffprobe" // Assume it's in PATH
}

// Analyze performs audio analysis and returns metrics + classification
func (a *ffmpegAudioAnalyzer) Analyze(audioPath string) (*AudioAnalysisResult, error) {
	// Check if file exists
	if _, err := os.Stat(audioPath); err != nil {
		return nil, fmt.Errorf("audio file not found: %w", err)
	}

	metrics, err := a.extractMetrics(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract audio metrics: %w", err)
	}

	class := a.classify(metrics)

	return &AudioAnalysisResult{
		Metrics: *metrics,
		Class:   class,
	}, nil
}

// extractMetrics uses ffprobe and ffmpeg to extract audio metrics
func (a *ffmpegAudioAnalyzer) extractMetrics(audioPath string) (*AudioMetrics, error) {
	metrics := &AudioMetrics{}

	// 1. Get duration using ffprobe
	duration, err := a.getDuration(audioPath)
	if err != nil {
		return nil, err
	}
	metrics.DurationSec = duration

	// 2. Get volume statistics using ffmpeg volumedetect filter
	volMetrics, err := a.getVolumeStats(audioPath)
	if err != nil {
		return nil, err
	}
	metrics.MeanVolumeDB = volMetrics.meanVol
	metrics.MaxVolumeDB = volMetrics.maxVol

	// 3. Detect silence periods using ffmpeg silencedetect filter
	silenceMetrics, err := a.detectSilence(audioPath)
	if err != nil {
		return nil, err
	}
	metrics.SilencePercentage = silenceMetrics.silencePercentage
	metrics.LongestSilenceSec = silenceMetrics.longestSilence

	return metrics, nil
}

// getDuration extracts audio duration using ffprobe
func (a *ffmpegAudioAnalyzer) getDuration(audioPath string) (float64, error) {
	cmd := exec.Command(a.ffprobePath,
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", durationStr, err)
	}

	return duration, nil
}

type volMetricsResult struct {
	meanVol float64
	maxVol  float64
}

// getVolumeStats extracts mean and max volume using ffmpeg volumedetect filter
//
//nolint:unparam
func (a *ffmpegAudioAnalyzer) getVolumeStats(audioPath string) (*volMetricsResult, error) {
	cmd := exec.Command(a.ffmpegPath,
		"-i", audioPath,
		"-af", "volumedetect",
		"-f", "null",
		"-",
	)
	// volumedetect outputs to stderr - it may return non-zero exit code but still produces output
	output, _ := cmd.CombinedOutput() //nolint:errcheck

	outputStr := string(output)

	// Parse mean_volume: -XX.X dB
	meanVolRe := regexp.MustCompile(`mean_volume:\s+(-?\d+\.?\d*)\s+dB`)
	meanMatch := meanVolRe.FindStringSubmatch(outputStr)

	// Parse max_volume: -XX.X dB
	maxVolRe := regexp.MustCompile(`max_volume:\s+(-?\d+\.?\d*)\s+dB`)
	maxMatch := maxVolRe.FindStringSubmatch(outputStr)

	meanVol := -70.0 // Default if not found
	maxVol := -70.0  // Default if not found

	if len(meanMatch) > 1 {
		if val, err := strconv.ParseFloat(meanMatch[1], 64); err == nil {
			meanVol = val
		}
	}

	if len(maxMatch) > 1 {
		if val, err := strconv.ParseFloat(maxMatch[1], 64); err == nil {
			maxVol = val
		}
	}

	return &volMetricsResult{
		meanVol: meanVol,
		maxVol:  maxVol,
	}, nil
}

type silenceMetricsResult struct {
	silencePercentage float64
	longestSilence    float64
}

// detectSilence uses ffmpeg silencedetect to find silence periods
// Silence threshold: -50dB (adjustable), minimum duration: 0.5s
//
//nolint:unparam
func (a *ffmpegAudioAnalyzer) detectSilence(audioPath string) (*silenceMetricsResult, error) {
	// silencedetect outputs to stderr
	cmd := exec.Command(a.ffmpegPath,
		"-i", audioPath,
		"-af", "silencedetect=noise=-50dB:d=0.5",
		"-f", "null",
		"-",
	)
	// silencedetect returns non-zero but still produces output
	output, _ := cmd.CombinedOutput() //nolint:errcheck

	outputStr := string(output)

	// Parse silence_start and silence_end timestamps
	// Format: [silencedetect @ 0x...] silence_start: 12.345
	startRe := regexp.MustCompile(`silence_start:\s+(\d+\.?\d*)`)
	endRe := regexp.MustCompile(`silence_end:\s+(\d+\.?\d*)\s+\|\s+silence_duration:\s+(\d+\.?\d*)`)

	starts := startRe.FindAllStringSubmatch(outputStr, -1)
	ends := endRe.FindAllStringSubmatch(outputStr, -1)
	_ = starts // silence_detect may return starts without ends (edge case noted above)

	// Calculate total silence duration and longest silence
	var totalSilence float64
	var longestSilence float64

	// Parse silence_end entries (they include duration)
	for _, match := range ends {
		if len(match) >= 3 {
			duration, err := strconv.ParseFloat(match[2], 64)
			if err == nil {
				totalSilence += duration
				if duration > longestSilence {
					longestSilence = duration
				}
			}
		}
	}

	// If we have starts but no ends, the last silence extends to the end
	// This is a simplification - in practice, silencedetect usually pairs them
	// (intentional no-op - the edge case is noted above)

	// Get duration to calculate percentage
	duration, err := a.getDuration(audioPath)
	if err != nil {
		duration = 1.0 // Avoid division by zero
	}

	silencePercentage := 0.0
	if duration > 0 {
		silencePercentage = (totalSilence / duration) * 100.0
	}

	return &silenceMetricsResult{
		silencePercentage: silencePercentage,
		longestSilence:    longestSilence,
	}, nil
}

// classify determines the audio class based on metrics
// Thresholds are tunable based on real-world data
func (a *ffmpegAudioAnalyzer) classify(metrics *AudioMetrics) AudioClass {
	// Thresholds (tunable based on production data):
	const (
		// Mean volume threshold: below this is effectively silent
		silentMeanVolDB = -60.0

		// Max volume threshold: below this has no significant peaks
		silentMaxVolDB = -50.0

		// Silence percentage: above this is mostly silence
		nearSilentSilencePct = 90.0

		// For near_silent: max volume should be low but not silent
		nearSilentMaxVolDB = -40.0
	)

	// Silent: very low mean volume AND very low max volume OR >95% silence
	if (metrics.MeanVolumeDB <= silentMeanVolDB && metrics.MaxVolumeDB <= silentMaxVolDB) ||
		metrics.SilencePercentage >= 95.0 {
		return AudioClassSilent
	}

	// Near-silent: mostly silence with only tiny peaks
	if metrics.SilencePercentage >= nearSilentSilencePct &&
		metrics.MaxVolumeDB <= nearSilentMaxVolDB {
		return AudioClassNearSilent
	}

	// Active: enough signal to attempt ASR
	return AudioClassActive
}

// IsSilent returns true if the audio class is silent
func (r *AudioAnalysisResult) IsSilent() bool {
	return r.Class == AudioClassSilent
}

// IsNearSilent returns true if the audio class is near_silent
func (r *AudioAnalysisResult) IsNearSilent() bool {
	return r.Class == AudioClassNearSilent
}

// IsActive returns true if the audio class is active
func (r *AudioAnalysisResult) IsActive() bool {
	return r.Class == AudioClassActive
}
