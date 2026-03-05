package orchestrator

import (
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/rag/skills"
	"strings"
)

// Router coordinates the execution of skills using LLM-based routing.
type Router struct {
	registry   *skills.SkillRegistry
	translator skills.TranslateService
	guard      *GuardOrchestrator
}

// NewRouter creates a new AI orchestrator with the provided registry.
func NewRouter(registry *skills.SkillRegistry, translator skills.TranslateService, guard *GuardOrchestrator) *Router {
	return &Router{
		registry:   registry,
		translator: translator,
		guard:      guard,
	}
}

// RouteAndExecute analyzes the prompt, picks the best skill, and executes it.
func (r *Router) RouteAndExecute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	// 0. Guard: Check for spam before wasting an LLM call
	if r.guard != nil {
		guardResult := r.guard.CheckPrompt(ctx)
		if guardResult != GuardClean {
			// Route to Identity skill for a more natural response
			skill, ok := r.registry.Get("Identity")
			if !ok {
				all := r.registry.GetAll()
				names := make([]string, 0, len(all))
				for _, s := range all {
					names = append(names, s.Name())
				}
				utils.LogError("Router Debug: Identity skill not found! Registered skills: %v", names)
				return nil, fmt.Errorf("identity skill not found for guard fallback")
			}
			res, err := skill.Execute(ctx)
			if err != nil {
				return nil, err
			}
			// Mark as blocked ONLY for PureSpam and Irrelevant (to hide user bubble)
			res.IsBlocked = (guardResult == GuardPureSpam || guardResult == GuardIrrelevant)
			return res, nil
		}
	}

	allSkills := r.registry.GetAll()
	if len(allSkills) == 0 {
		return nil, fmt.Errorf("no skills registered in the orchestrator")
	}

	// 1. Build the routing prompt
	skillDescriptions := make([]string, 0, len(allSkills))
	for _, s := range allSkills {
		skillDescriptions = append(skillDescriptions, fmt.Sprintf("- %s: %s", s.Name(), s.Description()))
	}

	routingPrompt := fmt.Sprintf(`You are the Brain of Sensio AI Assistant. Your task is to route the user's request to the correct skill.

Available Skills:
%s

User Request: "%s"

Rules:
1. Choose the single most appropriate Skill Name from the list above.
2. If the user wants to control something (lights, AC, music, playback, etc.), use the "Control" skill.
3. If the user is asking specifically about your identity, name, or general discovery of what you are, use the "Identity" skill.
4. If no specific skill matches but it's a general conversation, use the "Identity" skill.
5. ONLY return the Name of the chosen skill. No explanation.

Chosen Skill Name:`, strings.Join(skillDescriptions, "\n"), ctx.Prompt)

	// 2. Call LLM to decide
	model := "high"

	utils.LogDebug("Router: Routing prompt for '%s'", ctx.Prompt)
	chosenSkillName, err := ctx.LLM.CallModel(routingPrompt, model)
	if err != nil {
		return nil, fmt.Errorf("orchestrator routing failed: %w", err)
	}

	chosenSkillName = strings.TrimSpace(chosenSkillName)
	utils.LogDebug("Router: LLM chose skill '%s'", chosenSkillName)

	// 3. Execute the chosen skill
	skill, ok := r.registry.Get(chosenSkillName)
	if !ok {
		// Fallback to Identity skill if the LLM hallucinated a skill name or if we want a safe default
		utils.LogWarn("Router: Chosen skill '%s' not found in registry. Falling back to Identity.", chosenSkillName)
		skill, ok = r.registry.Get("Identity")
		if !ok {
			return nil, fmt.Errorf("skill '%s' not found and no Identity fallback available", chosenSkillName)
		}
	}

	res, err := skill.Execute(ctx)
	if err != nil {
		return nil, err
	}

	// 4. Translate response if needed
	if ctx.Language != "" && ctx.Language != "en" && r.translator != nil && res.Message != "" {
		utils.LogDebug("Router: Translating response to '%s'", ctx.Language)
		translated, err := r.translator.TranslateTextSync(res.Message, ctx.Language)
		if err == nil {
			res.Message = translated
		} else {
			utils.LogWarn("Router: Translation failed: %v", err)
		}
	}

	return res, nil
}
