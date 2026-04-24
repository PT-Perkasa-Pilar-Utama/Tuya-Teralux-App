package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/dtos"
	ragServices "sensio/domain/models/rag/services"
	"sensio/domain/models/rag/skills"
	"sensio/domain/pdf_dlq/entities"
	speechUsecases "sensio/domain/speech/usecases"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IntermediateSummaryNote represents structured extraction from a transcript window
// Used in hierarchical summarization map phase
type IntermediateSummaryNote struct {
	WindowID      int      `json:"window_id"`
	Topic         string   `json:"topic,omitempty"`
	Decisions     []string `json:"decisions,omitempty"`
	ActionItems   []string `json:"action_items,omitempty"`
	OpenQuestions []string `json:"open_questions,omitempty"`
	Risks         []string `json:"risks,omitempty"`
	SpeakerRefs   []string `json:"speaker_refs,omitempty"` // Speaker labels mentioned
	TimeRefs      []string `json:"time_refs,omitempty"`    // Timestamp references if available
	Summary       string   `json:"summary"`                // Brief narrative summary of this window
}

// securePDFProcessor defines an interface for PDF protection to avoid import cycle
type securePDFProcessor interface {
	ProtectAndStore(ctx context.Context, pdfPath string) (string, string, error)
}

// downloadTokenCreator defines interface for token creation to avoid import cycle
type downloadTokenCreator interface {
	CreateToken(recipient, objectKey, purpose string, password ...string) (string, error)
}

// pdfDeadLetterCreator defines interface for DLQ entry creation
type pdfDeadLetterCreator interface {
	Create(entry *entities.PDFDeadLetter) error
}

// HierarchicalSummaryResult encapsulates the output of hierarchical structured summarization
type HierarchicalSummaryResult struct {
	FinalSummary      string                    `json:"final_summary"`
	IntermediateNotes []IntermediateSummaryNote `json:"intermediate_notes"`
	CoverageStats     *dtos.CoverageStats       `json:"coverage_stats,omitempty"`
	ActionItems       []dtos.ActionItem         `json:"action_items,omitempty"`
	Decisions         []dtos.Decision           `json:"decisions,omitempty"`
	OpenIssues        []dtos.OpenIssue          `json:"open_issues,omitempty"`
	Risks             []dtos.Risk               `json:"risks,omitempty"`
}

type SummaryUseCase interface {
	SummarizeText(text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error)
	SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error)
	SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error)
	SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error)
	SummarizeTextSync(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress string) (*dtos.RAGSummaryResponseDTO, error)
}

// summaryUseCase orchestrates meeting summary generation.
//
// SRP/ARCHITECTURE NOTE: This use case currently has multiple responsibilities:
// - Metadata enrichment (fetching booking info from external service)
// - Provider resolution and fallback chain execution
// - Hierarchical map-reduce summarization orchestration
// - Prompt generation for map/reduce phases
// - Structured artifact extraction (decisions, action items, risks, etc.)
// - PDF rendering coordination
// - Agenda caching
//
// Future refactoring opportunities (Clean Architecture):
// 1. Extract metadata enrichment to separate service/interactor
// 2. Extract hierarchical summarization orchestration to dedicated orchestrator
// 3. Extract structured artifact extraction to parser/extractor component
// 4. Keep use case focused on orchestration flow only
//
// Current risk: Change cost and test coupling trending up. However, functionality
// is stable and backward-compatible. Refactor when adding significant new features.
type summaryUseCase struct {
	llm                       skills.LLMClient
	fallbackLLM               skills.LLMClient
	config                    *utils.Config
	cache                     *tasks.BadgerTaskCache
	store                     *tasks.StatusStore[dtos.RAGStatusDTO]
	renderer                  ragServices.SummaryPDFRenderer
	securePDFUC               securePDFProcessor
	tokenService              downloadTokenCreator
	bigExternal               *commonServices.DeviceInfoExternalService
	mqttSvc                   mqttPublisher
	llmTimeout                time.Duration
	renderTimeout             time.Duration
	skill                     skills.Skill
	chunkSkill                skills.Skill
	structuredExtractionSkill skills.Skill
	providerResolver          speechUsecases.ProviderResolver
	pdfDLQC                   pdfDeadLetterCreator
}

