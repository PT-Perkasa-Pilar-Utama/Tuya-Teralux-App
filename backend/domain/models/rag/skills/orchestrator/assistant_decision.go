package orchestrator

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/skills"
	"strings"
)

// AssistantDecision represents the structured output from the single LLM decision call.
type AssistantDecision struct {
	Intent        string            `json:"intent"`         // "chat" | "identity" | "control" | "blocked"
	Response      string            `json:"response,omitempty"`
	Operation     string            `json:"operation,omitempty"` // operational verb: "nyalakan"|"matikan"|"brightness"|"temperature"|"fan_speed"
	DeviceHints   []string          `json:"device_hints,omitempty"`
	ValueHints    map[string]string `json:"value_hints,omitempty"` // e.g., {"brightness": "50", "temperature": "24"}
	ControlPrompt string            `json:"control_prompt,omitempty"` // normalized control command if model wants to specify
	IsAmbiguous   bool              `json:"is_ambiguous,omitempty"`
	BlockReason   string            `json:"block_reason,omitempty"`
}

// AssistantDecisionEngine interface for the single-decision assistant flow.
type AssistantDecisionEngine interface {
	Decide(ctx *skills.SkillContext) (*AssistantDecision, error)
}

// AssistantDecisionEngineImpl implements the single-decision assistant flow.
type AssistantDecisionEngineImpl struct {
	llm skills.LLMClient
}

// NewAssistantDecisionEngine creates a new decision engine.
func NewAssistantDecisionEngine(llm skills.LLMClient) *AssistantDecisionEngineImpl {
	return &AssistantDecisionEngineImpl{llm: llm}
}

// SetLLM sets the LLM client for the decision engine.
func (e *AssistantDecisionEngineImpl) SetLLM(llm skills.LLMClient) {
	e.llm = llm
}

// Decide makes a single LLM call to determine intent and generate response.
func (e *AssistantDecisionEngineImpl) Decide(ctx *skills.SkillContext) (*AssistantDecision, error) {
	if ctx == nil || ctx.Prompt == "" {
		return nil, fmt.Errorf("empty prompt")
	}

	// Determine response language
	language := ctx.Language
	if language == "" {
		language = "id" // default to Indonesian
	}

	// Build the single decision prompt
	prompt := e.buildDecisionPrompt(ctx.Prompt, language, ctx.History)

	// Call LLM with strict JSON output requirement
	model := "high"
	response, err := e.llm.CallModel(ctx.Ctx, prompt, model)
	if err != nil {
		utils.LogError("AssistantDecisionEngine: LLM call failed: %v", err)
		return nil, err
	}

	// Parse JSON response
	decision, err := e.parseDecision(response)
	if err != nil {
		// Truncate raw response for logging
		rawLog := response
		if len(rawLog) > 200 {
			rawLog = rawLog[:200] + "..."
		}
		utils.LogError("AssistantDecisionEngine: Failed to parse decision JSON: %v | raw: %s | decision_validation_error=parse_failed", err, rawLog)
		return nil, err
	}

	// Validate decision
	if err := e.validateDecision(decision); err != nil {
		utils.LogError("AssistantDecisionEngine: Invalid decision: %v | decision_validation_error=%s", err, err.Error())
		return nil, err
	}

	return decision, nil
}

