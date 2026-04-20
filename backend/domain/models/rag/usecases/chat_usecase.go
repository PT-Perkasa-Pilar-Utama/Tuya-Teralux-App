package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/providers"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	"sensio/domain/models/rag/skills/orchestrator"
	tuyaDtos "sensio/domain/tuya/dtos"
	"strings"
	"time"
)

type ChatUseCase interface {
	Chat(ctx context.Context, uid, terminalID, prompt, language, requestID string) (*dtos.RAGChatResponseDTO, error)
}

type ChatUseCaseImpl struct {
	llm              skills.LLMClient
	fallbackLLM      skills.LLMClient
	config           *utils.Config
	badger           *infrastructure.BadgerService
	vector           *infrastructure.VectorService
	guard            *orchestrator.GuardOrchestrator
	fastIntentRouter *orchestrator.FastIntentRouter
	decisionEngine   *orchestrator.AssistantDecisionEngineImpl
	providerResolver providers.ProviderResolver
	controlUseCase   ControlUseCase // For actual device execution
	// Keep orchestrator for backward compatibility during migration
	orchestrator *orchestrator.Router
}

func NewChatUseCase(
	llm skills.LLMClient,
	fallbackLLM skills.LLMClient,
	cfg *utils.Config,
	badger *infrastructure.BadgerService,
	vector *infrastructure.VectorService,
	guard *orchestrator.GuardOrchestrator,
	fastIntentRouter *orchestrator.FastIntentRouter,
	decisionEngine *orchestrator.AssistantDecisionEngineImpl,
	providerResolver providers.ProviderResolver,
	controlUseCase ControlUseCase,
	orchestrator *orchestrator.Router, // kept for migration
) ChatUseCase {
	return &ChatUseCaseImpl{
		llm:              llm,
		fallbackLLM:      fallbackLLM,
		config:           cfg,
		badger:           badger,
		vector:           vector,
		guard:            guard,
		fastIntentRouter: fastIntentRouter,
		decisionEngine:   decisionEngine,
		providerResolver: providerResolver,
		controlUseCase:   controlUseCase,
		orchestrator:     orchestrator,
	}
}