func NewSummaryUseCase(
	llm skills.LLMClient,
	fallbackLLM skills.LLMClient,
	cfg *utils.Config,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[dtos.RAGStatusDTO],
	renderer ragServices.SummaryPDFRenderer,
	securePDFUC securePDFProcessor,
	tokenService downloadTokenCreator,
	bigExternal *commonServices.DeviceInfoExternalService,
	mqttSvc mqttPublisher,
	skill skills.Skill,
	chunkSkill skills.Skill,
	structuredExtractionSkill skills.Skill,
	providerResolver speechUsecases.ProviderResolver,
	pdfDLQC pdfDeadLetterCreator,
) SummaryUseCase {
	return &summaryUseCase{
		llm:                       llm,
		fallbackLLM:               fallbackLLM,
		config:                    cfg,
		cache:                     cache,
		store:                     store,
		renderer:                  renderer,
		securePDFUC:               securePDFUC,
		tokenService:              tokenService,
		bigExternal:               bigExternal,
		mqttSvc:                   mqttSvc,
		llmTimeout:                5 * time.Minute,
		renderTimeout:             30 * time.Second,
		skill:                     skill,
		chunkSkill:                chunkSkill,
		structuredExtractionSkill: structuredExtractionSkill,
		providerResolver:          providerResolver,
		pdfDLQC:                   pdfDLQC,
	}
}