// buildDecisionPrompt constructs the prompt for the single LLM decision call.
func (e *AssistantDecisionEngineImpl) buildDecisionPrompt(prompt, language string, history []string) string {
	// Language instruction
	var languageInstruction string
	if strings.EqualFold(language, "en") || strings.EqualFold(language, "english") {
		languageInstruction = "Respond in English."
	} else {
		languageInstruction = "Respond in Indonesian (Bahasa Indonesia)."
	}

	// History context (limited to last 4 exchanges)
	historyContext := ""
	if len(history) > 0 {
		start := len(history) - 4
		if start < 0 {
			start = 0
		}
		recentHistory := history[start:]
		historyContext = fmt.Sprintf(`

Recent Conversation History:
%s

`, strings.Join(recentHistory, "\n"))
	}

	return fmt.Sprintf(`You are Sensio, a smart home assistant. Analyze the user's request and respond appropriately.

%sUser Request: "%s"

Your task is to determine the intent and provide an appropriate response. You MUST output ONLY valid JSON with this exact structure:

{
  "intent": "chat" | "identity" | "control" | "blocked",
  "response": "your response text in the user's language",
  "operation": "nyalakan" | "matikan" | "brightness" | "temperature" | "fan_speed" (only for control intent),
  "device_hints": ["device name or type"] (optional, for control),
  "value_hints": {"key": "value"} (optional, e.g., {"brightness": "50", "temperature": "24"}),
  "control_prompt": "normalized control command" (optional, e.g., "nyalakan lampu ruang tamu"),
  "is_ambiguous": true/false (true if multiple devices match),
  "block_reason": "reason" (only for blocked intent)
}

Intent Guidelines:
- "identity": User asks who you are, what you can do, or general discovery
- "control": User wants to control a specific device (on/off, brightness, temperature, fan speed) - requires device name
- "chat": General conversation, questions, discovery ("what devices can I control?"), or tasks like summarization
- "blocked": Request is spam, promotional, sensitive topic, or irrelevant

For Control Intent:
- "operation": Use the operational verb that matches the command
  - "nyalakan" for turn on commands
  - "matikan" for turn off commands
  - "brightness" for brightness adjustment
  - "temperature" for temperature setting
  - "fan_speed" for fan speed adjustment
- "device_hints": MUST list the target device(s) - required for control
- "value_hints": Include numeric values if applicable (e.g., {"temperature": "24"}, {"brightness": "50"})
- "control_prompt": Optionally provide the full normalized command in Indonesian

Note: Discovery questions like "Apa aja device yang bisa saya kontrol?" should use intent="chat", not control.

Rules:
1. Output ONLY valid JSON, no markdown, no explanations
2. Follow the user's language: %s
3. For control commands, be specific about device and action
4. If ambiguous (multiple devices match), set is_ambiguous=true
5. Keep responses concise and helpful

Examples:

User: "Nyalakan lampu ruang tamu"
Output: {"intent":"control","response":"Baik, menyalakan lampu ruang tamu","operation":"nyalakan","device_hints":["lampu ruang tamu"]}

User: "Set AC 24 derajat"
Output: {"intent":"control","response":"Baik, mengatur AC ke 24 derajat","operation":"temperature","device_hints":["ac"],"value_hints":{"temperature":"24"}}

User: "Matikan kipas angin"
Output: {"intent":"control","response":"Baik, mematikan kipas angin","operation":"matikan","device_hints":["kipas angin"]}

User: "Lampu kamar 50 persen"
Output: {"intent":"control","response":"Baik, mengatur lampu kamar ke 50 persen","operation":"brightness","device_hints":["lampu kamar"],"value_hints":{"brightness":"50"}}

User: "Kamu siapa?"
Output: {"intent":"identity","response":"Hai! Saya Sensio, asisten rumah pintar Anda. Saya bisa membantu mengontrol perangkat smart home, merangkum rapat, dan menjawab pertanyaan. Ada yang bisa saya bantu?"}

User: "Apa aja device yang bisa saya kontrol?"
Output: {"intent":"chat","response":"Saya bisa membantu mengontrol berbagai perangkat smart home seperti lampu, AC, kipas angin, TV, dan speaker. Perangkat apa yang ingin Anda kontrol?"}

User: "Rangkum rapat tadi pagi"
Output: {"intent":"chat","response":"Tentu, saya bisa membantu merangkum rapat. Bisa berikan lebih detail tentang rapat mana yang ingin dirangkum?"}

Now analyze this request and output ONLY JSON:`, historyContext, prompt, languageInstruction)
}

// parseDecision parses the LLM response into an AssistantDecision.
func (e *AssistantDecisionEngineImpl) parseDecision(response string) (*AssistantDecision, error) {
	// Clean up response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var decision AssistantDecision
	if err := json.Unmarshal([]byte(response), &decision); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &decision, nil
}

// validateDecision validates the parsed decision.
func (e *AssistantDecisionEngineImpl) validateDecision(decision *AssistantDecision) error {
	if decision == nil {
		return fmt.Errorf("nil decision")
	}

	// Validate intent
	validIntents := map[string]bool{
		"chat":     true,
		"identity": true,
		"control":  true,
		"blocked":  true,
	}
	if !validIntents[decision.Intent] {
		return fmt.Errorf("invalid intent: %s", decision.Intent)
	}

	// Validate blocked intent requires block_reason
	if decision.Intent == "blocked" && decision.BlockReason == "" {
		return fmt.Errorf("blocked intent requires block_reason")
	}

	// Validate operation if present
	if decision.Operation != "" {
		validOperations := map[string]bool{
			"nyalakan":    true,
			"matikan":     true,
			"brightness":  true,
			"temperature": true,
			"fan_speed":   true,
		}
		if !validOperations[decision.Operation] {
			return fmt.Errorf("invalid operation: %s", decision.Operation)
		}
	}

	// Validate control intent
	if decision.Intent == "control" {
		// Control requires either:
		// 1. control_prompt (full command string), OR
		// 2. operation + device_hints (structured command)
		if decision.ControlPrompt != "" {
			// control_prompt is sufficient on its own
		} else if decision.Operation != "" {
			// operation requires device_hints to be valid
			if len(decision.DeviceHints) == 0 {
				return fmt.Errorf("control intent with 'operation' requires 'device_hints'")
			}
		} else {
			// Neither control_prompt nor operation provided
			return fmt.Errorf("control intent requires either 'operation' + 'device_hints' or 'control_prompt'")
		}
	}

	// Validate response is present for non-blocked intents
	if decision.Intent != "blocked" && decision.Response == "" {
		return fmt.Errorf("response required for non-blocked intent")
	}

	return nil
}
