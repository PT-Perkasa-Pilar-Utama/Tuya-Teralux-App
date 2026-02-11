package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
)

// Summary generates professional meeting minutes from the provided text using the LLM.
func (u *RAGUsecase) Summary(text string, language string, context string, style string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}
	
	targetLangName := "Indonesian"
	if strings.ToLower(language) == "en" {
		targetLangName = "English"
	}

	prompt := fmt.Sprintf(`You are an expert Executive Assistant specializing in distilling the essence of meetings and conversations into professional summaries.
Your goal is to extract CONCLUSIONS, DECISIONS, and KEY TAKEAWAYS. 
STRICT RULE: DO NOT retell the story chronologically. DO NOT menceritakan ulang isi teks secara naratif. Focus on what was achieved or decided.

STRICT RULES:
1. BE CONCISE. Use bullet points for details.
2. Remove all filler words (e.g., "uh", "um", "like", "so").
3. OUTPUT LANGUAGE: You must write the summary in %s even if the input is in a different language.

CONTEXT: %s
STYLE: %s

OUTPUT STRUCTURE:
# [Meeting Title / Topik Utama]

## 1. Summary (Ringkasan)
Summarize the essence of the conversation in 1-2 concise sentences.

## 2. Key Points (Poin Penting)
- **[Topic A]**: Main conclusion or result regarding this topic.
- **[Topic B]**: ...

## 3. Decisions (Keputusan)
- [Decision 1]
- [Decision 2]
(If none, state "No specific decisions recorded")

## 4. Action Items (Tindak Lanjut)
- [ ] **[PIC/Owner]**: [Task Description] (Deadline if any)

Text: "%s"
Summary (%s):`, targetLangName, context, style, text, targetLangName)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	summary, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return "", err
	}

	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, model='%s'", language, len(summary), model)
	utils.LogDebug("RAG Summary Result: %q", summary)
	return strings.TrimSpace(summary), nil
}