func (u *ChatUseCaseImpl) Chat(ctx context.Context, uid, terminalID, prompt, language, requestID string) (*dtos.RAGChatResponseDTO, error) {
	ucStart := time.Now()
	pipelinePath := "unknown"
	var controlDuration time.Duration // Track control execution time separately

	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	// Idempotency check: if requestID is provided, check cache for duplicate/completed requests
	// Key strategy: Use terminalID + requestID as primary basis (stable across MQTT/HTTP channels).
	// UID is excluded from key identity to prevent cross-channel mismatch when the same interaction
	// uses different auth contexts. Request ID must be unique per terminal interaction and reused
	// across MQTT/HTTP fallback for the same user interaction.
	if requestID != "" && u.badger != nil {
		idempotencyStart := time.Now()
		idempotencyKey := u.getChatIdempotencyKey(terminalID, requestID)

		// Try to acquire lock by setting in_progress marker
		isNew, err := u.badger.SetIfAbsentWithTTL(idempotencyKey, []byte("in_progress"), 30*time.Second)
		if err != nil {
			utils.LogError("ChatUseCase: Idempotency lock check failed | request_id=%s | error=%v", requestID, err)
			// Continue without idempotency check on error
		} else if !isNew {
			// Key exists - check if it's in_progress or completed
			existingData, ttlRemaining, err := u.badger.GetWithTTL(idempotencyKey)
			if err != nil {
				utils.LogError("ChatUseCase: Idempotency cache read failed | request_id=%s | error=%v", requestID, err)
			}
			if existingData != nil {
				if string(existingData) == "in_progress" {
					// Another request is still processing
					utils.LogInfo("ChatUseCase: Request already in progress | request_id=%s | duration_ms=%d", requestID, time.Since(idempotencyStart).Milliseconds())
					return &dtos.RAGChatResponseDTO{
						Response:       "Mengerti, saya sedang memproses permintaan Anda.",
						IsBlocked:      false,
						HTTPStatusCode: 200,
						RequestID:      requestID,
						Source:         "IDEMPOTENCY_IN_PROGRESS",
					}, nil
				}

				// Try to unmarshal completed response
				var cachedResponse dtos.RAGChatResponseDTO
				if err := json.Unmarshal(existingData, &cachedResponse); err == nil {
					utils.LogInfo("ChatUseCase: Returning cached response | request_id=%s | duration_ms=%d", requestID, time.Since(idempotencyStart).Milliseconds())
					// Mark as cached duplicate
					cachedResponse.Source = "IDEMPOTENCY_CACHED"
					return &cachedResponse, nil
				}

				// Cache corrupted, delete and continue
				utils.LogWarn("ChatUseCase: Idempotency cache corrupted | request_id=%s | error=%v", requestID, err)
				if err := u.badger.Delete(idempotencyKey); err != nil {
					utils.LogWarn("ChatUseCase: Failed to delete idempotency cache | request_id=%s | error=%v", requestID, err)
				}
			} else if ttlRemaining == 0 {
				// Key expired, clean up and continue
				if err := u.badger.Delete(idempotencyKey); err != nil {
					utils.LogWarn("ChatUseCase: Failed to delete expired idempotency key | request_id=%s | error=%v", requestID, err)
				}
			}
		}
	}

	// 1. Get History
	historyStart := time.Now()
	historyKey := fmt.Sprintf("chat_history:%s", terminalID)
	var history []string
	if u.badger != nil {
		data, err := u.badger.Get(historyKey)
		if err != nil {
			utils.LogWarn("ChatUseCase: Failed to get history | key=%s | error=%v", historyKey, err)
		} else if data != nil {
			unmarshalStart := time.Now()
			if err := json.Unmarshal(data, &history); err != nil {
				utils.LogWarn("ChatUseCase: Failed to unmarshal history | key=%s | error=%v", historyKey, err)
			}
			utils.LogDebug("ChatUseCase: History retrieved | key=%s | cache_duration_ms=%d | unmarshal_duration_ms=%d | history_size=%d", historyKey, time.Since(historyStart).Milliseconds()-time.Since(unmarshalStart).Milliseconds(), time.Since(unmarshalStart).Milliseconds(), len(history))
		} else {
			utils.LogDebug("ChatUseCase: History not found | key=%s | duration_ms=%d", historyKey, time.Since(historyStart).Milliseconds())
		}
	}
	historyDuration := time.Since(historyStart)

	// 3. Prepare Skill Context
	skillCtx := &skills.SkillContext{
		Ctx:        ctx,
		UID:        uid,
		TerminalID: terminalID,
		Prompt:     prompt,
		Language:   language,
		History:    history,
		LLM:        u.llm,
		Config:     u.config,
		Vector:     u.vector,
		Badger:     u.badger,
	}

	// 4. NEW PIPELINE: Guard -> Fast Intent -> Single Decision

	// 4a. Guard (rule-based, no LLM)
	guardStart := time.Now()
	guardResult := u.guard.CheckPrompt(skillCtx)
	guardDuration := time.Since(guardStart)

	if guardResult != orchestrator.GuardClean {
		// Handle blocked/dialog/irrelevant
		pipelinePath = "blocked"
		response := u.getGuardResponse(guardResult, language)
		isBlocked := guardResult == orchestrator.GuardPureSpam || guardResult == orchestrator.GuardIrrelevant

		u.saveHistoryIfNotBlocked(u.badger, historyKey, history, prompt, response, isBlocked)
		totalDuration := time.Since(ucStart)

		utils.LogInfo("ChatUseCase: Guard blocked request | pipeline_path=%s | guard_duration_ms=%d | total_duration_ms=%d", pipelinePath, guardDuration.Milliseconds(), totalDuration.Milliseconds())

		resp := &dtos.RAGChatResponseDTO{
			Response:       response,
			IsBlocked:      isBlocked,
			HTTPStatusCode: 200,
		}
		u.finalizeIdempotency(requestID, terminalID, resp)
		return resp, nil
	}

	// 4b. Fast Intent Router
	fastIntentStart := time.Now()
	fastIntentResult := u.fastIntentRouter.Classify(prompt)
	fastIntentDuration := time.Since(fastIntentStart)

	// Handle fast-routed intents
	if fastIntentResult.Intent != orchestrator.FastIntentNone {
		switch fastIntentResult.Intent {
		case orchestrator.FastIntentIdentity:
			pipelinePath = "fast_identity"
			response := u.getIdentityResponse(language)
			u.saveHistoryIfNotBlocked(u.badger, historyKey, history, prompt, response, false)
			totalDuration := time.Since(ucStart)
			utils.LogInfo("ChatUseCase: Fast identity route | pipeline_path=%s | fast_intent_duration_ms=%d | total_duration_ms=%d", pipelinePath, fastIntentDuration.Milliseconds(), totalDuration.Milliseconds())
			resp := &dtos.RAGChatResponseDTO{
				Response:       response,
				IsControl:      false,
				IsBlocked:      false,
				HTTPStatusCode: 200,
			}
			u.finalizeIdempotency(requestID, terminalID, resp)
			return resp, nil

		case orchestrator.FastIntentDiscovery:
			pipelinePath = "fast_discovery"
			response := u.getDiscoveryResponse(skillCtx, language)
			u.saveHistoryIfNotBlocked(u.badger, historyKey, history, prompt, response, false)
			totalDuration := time.Since(ucStart)
			utils.LogInfo("ChatUseCase: Fast discovery route | pipeline_path=%s | fast_intent_duration_ms=%d | total_duration_ms=%d", pipelinePath, fastIntentDuration.Milliseconds(), totalDuration.Milliseconds())
			resp := &dtos.RAGChatResponseDTO{
				Response:       response,
				IsControl:      true,
				IsBlocked:      false,
				HTTPStatusCode: 200,
			}
			u.finalizeIdempotency(requestID, terminalID, resp)
			return resp, nil

		case orchestrator.FastIntentControl:
			pipelinePath = "fast_control"
			// Execute control directly
			controlResult, err := u.executeFastControl(skillCtx, fastIntentResult)
			controlDuration := time.Since(fastIntentStart)
			if err != nil {
				utils.LogWarn("ChatUseCase: Fast control execution failed: %v, falling back to decision engine", err)
				// Fall through to decision engine
			} else {
				u.saveHistoryIfNotBlocked(u.badger, historyKey, history, prompt, controlResult.Message, false)
				totalDuration := time.Since(ucStart)
				utils.LogInfo("ChatUseCase: Fast control executed | pipeline_path=%s | control_duration_ms=%d | total_duration_ms=%d", pipelinePath, controlDuration.Milliseconds(), totalDuration.Milliseconds())
				resp := &dtos.RAGChatResponseDTO{
					Response:       controlResult.Message,
					IsControl:      controlResult.IsControl,
					IsBlocked:      false,
					HTTPStatusCode: controlResult.HTTPStatusCode,
				}
				u.finalizeIdempotency(requestID, terminalID, resp)
				return resp, nil
			}
		}
	}

	// 4c. Single Decision Engine (for non-fast-routed requests)
	pipelinePath = "single_decision_chat"
	decisionStart := time.Now()
	totalDecisionDuration := time.Duration(0)

	utils.LogDebug("ChatUseCase: DecisionEngine.Decide starting | llm_provider=%T", u.llm)
	u.decisionEngine.SetLLM(u.llm)
	decision, err := u.executeDecisionWithFallback(skillCtx, &totalDecisionDuration)
	totalDecisionDuration += time.Since(decisionStart)

	utils.LogDebug("ChatUseCase: DecisionEngine.Decide completed | duration_ms=%d | err=%v", totalDecisionDuration.Milliseconds(), err)

	if err != nil {
		// Still failed, use service issue fallback
		pipelinePath = "service_issue"
		response := u.getServiceIssueResponse(language)
		u.saveHistoryIfNotBlocked(u.badger, historyKey, history, prompt, response, false)
		totalDuration := time.Since(ucStart)
		utils.LogError("ChatUseCase: Decision engine failed completely | pipeline_path=%s | total_duration_ms=%d", pipelinePath, totalDuration.Milliseconds())
		resp := &dtos.RAGChatResponseDTO{
			Response:       response,
			IsControl:      false,
			IsBlocked:      false,
			HTTPStatusCode: 200,
		}
		u.finalizeIdempotency(requestID, terminalID, resp)
		return resp, nil
	}
	decisionDuration := totalDecisionDuration

	// Handle decision intent
	var result *skills.SkillResult
	switch decision.Intent {
	case "blocked":
		pipelinePath = "blocked_decision"
		result = &skills.SkillResult{
			Message:   u.getGuardResponse(orchestrator.GuardPureSpam, language),
			IsBlocked: true,
		}

	case "identity":
		pipelinePath = "single_decision_identity"
		result = &skills.SkillResult{
			Message:   decision.Response,
			IsBlocked: false,
		}

	case "control":
		pipelinePath = "single_decision_control"
		// Execute control based on decision hints
		controlStart := time.Now()
		controlResult, err := u.executeDecisionControl(skillCtx, decision)
		controlDuration := time.Since(controlStart)

		if err != nil {
			utils.LogWarn("ChatUseCase: Decision control execution failed | duration_ms=%d | error=%v", controlDuration.Milliseconds(), err)
			result = &skills.SkillResult{
				Message:   "Maaf, saya tidak dapat memproses perintah kontrol tersebut.",
				IsControl: true,
			}
		} else {
			result = controlResult
			utils.LogDebug("ChatUseCase: Decision control executed | duration_ms=%d | device_id=%v", controlDuration.Milliseconds(), result.Data)
		}

	case "chat":
		fallthrough
	default:
		pipelinePath = "single_decision_chat"
		result = &skills.SkillResult{
			Message:   decision.Response,
			IsBlocked: false,
		}
	}

	// 5. Save History (skip if blocked)
	historySaveStart := time.Now()
	if u.badger != nil && !result.IsBlocked {
		history = append(history, "User: "+prompt, "Assistant: "+result.Message)
		if len(history) > 20 {
			history = history[len(history)-20:]
		}
		marshalStart := time.Now()
		data, _ := json.Marshal(history)
		marshalDuration := time.Since(marshalStart)

		setStart := time.Now()
		if err := u.badger.Set(historyKey, data); err != nil {
			utils.LogWarn("ChatUseCase: Failed to save history | key=%s | error=%v", historyKey, err)
		}
		setDuration := time.Since(setStart)

		utils.LogDebug("ChatUseCase: History saved | key=%s | marshal_duration_ms=%d | set_duration_ms=%d | history_size=%d", historyKey, marshalDuration.Milliseconds(), setDuration.Milliseconds(), len(history))
	}
	historySaveDuration := time.Since(historySaveStart)

	totalDuration := time.Since(ucStart)
	utils.LogInfo("ChatUseCase: Chat completed | pipeline_path=%s | history_duration_ms=%d | guard_duration_ms=%d | fast_intent_duration_ms=%d | decision_duration_ms=%d | control_duration_ms=%d | history_save_duration_ms=%d | total_duration_ms=%d",
		pipelinePath, historyDuration.Milliseconds(), guardDuration.Milliseconds(), fastIntentDuration.Milliseconds(), decisionDuration.Milliseconds(), controlDuration.Milliseconds(), historySaveDuration.Milliseconds(), totalDuration.Milliseconds())

	// Handle Redirect for Control
	var redirect *dtos.RedirectDTO
	if result.IsControl && decision != nil && decision.Intent == "control" {
		redirect = &dtos.RedirectDTO{
			Endpoint: "/api/rag/control",
			Method:   "POST",
			Body: dtos.RAGControlRequestDTO{
				Prompt:     prompt,
				TerminalID: terminalID,
			},
		}
	}

	// Update idempotency cache with completed response
	resp := &dtos.RAGChatResponseDTO{
		Response:       result.Message,
		IsControl:      result.IsControl,
		IsBlocked:      result.IsBlocked,
		Redirect:       redirect,
		HTTPStatusCode: result.HTTPStatusCode,
	}
	u.finalizeIdempotency(requestID, terminalID, resp)

	return resp, nil
}