func (u *summaryUseCase) summaryInternal(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress string) (*dtos.RAGSummaryResponseDTO, error) {
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

	var room string
	// Fetch booking info via BigExternal if MAC provided
	if macAddress != "" {
		macAddress = strings.ToUpper(strings.TrimSpace(macAddress))
		if u.bigExternal != nil {
			info, err := u.bigExternal.GetDeviceInfoByMac(macAddress)
			if err == nil {
				if date == "" {
					date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxByendDate"])
					if date == "" || date == "<nil>" {
						date = fmt.Sprintf("%v", info["SDTGetRoomTeraluxtimeendDate"])
					}
				}

				roomRaw := fmt.Sprintf("%v", info["SDTGetRoomTeraluxRoomName"])
				if roomRaw != "<nil>" && roomRaw != "" {
					room = roomRaw
				}

				// Always override location with building name from device info when macAddress is provided
				building := fmt.Sprintf("%v", info["SDTGetRoomTeraluxBuildingsName"])
				if building != "" && building != "<nil>" {
					location = building
				} else {
					location = room
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
			}
		}
	}

	if u.skill == nil {
		return nil, fmt.Errorf("summary skill not configured")
	}

	// Use health-aware fallback chain for summarization with terminal preference
	executeWithFallback := func(ctx context.Context, prompt string, language string, meetingContext string, style string, date string, location string, participants string) (*skills.SkillResult, error) {
		var result *skills.SkillResult
		var err error

		if macAddress != "" {
			// Use terminal-specific provider preference
			err = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				skillCtx := &skills.SkillContext{
					Ctx:          ctx,
					Prompt:       prompt,
					Language:     language,
					LLM:          resolvedSet.LLM,
					Config:       u.config,
					Date:         date,
					Location:     location,
					Participants: participants,
					Style:        style,
					Context:      meetingContext,
				}
				res, execErr := u.skill.Execute(skillCtx)
				if execErr == nil {
					result = res
				}
				return execErr
			})
		} else {
			// Use standard health-aware fallback
			err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				skillCtx := &skills.SkillContext{
					Ctx:          ctx,
					Prompt:       prompt,
					Language:     language,
					LLM:          resolvedSet.LLM,
					Config:       u.config,
					Date:         date,
					Location:     location,
					Participants: participants,
					Style:        style,
					Context:      meetingContext,
				}
				res, execErr := u.skill.Execute(skillCtx)
				if execErr == nil {
					result = res
				}
				return execErr
			})
		}
		return result, err
	}

	var trimmedSummary string
	var inferredAgenda string
	var res *skills.SkillResult
	var err error

	// Track summary mode for observability
	summaryMode := "single_pass"
	var hierarchicalResult *HierarchicalSummaryResult

	// Hierarchical Structured Summarization for long transcripts
	// Threshold: >16,000 chars (~4000 tokens) triggers hierarchical path
	if len(text) > 4000*4 {
		utils.LogInfo("SummaryUseCase: Text too long (%d chars), using hierarchical structured summarization", len(text))
		hierarchicalResult, err = u.summarizeHierarchical(ctx, text, language, meetingContext, style, date, location, participants, macAddress)
		if err == nil {
			summaryMode = "hierarchical_structured"
			trimmedSummary = hierarchicalResult.FinalSummary
			utils.LogInfo("SummaryUseCase: Hierarchical summarization completed with %d intermediate notes", len(hierarchicalResult.IntermediateNotes))
		} else {
			utils.LogError("SummaryUseCase: Hierarchical summarization failed: %v. Falling back to chunked.", err)
		}
	}

	// Fallback to legacy chunked summarization if hierarchical failed
	if trimmedSummary == "" && len(text) > 4000*4 {
		utils.LogInfo("SummaryUseCase: Falling back to chunked summarization")
		chunkedSummary, chunkErr := u.summarizeInChunks(ctx, text, language, meetingContext, macAddress)
		if chunkErr == nil {
			res, err = executeWithFallback(ctx, chunkedSummary, language, meetingContext, style, date, location, participants)
			if err == nil {
				trimmedSummary = res.Message
			}
		} else {
			utils.LogError("SummaryUseCase: Chunked summarization failed: %v. Falling back to single-pass.", chunkErr)
		}
	}

	// Single-pass execution for short transcripts or fallback
	if trimmedSummary == "" {
		res, err = executeWithFallback(ctx, text, language, meetingContext, style, date, location, participants)
		if err != nil {
			return nil, err
		}
		trimmedSummary = res.Message
	}

	// Post-processing: Extract Agenda
	if strings.Contains(trimmedSummary, "# AGENDA:") {
		lines := strings.Split(trimmedSummary, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "# AGENDA:") {
				inferredAgenda = strings.TrimSpace(strings.TrimPrefix(line, "# AGENDA:"))
				trimmedSummary = strings.Join(append(lines[:i], lines[i+1:]...), "\n")
				break
			}
		}
	}

	if meetingContext == "" && inferredAgenda != "" {
		meetingContext = inferredAgenda
	}

	// PDF Generation
	uuidStr, _ := uuid.NewV7()
	tmpPdfFilename := fmt.Sprintf("summary_%s.pdf", uuidStr.String())
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	tmpPdfPath := filepath.Join(basePath, "uploads", "reports", tmpPdfFilename)
	_ = os.MkdirAll(filepath.Dir(tmpPdfPath), 0755)

	pdfUrl := ""
	if u.renderer != nil {
		meta := ragServices.SummaryPDFMeta{
			Language:     targetLangName,
			Context:      meetingContext,
			Style:        style,
			Date:         date,
			Location:     location,
			Room:         room,
			Participants: participants,
			CustomerName: "Internal User",
			CompanyName:  "Sensio",
		}
		if err := u.renderer.Render(trimmedSummary, tmpPdfPath, meta); err != nil {
			utils.LogError("PDF: generation failed: %v", err)
			pdfUrl = ""
			u.recordPDFDLQ(uuidStr.String(), fmt.Sprintf("pdf_render: %v", err))
		} else if u.securePDFUC != nil {
			s3Key, password, err := u.securePDFUC.ProtectAndStore(ctx, tmpPdfPath)
			if err != nil {
				utils.LogError("SecurePDF: failed to protect PDF: %v", err)
				pdfUrl = ""
				u.recordPDFDLQ(uuidStr.String(), fmt.Sprintf("s3_upload: %v", err))
			} else {
				tokenID, err := u.tokenService.CreateToken("", s3Key, "summary_pdf", password)
				if err != nil {
					utils.LogError("SecurePDF: failed to create token: %v", err)
					pdfUrl = ""
					u.recordPDFDLQ(uuidStr.String(), fmt.Sprintf("token_create: %v", err))
				} else {
					pdfUrl = fmt.Sprintf("/api/download/resolve/%s", tokenID)
				}
			}
			os.Remove(tmpPdfPath) //nolint:errcheck
		} else {
			_ = os.Remove(tmpPdfPath)
		}
	}

	// Cache inferred agenda
	if macAddress != "" && inferredAgenda != "" {
		_ = u.cache.Set(fmt.Sprintf("agenda_mac_%s", macAddress), inferredAgenda)
	}

	// Build response with structured fields (backward compatible)
	response := &dtos.RAGSummaryResponseDTO{
		Summary:       trimmedSummary,
		PDFUrl:        pdfUrl,
		AgendaContext: meetingContext,
		SummaryMode:   summaryMode,
	}

	// Populate structured fields if hierarchical summarization was used
	if hierarchicalResult != nil {
		response.SummaryVersion = "2.0-structured"
		response.ActionItems = hierarchicalResult.ActionItems
		response.Decisions = hierarchicalResult.Decisions
		response.OpenIssues = hierarchicalResult.OpenIssues
		response.Risks = hierarchicalResult.Risks
		response.CoverageStats = hierarchicalResult.CoverageStats
	}

	return response, nil
}

