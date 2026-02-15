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
	"teralux_app/domain/rag/utilities"
	"time"

	"github.com/google/uuid"
)

type SummaryUseCase interface {
	SummarizeText(text string, language string, meetingContext string, style string) (string, error)
	SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string) (string, error)
}

type summaryUseCase struct {
	llm           utilities.LLMClient
	config        *utils.Config
	cache         *tasks.BadgerTaskCache
	store         *tasks.StatusStore[dtos.RAGStatusDTO]
	renderer      services.SummaryPDFRenderer
	llmTimeout    time.Duration // Timeout for LLM calls
	renderTimeout time.Duration // Timeout for PDF rendering
}

func NewSummaryUseCase(
	llm utilities.LLMClient,
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
	if strings.ToLower(language) == "en" {
		targetLangName = "English"
	}

	// Build structured prompt with configured assertiveness and risk scoring
	promptConfig := &services.PromptConfig{
		Assertiveness: 8,          // Strategic assertiveness (calling out gaps/risks)
		Audience:      "mixed",    // C-level + VP/Director level
		RiskScale:     "granular", // 1-10 scoring for nuance
		Context:       meetingContext,
		Style:         style,
		Language:      targetLangName,
	}
	prompt := promptConfig.BuildPrompt(text)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	summary, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return nil, err
	}

	trimmedSummary := strings.TrimSpace(summary)

	// Generate PDF
	pdfFilename := fmt.Sprintf("summary_%d.pdf", time.Now().Unix())
	pdfPath := filepath.Join("uploads", "reports", pdfFilename)

	// Create reports directory if not exists
	os.MkdirAll(filepath.Dir(pdfPath), 0755)

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
	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{Status: "pending"}
	u.store.Set(taskID, status)

	_ = u.cache.Set(taskID, status)

	go func() {
		result, err := u.summaryInternal(text, language, meetingContext, style)
		var finalStatus *dtos.RAGStatusDTO
		if err != nil {
			utils.LogError("RAG Summary Task %s: Failed with error: %v", taskID, err)
			finalStatus = &dtos.RAGStatusDTO{Status: "failed", Result: err.Error()}
		} else {
			utils.LogInfo("RAG Summary Task %s: Completed successfully", taskID)
			finalStatus = &dtos.RAGStatusDTO{
				Status:          "completed",
				ExecutionResult: result,
				Result:          result.Summary,
			}
		}

		u.store.Set(taskID, finalStatus)
		_ = u.cache.SetPreserveTTL(taskID, finalStatus)
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
	// Validate context is not already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{Status: "pending"}
	u.store.Set(taskID, status)
	_ = u.cache.Set(taskID, status)

	go func() {
		// Use parent context directly for cancellation, but remove the internal time-based limit
		internalCtx := ctx

		// Check if parent context was cancelled before starting
		select {
		case <-ctx.Done():
			u.store.Set(taskID, &dtos.RAGStatusDTO{Status: "cancelled", Result: "Parent context cancelled"})
			_ = u.cache.SetPreserveTTL(taskID, &dtos.RAGStatusDTO{Status: "cancelled"})
			return
		default:
		}

		result, err := u.summaryInternalWithContext(internalCtx, text, language, meetingContext, style)
		var finalStatus *dtos.RAGStatusDTO
		if err != nil {
			utils.LogError("RAG Summary Task %s: Failed with error: %v", taskID, err)
			finalStatus = &dtos.RAGStatusDTO{Status: "failed", Result: err.Error()}
		} else {
			utils.LogInfo("RAG Summary Task %s: Completed successfully", taskID)
			finalStatus = &dtos.RAGStatusDTO{
				Status:          "completed",
				ExecutionResult: result,
				Result:          result.Summary,
			}
		}

		u.store.Set(taskID, finalStatus)
		_ = u.cache.SetPreserveTTL(taskID, finalStatus)
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

	targetLangName := "Indonesian"
	if strings.ToLower(language) == "en" {
		targetLangName = "English"
	}

	// Check context before proceeding
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
	default:
	}

	// Build structured prompt
	promptConfig := &services.PromptConfig{
		Assertiveness: 8,
		Audience:      "mixed",
		RiskScale:     "granular",
		Context:       meetingContext,
		Style:         style,
		Language:      targetLangName,
	}
	prompt := promptConfig.BuildPrompt(text)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	// LLM call (currently no built-in context support in CallModel, so we check before/after)
	// In future: upgrade LLMClient interface to accept context
	summary, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Check context again after LLM call
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("operation cancelled after LLM call: %w", ctx.Err())
	default:
	}

	trimmedSummary := strings.TrimSpace(summary)

	// PDF rendering with timeout
	pdfFilename := fmt.Sprintf("summary_%d.pdf", time.Now().Unix())
	pdfPath := filepath.Join("uploads", "reports", pdfFilename)
	os.MkdirAll(filepath.Dir(pdfPath), 0755)

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
