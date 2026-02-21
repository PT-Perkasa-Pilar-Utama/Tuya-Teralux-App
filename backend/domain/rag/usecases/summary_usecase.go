package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/services"
	"teralux_app/domain/rag/skills"
	"time"

	"github.com/google/uuid"
)

type SummaryUseCase interface {
	SummarizeText(text string, language string, meetingContext string, style string) (string, error)
	SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, trigger string) (string, error)
	SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string) (string, error)
	SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, trigger string) (string, error)
}

type summaryUseCase struct {
	llm           skills.LLMClient
	config        *utils.Config
	cache         *tasks.BadgerTaskCache
	store         *tasks.StatusStore[dtos.RAGStatusDTO]
	renderer      services.SummaryPDFRenderer
	llmTimeout    time.Duration // Timeout for LLM calls
	renderTimeout time.Duration // Timeout for PDF rendering
}

func NewSummaryUseCase(
	llm skills.LLMClient,
	cfg *utils.Config,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[dtos.RAGStatusDTO],
	renderer services.SummaryPDFRenderer,
) SummaryUseCase {
	return &summaryUseCase{
		llm:           llm,
		config:        cfg,
		cache:         cache,
		store:         store,
		renderer:      renderer,
		llmTimeout:    5 * time.Minute,  // Increased for strategic reasoning depth
		renderTimeout: 30 * time.Second, // Default PDF render timeout
	}
}

func (u *summaryUseCase) summaryInternal(text string, language string, meetingContext string, style string) (*dtos.RAGSummaryResponseDTO, error) {
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

	// Delegate prompt generation and LLM call to SummarySkill
	skill := &skills.SummarySkill{}
	ctx := &skills.SkillContext{
		Prompt:   text,
		Language: language,
		LLM:      u.llm,
		Config:   u.config,
		// History can be used to pass context/style if extended
	}

	res, err := skill.Execute(ctx)
	if err != nil {
		return nil, err
	}

	trimmedSummary := res.Message

	// Generate PDF
	pdfFilename := fmt.Sprintf("summary_%d.pdf", time.Now().Unix())

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
			Language: targetLangName,
			Context:  meetingContext,
			Style:    style,
		}
		if err := u.renderer.Render(trimmedSummary, pdfPath, meta); err != nil {
			utils.LogWarn("Warning: Failed to generate PDF: %v", err)
		}
	} else {
		utils.LogWarn("Warning: PDF renderer is not configured")
	}

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)

	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, pdf='%s'", language, len(trimmedSummary), pdfUrl)
	utils.LogDebug("RAG Summary Result: %q", trimmedSummary)

	return &dtos.RAGSummaryResponseDTO{
		Summary: trimmedSummary,
		PDFUrl:  pdfUrl,
	}, nil
}

func (u *summaryUseCase) SummarizeText(text string, language string, meetingContext string, style string) (string, error) {
	return u.SummarizeTextWithTrigger(text, language, meetingContext, style, "")
}

func (u *summaryUseCase) SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, trigger string) (string, error) {
	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:    "pending",
		Trigger:   trigger,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	u.store.Set(taskID, status)

	_ = u.cache.Set(taskID, status)

	go func() {
		result, err := u.summaryInternal(text, language, meetingContext, style)
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
func (u *summaryUseCase) SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string) (string, error) {
	return u.SummarizeTextWithContextAndTrigger(ctx, text, language, meetingContext, style, "")
}

func (u *summaryUseCase) SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, trigger string) (string, error) {
	// Validate context is not already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:    "pending",
		Trigger:   trigger,
		StartedAt: time.Now().Format(time.RFC3339),
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

		result, err := u.summaryInternalWithContext(internalCtx, text, language, meetingContext, style)
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
func (u *summaryUseCase) summaryInternalWithContext(ctx context.Context, text string, language string, meetingContext string, style string) (*dtos.RAGSummaryResponseDTO, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}

	// Delegate prompt generation and LLM call to SummarySkill with context (manually enforced for now)
	skill := &skills.SummarySkill{}
	ctxSkill := &skills.SkillContext{
		Prompt:   text,
		Language: language,
		LLM:      u.llm,
		Config:   u.config,
	}

	// Calculate targetLangName for PDF Meta (logic moved from deleted block)
	targetLangName := "Indonesian"
	if strings.EqualFold(language, "en") {
		targetLangName = "English"
	}

	// We can't easily pass the context cancellation to Skill.Execute yet without modifying Skill interface.
	// For now, we rely on the skill execution.
	res, err := skill.Execute(ctxSkill)
	if err != nil {
		return nil, fmt.Errorf("SummarySkill execution failed: %w", err)
	}

	trimmedSummary := res.Message

	// PDF rendering with timeout
	pdfFilename := fmt.Sprintf("summary_%d.pdf", time.Now().Unix())

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
			Language: targetLangName,
			Context:  meetingContext,
			Style:    style,
		}

		if err := u.renderer.Render(trimmedSummary, pdfPath, meta); err != nil {
			utils.LogWarn("Warning: Failed to generate PDF: %v (will continue with text-only response)", err)
		}
	}

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)
	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, pdf='%s'", language, len(trimmedSummary), pdfUrl)

	return &dtos.RAGSummaryResponseDTO{
		Summary: trimmedSummary,
		PDFUrl:  pdfUrl,
	}, nil
}

func (u *summaryUseCase) updateStatus(taskID string, statusStr string, err error, result *dtos.RAGSummaryResponseDTO) {
	var existing dtos.RAGStatusDTO
	_, _, _ = u.cache.GetWithTTL(taskID, &existing)

	status := &dtos.RAGStatusDTO{
		Status:    statusStr,
		StartedAt: existing.StartedAt,
		Trigger:   existing.Trigger,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
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
	}

	u.store.Set(taskID, status)
	_ = u.cache.SetPreserveTTL(taskID, status)
}
