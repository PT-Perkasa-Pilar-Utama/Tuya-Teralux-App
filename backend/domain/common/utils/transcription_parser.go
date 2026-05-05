package utils

import (
	"regexp"
	"sensio/domain/models/whisper/dtos"
	"strconv"
	"strings"
)

// ParseUtterancesFromText extracts structured utterances from plain-text transcription
// that contains speaker labels like "[Speaker 1]", "Speaker 1:", etc.
// This is a best-effort parser for providers that don't return structured diarization.
// Returns nil if no speaker labels are detected (do NOT fabricate utterances).
//
// NOTE: Removed generic `([A-Z][a-z]+):` pattern to avoid false positives from
// headings like "Agenda:", "Risk:", "Decision:", etc. Only explicit speaker
// labels ([Speaker N], Speaker N:) are recognized.
func ParseUtterancesFromText(transcription string) []dtos.Utterance {
	if strings.TrimSpace(transcription) == "" {
		return nil
	}

	var utterances []dtos.Utterance

	// High-confidence speaker label patterns (explicit diarization markers)
	patterns := []string{
		`\[Speaker\s*(\d+)\]:?`,       // [Speaker 1]:
		`\[Speaker\s*([A-Za-z]+)\]:?`, // [Speaker A]:
		`Speaker\s*(\d+):`,            // Speaker 1:
		`Speaker\s*([A-Za-z]+):`,      // Speaker A:
		// NOTE: Generic name pattern removed to avoid false positives
	}

	// Combine patterns into one regex
	combinedPattern := strings.Join(patterns, "|")
	re := regexp.MustCompile(combinedPattern)

	// Find all speaker label matches with their positions
	matches := re.FindAllStringSubmatchIndex(transcription, -1)
	if len(matches) == 0 {
		// NO speaker labels found - return nil, do NOT fabricate utterances
		// This prevents false-positive diarization reporting
		return nil
	}

	// Extract utterances between speaker labels
	for i, match := range matches {
		startIdx := match[0]
		endIdx := match[1]

		// Extract speaker label
		labelStart := match[2]
		labelEnd := match[3]
		speakerLabel := "Unknown"
		if labelStart >= 0 && labelEnd >= 0 {
			speakerNum := transcription[labelStart:labelEnd]
			// Normalize to "Speaker X" format
			if num, err := strconv.Atoi(speakerNum); err == nil {
				speakerLabel = "Speaker " + strconv.Itoa(num)
			} else {
				speakerLabel = "Speaker " + speakerNum
			}
		}

		// Extract text for this utterance (from end of this label to start of next label or end)
		textStart := endIdx
		textEnd := len(transcription)
		if i+1 < len(matches) {
			textEnd = matches[i+1][0]
		}

		text := strings.TrimSpace(transcription[textStart:textEnd])
		if text == "" {
			continue
		}

		// Estimate timing (rough heuristic: 1 char ≈ 100ms, 10 chars per second)
		// Cast to int64 BEFORE multiplication to avoid overflow
		startMs := int64(startIdx) * 100
		endMs := int64(textEnd) * 100

		utterances = append(utterances, dtos.Utterance{
			SpeakerLabel: speakerLabel,
			StartMs:      startMs,
			EndMs:        endMs,
			Text:         text,
			Confidence:   0.0, // Unknown without provider metadata
		})
	}

	return utterances
}

// BuildConfidenceSummary calculates confidence statistics from utterances
func BuildConfidenceSummary(utterances []dtos.Utterance, segmentsCount int) *dtos.ConfidenceSummary {
	if len(utterances) == 0 {
		return &dtos.ConfidenceSummary{
			SegmentsCount:   segmentsCount,
			UtterancesCount: 0,
		}
	}

	var sum, min, max float64
	min = 1.0
	hasConfidence := false

	for _, u := range utterances {
		if u.Confidence > 0 {
			sum += u.Confidence
			hasConfidence = true
			if u.Confidence < min {
				min = u.Confidence
			}
			if u.Confidence > max {
				max = u.Confidence
			}
		}
	}

	avg := 0.0
	if hasConfidence {
		avg = sum / float64(len(utterances))
	}

	return &dtos.ConfidenceSummary{
		AverageConfidence: avg,
		MinConfidence:     min,
		MaxConfidence:     max,
		SegmentsCount:     segmentsCount,
		UtterancesCount:   len(utterances),
	}
}

