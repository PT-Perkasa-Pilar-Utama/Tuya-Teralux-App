package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	mailServices "sensio/domain/mail/services"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/services"
	"sensio/domain/rag/skills"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SummaryUseCase interface {
	SummarizeText(text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (string, error)
	SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL, trigger string) (string, error)
	SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (string, error)
	SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL, trigger string) (string, error)
}

type summaryUseCase struct {
	llm           skills.LLMClient
	fallbackLLM   skills.LLMClient
	config        *utils.Config
	cache         *tasks.BadgerTaskCache
	store         *tasks.StatusStore[dtos.RAGStatusDTO]
	renderer      services.SummaryPDFRenderer
	mailExternal  *mailServices.MailExternalService
	mqttSvc       mqttPublisher
	llmTimeout    time.Duration // Timeout for LLM calls
	renderTimeout time.Duration // Timeout for PDF rendering
}

func NewSummaryUseCase(
	llm skills.LLMClient,
	fallbackLLM skills.LLMClient,
	cfg *utils.Config,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[dtos.RAGStatusDTO],
	renderer services.SummaryPDFRenderer,
	mailExternal *mailServices.MailExternalService,
	mqttSvc mqttPublisher,
) SummaryUseCase {
	return &summaryUseCase{
		llm:           llm,
		fallbackLLM:   fallbackLLM,
		config:        cfg,
		cache:         cache,
		store:         store,
		renderer:      renderer,
		mailExternal:  mailExternal,
		mqttSvc:       mqttSvc,
		llmTimeout:    5 * time.Minute,  // Increased for strategic reasoning depth
		renderTimeout: 30 * time.Second, // Default PDF render timeout
	}
}

func (u *summaryUseCase) summaryInternal(text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (*dtos.RAGSummaryResponseDTO, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}

	targetLangName := "Indonesian"
	if strings.EqualFold(language, "en") {
		targetLangName = "English"
	}

	// If MacAddress is provided, fetch booking info to complement metadata
	if macAddress != "" {
		macAddress = strings.ToUpper(strings.TrimSpace(macAddress)) // Normalize MAC
		if u.mailExternal != nil {
			info, err := u.mailExternal.GetDeviceInfoByMac(macAddress)
			if err == nil {
				if date == "" {
					date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxByendDate"])
					if date == "" || date == "<nil>" {
						date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxtimeendDate"]) // Fallback to other key
					}
				}
				if location == "" {
					location = fmt.Sprintf("%v", info["SDTGetRoomTeraluxRoomName"])
				}
				if participants == "" {
					participants = fmt.Sprintf("%v", info["SDTGetRoomTeraluxCustomerName"])
				}
				if meetingContext == "" {
					apiAgenda := fmt.Sprintf("%v", info["SDTGetRoomTeraluxMeetingAgenda"])
					if apiAgenda != "" && apiAgenda != "<nil>" {
						meetingContext = apiAgenda
					}
				}
				utils.LogDebug("SummaryUseCase: Fetched booking metadata for MAC %s: Date=%s, Location=%s", macAddress, date, location)
			} else {
				utils.LogWarn("SummaryUseCase: Failed to fetch booking metadata for MAC %s: %v", macAddress, err)
			}
		}
	}

	// Delegate prompt generation and LLM call to SummarySkill
	skill := &skills.SummarySkill{}
	ctx := &skills.SkillContext{
		Prompt:       text,
		Language:     language,
		LLM:          u.llm,
		Config:       u.config,
		Date:         date,
		Location:     location,
		Participants: participants,
		Style:        style,
		Context:      meetingContext,
	}

	res, err := skill.Execute(ctx)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("SummaryTask: Primary LLM failed, falling back to local model: %v", err)
		ctx.LLM = u.fallbackLLM
		res, err = skill.Execute(ctx)
	}

	if err != nil {
		return nil, err
	}

	trimmedSummary := res.Message
	inferredAgenda := ""

	// Extract # AGENDA: [content] if present
	if strings.Contains(trimmedSummary, "# AGENDA:") {
		lines := strings.Split(trimmedSummary, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "# AGENDA:") {
				inferredAgenda = strings.TrimSpace(strings.TrimPrefix(line, "# AGENDA:"))
				// Remove the specific agenda line from the summary body
				trimmedSummary = strings.Join(append(lines[:i], lines[i+1:]...), "\n")
				break
			}
		}
	}

	if meetingContext == "" && inferredAgenda != "" {
		meetingContext = inferredAgenda
	}

	// Generate PDF
	uuidStr, _ := uuid.NewV7()
	pdfFilename := fmt.Sprintf("summary_%s.pdf", uuidStr.String())

	// Determine backend root to ensure uploads are correctly placed
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	pdfPath := filepath.Join(basePath, "uploads", "reports", pdfFilename)

	// Create reports directory if not exists
	_ = os.MkdirAll(filepath.Dir(pdfPath), 0755)

	if u.renderer != nil {
		meta := services.SummaryPDFMeta{
			Language:     targetLangName,
			Context:      meetingContext,
			Style:        style,
			Date:         date,
			Location:     location,
			Participants: participants,
			CustomerName: "Internal User", // Default or fetch from somewhere if available
			CompanyName:  "Sensio",        // Matching mail domain default
		}
		if err := u.renderer.Render(trimmedSummary, pdfPath, meta); err != nil {
			utils.LogWarn("Warning: Failed to generate PDF: %v", err)
		}
	} else {
		utils.LogWarn("Warning: PDF renderer is not configured")
	}

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)
	if baseURL != "" {
		pdfUrl = fmt.Sprintf("%s%s", baseURL, pdfUrl)
	}

	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, pdf='%s'", language, len(trimmedSummary), pdfUrl)
	utils.LogDebug("RAG Summary Result: %q", trimmedSummary)

	// Cache inferred agenda if MAC is provided to share with mail domain
	if macAddress != "" && inferredAgenda != "" {
		cacheKey := fmt.Sprintf("agenda_mac_%s", macAddress)
		_ = u.cache.Set(cacheKey, inferredAgenda)
		utils.LogDebug("SummaryUseCase: Cached inferred agenda for MAC %s: %s", macAddress, inferredAgenda)
	}

	return &dtos.RAGSummaryResponseDTO{
		Summary:       trimmedSummary,
		PDFUrl:        pdfUrl,
		AgendaContext: meetingContext,
	}, nil
}

