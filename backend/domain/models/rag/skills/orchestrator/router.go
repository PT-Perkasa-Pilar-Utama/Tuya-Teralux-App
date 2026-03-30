package orchestrator

import (
	"context"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	"strings"
	"time"
)

// Router coordinates the execution of skills using LLM-based routing.
type Router struct {
	registry   *skills.SkillRegistry
	translator skills.TranslateService
	guard      *GuardOrchestrator
}

func isDeviceDiscoveryPrompt(prompt string) bool {
	prompt = strings.ToLower(strings.TrimSpace(prompt))
	if prompt == "" {
		return false
	}

	deviceWords := []string{"device", "perangkat", "alat", "lampu", "ac", "tv"}
	discoveryWords := []string{
		"what can i control", "what devices", "which devices", "available devices", "connected devices",
		"bisa control", "bisa saya control", "bisa aku control", "bisa dikontrol",
		"apa aja", "apa saja", "daftar", "list", "terdaftar", "tersambung", "konek", "sambung",
	}

	hasDeviceWord := false
	for _, w := range deviceWords {
		if strings.Contains(prompt, w) {
			hasDeviceWord = true
			break
		}
	}

	if !hasDeviceWord {
		return false
	}

	for _, w := range discoveryWords {
		if strings.Contains(prompt, w) {
			return true
		}
	}

	return false
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
	routerStart := time.Now()

	// 0. Guard: Check for spam before wasting an LLM call
	guardStart := time.Now()
	guardDuration := time.Duration(0)
	if r.guard != nil {
		guardResult := r.guard.CheckPrompt(ctx)
		guardDuration = time.Since(guardStart)
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
			utils.LogInfo("Router: Guard blocked prompt | guard_duration_ms=%d", guardDuration.Milliseconds())
			return res, nil
		}
		utils.LogDebug("Router: Guard check passed | guard_duration_ms=%d", guardDuration.Milliseconds())
	}

	allSkills := r.registry.GetAll()
	if len(allSkills) == 0 {
		return nil, fmt.Errorf("no skills registered in the orchestrator")
	}

	// Deterministic override for explicit device discovery/list intent.
	if isDeviceDiscoveryPrompt(ctx.Prompt) {
		if controlSkill, ok := r.registry.Get("Control"); ok {
			utils.LogDebug("Router: deterministic routing to 'Control' for device discovery prompt '%s'", ctx.Prompt)
			skillStart := time.Now()
			res, err := controlSkill.Execute(ctx)
			skillDuration := time.Since(skillStart)
			if err != nil {
				return nil, err
			}

			translateDuration := time.Duration(0)
			if ctx.Language != "" && ctx.Language != "en" && r.translator != nil && res.Message != "" {
				utils.LogDebug("Router: Translating response to '%s'", ctx.Language)
				translateStart := time.Now()
				translated, err := r.translator.TranslateTextSync(context.Background(), res.Message, ctx.Language, ctx.TerminalID)
				translateDuration = time.Since(translateStart)
				if err == nil {
					res.Message = translated
				} else {
					utils.LogWarn("Router: Translation failed: %v", err)
				}
			}
			utils.LogInfo("Router: Deterministic Control skill completed | skill_duration_ms=%d | translate_duration_ms=%d", skillDuration.Milliseconds(), translateDuration.Milliseconds())
			return res, nil
		}
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
3. If the user asks about which/what devices are available, connected, registered, or controllable, use the "Control" skill.
4. If the user is asking specifically about your identity, name, or general discovery of what you are, use the "Identity" skill.
5. If no specific skill matches but it's a general conversation, use the "Identity" skill.
6. ONLY return the Name of the chosen skill. No explanation.

Chosen Skill Name:`, strings.Join(skillDescriptions, "\n"), ctx.Prompt)

	// 2. Call LLM to decide
	model := "high"

	utils.LogDebug("Router: Routing prompt for '%s'", ctx.Prompt)
	routeSelectionStart := time.Now()
	chosenSkillName, err := ctx.LLM.CallModel(ctx.Ctx, routingPrompt, model)
	routeSelectionDuration := time.Since(routeSelectionStart)
	if err != nil {
		return nil, fmt.Errorf("orchestrator routing failed: %w | route_selection_duration_ms=%d", err, routeSelectionDuration.Milliseconds())
	}

	chosenSkillName = strings.TrimSpace(chosenSkillName)
	utils.LogDebug("Router: LLM chose skill '%s' | route_selection_duration_ms=%d", chosenSkillName, routeSelectionDuration.Milliseconds())

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

	skillStart := time.Now()
	res, err := skill.Execute(ctx)
	skillDuration := time.Since(skillStart)
	if err != nil {
		return nil, err
	}

	// 4. Translate response if needed
	translateDuration := time.Duration(0)
	if ctx.Language != "" && ctx.Language != "en" && r.translator != nil && res.Message != "" {
		utils.LogDebug("Router: Translating response to '%s'", ctx.Language)
		translateStart := time.Now()
		translated, err := r.translator.TranslateTextSync(context.Background(), res.Message, ctx.Language, ctx.TerminalID)
		translateDuration = time.Since(translateStart)
		if err == nil {
			res.Message = translated
		} else {
			utils.LogWarn("Router: Translation failed: %v", err)
		}
	}

	totalDuration := time.Since(routerStart)
	utils.LogInfo("Router: Routing completed | guard_duration_ms=%d | route_selection_duration_ms=%d | skill_duration_ms=%d | translate_duration_ms=%d | total_duration_ms=%d", guardDuration.Milliseconds(), routeSelectionDuration.Milliseconds(), skillDuration.Milliseconds(), translateDuration.Milliseconds(), totalDuration.Milliseconds())

	return res, nil
}

// GetSkillRegistry returns the skill registry for direct skill access (e.g., fallback to Identity)
func (r *Router) GetSkillRegistry() *skills.SkillRegistry {
	return r.registry
}
