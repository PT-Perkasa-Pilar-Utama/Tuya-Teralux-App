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

	prompt := fmt.Sprintf(`You are an expert Executive Assistant specializing in creating concise, accurate, and professional meeting minutes from raw transcriptions.
Your goal is to transform the provided text into a structured Markdown document.

STRICT RULES:
1. Remove all filler words (e.g., "uh", "um", "like", "so").
2. Do not include casual banter or off-topic remarks unless they result in a decision.
3. Group discussions by topic.
4. Ensure every "Action Item" has a clear owner (if mentioned) and objective.
5. OUTPUT LANGUAGE: You must write the summary in %s even if the input is in a different language.

CONTEXT: %s
STYLE: %s

OUTPUT STRUCTURE:
# [Meeting Title/Subject]
**Date**: [Detected or Placeholder]
**Participants**: [List if mentioned]

## 1. Agenda & Objectives
Summarize the main purpose of the meeting in 1-2 sentences.

## 2. Key Discussion Points
- **[Topic A]**: Brief summary of arguments or information shared.
- **[Topic B]**: ...

## 3. Decisions Made
- [Decision 1]
- [Decision 2]

## 4. Action Items
- [ ] **[Owner]**: [Task Description] (Deadline if any)

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