func (u *summaryUseCase) SummarizeText(text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (string, error) {
	return u.SummarizeTextWithTrigger(text, language, meetingContext, style, date, location, participants, macAddress, baseURL, "")
}

func (u *summaryUseCase) SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL, trigger string) (string, error) {
	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:     "pending",
		Trigger:    trigger,
		MacAddress: macAddress,
		StartedAt:  time.Now().Format(time.RFC3339),
	}
	u.store.Set(taskID, status)

	_ = u.cache.Set(taskID, status)

	go func() {
		result, err := u.summaryInternal(text, language, meetingContext, style, date, location, participants, macAddress, baseURL)
		if err != nil {
			utils.LogError("RAG Summary Task %s: Failed with error: %v", taskID, err)
			u.updateStatus(taskID, "failed", err, nil)
		} else {
			utils.LogInfo("RAG Summary Task %s: Completed successfully", taskID)
			u.updateStatus(taskID, "completed", nil, result)
		}
	}()

	return taskID, nil
}

// SummarizeTextWithContext provides context-aware processing with built-in timeout and cancellation support.
// This method should be preferred over SummarizeText for API-driven use cases.
//
// Parameters:
//   - ctx: Cancellation context (can define timeout: ctx, cancel := context.WithTimeout(...))
//   - text, language, context, style: Same as SummarizeText
//
// Returns:
//   - Task ID (use status polling)
//   - Error if context is invalid or task startup fails
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	taskID, err := useCase.SummarizeTextWithContext(ctx, transcript, "en", "meeting", "executive")
func (u *summaryUseCase) SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (string, error) {
	return u.SummarizeTextWithContextAndTrigger(ctx, text, language, meetingContext, style, date, location, participants, macAddress, baseURL, "")
}

