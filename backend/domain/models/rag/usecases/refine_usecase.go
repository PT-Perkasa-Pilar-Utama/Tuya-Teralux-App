package usecases

import (
	"context"
	"fmt"
	commonUtils "sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	whisperdtos "sensio/domain/models/whisper/dtos"
	"sensio/domain/speech/providers"
	"sensio/domain/speech/utils"
	"strings"
	"time"
)

type RefineUseCase interface {
	RefineText(ctx context.Context, text string, lang string, args ...string) (string, error)
	NormalizeText(text string) string
	NormalizeUtterances(utterances []whisperdtos.Utterance) []whisperdtos.Utterance
}

// TODO(#dead-code): NormalizeText and NormalizeUtterances are currently unused.
// They provide safe punctuation/casing normalization WITHOUT semantic rewriting,
// which preserves uncertainty markers that full refine may smooth away.
// Integration path: Add opts.Normalize option to TranscribeOptions and call
// NormalizeUtterances() when opts.Normalize=true && opts.Refine=false.
// See: transcript_normalizer.go for implementation.

type refineUseCase struct {
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *commonUtils.Config
	skill            skills.Skill
	providerResolver providers.ProviderResolver
}

func NewRefineUseCase(llm skills.LLMClient, fallbackLLM skills.LLMClient, cfg *commonUtils.Config, skill skills.Skill, providerResolver providers.ProviderResolver) RefineUseCase {
	return &refineUseCase{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		skill:            skill,
		providerResolver: providerResolver,
	}
}

// RefineText improves the grammar and clarity of the transcribed text based on the detected language.
// Optional: pass macAddress as third argument for terminal-specific provider resolution
func (u *refineUseCase) RefineText(ctx context.Context, text string, lang string, args ...string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	textChars := len(text)
	startTime := time.Now()
	commonUtils.LogDebug("Refine: starting (lang=%s chars=%d)", lang, textChars)

	if u.skill == nil {
		return "", fmt.Errorf("refine skill not configured")
	}

	// Use centralized health-aware fallback chain with terminal preference if macAddress provided
	var result string
	var err error

	if len(args) > 0 && args[0] != "" {
		// Use terminal-specific provider preference
		macAddress := args[0]
		err = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx := &skills.SkillContext{
				Ctx:      ctx,
				Prompt:   text,
				Language: lang,
				LLM:      resolvedSet.LLM,
				Config:   u.config,
			}
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res.Message
			}
			return execErr
		})
	} else {
		// Use standard health-aware fallback
		err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx := &skills.SkillContext{
				Ctx:      ctx,
				Prompt:   text,
				Language: lang,
				LLM:      resolvedSet.LLM,
				Config:   u.config,
			}
			res, execErr := u.skill.Execute(skillCtx)
			if execErr == nil {
				result = res.Message
			}
			return execErr
		})
	}

	if err != nil {
		commonUtils.LogWarn("Refine: failed (lang=%s chars=%d duration=%s) err=%v", lang, textChars, time.Since(startTime), err)
		return "", err
	}

	commonUtils.LogDebug("Refine: completed (lang=%s chars=%d duration=%s output_chars=%d)", lang, textChars, time.Since(startTime), len(result))
	commonUtils.LogDebug("RAG Refine: lang='%s', original='%s', refined='%s'", lang, text, result)
	return result, nil
}

// NormalizeText performs safe normalization without semantic rewriting
// This preserves uncertainty markers and original wording while fixing punctuation/casing
func (u *refineUseCase) NormalizeText(text string) string {
	return utils.NormalizeTranscript(text)
}

// NormalizeUtterances applies normalization per-utterance instead of whole-transcript
// This preserves the structure while cleaning up each utterance
func (u *refineUseCase) NormalizeUtterances(utterances []whisperdtos.Utterance) []whisperdtos.Utterance {
	if len(utterances) == 0 {
		return utterances
	}

	normalized := make([]whisperdtos.Utterance, len(utterances))
	for i, u := range utterances {
		normalized[i] = whisperdtos.Utterance{
			SpeakerLabel: u.SpeakerLabel,
			StartMs:      u.StartMs,
			EndMs:        u.EndMs,
			Text:         utils.NormalizeTranscript(u.Text),
			Confidence:   u.Confidence,
		}
	}

	return normalized
}