// getGuardResponse returns the appropriate response for guard results.
func (u *ChatUseCaseImpl) getGuardResponse(result orchestrator.GuardResult, language string) string {
	switch result {
	case orchestrator.GuardPureSpam:
		return "" // No response for pure spam
	case orchestrator.GuardDialogWithPromo:
		return u.guard.IdentityResponse(language)
	case orchestrator.GuardIrrelevant:
		if strings.EqualFold(language, "en") {
			return "I'm Sensio, your smart home assistant. I can help you control devices, summarize meetings, and answer questions about your smart home. How can I help you today?"
		}
		return "Hai! Saya Sensio, asisten rumah pintar kamu. Saya bisa bantu kontrol perangkat, merangkum rapat, dan menjawab pertanyaan seputar smart home kamu. Ada yang bisa saya bantu?"
	default:
		return ""
	}
}

// getIdentityResponse returns the identity response.
func (u *ChatUseCaseImpl) getIdentityResponse(language string) string {
	if strings.EqualFold(language, "en") {
		return "Hi! I'm Sensio, your smart home assistant. I can help you control devices, summarize meetings, and answer questions about your smart home. How can I help you today?"
	}
	return "Hai! Saya Sensio, asisten rumah pintar kamu. Saya bisa bantu kontrol perangkat, merangkum rapat, dan menjawab pertanyaan seputar smart home kamu. Ada yang bisa saya bantu?"
}