func (u *summaryUseCase) SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL, trigger string) (string, error) {
	// Validate context is not already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:     "pending",
		Trigger:    trigger,
		MacAddress: macAddress,
		StartedAt:  time.Now().Format(time.RFC3339),
	}
	u.store.Set(taskID, status)
	_ = u.cache.Set(taskID, status)

	go func() {
		// Use parent context directly for cancellation, but remove the internal time-based limit
		internalCtx := ctx

		// Check if parent context was cancelled before starting
		select {
		case <-ctx.Done():
			u.updateStatus(taskID, "cancelled", ctx.Err(), nil)
			return
		default:
		}

		result, err := u.summaryInternalWithContext(internalCtx, text, language, meetingContext, style, date, location, participants, macAddress, baseURL)
		if err != nil {
			utils.LogError("RAG Summary Task %s: Failed with error: %v", taskID, err)
			u.updateStatus(taskID, "failed", err, nil)
		} else {
			utils.LogInfo("RAG Summary Task %s: Completed successfully", taskID)
			u.updateStatus(taskID, "completed", nil, result)
		}
	}()

	return taskID, nil
}

// SetLLMTimeout configures the timeout for LLM calls
func (u *summaryUseCase) SetLLMTimeout(d time.Duration) {
	if d > 0 {
		u.llmTimeout = d
	}
}

// SetRenderTimeout configures the timeout for PDF rendering
func (u *summaryUseCase) SetRenderTimeout(d time.Duration) {
	if d > 0 {
		u.renderTimeout = d
	}
}