func (u *summaryUseCase) summarizeInChunks(ctx context.Context, text string, language string, meetingContext string, args ...string) (string, error) {
	if u.chunkSkill == nil {
		return "", fmt.Errorf("chunk summary skill not configured")
	}

	chunks := u.splitText(text, 16000) // ~4000 tokens
	//nolint:prealloc
	var intermediateSummaries []string

	// Check for macAddress in args for terminal-specific provider preference
	macAddress := ""
	if len(args) > 0 {
		macAddress = args[0]
	}

	for idx, chunk := range chunks {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		utils.LogInfo("summarizeInChunks: Processing chunk %d/%d", idx+1, len(chunks))

		var chunkSummary string
		var err error

		if macAddress != "" {
			// Use terminal-specific provider preference
			err = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				sCtx := &skills.SkillContext{
					Ctx:      ctx,
					Prompt:   chunk,
					Language: language,
					LLM:      resolvedSet.LLM,
					Config:   u.config,
					Context:  meetingContext,
				}
				res, execErr := u.chunkSkill.Execute(sCtx)
				if execErr == nil {
					chunkSummary = res.Message
				}
				return execErr
			})
		} else {
			// Use standard health-aware fallback
			err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				sCtx := &skills.SkillContext{
					Ctx:      ctx,
					Prompt:   chunk,
					Language: language,
					LLM:      resolvedSet.LLM,
					Config:   u.config,
					Context:  meetingContext,
				}
				res, execErr := u.chunkSkill.Execute(sCtx)
				if execErr == nil {
					chunkSummary = res.Message
				}
				return execErr
			})
		}

		if err != nil {
			utils.LogError("summarizeInChunks: Chunk %d failed: %v", idx+1, err)
			continue
		}
		intermediateSummaries = append(intermediateSummaries, chunkSummary)
	}

	if len(intermediateSummaries) == 0 {
		return "", fmt.Errorf("all chunk summaries failed")
	}

	return strings.Join(intermediateSummaries, "\n\n---\n\n"), nil
}

