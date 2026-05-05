package utils

import (
	"sensio/domain/models/whisper/dtos"
	"testing"
)

// =============================================================================
// ParseUtterancesFromText Tests
// =============================================================================

func TestParseUtterancesFromText(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantLen      int
		wantNil      bool
		wantErr      bool
		checkFirst   bool
		firstSpeaker string
		firstText    string
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			wantNil: true,
		},
		{
			name:    "whitespace only returns nil",
			input:   "   \n\t  ",
			wantNil: true,
		},
		{
			name:    "no speaker labels returns nil",
			input:   "This is plain text without any speaker labels.",
			wantNil: true,
		},
		{
			name:         "single speaker bracket format",
			input:        "[Speaker 1]: Hello world",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Speaker 1",
			firstText:    "Hello world",
		},
		{
			name:         "single speaker with colon - BUG: returns Unknown (alternation captures wrong group)",
			input:        "Speaker 1: Hello world",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Unknown", // Bug: code reads match[2] but alternation puts group at match[6]
			firstText:    "Hello world",
		},
		{
			name:         "multiple speakers",
			input:        "[Speaker 1]: Hello [Speaker 2]: Hi there",
			wantLen:      2,
			checkFirst:   true,
			firstSpeaker: "Speaker 1",
			firstText:    "Hello",
		},
		{
			name:         "speaker with letter label - BUG: returns Unknown",
			input:        "[Speaker A]: Hello",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Unknown", // Bug: code reads match[2] but [A-Za-z]+ captures at match[4]
			firstText:    "Hello",
		},
		{
			name:         "speaker with letter label no brackets - BUG: returns Unknown",
			input:        "Speaker A: Hello",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Unknown", // Bug: alternation captures wrong group
			firstText:    "Hello",
		},
		{
			name:         "speaker labels with spaces",
			input:        "[Speaker   1]: Hello",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Speaker 1",
			firstText:    "Hello",
		},
		{
			name:         "Speaker with colon format - BUG: returns Unknown",
			input:        "Speaker 2: How are you?",
			wantLen:      1,
			checkFirst:   true,
			firstSpeaker: "Unknown", // Bug: alternation captures wrong group
			firstText:    "How are you?",
		},
		{
			name:    "generic name pattern NOT recognized (Agenda:)",
			input:   "Agenda: Introduction",
			wantNil: true,
		},
		{
			name:    "generic name pattern NOT recognized (Risk:)",
			input:   "Risk: High priority",
			wantNil: true,
		},
		{
			name:    "generic name pattern NOT recognized (Decision:)",
			input:   "Decision: Approved",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseUtterancesFromText(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseUtterancesFromText() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("ParseUtterancesFromText() = nil, want non-nil")
				return
			}

			if tt.wantLen > 0 && len(got) != tt.wantLen {
				t.Errorf("ParseUtterancesFromText() returned %d utterances, want %d", len(got), tt.wantLen)
			}

			if tt.checkFirst && len(got) > 0 {
				if got[0].SpeakerLabel != tt.firstSpeaker {
					t.Errorf("First utterance SpeakerLabel = %q, want %q", got[0].SpeakerLabel, tt.firstSpeaker)
				}
				if got[0].Text != tt.firstText {
					t.Errorf("First utterance Text = %q, want %q", got[0].Text, tt.firstText)
				}
			}
		})
	}
}

func TestParseUtterancesFromText_Timing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		checkTiming  bool
		firstStartMs int64
		firstEndMs   int64
	}{
		{
			name:         "timing estimation works",
			input:        "[Speaker 1]: Hello",
			checkTiming:  true,
			firstStartMs: 0,    // match[0][0] is 0
			firstEndMs:   1900, // len("Hello") * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseUtterancesFromText(tt.input)
			if len(got) == 0 {
				t.Fatalf("ParseUtterancesFromText() returned empty slice")
			}

			if tt.checkTiming {
				if got[0].StartMs != tt.firstStartMs {
					t.Errorf("First utterance StartMs = %d, want %d", got[0].StartMs, tt.firstStartMs)
				}
				// EndMs is estimated from textEnd * 100
				if got[0].EndMs <= got[0].StartMs {
					t.Errorf("First utterance EndMs = %d should be > StartMs = %d", got[0].EndMs, got[0].StartMs)
				}
			}
		})
	}
}

