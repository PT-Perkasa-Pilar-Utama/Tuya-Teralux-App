package orchestrator

import (
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	"strings"
)

// GuardResult represents the classification of a prompt by the guard.
type GuardResult int

const (
	GuardClean           GuardResult = iota // Normal input, proceed
	GuardPureSpam                           // Block entirely (no UI)
	GuardDialogWithPromo                    // Show request, respond with identity
	GuardIrrelevant                         // Rambling/Irrelevant, block
)

// GuardOrchestrator detects promotional spam from voice commands or typed input.
type GuardOrchestrator struct {
	guardSkill skills.Skill
}

func NewGuardOrchestrator(guardSkill skills.Skill) *GuardOrchestrator {
	return &GuardOrchestrator{guardSkill: guardSkill}
}

// spamPatterns are exact phrases that indicate pure promotional text.
var spamPatterns = []string{
	"terima kasih telah menonton",
	"terima kasih sudah menonton",
	"terima kasih mononton", // typo common in ASR
	"thanks for watching",
	"thank you for watching",
	"jangan lupa subscribe",
	"jangan lupa untuk subscribe",
	"don't forget to subscribe",
	"like comment share",
	"like komen share",
	"like, comment, and share",
	"like, komen, dan share",
	"subscribe like komen",
	"subscribe, like, komen",
	"sampai jumpa di video",
	"see you next time",
	"see you in the next",
	"follow us on",
	"follow kami di",
	"kunjungi channel",
	"visit our channel",
	"hit the bell",
	"tekan tombol lonceng",
	"aktifkan notifikasi",
	"turn on notifications",
	"link di deskripsi",
	"link in the description",
	"check the description",
	"cek deskripsi",
}

// spamKeywords are individual keywords that, when combined, suggest spam content.
var spamKeywords = []string{
	"subscribe",
	"subscriber",
	"unsubscribe",
	"dislike",
	"komen",
	"notification",
	"notifikasi",
	"lonceng",
	"bell",
	"channel",
	"menonton",
	"watching",
}

// sensitiveTopicKeywords indicate off-topic political, religious, or controversial topics.
var sensitiveTopicKeywords = []string{
	// Politics (ID)
	"jokowi", "prabowo", "megawati", "anies", "ganjar", "gibran",
	"presiden", "menteri", "dpr", "dprd", "mpr",
	"apbn", "apbd", "korupsi", "pemilu", "pilkada", "pilpres",
	"partai", "parpol", "golkar", "pdip", "gerindra", "nasdem", "demokrat", "pks", "psi",
	"undang-undang", "ruu", "perppu",
	"makan bergizi gratis",
	// Politics (EN)
	"trump", "biden", "obama", "putin", "election", "congress", "senate",
	"republican", "democrat", "impeach",
	// Religion
	"agama", "islam", "kristen", "hindu", "buddha", "katolik",
	"gereja", "masjid", "vihara", "pura",
	"alkitab", "alquran", "injil", "hadits",
	"haram", "halal", "kafir", "murtad",
	// Controversial
	"lgbt", "aborsi", "abortion", "narkoba", "terorisme", "teroris",
	"ijazah", "ijasah",
	"sawit", "nyawit",
	// War & Conflict
	"perang", "war", "invasi", "invasion", "bom", "bombing",
	"militer", "military", "tentara", "army", "senjata", "weapon",
	"nuklir", "nuclear", "rudal", "missile", "sanksi", "sanction",
	"gaza", "palestina", "palestine", "israel", "ukraine", "ukraina", "rusia", "russia",
	"nato", "isis", "taliban",
	// Celebrity Gossip
	"artis", "selebriti", "celebrity", "gosip", "skandal",
}

// smartHomeKeywords indicate the prompt IS relevant to Sensio's scope.
var smartHomeKeywords = []string{
	"lampu", "light", "ac", "kipas", "fan", "tv", "speaker",
	"nyalakan", "matikan", "hidupkan", "turn on", "turn off",
	"suhu", "temperature", "brightness", "kecerahan",
	"perangkat", "device", "sensor", "smart home",
	"sensio", "asisten", "assistant",
	"rapat", "meeting", "notulen", "summary", "rangkum", "ringkas",
	"terjemah", "translate", "translation",
	"rekam", "record", "audio", "transcri",
}