// summarizeHierarchical performs structured hierarchical summarization using map-reduce
// Map: Extract structured notes from transcript windows
// Reduce: Synthesize final summary from intermediate notes
func (u *summaryUseCase) summarizeHierarchical(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress string) (*HierarchicalSummaryResult, error) {
	if u.chunkSkill == nil {
		return nil, fmt.Errorf("chunk summary skill not configured")
	}

	// Use structured extraction skill for map phase if available, fallback to chunkSkill
	mapSkill := u.structuredExtractionSkill
	if mapSkill == nil {
		utils.LogWarn("summarizeHierarchical: structuredExtractionSkill not configured, falling back to chunkSkill (JSON extraction may fail)")
		mapSkill = u.chunkSkill
	}

	// Split into utterance-aware windows (not raw character chunks)
	windows := u.splitTextUtteranceAware(text, 16000) // ~4000 tokens per window
	utils.LogInfo("summarizeHierarchical: Split transcript into %d windows", len(windows))

	//nolint:prealloc
	var intermediateNotes []IntermediateSummaryNote
	var processedWindows int
	var emptyWindows int

	// MAP PHASE: Extract structured notes from each window
	for idx, window := range windows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		utils.LogInfo("summarizeHierarchical: Processing window %d/%d", idx+1, len(windows))

		var note IntermediateSummaryNote
		note.WindowID = idx

		// Build structured extraction prompt (JSON format)
		extractPrompt := u.buildMapPhasePrompt(window, language, idx)

		var err error
		if macAddress != "" {
			err = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				sCtx := &skills.SkillContext{
					Ctx:      ctx,
					Prompt:   extractPrompt,
					Language: language,
					LLM:      resolvedSet.LLM,
					Config:   u.config,
					Context:  meetingContext,
					Metadata: map[string]string{"window_id": fmt.Sprintf("%d", idx)},
				}
				res, execErr := mapSkill.Execute(sCtx)
				if execErr == nil {
					// Parse JSON response into structured note
					note, execErr = u.parseIntermediateNote(res.Message, idx)
				}
				return execErr
			})
		} else {
			err = u.providerResolver.ExecuteWithFallback(func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
				sCtx := &skills.SkillContext{
					Ctx:      ctx,
					Prompt:   extractPrompt,
					Language: language,
					LLM:      resolvedSet.LLM,
					Config:   u.config,
					Context:  meetingContext,
					Metadata: map[string]string{"window_id": fmt.Sprintf("%d", idx)},
				}
				res, execErr := mapSkill.Execute(sCtx)
				if execErr == nil {
					note, execErr = u.parseIntermediateNote(res.Message, idx)
				}
				return execErr
			})
		}

		if err != nil {
			// JSON parse failed but fallback may have produced a summary
			// Don't skip the window entirely if we have fallback content
			utils.LogWarn("summarizeHierarchical: Window %d JSON parse failed, using fallback summary", idx+1)

			// Check if fallback produced meaningful content
			if note.Summary == "" {
				utils.LogError("summarizeHierarchical: Window %d produced no content even in fallback", idx+1)
				emptyWindows++
				continue
			}
			// Include the note with just summary populated (structured fields will be empty)
			// This preserves coverage even for models that don't follow JSON contract
		}

		// Check if note has meaningful content
		if note.Summary == "" && len(note.Decisions) == 0 && len(note.ActionItems) == 0 {
			emptyWindows++
			continue
		}

		intermediateNotes = append(intermediateNotes, note)
		processedWindows++
	}

	if len(intermediateNotes) == 0 {
		return nil, fmt.Errorf("all windows failed to produce meaningful notes")
	}

	// REDUCE PHASE: Synthesize final summary from intermediate notes
	utils.LogInfo("summarizeHierarchical: Reducing %d intermediate notes into final summary", len(intermediateNotes))

	reducePrompt := u.buildReducePhasePrompt(intermediateNotes, language, meetingContext, style, date, location, participants)

	var finalResult *skills.SkillResult
	var reduceErr error

	if macAddress != "" {
		reduceErr = u.providerResolver.ExecuteWithFallbackByMac(macAddress, func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
			sCtx := &skills.SkillContext{
				Ctx:          ctx,
				Prompt:       reducePrompt,
				Language:     language,
				LLM:          resolvedSet.LLM,
				Config:       u.config,
				Date:         date,
				Location:     location,
				Participants: participants,
				Style:        style,
				Context:      meetingContext,
			}
			res, execErr := u.skill.Execute(sCtx)
			if execErr == nil {
				finalResult = res
			}
			return execErr
		})
	} else {
		reduceErr = u.providerResolver.ExecuteWithFallback(func(resolvedSet *speechUsecases.ResolvedProviderSet) error {
			sCtx := &skills.SkillContext{
				Ctx:          ctx,
				Prompt:       reducePrompt,
				Language:     language,
				LLM:          resolvedSet.LLM,
				Config:       u.config,
				Date:         date,
				Location:     location,
				Participants: participants,
				Style:        style,
				Context:      meetingContext,
			}
			res, execErr := u.skill.Execute(sCtx)
			if execErr == nil {
				finalResult = res
			}
			return execErr
		})
	}

	if reduceErr != nil {
		return nil, fmt.Errorf("reduce phase failed: %w", reduceErr)
	}

	// Build coverage stats
	coverageStats := &dtos.CoverageStats{
		TotalWindows:     len(windows),
		ProcessedWindows: processedWindows,
		EmptyWindows:     emptyWindows,
		CoverageRatio:    float64(processedWindows) / float64(len(windows)),
		SourceChars:      len(text),
		SummaryChars:     len(finalResult.Message),
		CompressionRatio: float64(len(finalResult.Message)) / float64(len(text)),
	}

	// Extract structured artifacts from intermediate notes
	actionItems := u.extractActionItemsFromNotes(intermediateNotes)
	decisions := u.extractDecisionsFromNotes(intermediateNotes)
	openIssues := u.extractOpenIssuesFromNotes(intermediateNotes)
	risks := u.extractRisksFromNotes(intermediateNotes)

	return &HierarchicalSummaryResult{
		FinalSummary:      finalResult.Message,
		IntermediateNotes: intermediateNotes,
		CoverageStats:     coverageStats,
		ActionItems:       actionItems,
		Decisions:         decisions,
		OpenIssues:        openIssues,
		Risks:             risks,
	}, nil
}

