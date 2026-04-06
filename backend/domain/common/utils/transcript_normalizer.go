package utils

import (
	"regexp"
	"strings"
)

// NormalizeTranscript performs safe normalization on transcript text
// It fixes obvious punctuation/casing/noise WITHOUT paraphrasing or smoothing away uncertainty
// This is safer than full refinement for preserving original meaning
func NormalizeTranscript(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}

	// 1. Fix multiple spaces -> single space
	reSpace := regexp.MustCompile(`\s+`)
	text = reSpace.ReplaceAllString(text, " ")

	// 2. Fix common punctuation issues
	// Add space after punctuation if missing (but preserve intentional formatting)
	text = fixPunctuationSpacing(text)

	// 3. Fix capitalization at sentence starts (conservative)
	text = fixSentenceCapitalization(text)

	// 4. Remove filler words that are clearly transcription artifacts
	// But preserve uncertainty markers like "might", "possibly", "I think"
	text = removeFillerWords(text)

	// 5. Fix repeated words (common ASR error)
	text = fixRepeatedWords(text)

	return strings.TrimSpace(text)
}

// NormalizeUtterance applies normalization to a single utterance
func NormalizeUtterance(text string) string {
	return NormalizeTranscript(text)
}

// fixPunctuationSpacing adds proper spacing after punctuation
func fixPunctuationSpacing(text string) string {
	// Add space after comma if missing (but not before)
	reComma := regexp.MustCompile(`,(\S)`)
	text = reComma.ReplaceAllString(text, ", $1")

	// Add space after period if followed by non-space (end of sentence)
	rePeriod := regexp.MustCompile(`\.(\S)`)
	text = rePeriod.ReplaceAllString(text, ". $1")

	// Remove space before punctuation
	reBeforePunct := regexp.MustCompile(`\s+([,.!?;:])`)
	text = reBeforePunct.ReplaceAllString(text, "$1")

	return text
}

// fixSentenceCapitalization ensures sentences start with capital letters
func fixSentenceCapitalization(text string) string {
	// Very conservative: only capitalize first letter of text and after .!?
	if len(text) == 0 {
		return text
	}

	// Capitalize first character
	runes := []rune(text)
	if len(runes) > 0 && runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 'a' + 'A'
	}

	// Capitalize after sentence-ending punctuation
	reCap := regexp.MustCompile(`([.!?])\s+([a-z])`)
	text = string(runes)
	text = reCap.ReplaceAllStringFunc(text, func(match string) string {
		parts := reCap.FindStringSubmatch(match)
		if len(parts) == 3 {
			return parts[1] + " " + strings.ToUpper(parts[2])
		}
		return match
	})

	return text
}

// removeFillerWords removes common filler words that are typically ASR artifacts
// Preserves uncertainty markers that may affect meaning
func removeFillerWords(text string) string {
	// Common filler words to remove (context-dependent, use sparingly)
	fillerPatterns := []string{
		`\buh\b`,
		`\bum\b`,
		`\ber\b`,
		`\bmm\b`,
		`\buhm\b`,
	}

	for _, pattern := range fillerPatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	// Clean up extra spaces from removal
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return text
}

// fixRepeatedWords fixes common ASR error where words are repeated
func fixRepeatedWords(text string) string {
	// Go's regexp doesn't support backreferences, so we use a manual approach
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	result := make([]string, 0, len(words))
	result = append(result, words[0])

	for i := 1; i < len(words); i++ {
		prev := strings.ToLower(result[len(result)-1])
		curr := strings.ToLower(words[i])

		// Skip if current word is same as previous (handles exact repeats)
		if curr == prev {
			continue
		}

		// Skip if current word starts with previous word (handles slight variations)
		if len(curr) > len(prev) && strings.HasPrefix(curr, prev) {
			continue
		}

		result = append(result, words[i])
	}

	return strings.Join(result, " ")
}

// ShouldPreserveUncertainty checks if a phrase contains uncertainty markers
// that should NOT be smoothed away during normalization
func ShouldPreserveUncertainty(text string) bool {
	uncertaintyMarkers := []string{
		"might", "may", "could", "possibly", "perhaps",
		"i think", "i guess", "i suppose", "maybe",
		"should", "would", "could",
		"mungkin", "barangkali", "sepertinya", "kayaknya",
		"saya rasa", "saya pikir", "saya kira",
	}

	lowerText := strings.ToLower(text)
	for _, marker := range uncertaintyMarkers {
		if strings.Contains(lowerText, marker) {
			return true
		}
	}

	return false
}