// getDiscoveryResponse returns the device discovery response.
func (u *ChatUseCaseImpl) getDiscoveryResponse(ctx *skills.SkillContext, language string) string {
	// Try to get actual devices from vector store - user-scoped like control path
	if u.vector != nil && ctx != nil && ctx.UID != "" {
		// Use same key pattern as control path: tuya:devices:uid:{uid}
		deviceKey := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
		deviceJSON, found := u.vector.Get(deviceKey)
		if found && deviceJSON != "" {
			// Unmarshal into assistant-safe snapshot DTO
			var snapshot tuyaDtos.AssistantSafeDevicesSnapshotDTO
			if err := json.Unmarshal([]byte(deviceJSON), &snapshot); err == nil && len(snapshot.Devices) > 0 {
				// Build response from actual parsed devices
				deviceTypes := make(map[string]int) // category -> count
				deviceNames := make([]string, 0, len(snapshot.Devices))

				for _, dev := range snapshot.Devices {
					// Normalize category
					normalizedCat := u.normalizeDeviceCategory(dev.Category, dev.ProductName)
					deviceTypes[normalizedCat]++
					if dev.Name != "" {
						deviceNames = append(deviceNames, dev.Name)
					}
				}

				if len(deviceTypes) > 0 {
					var categories []string
					for cat, count := range deviceTypes {
						categories = append(categories, fmt.Sprintf("%s (%d)", cat, count))
					}

					// If <= 8 devices, list names; otherwise summarize by category
					if len(snapshot.Devices) <= 8 {
						if strings.EqualFold(language, "en") {
							return fmt.Sprintf("I found %d devices in your smart home: %s. Which device would you like to control?", len(snapshot.Devices), strings.Join(deviceNames, ", "))
						}
						return fmt.Sprintf("Saya menemukan %d perangkat di smart home Anda: %s. Perangkat apa yang ingin Anda kontrol?", len(snapshot.Devices), strings.Join(deviceNames, ", "))
					} else {
						if strings.EqualFold(language, "en") {
							return fmt.Sprintf("I found %d devices in your smart home including: %s. Which device would you like to control?", len(snapshot.Devices), strings.Join(categories, ", "))
						}
						return fmt.Sprintf("Saya menemukan %d perangkat di smart home Anda termasuk: %s. Perangkat apa yang ingin Anda kontrol?", len(snapshot.Devices), strings.Join(categories, ", "))
					}
				}
			} else {
				utils.LogWarn("getDiscoveryResponse: failed to unmarshal device snapshot for user %s: %v", ctx.UID, err)
			}
		}
		utils.LogDebug("getDiscoveryResponse: No devices found for user %s", ctx.UID)
	}

	// Fallback to generic response
	if strings.EqualFold(language, "en") {
		return "I can help you control various smart home devices including lights, air conditioners, fans, TVs, and speakers. Which device would you like to control?"
	}
	return "Saya bisa membantu mengontrol berbagai perangkat smart home seperti lampu, AC, kipas angin, TV, dan speaker. Perangkat apa yang ingin Anda kontrol?"
}

