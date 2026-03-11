package orchestrator

import (
	"encoding/json"
	"fmt"
	"sensio/domain/models/rag/skills"
	tuyaDtos "sensio/domain/tuya/dtos"
	"strings"
)

// BaseOrchestrator provides default implementation for placeholder replacement and LLM execution.
type BaseOrchestrator struct{}

func NewBaseOrchestrator() *BaseOrchestrator {
	return &BaseOrchestrator{}
}

func renderRegisteredDevices(ctx *skills.SkillContext) string {
	if ctx == nil || ctx.Vector == nil || strings.TrimSpace(ctx.UID) == "" {
		return "No devices connected."
	}

	aggKey := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
	aggJSON, ok := ctx.Vector.Get(aggKey)
	if !ok || strings.TrimSpace(aggJSON) == "" {
		return "No devices connected."
	}

	var agg tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &agg); err != nil {
		return "No devices connected."
	}
	if len(agg.Devices) == 0 {
		return "No devices connected."
	}

	names := make([]string, 0, len(agg.Devices))
	for _, d := range agg.Devices {
		names = append(names, fmt.Sprintf("- %s (ID: %s)", d.Name, d.ID))
	}
	return strings.Join(names, "\n")
}

func (b *BaseOrchestrator) Execute(ctx *skills.SkillContext, prompt string) (*skills.SkillResult, error) {
	// 1. Identify model to use
	model := "high"
	// Get skill name from context if possible, but orchestrator shouldn't strictly depend on name
	// However, for certain logic we might need hints.
	// For now let's just use generic logic.

	finalPrompt := prompt
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{prompt}}", ctx.Prompt)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{history}}", strings.Join(ctx.History, "\n"))
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{devices}}", renderRegisteredDevices(ctx))

	// Special handling for Translation placeholders if present
	if strings.Contains(finalPrompt, "{{target_lang}}") {
		targetLang := "English"
		if ctx.Language != "" {
			switch {
			case strings.EqualFold(ctx.Language, "id") || strings.EqualFold(ctx.Language, "indonesian"):
				targetLang = "Indonesian"
			case strings.EqualFold(ctx.Language, "en") || strings.EqualFold(ctx.Language, "english"):
				targetLang = "English"
			default:
				targetLang = ctx.Language
			}
		} else if strings.Contains(strings.ToLower(ctx.Prompt), "indonesia") || strings.Contains(strings.ToLower(ctx.Prompt), "indo") {
			targetLang = "Indonesian"
		}
		finalPrompt = strings.ReplaceAll(finalPrompt, "{{target_lang}}", targetLang)
		model = "low"
	}

	// 2. Call LLM
	res, err := ctx.LLM.CallModel(ctx.Ctx, finalPrompt, model)
	if err != nil {
		return nil, err
	}

	return &skills.SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