// summaryInternalWithContext adds context-aware timeout enforcement to PDF rendering and LLM calls.
func (u *summaryUseCase) summaryInternalWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress, baseURL string) (*dtos.RAGSummaryResponseDTO, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}

	// If MacAddress is provided, fetch booking info
	if macAddress != "" {
		macAddress = strings.ToUpper(strings.TrimSpace(macAddress)) // Normalize MAC
		if u.mailExternal != nil {
			info, err := u.mailExternal.GetDeviceInfoByMac(macAddress)
			if err == nil {
				if date == "" {
					date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxByendDate"])
					if date == "" || date == "<nil>" {
						date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxtimeendDate"]) // Fallback
					}
				}
				if location == "" {
					location = fmt.Sprintf("%v", info["SDTGetRoomTeraluxRoomName"])
				}
				if participants == "" {
					participants = fmt.Sprintf("%v", info["SDTGetRoomTeraluxCustomerName"])
				}
				if meetingContext == "" {
					apiAgenda := fmt.Sprintf("%v", info["SDTGetRoomTeraluxMeetingAgenda"])
					if apiAgenda != "" && apiAgenda != "<nil>" {
						meetingContext = apiAgenda
					}
				}
				utils.LogDebug("SummaryUseCase (WithContext): Fetched booking metadata for MAC %s: Date=%s, Location=%s", macAddress, date, location)
			} else {
				utils.LogWarn("SummaryUseCase (WithContext): Failed to fetch booking metadata for MAC %s: %v", macAddress, err)
			}
		}
	}

	// Delegate prompt generation and LLM call to SummarySkill with context (manually enforced for now)
	skill := &skills.SummarySkill{}
	ctxSkill := &skills.SkillContext{
		Prompt:       text,
		Language:     language,
		LLM:          u.llm,
		Config:       u.config,
		Date:         date,
		Location:     location,
		Participants: participants,
		Style:        style,
		Context:      meetingContext,
	}

	// Calculate targetLangName for PDF Meta (logic moved from deleted block)
	targetLangName := "Indonesian"
	if strings.EqualFold(language, "en") {
		targetLangName = "English"
	}

	// We can't easily pass the context cancellation to Skill.Execute yet without modifying Skill interface.
	// For now, we rely on the skill execution.
	res, err := skill.Execute(ctxSkill)
	if err != nil && u.fallbackLLM != nil {
		utils.LogWarn("SummaryTask (WithContext): Primary LLM failed, falling back to local model: %v", err)
		ctxSkill.LLM = u.fallbackLLM
		res, err = skill.Execute(ctxSkill)
	}

	if err != nil {
		return nil, fmt.Errorf("SummarySkill execution failed: %w", err)
	}

	trimmedSummary := res.Message
	inferredAgenda := ""

	// Extract # AGENDA: Tag if present
	if strings.Contains(trimmedSummary, "# AGENDA:") {
		lines := strings.Split(trimmedSummary, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "# AGENDA:") {
				inferredAgenda = strings.TrimSpace(strings.TrimPrefix(line, "# AGENDA:"))
				// Remove the header from result
				trimmedSummary = strings.Join(append(lines[:i], lines[i+1:]...), "\n")
				break
			}
		}
	}

	if meetingContext == "" && inferredAgenda != "" {
		meetingContext = inferredAgenda
	}

	// PDF rendering with timeout
	uuidStr, _ := uuid.NewV7()
	pdfFilename := fmt.Sprintf("summary_%s.pdf", uuidStr.String())

	// Determine backend root
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	pdfPath := filepath.Join(basePath, "uploads", "reports", pdfFilename)
	_ = os.MkdirAll(filepath.Dir(pdfPath), 0755)

	if u.renderer != nil {
		// Use parent context directly
		renderCtx := ctx

		// Check render context before starting
		select {
		case <-renderCtx.Done():
			utils.LogWarn("Warning: PDF render timeout will trigger during rendering")
		default:
		}

		meta := services.SummaryPDFMeta{
			Language:     targetLangName,
			Context:      meetingContext,
			Style:        style,
			Date:         date,
			Location:     location,
			Participants: participants,
			CustomerName: "Internal User",
			CompanyName:  "Sensio",
		}

		if err := u.renderer.Render(trimmedSummary, pdfPath, meta); err != nil {
			utils.LogWarn("Warning: Failed to generate PDF: %v (will continue with text-only response)", err)
		}
	}

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)
	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, pdf='%s'", language, len(trimmedSummary), pdfUrl)

	// Cache inferred agenda if MAC is provided
	if macAddress != "" && inferredAgenda != "" {
		cacheKey := fmt.Sprintf("agenda_mac_%s", macAddress)
		_ = u.cache.Set(cacheKey, inferredAgenda)
		utils.LogDebug("SummaryUseCase (WithContext): Cached inferred agenda for MAC %s: %s", macAddress, inferredAgenda)
	}

	return &dtos.RAGSummaryResponseDTO{
		Summary:       trimmedSummary,
		PDFUrl:        pdfUrl,
		AgendaContext: meetingContext,
	}, nil
}

func (u *summaryUseCase) updateStatus(taskID string, statusStr string, err error, result *dtos.RAGSummaryResponseDTO) {
	var existing dtos.RAGStatusDTO
	_, _, _ = u.cache.GetWithTTL(taskID, &existing)

	status := &dtos.RAGStatusDTO{
		Status:     statusStr,
		StartedAt:  existing.StartedAt,
		Trigger:    existing.Trigger,
		MacAddress: existing.MacAddress,
		ExpiresAt:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.Result = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	}

	if result != nil {
		status.ExecutionResult = result
		status.Result = result.Summary
		status.HTTPStatusCode = 200
	}

	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}

		// Send MQTT "stop" signal if MacAddress is available
		if status.MacAddress != "" && u.mqttSvc != nil {
			taskTopic := fmt.Sprintf("users/%s/task", status.MacAddress)
			msg := map[string]string{
				"event": "stop",
				"task":  "RAG",
			}
			payload, _ := json.Marshal(msg)
			if err := u.mqttSvc.Publish(taskTopic, 0, false, payload); err != nil {
				utils.LogError("RAG Summary Task %s: Failed to publish stop signal to MQTT: %v", taskID, err)
			} else {
				utils.LogInfo("RAG Summary Task %s: Published stop signal to %s", taskID, taskTopic)
			}
		}
	}

	u.store.Set(taskID, status)
	_ = u.cache.SetPreserveTTL(taskID, status)
}