// normalizeDeviceCategory maps Tuya category codes to assistant-friendly labels.
func (u *ChatUseCaseImpl) normalizeDeviceCategory(category, productName string) string {
	cat := strings.ToLower(strings.TrimSpace(category))

	// AC categories
	if cat == "wnykq" || cat == "ac" || cat == "cl" || strings.Contains(cat, "air conditioner") {
		return "AC"
	}

	// Light categories
	if cat == "dj" || cat == "kg" || cat == "ty" || cat == "xdd" || cat == "fwd" || strings.Contains(cat, "light") || strings.Contains(cat, "lamp") {
		return "Lampu"
	}

	// Switch/outlet categories
	if cat == "dlq" || cat == "pc" || cat == "cz" || strings.Contains(cat, "switch") || strings.Contains(cat, "outlet") {
		return "Switch"
	}

	// TV categories
	if cat == "infrared_tv" || cat == "tv" || strings.Contains(cat, "television") {
		return "TV"
	}

	// Fan categories
	if cat == "fs" || cat == "fskg" || strings.Contains(cat, "fan") {
		return "Kipas"
	}

	// Sensor categories
	if cat == "wsdcg" || strings.Contains(cat, "sensor") {
		return "Sensor"
	}

	// Panel/terminal
	if cat == "dgnzk" || strings.Contains(cat, "panel") || strings.Contains(cat, "terminal") {
		return "Panel"
	}

	// Speaker
	if strings.Contains(cat, "speaker") {
		return "Speaker"
	}

	// Fallback to product name or category
	if productName != "" {
		return productName
	}
	if category != "" {
		return category
	}
	return "Device"
}