// CheckPrompt analyzes a prompt and returns a GuardResult.
// Uses a fast keyword path first, then falls back to LLM for ALL unmatched prompts.
func (g *GuardOrchestrator) CheckPrompt(ctx *skills.SkillContext) GuardResult {
	prompt := strings.TrimSpace(ctx.Prompt)
	if prompt == "" {
		return GuardClean
	}

	promptLower := strings.ToLower(prompt)
	wordCount := len(strings.Fields(prompt))

	// --- Layer 1: Fast Spam Pattern Matching ---
	for _, pattern := range spamPatterns {
		if strings.Contains(promptLower, pattern) {
			if wordCount <= 15 {
				utils.LogInfo("Guard: Fast-match SPAM: '%s'", prompt)
				return GuardPureSpam
			}
			utils.LogInfo("Guard: Fast-match DIALOG_WITH_PROMO: '%s'", prompt)
			return GuardDialogWithPromo
		}
	}

	// Spam keyword density check
	spamKeywordCount := 0
	for _, kw := range spamKeywords {
		if strings.Contains(promptLower, kw) {
			spamKeywordCount++
		}
	}

	if spamKeywordCount >= 3 && wordCount <= 20 {
		utils.LogInfo("Guard: Keyword density SPAM (%d kw, %d words): '%s'", spamKeywordCount, wordCount, prompt)
		return GuardPureSpam
	}
	if spamKeywordCount >= 2 && wordCount <= 10 {
		utils.LogInfo("Guard: Keyword density SPAM (%d kw, %d words): '%s'", spamKeywordCount, wordCount, prompt)
		return GuardPureSpam
	}
	if spamKeywordCount >= 2 && wordCount > 20 {
		utils.LogInfo("Guard: Keyword density DIALOG_WITH_PROMO: '%s'", prompt)
		return GuardDialogWithPromo
	}

	// --- Layer 2: Sensitive Topic Fast-Match ---
	sensitiveCount := 0
	for _, kw := range sensitiveTopicKeywords {
		if strings.Contains(promptLower, kw) {
			sensitiveCount++
		}
	}

	if sensitiveCount >= 1 {
		// Check if there's ALSO a smart home keyword (e.g., "jokowi bilang matikan lampu")
		for _, kw := range smartHomeKeywords {
			if strings.Contains(promptLower, kw) {
				utils.LogInfo("Guard: Sensitive topic but has smart home intent, allowing: '%s'", prompt)
				return GuardClean
			}
		}

		// Pure sensitive topic, no smart home relevance → irrelevant
		utils.LogInfo("Guard: Sensitive topic detected (%d hits): '%s'", sensitiveCount, prompt)
		return GuardIrrelevant
	}

	// --- Layer 3: LLM Classification (ALL remaining unmatched prompts) ---
	// If the prompt clearly has smart home keywords → skip LLM, it's clean
	for _, kw := range smartHomeKeywords {
		if strings.Contains(promptLower, kw) {
			return GuardClean
		}
	}

	// For everything else → ask the LLM to classify
	if g.guardSkill != nil {
		utils.LogDebug("Guard: No fast-match, using LLM for: '%s'", prompt)
		result, err := g.guardSkill.Execute(ctx)
		if err != nil {
			utils.LogWarn("Guard: LLM classification failed, defaulting to CLEAN: %v", err)
			return GuardClean
		}

		classification := strings.TrimSpace(strings.ToUpper(result.Message))
		switch classification {
		case "SPAM":
			utils.LogInfo("Guard: LLM classified as SPAM: '%s'", prompt)
			return GuardPureSpam
		case "IRRELEVANT":
			utils.LogInfo("Guard: LLM classified as IRRELEVANT: '%s'", prompt)
			return GuardIrrelevant
		case "DIALOG":
			utils.LogInfo("Guard: LLM classified as DIALOG: '%s'", prompt)
			return GuardDialogWithPromo
		default:
			return GuardClean
		}
	}

	return GuardClean
}

// IdentityResponse returns the Sensio identity message for dialog-with-promo scenarios.
func (g *GuardOrchestrator) IdentityResponse(language string) string {
	if strings.EqualFold(language, "en") || strings.EqualFold(language, "english") {
		return "Hi! I'm Sensio, your smart home assistant. I can help you control devices, summarize meetings, and answer questions about your smart home. How can I help you today?"
	}
	return "Hai! Saya Sensio, asisten rumah pintar kamu. Saya bisa bantu kontrol perangkat, merangkum rapat, dan menjawab pertanyaan seputar smart home kamu. Ada yang bisa saya bantu?"
}