// buildMapPhasePrompt creates the prompt for extracting structured notes from a transcript window
func (u *summaryUseCase) buildMapPhasePrompt(windowText string, language string, windowID int) string {
	targetLangName := "Indonesian"
	if strings.EqualFold(language, "en") {
		targetLangName = "English"
	}

	return fmt.Sprintf(`You are extracting structured meeting notes from a transcript window.

**Task**: Read the transcript segment below and extract ONLY the following structured information. Output MUST be valid JSON.

**Output Format** (JSON):
{
  "topic": "Main topic discussed in this segment",
  "decisions": ["Decision 1", "Decision 2"],
  "action_items": ["Action item 1", "Action item 2"],
  "open_questions": ["Unresolved question 1"],
  "risks": ["Identified risk 1"],
  "speaker_refs": ["Speaker 1", "Speaker 2"],
  "summary": "2-3 sentence narrative summary of this segment"
}

**Rules**:
- If a field has no content, use empty array []
- Do NOT invent information - only extract what is explicitly stated
- Preserve uncertainty markers (e.g., "might", "possibly")
- Keep speaker references as they appear
- Write output in %s

**Transcript Segment (Window %d)**:
%s`, targetLangName, windowID, windowText)
}

// buildReducePhasePrompt creates the prompt for synthesizing final summary from intermediate notes
func (u *summaryUseCase) buildReducePhasePrompt(notes []IntermediateSummaryNote, language string, meetingContext string, style string, date string, location string, participants string) string {
	targetLangName := "Indonesian"
	if strings.EqualFold(language, "en") {
		targetLangName = "English"
	}

	// Serialize notes to JSON for the prompt
	notesJSON, _ := json.MarshalIndent(notes, "", "  ")

	return fmt.Sprintf(`You are creating a final meeting summary from structured intermediate notes.

**Context**:
- Meeting Context: %s
- Style: %s
- Date: %s
- Location: %s
- Participants: %s
- Language: %s

**Intermediate Notes** (from transcript windows):
%s

**Task**: Synthesize a comprehensive meeting summary using the intermediate notes above.

**Requirements**:
1. Preserve all decisions, action items, open issues, and risks from the notes
2. Do NOT invent ownership - if PIC is not specified, leave it blank
3. Preserve unresolved disagreements
4. Synthesize a comprehensive narrative that preserves discussion context and implications
5. Include important decisions, action items, open issues, and risks when present
6. Do not over-optimize for terseness - preserve important discussion substance
7. Balance structured sections with analytical narrative context
8. Write in %s

**Output**: Meeting summary in Markdown format.`, meetingContext, style, date, location, participants, targetLangName, string(notesJSON), targetLangName)
}