// getServiceIssueResponse returns a service issue fallback response.
func (u *ChatUseCaseImpl) getServiceIssueResponse(language string) string {
	if strings.EqualFold(language, "en") {
		return "Sorry, the AI service is temporarily unavailable. Please try again in a moment."
	}
	return "Maaf, layanan AI sedang gangguan. Silakan coba lagi sebentar."
}

// saveHistoryIfNotBlocked saves history if not blocked.
func (u *ChatUseCaseImpl) saveHistoryIfNotBlocked(badger *infrastructure.BadgerService, historyKey string, history []string, prompt, response string, isBlocked bool) time.Duration {
	if badger == nil || isBlocked {
		return 0
	}
	historyStart := time.Now()
	history = append(history, "User: "+prompt, "Assistant: "+response)
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	data, err := json.Marshal(history)
	if err != nil {
		utils.LogWarn("ChatUseCase: Failed to marshal history | key=%s | error=%v", historyKey, err)
		return time.Since(historyStart)
	}
	if err := badger.Set(historyKey, data); err != nil {
		utils.LogWarn("ChatUseCase: Failed to save history | key=%s | error=%v", historyKey, err)
	}
	return time.Since(historyStart)
}

// executeFastControl executes a fast-routed control command.
func (u *ChatUseCaseImpl) executeFastControl(ctx *skills.SkillContext, intent orchestrator.FastIntentResult) (*skills.SkillResult, error) {
	// Use the original user prompt directly to preserve quantifiers like "semua" (all)
	// and ordinal hints that would be lost if we reconstructed from intent
	controlPrompt := ctx.Prompt

	// Call actual control use case for device execution
	if u.controlUseCase != nil {
		controlResult, err := u.controlUseCase.ProcessControl(ctx.Ctx, ctx.UID, ctx.TerminalID, controlPrompt)
		if err == nil {
			// Return actual execution result with preserved status code
			return &skills.SkillResult{
				Message:        controlResult.Message,
				IsControl:      true,
				IsBlocked:      false,
				HTTPStatusCode: controlResult.HTTPStatusCode, // Preserve actual status code
			}, nil
		}
		utils.LogWarn("executeFastControl: Control execution failed: %v", err)

		status, message := u.mapControlRuntimeError(err, languageFromSkillContext(ctx))
		return &skills.SkillResult{
			Message:        message,
			IsControl:      true,
			IsBlocked:      false,
			HTTPStatusCode: status,
		}, nil
	}

	// Control use case not configured - this is a system error
	utils.LogWarn("executeFastControl: Control use case not configured")
	return &skills.SkillResult{
		Message:        u.getControlUnavailableResponse(languageFromSkillContext(ctx)),
		IsControl:      true,
		IsBlocked:      false,
		HTTPStatusCode: 503,
	}, nil
}