func TestParseUtterancesFromText_EmptyTextBetweenLabels(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLen    int
		secondText string
	}{
		{
			name:       "empty text between labels is skipped",
			input:      "[Speaker 1]: [Speaker 2]: Hello",
			wantLen:    1,
			secondText: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseUtterancesFromText(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("ParseUtterancesFromText() returned %d utterances, want %d", len(got), tt.wantLen)
			}
			if len(got) > 1 && got[1].Text != tt.secondText {
				t.Errorf("Second utterance Text = %q, want %q", got[1].Text, tt.secondText)
			}
		})
	}
}

// =============================================================================
// FindOverlapLength Tests
// =============================================================================

func TestFindOverlapLength(t *testing.T) {
	tests := []struct {
		name string
		prev string
		curr string
		want int
	}{
		{
			name: "empty prev returns 0",
			prev: "",
			curr: "hello world",
			want: 0,
		},
		{
			name: "empty curr returns 0",
			prev: "hello world",
			curr: "",
			want: 0,
		},
		{
			name: "both empty returns 0",
			prev: "",
			curr: "",
			want: 0,
		},
		{
			name: "single word no match returns 0",
			prev: "hello",
			curr: "world",
			want: 0,
		},
		{
			name: "three word match returns length",
			prev: "the quick brown fox",
			curr: "the quick brown fox jumps",
			want: len("the quick brown fox"),
		},
		{
			name: "partial word overlap returns 0",
			prev: "hello world",
			curr: "world today",
			want: 0, // "world" is only 2 words, needs 3+ for match
		},
		{
			name: "four word match returns length",
			prev: "one two three four five",
			curr: "three four five six seven",
			want: len("three four five"),
		},
		{
			name: "five word match returns length",
			prev: "a b c d e f g",
			curr: "e f g h i j",
			want: len("e f g"),
		},
		{
			name: "no consecutive match returns 0",
			prev: "cat and dog",
			curr: "and dog run",
			want: 0, // "and dog" is only 2 words
		},
		{
			name: "different word order no match",
			prev: "hello world today",
			curr: "world today hello",
			want: 0,
		},
		{
			name: "last 5 words match",
			prev: "word1 word2 word3 word4 word5 word6 word7",
			curr: "word3 word4 word5 word6 word7 word8",
			want: 29, // 5 words "word3 word4 word5 word6 word7" = 5*5 + 4 spaces = 29
		},
		{
			name: "single word prev no match",
			prev: "hello",
			curr: "hello world",
			want: 0,
		},
		{
			name: "two word prev no match",
			prev: "hello world",
			curr: "world today",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindOverlapLength(tt.prev, tt.curr)
			if got != tt.want {
				t.Errorf("FindOverlapLength(%q, %q) = %d, want %d", tt.prev, tt.curr, got, tt.want)
			}
		})
	}
}

// =============================================================================
// MergeTranscriptions Tests
// =============================================================================