// parseIntermediateNote parses LLM response into structured IntermediateSummaryNote
// If JSON parsing fails, logs warning and returns fallback note with full response as summary
func (u *summaryUseCase) parseIntermediateNote(response string, windowID int) (IntermediateSummaryNote, error) {
	var note IntermediateSummaryNote

	// Try to parse as JSON first
	if err := json.Unmarshal([]byte(response), &note); err != nil {
		// JSON parse failed - log warning for observability
		utils.LogWarn("parseIntermediateNote: JSON parse failed for window %d (response truncated to 200 chars): %s", windowID, truncateString(response, 200))

		// Fallback: create minimal note with full response as summary
		// This allows pipeline to continue but structured fields will be empty
		note.WindowID = windowID
		note.Summary = response
		note.Decisions = []string{}
		note.ActionItems = []string{}
		note.OpenQuestions = []string{}
		note.Risks = []string{}

		return note, fmt.Errorf("JSON parse failed: %w", err)
	}

	// Ensure arrays are not nil
	if note.Decisions == nil {
		note.Decisions = []string{}
	}
	if note.ActionItems == nil {
		note.ActionItems = []string{}
	}
	if note.OpenQuestions == nil {
		note.OpenQuestions = []string{}
	}
	if note.Risks == nil {
		note.Risks = []string{}
	}

	return note, nil
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractActionItemsFromNotes converts intermediate notes to structured action items
func (u *summaryUseCase) extractActionItemsFromNotes(notes []IntermediateSummaryNote) []dtos.ActionItem {
	var items []dtos.ActionItem
	id := 1

	for _, note := range notes {
		for _, action := range note.ActionItems {
			items = append(items, dtos.ActionItem{
				ID:     id,
				Task:   action,
				PIC:    "", // Will be populated if speaker refs exist
				Status: "pending",
			})
			id++
		}
	}

	return items
}

// extractDecisionsFromNotes converts intermediate notes to structured decisions
func (u *summaryUseCase) extractDecisionsFromNotes(notes []IntermediateSummaryNote) []dtos.Decision {
	var decisions []dtos.Decision
	id := 1

	for _, note := range notes {
		for _, decision := range note.Decisions {
			decisions = append(decisions, dtos.Decision{
				ID:          id,
				Description: decision,
			})
			id++
		}
	}

	return decisions
}

// extractOpenIssuesFromNotes converts intermediate notes to structured open issues
func (u *summaryUseCase) extractOpenIssuesFromNotes(notes []IntermediateSummaryNote) []dtos.OpenIssue {
	var issues []dtos.OpenIssue
	id := 1

	for _, note := range notes {
		for _, question := range note.OpenQuestions {
			issues = append(issues, dtos.OpenIssue{
				ID:          id,
				Description: question,
			})
			id++
		}
	}

	return issues
}

// extractRisksFromNotes converts intermediate notes to structured risks
func (u *summaryUseCase) extractRisksFromNotes(notes []IntermediateSummaryNote) []dtos.Risk {
	var risks []dtos.Risk
	id := 1

	for _, note := range notes {
		for _, risk := range note.Risks {
			risks = append(risks, dtos.Risk{
				ID:          id,
				Description: risk,
			})
			id++
		}
	}

	return risks
}

// splitTextUtteranceAware splits text at utterance/paragraph boundaries instead of raw character positions
func (u *summaryUseCase) splitTextUtteranceAware(text string, maxChars int) []string {
	var windows []string

	for len(text) > maxChars {
		// Try to split at paragraph boundary first
		splitIdx := strings.LastIndex(text[:maxChars], "\n\n")

		// Fallback to sentence boundary
		if splitIdx == -1 {
			splitIdx = strings.LastIndex(text[:maxChars], ". ")
		}

		// Fallback to line boundary
		if splitIdx == -1 {
			splitIdx = strings.LastIndex(text[:maxChars], "\n")
		}

		// Hard cut as last resort
		if splitIdx == -1 {
			splitIdx = maxChars
		} else {
			splitIdx += 1 // Include the delimiter
		}

		windows = append(windows, text[:splitIdx])
		text = text[splitIdx:]
	}

	if len(text) > 0 {
		windows = append(windows, text)
	}

	return windows
}

func (u *summaryUseCase) splitText(text string, maxChars int) []string {
	var chunks []string
	for len(text) > maxChars {
		splitIdx := strings.LastIndex(text[:maxChars], "\n")
		if splitIdx == -1 {
			splitIdx = strings.LastIndex(text[:maxChars], ". ")
		}
		if splitIdx == -1 {
			splitIdx = maxChars
		} else {
			splitIdx += 1
		}
		chunks = append(chunks, text[:splitIdx])
		text = text[splitIdx:]
	}
	if len(text) > 0 {
		chunks = append(chunks, text)
	}
	return chunks
}

func (u *summaryUseCase) SummarizeTextSync(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress string) (*dtos.RAGSummaryResponseDTO, error) {
	return u.summaryInternal(ctx, text, language, meetingContext, style, date, location, participants, macAddress)
}

func (u *summaryUseCase) SummarizeText(text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error) {
	return u.summarizeWithTrigger(context.Background(), text, language, meetingContext, style, date, location, participants, "", args...)
}

func (u *summaryUseCase) SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error) {
	return u.summarizeWithTrigger(ctx, text, language, meetingContext, style, date, location, participants, "", args...)
}

func (u *summaryUseCase) SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error) {
	return u.summarizeWithTrigger(context.Background(), text, language, meetingContext, style, date, location, participants, trigger, args...)
}

func (u *summaryUseCase) SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error) {
	return u.summarizeWithTrigger(ctx, text, language, meetingContext, style, date, location, participants, trigger, args...)
}