// executeDecisionControl executes control based on decision engine hints.
func (u *ChatUseCaseImpl) executeDecisionControl(ctx *skills.SkillContext, decision *orchestrator.AssistantDecision) (*skills.SkillResult, error) {
	// Reconstruct deterministic control prompt
	controlPrompt, err := u.buildControlPromptFromDecision(decision)
	if err != nil {
		utils.LogWarn("executeDecisionControl: Prompt reconstruction failed: %v", err)
		return &skills.SkillResult{
			Message:        "Maaf, saya tidak dapat memproses perintah kontrol tersebut.",
			IsControl:      true,
			IsBlocked:      false,
			HTTPStatusCode: 400,
		}, nil
	}

	// Call actual control use case for device execution
	if u.controlUseCase != nil && controlPrompt != "" {
		controlResult, err := u.controlUseCase.ProcessControl(ctx.Ctx, ctx.UID, ctx.TerminalID, controlPrompt)
		if err == nil {
			// Return actual execution result with preserved status code
			return &skills.SkillResult{
				Message:        controlResult.Message,
				IsControl:      true,
				IsBlocked:      false,
				HTTPStatusCode: controlResult.HTTPStatusCode, // Preserve actual status code
			}, nil
		}
		utils.LogWarn("executeDecisionControl: Control execution failed: %v", err)

		status, message := u.mapControlRuntimeError(err, languageFromSkillContext(ctx))
		return &skills.SkillResult{
			Message:        message,
			IsControl:      true,
			IsBlocked:      false,
			HTTPStatusCode: status,
		}, nil
	}

	// No control prompt available - this is a decision engine error
	utils.LogWarn("executeDecisionControl: No control prompt available from decision")
	return &skills.SkillResult{
		Message:        "Maaf, saya tidak dapat memproses perintah kontrol tersebut.",
		IsControl:      true,
		IsBlocked:      false,
		HTTPStatusCode: 400,
	}, nil
}

// buildControlPromptFromDecision reconstructs a deterministic control prompt from structured decision.
func (u *ChatUseCaseImpl) buildControlPromptFromDecision(decision *orchestrator.AssistantDecision) (string, error) {
	// 1. Use explicit control_prompt if present (high confidence override)
	if decision.ControlPrompt != "" {
		return decision.ControlPrompt, nil
	}

	// 2. Fallback to structured reconstruction if operation and devices are present
	if decision.Operation != "" && len(decision.DeviceHints) > 0 {
		device := decision.DeviceHints[0]

		// Map values deterministically based on operation type
		var value string
		if decision.ValueHints != nil {
			switch decision.Operation {
			case "brightness":
				value = decision.ValueHints["brightness"]
			case "temperature":
				value = decision.ValueHints["temperature"]
			case "fan_speed":
				value = decision.ValueHints["fan_speed"]
			}
		}

		if value != "" {
			return fmt.Sprintf("%s %s %s", decision.Operation, device, value), nil
		}
		return fmt.Sprintf("%s %s", decision.Operation, device), nil
	}

	return "", fmt.Errorf("insufficient decision data for prompt reconstruction")
}

