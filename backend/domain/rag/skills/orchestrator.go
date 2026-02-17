package skills

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
)

// Orchestrator coordinates the execution of skills using LLM-based routing.
type Orchestrator struct {
	registry   *SkillRegistry
	translator TranslateService
}

// NewOrchestrator creates a new AI orchestrator with the provided registry.
func NewOrchestrator(registry *SkillRegistry, translator TranslateService) *Orchestrator {
	return &Orchestrator{
		registry:   registry,
		translator: translator,
	}
}

// RouteAndExecute analyzes the prompt, picks the best skill, and executes it.
func (o *Orchestrator) RouteAndExecute(ctx *SkillContext) (*SkillResult, error) {
	allSkills := o.registry.GetAll()
	if len(allSkills) == 0 {
		return nil, fmt.Errorf("no skills registered in the orchestrator")
	}

	// 1. Build the routing prompt
	var skillDescriptions []string
	for _, s := range allSkills {
		skillDescriptions = append(skillDescriptions, fmt.Sprintf("- %s: %s", s.Name(), s.Description()))
	}

	routingPrompt := fmt.Sprintf(`You are the Brain of Sensio AI Assistant. Your task is to route the user's request to the correct skill.

Available Skills:
%s

User Request: "%s"

Rules:
1. Choose the single most appropriate Skill Name from the list above.
2. If the user is asking about your identity or what you can do, use the "Identity" skill.
3. If the user wants to control something (lights, AC, etc.), use the "Control" skill.
4. If no specific skill matches but it's a general question, use the "Identity" skill to provide a helpful response.
5. ONLY return the Name of the chosen skill. No explanation.

Chosen Skill Name:`, strings.Join(skillDescriptions, "\n"), ctx.Prompt)

	// 2. Call LLM to decide
	model := ctx.Config.LLMModel
	if model == "" {
		model = "default"
	}

	utils.LogDebug("Orchestrator: Routing prompt for '%s'", ctx.Prompt)
	chosenSkillName, err := ctx.LLM.CallModel(routingPrompt, model)
	if err != nil {
		return nil, fmt.Errorf("orchestrator routing failed: %w", err)
	}

	chosenSkillName = strings.TrimSpace(chosenSkillName)
	utils.LogDebug("Orchestrator: LLM chose skill '%s'", chosenSkillName)

	// 3. Execute the chosen skill
	skill, ok := o.registry.Get(chosenSkillName)
	if !ok {
		// Fallback to Identity skill if the LLM hallucinated a skill name or if we want a safe default
		utils.LogWarn("Orchestrator: Chosen skill '%s' not found in registry. Falling back to Identity.", chosenSkillName)
		skill, ok = o.registry.Get("Identity")
		if !ok {
			return nil, fmt.Errorf("skill '%s' not found and no Identity fallback available", chosenSkillName)
		}
	}

	res, err := skill.Execute(ctx)
	if err != nil {
		return nil, err
	}

	// 4. Translate response if needed
	// If the user requested a specific language (e.g. "id") and it's not English ("en"),
	// and we have a translator available, we translate the response.
	if ctx.Language != "" && ctx.Language != "en" && o.translator != nil && res.Message != "" {
		utils.LogDebug("Orchestrator: Translating response to '%s'", ctx.Language)
		translated, err := o.translator.TranslateTextSync(res.Message, ctx.Language)
		if err == nil {
			res.Message = translated
		} else {
			utils.LogWarn("Orchestrator: Translation failed: %v", err)
		}
	}

	return res, nil
}
