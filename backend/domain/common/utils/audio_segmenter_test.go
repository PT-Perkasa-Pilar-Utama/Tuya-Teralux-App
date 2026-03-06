package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSplitAudioSegments_DurationAndOverlap(t *testing.T) {
	// Create a dummy 12-second audio file using ffmpeg lavfi
	testDir := t.TempDir()
	inputPath := filepath.Join(testDir, "test_audio.wav")

	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "anullsrc=r=16000:cl=mono", "-t", "12", inputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create dummy audio: %v, output: %s", err, string(out))
	}

	// Test segmenting with 5s segments and 1s overlap
	segments, err := SplitAudioSegments(inputPath, 5, 1)
	if err != nil {
		t.Fatalf("SplitAudioSegments failed: %v", err)
	}

	// cleanup on defer
	defer CleanupSegments(segments)

	// Expected segments:
	// idx 0: starts 0
	// idx 1: starts 4 (5-1)
	// idx 2: starts 9 (10-1)
	// loop breaks at idx=3 (3*5=15 >= 12)
	// Total exactly 3 segments
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	for i, seg := range segments {
		if seg.Index != i {
			t.Errorf("Expected segment index %d, got %d", i, seg.Index)
		}

		info, err := os.Stat(seg.Path)
		if err != nil {
			t.Errorf("Segment file %d does not exist: %v", i, err)
		} else if info.Size() == 0 {
			t.Errorf("Segment file %d is empty", i)
		}
	}
}

func TestSplitAudioSegments_ShortEndSegment(t *testing.T) {
	// 5.5s audio with 5s segment, 2s overlap
	testDir := t.TempDir()
	inputPath := filepath.Join(testDir, "short_audio.wav")

	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "anullsrc=r=16000:cl=mono", "-t", "5.5", inputPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create dummy audio: %v", err)
	}

	segments, err := SplitAudioSegments(inputPath, 5, 2)
	if err != nil {
		t.Fatalf("SplitAudioSegments failed: %v", err)
	}
	defer CleanupSegments(segments)

	// idx 0: starts 0, dur=5+2=7 > total -> actually total is 5.5, start+dur(5)<total? false (5 < 5.5). dur=5.
	// idx 1: starts 5-2=3, loop break when idx=2 (10>=5.5)
	// So 2 segments
	if len(segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(segments))
	}
}
