package utils

import (
	"testing"
)

func TestAudioAnalyzer_Classify(t *testing.T) {
	analyzer := NewAudioAnalyzer().(*ffmpegAudioAnalyzer)

	tests := []struct {
		name          string
		metrics       AudioMetrics
		expectedClass AudioClass
	}{
		{
			name: "silent_low_mean_and_max",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -70.0,
				MaxVolumeDB:       -60.0,
				SilencePercentage: 98.0,
				LongestSilenceSec: 4.5,
			},
			expectedClass: AudioClassSilent,
		},
		{
			name: "silent_high_silence_percentage",
			metrics: AudioMetrics{
				DurationSec:       10.0,
				MeanVolumeDB:      -55.0,
				MaxVolumeDB:       -45.0,
				SilencePercentage: 96.0,
				LongestSilenceSec: 9.0,
			},
			expectedClass: AudioClassSilent,
		},
		{
			name: "near_silent_mostly_silence",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -55.0,
				MaxVolumeDB:       -42.0,
				SilencePercentage: 92.0,
				LongestSilenceSec: 4.0,
			},
			expectedClass: AudioClassNearSilent,
		},
		{
			name: "near_silent_room_noise",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -52.0,
				MaxVolumeDB:       -41.0,
				SilencePercentage: 91.0,
				LongestSilenceSec: 3.5,
			},
			expectedClass: AudioClassNearSilent,
		},
		{
			name: "active_speech",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -35.0,
				MaxVolumeDB:       -20.0,
				SilencePercentage: 40.0,
				LongestSilenceSec: 1.0,
			},
			expectedClass: AudioClassActive,
		},
		{
			name: "active_clear_voice",
			metrics: AudioMetrics{
				DurationSec:       3.0,
				MeanVolumeDB:      -30.0,
				MaxVolumeDB:       -15.0,
				SilencePercentage: 20.0,
				LongestSilenceSec: 0.5,
			},
			expectedClass: AudioClassActive,
		},
		{
			name: "boundary_mean_volume",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -60.0, // Exactly at threshold
				MaxVolumeDB:       -50.0, // Exactly at threshold
				SilencePercentage: 94.0,
				LongestSilenceSec: 4.0,
			},
			expectedClass: AudioClassSilent, // At threshold, classified as silent
		},
		{
			name: "boundary_silence_percentage",
			metrics: AudioMetrics{
				DurationSec:       5.0,
				MeanVolumeDB:      -55.0,
				MaxVolumeDB:       -42.0,
				SilencePercentage: 90.0, // Exactly at threshold
				LongestSilenceSec: 4.0,
			},
			expectedClass: AudioClassNearSilent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.classify(&tt.metrics)
			if got != tt.expectedClass {
				t.Errorf("classify(metrics) = %v, want %v", got, tt.expectedClass)
			}
		})
	}
}

func TestAudioAnalysisResult_Helpers(t *testing.T) {
	tests := []struct {
		name           string
		class          AudioClass
		wantSilent     bool
		wantNearSilent bool
		wantActive     bool
	}{
		{
			name:           "silent_class",
			class:          AudioClassSilent,
			wantSilent:     true,
			wantNearSilent: false,
			wantActive:     false,
		},
		{
			name:           "near_silent_class",
			class:          AudioClassNearSilent,
			wantSilent:     false,
			wantNearSilent: true,
			wantActive:     false,
		},
		{
			name:           "active_class",
			class:          AudioClassActive,
			wantSilent:     false,
			wantNearSilent: false,
			wantActive:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudioAnalysisResult{Class: tt.class}

			if result.IsSilent() != tt.wantSilent {
				t.Errorf("IsSilent() = %v, want %v", result.IsSilent(), tt.wantSilent)
			}
			if result.IsNearSilent() != tt.wantNearSilent {
				t.Errorf("IsNearSilent() = %v, want %v", result.IsNearSilent(), tt.wantNearSilent)
			}
			if result.IsActive() != tt.wantActive {
				t.Errorf("IsActive() = %v, want %v", result.IsActive(), tt.wantActive)
			}
		})
	}
}

// Note: Integration tests for actual FFmpeg/ffprobe execution require test audio files
// and should be run separately with test fixtures. Example test structure:
//
// func TestAudioAnalyzer_Analyze_RealFile(t *testing.T) {
//     // Skip if ffmpeg not available
//     if _, err := exec.LookPath("ffmpeg"); err != nil {
//         t.Skip("ffmpeg not available")
//     }
//
//     analyzer := NewAudioAnalyzer()
//     result, err := analyzer.Analyze("testdata/silence_5s.wav")
//     if err != nil {
//         t.Fatalf("Analyze() error = %v", err)
//     }
//
//     if result.Class != AudioClassSilent {
//         t.Errorf("Class = %v, want %v", result.Class, AudioClassSilent)
//     }
//     if result.Metrics.SilencePercentage < 95.0 {
//         t.Errorf("SilencePercentage = %.1f, want >= 95", result.Metrics.SilencePercentage)
//     }
// }