// MergeUtterances merges utterances from multiple segments into a single ordered list
// while adjusting timestamps to account for segment offsets.
func MergeUtterances(segments []dtos.TranscriptSegment) []dtos.Utterance {
	var allUtterances []dtos.Utterance

	for _, seg := range segments {
		if len(seg.Utterances) > 0 {
			// Adjust timestamps based on segment start time
			segmentOffset := seg.StartMs
			for _, u := range seg.Utterances {
				// Only adjust if utterance timestamps are relative to segment
				// If they're already absolute, skip adjustment
				if u.StartMs < segmentOffset {
					u.StartMs += segmentOffset
					u.EndMs += segmentOffset
				}
				allUtterances = append(allUtterances, u)
			}
		}
	}

	return allUtterances
}

// MergeTranscriptions merges text from multiple segments, handling overlap regions
// to avoid duplicate text at boundaries.
func MergeTranscriptions(segments []dtos.TranscriptSegment) string {
	if len(segments) == 0 {
		return ""
	}

	if len(segments) == 1 {
		return segments[0].Text
	}

	var texts []string
	for i, seg := range segments {
		text := strings.TrimSpace(seg.Text)
		if text == "" {
			continue
		}

		// Check for overlap with previous segment
		if i > 0 && len(texts) > 0 {
			prevText := texts[len(texts)-1]
			overlap := findOverlap(prevText, text)
			if overlap > 0 {
				// Remove overlapping portion from current text
				text = text[overlap:]
			}
		}

		texts = append(texts, text)
	}

	return strings.Join(texts, " ")
}

// findOverlap finds the length of overlapping text between end of prev and start of curr
func findOverlap(prev, curr string) int {
	// Simple heuristic: check if last 20-50 chars of prev match first 20-50 chars of curr
	prevWords := strings.Fields(prev)
	currWords := strings.Fields(curr)

	if len(prevWords) == 0 || len(currWords) == 0 {
		return 0
	}

	// Check last 3-5 words of prev against first 3-5 words of curr
	maxCheck := 5
	if len(prevWords) < maxCheck {
		maxCheck = len(prevWords)
	}

	for n := maxCheck; n >= 3; n-- {
		prevEnd := prevWords[len(prevWords)-n:]
		currStart := currWords[:n]

		if matchWords(prevEnd, currStart) {
			// Calculate character offset
			overlapText := strings.Join(prevEnd, " ")
			return len(overlapText)
		}
	}

	return 0
}

// matchWords checks if two word slices are identical
func matchWords(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FindOverlapLength finds the character length of overlapping text between end of prev and start of curr
// This is a simpler character-based version for use in transcription merging
func FindOverlapLength(prev, curr string) int {
	if len(prev) == 0 || len(curr) == 0 {
		return 0
	}

	prevWords := strings.Fields(prev)
	currWords := strings.Fields(curr)

	if len(prevWords) == 0 || len(currWords) == 0 {
		return 0
	}

	// Check last 3-5 words of prev against first 3-5 words of curr
	maxCheck := 5
	if len(prevWords) < maxCheck {
		maxCheck = len(prevWords)
	}

	for n := maxCheck; n >= 3; n-- {
		prevEnd := prevWords[len(prevWords)-n:]
		currStart := currWords[:n]

		if matchWords(prevEnd, currStart) {
			// Calculate character offset
			overlapText := strings.Join(prevEnd, " ")
			return len(overlapText)
		}
	}

	return 0
}