// mapControlRuntimeError maps ProcessControl runtime errors to appropriate HTTP status and message.
func (u *ChatUseCaseImpl) mapControlRuntimeError(err error, language string) (status int, message string) {
	errStr := strings.ToLower(err.Error())

	// Default to Indonesian
	isEn := strings.EqualFold(language, "en")

	// 1. Service/Provider Unavailable (503)
	if strings.Contains(errStr, "auth") ||
		strings.Contains(errStr, "token") ||
		strings.Contains(errStr, "provider") ||
		strings.Contains(errStr, "unavailable") ||
		strings.Contains(errStr, "timeout") {
		if isEn {
			return 503, "Sorry, the control system is temporarily unavailable. Please try again in a moment."
		}
		return 503, "Maaf, sistem kontrol sedang tidak tersedia. Silakan coba lagi sebentar."
	}

	// 2. Fallback for other unexpected internal errors (500)
	if isEn {
		return 500, "Sorry, an internal error occurred while processing your control request."
	}
	return 500, "Maaf, terjadi gangguan internal saat memproses perintah kontrol Anda."
}

// getControlUnavailableResponse returns a generic unavailable message.
func (u *ChatUseCaseImpl) getControlUnavailableResponse(language string) string {
	if strings.EqualFold(language, "en") {
		return "Maaf, sistem kontrol tidak tersedia."
	}
	return "Maaf, sistem kontrol tidak tersedia."
}

func languageFromSkillContext(ctx *skills.SkillContext) string {
	if ctx == nil {
		return "id"
	}
	return ctx.Language
}

// executeDecisionWithFallback executes the decision engine with health-aware remote provider fallback
func (u *ChatUseCaseImpl) executeDecisionWithFallback(skillCtx *skills.SkillContext, totalDuration *time.Duration) (*orchestrator.AssistantDecision, error) {
	var finalDecision *orchestrator.AssistantDecision
	var err error

	if skillCtx.TerminalID != "" {
		// Use terminal-specific provider preference
		err = u.providerResolver.ExecuteWithFallbackByTerminal(skillCtx.TerminalID, func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx.LLM = resolvedSet.LLM
			u.decisionEngine.SetLLM(resolvedSet.LLM)

			decisionStart := time.Now()
			decision, execErr := u.decisionEngine.Decide(skillCtx)
			*totalDuration += time.Since(decisionStart)

			if execErr == nil {
				finalDecision = decision
			}
			return execErr
		})
	} else {
		// Use standard health-aware fallback
		err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *providers.ResolvedProviderSet) error {
			skillCtx.LLM = resolvedSet.LLM
			u.decisionEngine.SetLLM(resolvedSet.LLM)

			decisionStart := time.Now()
			decision, execErr := u.decisionEngine.Decide(skillCtx)
			*totalDuration += time.Since(decisionStart)

			if execErr == nil {
				finalDecision = decision
			}
			return execErr
		})
	}

	return finalDecision, err
}

// finalizeIdempotency persists the response to the idempotency cache if requestID is provided.
// This ensures all return paths (early returns and normal completion) maintain consistent idempotency state.
// Idempotency key uses terminalID + requestID (UID excluded for cross-channel stability).
func (u *ChatUseCaseImpl) finalizeIdempotency(requestID, terminalID string, response *dtos.RAGChatResponseDTO) {
	if requestID == "" || u.badger == nil {
		return
	}

	idempotencyKey := u.getChatIdempotencyKey(terminalID, requestID)
	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		utils.LogError("ChatUseCase: Failed to marshal response for idempotency cache | request_id=%s | error=%v", requestID, marshalErr)
		return
	}

	// Cache completed response for 5 minutes
	if err := u.badger.SetWithTTL(idempotencyKey, responseData, 5*time.Minute); err != nil {
		utils.LogError("ChatUseCase: Failed to update idempotency cache | request_id=%s | error=%v", requestID, err)
	} else {
		utils.LogDebug("ChatUseCase: Response cached for idempotency | request_id=%s", requestID)
	}
}

// getChatIdempotencyKey returns the idempotency cache key for chat requests.
// Key format: chat:idempotency:{terminalID}:{requestID}
// UID is intentionally excluded to ensure cross-channel stability (MQTT/HTTP fallback).
func (u *ChatUseCaseImpl) getChatIdempotencyKey(terminalID, requestID string) string {
	return fmt.Sprintf("chat:idempotency:%s:%s", terminalID, requestID)
}