func (u *summaryUseCase) summarizeWithTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	macAddress := ""
	idempotencyKey := ""
	if len(args) > 0 {
		macAddress = args[0]
	}
	if len(args) > 1 {
		idempotencyKey = args[1]
	}

	// Idempotency check with content hash
	var idempHash string
	if idempotencyKey != "" {
		idempHash = "idemp_summary_" + utils.HashString(fmt.Sprintf("%s_%s_%s_%s", idempotencyKey, language, macAddress, utils.HashString(text)))
		var existingID string
		if _, exists, _ := u.cache.GetWithTTL(idempHash, &existingID); exists && existingID != "" {
			// Check task state - only return if NOT failed
			status, ok := u.store.Get(existingID)
			if !ok || status == nil {
				// Fallback to cache if memory store is empty (after restart)
				var cachedStatus dtos.RAGStatusDTO
				if _, cachedExists, _ := u.cache.GetWithTTL(existingID, &cachedStatus); cachedExists {
					status = &cachedStatus
					u.store.Set(existingID, status)
					ok = true
				}
			}

			if ok && status != nil && status.Status != "failed" {
				utils.LogInfo("Summary Task: Duplicate request detected for IdempotencyKey %s. Returning existing TaskID %s (Status: %s)", idempotencyKey, existingID, status.Status)
				return existingID, nil
			}
			utils.LogInfo("Summary Task: Found existing task %s for key %s but it failed or could not be loaded. Proceeding with new task.", existingID, idempotencyKey)
		}
	}

	taskID := uuid.New().String()
	status := &dtos.RAGStatusDTO{
		Status:         "pending",
		Trigger:        trigger,
		MeetingContext: meetingContext,
		Language:       language,
		MacAddress:     macAddress,
		StartedAt:      time.Now().Format(time.RFC3339),
	}

	u.store.Set(taskID, status)
	if idempHash != "" {
		_ = u.cache.Set(idempHash, taskID)
	}

	// Use a background context as parent if none provided
	if ctx == nil {
		ctx = context.Background()
	}

	go u.runSummaryAsync(ctx, taskID, text, language, meetingContext, style, date, location, participants)

	return taskID, nil
}

func (u *summaryUseCase) runSummaryAsync(ctx context.Context, taskID string, text string, language string, meetingContext string, style string, date string, location string, participants string) {
	status, _ := u.store.Get(taskID)
	if status == nil {
		return
	}

	// Check for early cancellation
	select {
	case <-ctx.Done():
		status.Status = "failed"
		status.Error = "task cancelled: " + ctx.Err().Error()
		u.store.Set(taskID, status)
		return
	default:
	}

	status.Status = "processing"
	u.store.Set(taskID, status)

	res, err := u.summaryInternal(ctx, text, language, meetingContext, style, date, location, participants, status.MacAddress)
	if err != nil {
		utils.LogError("RAG Summary Task %s failed: %v", taskID, err)
		status.Status = "failed"
		status.Error = err.Error()
	} else {
		status.Status = "completed"
		// Propagate ALL structured summary fields for E2E consistency
		status.Summary = res.Summary
		status.PDFUrl = res.PDFUrl
		status.AgendaContext = res.AgendaContext
		status.SummaryVersion = res.SummaryVersion
		status.SummaryMode = res.SummaryMode
		status.ActionItems = res.ActionItems
		status.Decisions = res.Decisions
		status.OpenIssues = res.OpenIssues
		status.Risks = res.Risks
		status.CoverageStats = res.CoverageStats
		status.SpeakerCoverage = res.SpeakerCoverage
		status.SourceLanguage = res.SourceLanguage
		status.TranslatedFromLanguage = res.TranslatedFromLanguage

		if status.MacAddress != "" && u.mqttSvc != nil {
			topic := fmt.Sprintf("users/%s/%s/task", status.MacAddress, u.config.ApplicationEnvironment)
			msg := map[string]interface{}{"event": "stop", "task": "RAG"}
			payload, _ := json.Marshal(msg)
			_ = u.mqttSvc.Publish(topic, 0, false, payload)
		}
	}

	u.store.Set(taskID, status)
	_ = u.cache.SetPreserveTTL(taskID, status)
}

func (u *summaryUseCase) recordPDFDLQ(jobID string, failureReason string) {
	if u.pdfDLQC == nil {
		return
	}
	entry := &entities.PDFDeadLetter{
		JobID:         jobID,
		FailureReason: failureReason,
		RetryCount:    0,
	}
	if err := u.pdfDLQC.Create(entry); err != nil {
		utils.LogError("PDFDLQ: failed to record dead letter entry: %v", err)
	} else {
		utils.LogInfo("PDFDLQ: recorded dead letter entry for job %s", jobID)
	}
}
