package utils

import (
	"regexp"
	"strings"
)

// TranscriptValidationResult holds the result of transcript validation
type TranscriptValidationResult struct {
	OriginalText    string // Original transcript from provider
	NormalizedText  string // Normalized for comparison
	IsValid         bool   // Whether transcript should be used
	RejectionReason string // Reason for rejection if invalid
	AudioClass      string // Audio class that produced this transcript
	IsCommandLike   bool   // Whether transcript looks like a command
}

// TranscriptValidator validates transcripts for hallucinations and low-quality output
type TranscriptValidator interface {
	Validate(transcript string, audioClass string) *TranscriptValidationResult
}

type transcriptValidator struct {
	// Known hallucination phrases (lowercase for comparison)
	hallucinationPhrases []string

	// Subtitle/credit patterns (regex)
	creditPatterns []*regexp.Regexp

	// Smart home / command keywords (lowercase)
	commandKeywords []string

	// Interrogative control phrases
	controlPhrases []string
}

// NewTranscriptValidator creates a new transcript validator
func NewTranscriptValidator() TranscriptValidator {
	return &transcriptValidator{
		hallucinationPhrases: []string{
			// Gratitude / thanks (Indonesian + English)
			"terima kasih",
			"thank you",
			"thanks",
			"thanks for watching",
			"thank you for watching",
			"terima kasih telah menonton",
			"terima kasih sudah menonton",
			"terima kasih mononton", // ASR typo

			// Subscription / engagement prompts
			"jangan lupa subscribe",
			"jangan lupa untuk subscribe",
			"don't forget to subscribe",
			"like comment share",
			"like komen share",
			"like, comment, and share",
			"like, komen, dan share",
			"subscribe like komen",
			"subscribe, like, komen",
			"subscribe, like, comment",
			"jangan lupa like dan share",
			"jangan lupa komen",

			// Closing / outro phrases
			"sampai jumpa di video",
			"see you next time",
			"see you in the next",
			"selamat menonton",
			"selamat menikmati",
			"goodbye",
			"bye",

			// Credit / attribution
			"translated by",
			"subtitles by",
			"diterjemahkan oleh",
			"subtitle oleh",

			// Follow / visit prompts
			"follow us on",
			"follow kami di",
			"kunjungi channel",
			"visit our channel",
			"cek deskripsi",
			"link di deskripsi",
			"link in the description",
			"check the description",

			// Bell / notification prompts
			"hit the bell",
			"tekan tombol lonceng",
			"aktifkan notifikasi",
			"turn on notifications",
			"pasang notifikasi",

			// Generic courtesy from low-signal audio
			"halo",
			"hai",
			"hi",
			"oke",
			"okay",
			"ya",
			"yes",
			"no",
			"tidak",
		},

		creditPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)subscribe\s+\d+`), // "subscribe 1000"
			regexp.MustCompile(`(?i)subscriber`),
			regexp.MustCompile(`(?i)views?`),
			regexp.MustCompile(`(?i)watch time`),
			regexp.MustCompile(`(?i)youtube\s+studio`),
			regexp.MustCompile(`(?i)analytics`),
			regexp.MustCompile(`(?i)monetisasi`),
			regexp.MustCompile(`(?i)monetization`),
			regexp.MustCompile(`(?i)translated\s+by`), // "translated by John"
			regexp.MustCompile(`(?i)subtitles?\s+by`), // "subtitles by"
		},

		commandKeywords: []string{
			// Device control (Indonesian + English)
			"lampu", "light", "lights",
			"ac", "air conditioner",
			"kipas", "fan",
			"tv", "television",
			"speaker",
			"curtain", "gorden", "tirai",
			"door", "pintu",
			"lock", "kunci",
			"switch", "saklar",
			"socket", "outlet",

			// Actions
			"nyalakan", "hidupkan", "turn on", "activate",
			"matikan", "turn off", "deactivate",
			"setel", "set", "atur", "adjust",
			"ubah", "change",
			"naikkan", "increase", "turn up",
			"turunkan", "decrease", "turn down",
			"brighten", "terangkan",
			"dim", "redupkan",

			// Smart home context
			"sensio",
			"asisten", "assistant",
			"smart home",
			"ruangan", "room",
			"mode",

			// Meeting / productivity
			"rapat", "meeting",
			"notulen", "minutes",
			"summary", "rangkum", "ringkas", "summarize",
			"terjemah", "translate", "translation",
			"rekam", "record",
			"audio",
		},

		controlPhrases: []string{
			"berapa", "how much", "how many", "what is",
			"apa", "siapa", "who", "where", "kapan", "when",
			"kenapa", "why", "bagaimana", "how",
			"tolong", "please",
			"bisa", "can", "could",
			"tolong nyalakan", "please turn on",
			"tolong matikan", "please turn off",
		},
	}
}

// Validate checks if a transcript is valid or should be rejected
func (v *transcriptValidator) Validate(transcript string, audioClass string) *TranscriptValidationResult {
	result := &TranscriptValidationResult{
		OriginalText: transcript,
		AudioClass:   audioClass,
	}

	// Step 1: Normalize transcript
	normalized := v.normalize(transcript)
	result.NormalizedText = normalized

	// Step 2: Check if empty after normalization
	if normalized == "" {
		result.IsValid = false
		result.RejectionReason = "empty_after_normalization"
		return result
	}

	// Step 3: Check if command-like (preserve even if short)
	result.IsCommandLike = v.isCommandLike(normalized)

	// Step 4: Check subtitle/credit patterns FIRST (these should never be preserved)
	if v.matchesCreditPattern(normalized) {
		result.IsValid = false
		result.RejectionReason = "subtitle_credit_pattern"
		return result
	}

	// Step 5: Check known hallucination phrases
	// Exception: if it's command-like, preserve it (unless it matched credit patterns above)
	// Also preserve short generic words for active audio (e.g., "halo", "oke" spoken clearly)
	if v.matchesHallucinationPhrase(normalized) {
		if !result.IsCommandLike {
			// For active audio, only reject longer phrases, not short generic words
			if audioClass == "active" && v.isShortGenericWord(normalized) {
				// Allow short generic words for active audio
			} else {
				result.IsValid = false
				result.RejectionReason = "known_hallucination_phrase"
				return result
			}
		}
	}

	// Step 6: Check low-signal generic short phrase
	if audioClass == "near_silent" && v.isGenericShortPhrase(normalized) {
		result.IsValid = false
		result.RejectionReason = "low_signal_generic_phrase"
		return result
	}

	// Passed all checks
	result.IsValid = true
	return result
}

// normalize prepares transcript for comparison
func (v *transcriptValidator) normalize(text string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Lowercase
	text = strings.ToLower(text)

	// Collapse repeated whitespace
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	// Strip surrounding punctuation-only noise
	// Remove leading/trailing punctuation
	text = strings.Trim(text, ".,!?;:'\"()-[]{}")

	// Remove standalone punctuation
	text = spaceRe.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// isCommandLike checks if transcript contains command-like vocabulary
func (v *transcriptValidator) isCommandLike(normalized string) bool {
	// Check for command keywords
	for _, kw := range v.commandKeywords {
		if strings.Contains(normalized, kw) {
			return true
		}
	}

	// Check for interrogative control phrases
	for _, phrase := range v.controlPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}

	// Check for device action patterns (verb + noun)
	// e.g., "nyalakan lampu", "turn on light"
	verbPatterns := []string{"nyalakan", "matikan", "hidupkan", "turn on", "turn off", "setel", "atur"}
	for _, verb := range verbPatterns {
		if strings.Contains(normalized, verb) {
			// Has verb, likely a command
			return true
		}
	}

	return false
}

// matchesHallucinationPhrase checks for known hallucination phrases
func (v *transcriptValidator) matchesHallucinationPhrase(normalized string) bool {
	for _, phrase := range v.hallucinationPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

// matchesCreditPattern checks for subtitle/credit/outro patterns
func (v *transcriptValidator) matchesCreditPattern(normalized string) bool {
	for _, pattern := range v.creditPatterns {
		if pattern.MatchString(normalized) {
			return true
		}
	}
	return false
}

// isGenericShortPhrase checks if transcript is a generic short phrase
// Used for near_silent audio classification
func (v *transcriptValidator) isGenericShortPhrase(normalized string) bool {
	// Count words
	words := strings.Fields(normalized)
	if len(words) > 5 {
		return false // Not short
	}

	// Check if it's just a generic courtesy phrase
	genericPhrases := []string{
		"terima kasih", "thank you", "thanks",
		"halo", "hai", "hi",
		"oke", "okay", "ya", "yes",
		"no", "tidak",
		"selamat", "goodbye", "bye",
	}

	for _, phrase := range genericPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}

	// Very short (1-2 words) without command keywords
	if len(words) <= 2 && !v.isCommandLike(normalized) {
		return true
	}

	return false
}

// isShortGenericWord checks if transcript is a very short generic word (1-2 chars or common greeting)
// Used to preserve legitimate active-speech greetings
func (v *transcriptValidator) isShortGenericWord(normalized string) bool {
	words := strings.Fields(normalized)
	if len(words) > 1 {
		return false // Not a single word
	}

	// Short generic words that are OK for active audio
	shortGenericWords := []string{
		"halo", "hai", "hi",
		"oke", "okay",
		"ya", "yes",
		"no", "tidak",
		"bye",
	}

	for _, word := range shortGenericWords {
		if normalized == word {
			return true
		}
	}

	return false
}

// Helper functions for convenience

// IsValidTranscript quickly checks if a transcript is valid
func IsValidTranscript(transcript string, audioClass string) bool {
	validator := NewTranscriptValidator()
	result := validator.Validate(transcript, audioClass)
	return result.IsValid
}

// NormalizeTranscriptForValidation normalizes a transcript for comparison (validator-specific)
// Note: This is different from utils.NormalizeTranscript which does punctuation/casing fixes.
// This function is specifically for hallucination detection (lowercase, trim, collapse).
func NormalizeTranscriptForValidation(transcript string) string {
	validator := NewTranscriptValidator()
	return validator.(*transcriptValidator).normalize(transcript)
}