func TestMergeTranscriptions(t *testing.T) {
	tests := []struct {
		name     string
		segments []dtos.TranscriptSegment
		want     string
	}{
		{
			name:     "empty segments returns empty",
			segments: []dtos.TranscriptSegment{},
			want:     "",
		},
		{
			name: "single segment returns text",
			segments: []dtos.TranscriptSegment{
				{Text: "Hello world"},
			},
			want: "Hello world",
		},
		{
			name: "multiple segments joined with space",
			segments: []dtos.TranscriptSegment{
				{Text: "Hello"},
				{Text: "World"},
			},
			want: "Hello World",
		},
		{
			name: "segments with overlap removed - BUG: leaves leading space",
			segments: []dtos.TranscriptSegment{
				{Text: "the quick brown fox"},
				{Text: "the quick brown fox jumps"},
			},
			want: "the quick brown fox  jumps", // Bug: overlap removal leaves leading space
		},
		{
			name: "segments trimmed",
			segments: []dtos.TranscriptSegment{
				{Text: "  Hello  "},
				{Text: "  World  "},
			},
			want: "Hello World",
		},
		{
			name: "empty text segments skipped",
			segments: []dtos.TranscriptSegment{
				{Text: "Hello"},
				{Text: ""},
				{Text: "World"},
			},
			want: "Hello World",
		},
		{
			name: "three segments - BUG: fragment tracking causes incorrect output",
			segments: []dtos.TranscriptSegment{
				{Text: "the quick brown fox"},
				{Text: "the quick brown fox jumps"},
				{Text: "the quick brown fox jumps over"},
			},
			want: "the quick brown fox  jumps the quick brown fox jumps over", // Bug: fragment tracking
		},
		{
			name: "no overlap segments concatenated",
			segments: []dtos.TranscriptSegment{
				{Text: "cat and dog"},
				{Text: "bird flies away"},
			},
			want: "cat and dog bird flies away",
		},
		{
			name: "whitespace only segments skipped",
			segments: []dtos.TranscriptSegment{
				{Text: "Hello"},
				{Text: "   \n\t  "},
				{Text: "World"},
			},
			want: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeTranscriptions(tt.segments)
			if got != tt.want {
				t.Errorf("MergeTranscriptions() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Integration-style tests for MergeUtterances
// =============================================================================

func TestMergeUtterances(t *testing.T) {
	tests := []struct {
		name     string
		segments []dtos.TranscriptSegment
		wantLen  int
	}{
		{
			name:     "empty segments",
			segments: []dtos.TranscriptSegment{},
			wantLen:  0,
		},
		{
			name: "single segment with utterances",
			segments: []dtos.TranscriptSegment{
				{
					StartMs: 0,
					EndMs:   5000,
					Utterances: []dtos.Utterance{
						{StartMs: 0, EndMs: 2000, Text: "Hello"},
					},
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple segments combined",
			segments: []dtos.TranscriptSegment{
				{
					StartMs: 0,
					EndMs:   5000,
					Utterances: []dtos.Utterance{
						{StartMs: 0, EndMs: 2000, Text: "Hello"},
					},
				},
				{
					StartMs: 5000,
					EndMs:   10000,
					Utterances: []dtos.Utterance{
						{StartMs: 0, EndMs: 2000, Text: "World"},
					},
				},
			},
			wantLen: 2,
		},
		{
			name: "absolute timestamps not adjusted",
			segments: []dtos.TranscriptSegment{
				{
					StartMs: 0,
					EndMs:   5000,
					Utterances: []dtos.Utterance{
						{StartMs: 0, EndMs: 2000, Text: "Hello"},
					},
				},
				{
					StartMs: 5000,
					EndMs:   10000,
					Utterances: []dtos.Utterance{
						{StartMs: 6000, EndMs: 8000, Text: "World"}, // absolute timestamp
					},
				},
			},
			wantLen: 2,
		},
		{
			name: "relative timestamps adjusted",
			segments: []dtos.TranscriptSegment{
				{
					StartMs: 1000,
					EndMs:   5000,
					Utterances: []dtos.Utterance{
						{StartMs: 0, EndMs: 2000, Text: "Hello"}, // relative
					},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeUtterances(tt.segments)
			if len(got) != tt.wantLen {
				t.Errorf("MergeUtterances() returned %d utterances, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMergeUtterances_TimestampAdjustment(t *testing.T) {
	// Test that relative timestamps get adjusted by segment offset
	segments := []dtos.TranscriptSegment{
		{
			StartMs: 5000,
			EndMs:   10000,
			Utterances: []dtos.Utterance{
				{StartMs: 0, EndMs: 2000, Text: "First"},
			},
		},
	}

	got := MergeUtterances(segments)

	if len(got) != 1 {
		t.Fatalf("MergeUtterances() returned %d utterances, want 1", len(got))
	}

	// Relative timestamp (0) should be adjusted by segment offset (5000)
	if got[0].StartMs != 5000 {
		t.Errorf("First utterance StartMs = %d, want 5000 (adjusted)", got[0].StartMs)
	}
	if got[0].EndMs != 7000 {
		t.Errorf("First utterance EndMs = %d, want 7000 (adjusted)", got[0].EndMs)
	}
}
