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
	"sensio/domain/models/rag/services"
	"sensio/domain/models/rag/skills"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SummaryUseCase interface {
	SummarizeText(text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error)
	SummarizeTextWithTrigger(text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error)
	SummarizeTextWithContext(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, args ...string) (string, error)
	SummarizeTextWithContextAndTrigger(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, trigger string, args ...string) (string, error)
	SummarizeTextSync(ctx context.Context, text string, language string, meetingContext string, style string, date string, location string, participants string, macAddress string) (*dtos.RAGSummaryResponseDTO, error)
}

type summaryUseCase struct {
	llm           skills.LLMClient
	fallbackLLM   skills.LLMClient
	config        *utils.Config
	cache         *tasks.BadgerTaskCache
	store         *tasks.StatusStore[dtos.RAGStatusDTO]
	renderer      services.SummaryPDFRenderer
	bigExternal   *commonServices.BigExternalService
	mqttSvc       mqttPublisher
	llmTimeout    time.Duration
	renderTimeout time.Duration
	skill         skills.Skill
	chunkSkill    skills.Skill
}

func NewSummaryUseCase(
	llm skills.LLMClient,
	fallbackLLM skills.LLMClient,
	cfg *utils.Config,
	cache *tasks.BadgerTaskCache,
	store *tasks.StatusStore[dtos.RAGStatusDTO],
	renderer services.SummaryPDFRenderer,
	bigExternal *commonServices.BigExternalService,
	mqttSvc mqttPublisher,
	skill skills.Skill,
	chunkSkill skills.Skill,
) SummaryUseCase {
	return &summaryUseCase{
		llm:           llm,
		fallbackLLM:   fallbackLLM,
		config:        cfg,
		cache:         cache,
		store:         store,
		renderer:      renderer,
		bigExternal:   bigExternal,
		mqttSvc:       mqttSvc,
		llmTimeout:    5 * time.Minute,
		renderTimeout: 30 * time.Second,
		skill:         skill,
		chunkSkill:    chunkSkill,
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
			}
		}
	}

	if u.skill == nil {
		return nil, fmt.Errorf("summary skill not configured")
	}

	skillCtx := &skills.SkillContext{
		Ctx:          ctx,
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

	var trimmedSummary string
	var inferredAgenda string
	var res *skills.SkillResult
	var err error

	// Phase 3: Chunked Summarization for long transcripts
	if len(text) > 4000*4 { // Heuristic: ~4000 tokens
		utils.LogInfo("SummaryUseCase: Text too long (%d chars), using chunked summarization", len(text))
		chunkedSummary, chunkErr := u.summarizeInChunks(ctx, text, language, meetingContext)
		if chunkErr == nil {
			skillCtx.Prompt = chunkedSummary
			res, err = u.skill.Execute(skillCtx)
			if err == nil {
				trimmedSummary = res.Message
			}
		} else {
			utils.LogError("SummaryUseCase: Chunked summarization failed: %v. Falling back to single-pass.", chunkErr)
		}
	}

	// Single-pass execution
	if trimmedSummary == "" {
		res, err = u.skill.Execute(skillCtx)
		if err != nil && u.fallbackLLM != nil {
			utils.LogWarn("SummaryTask: Primary LLM failed, falling back to local: %v", err)
			skillCtx.LLM = u.fallbackLLM
			res, err = u.skill.Execute(skillCtx)
		}
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
	pdfFilename := fmt.Sprintf("summary_%s.pdf", uuidStr.String())
	basePath := "."
	if envPath := utils.FindEnvFile(); envPath != "" {
		basePath = filepath.Dir(envPath)
	}
	pdfPath := filepath.Join(basePath, "uploads", "reports", pdfFilename)
	_ = os.MkdirAll(filepath.Dir(pdfPath), 0755)

	if u.renderer != nil {
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
			return nil, fmt.Errorf("pdf generation failed: %w", err)
		}
	}

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)

	// Cache inferred agenda
	if macAddress != "" && inferredAgenda != "" {
		_ = u.cache.Set(fmt.Sprintf("agenda_mac_%s", macAddress), inferredAgenda)
	}

	return &dtos.RAGSummaryResponseDTO{
		Summary:       trimmedSummary,
		PDFUrl:        pdfUrl,
		AgendaContext: meetingContext,
	}, nil
}

func (u *summaryUseCase) summarizeInChunks(ctx context.Context, text string, language string, meetingContext string) (string, error) {
	if u.chunkSkill == nil {
		return "", fmt.Errorf("chunk summary skill not configured")
	}

	chunks := u.splitText(text, 16000) // ~4000 tokens
	var intermediateSummaries []string
	for idx, chunk := range chunks {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		utils.LogInfo("summarizeInChunks: Processing chunk %d/%d", idx+1, len(chunks))
		sCtx := &skills.SkillContext{
			Ctx:      ctx,
			Prompt:   chunk,
			Language: language,
			LLM:      u.llm,
			Config:   u.config,
			Context:  meetingContext,
		}
		res, err := u.chunkSkill.Execute(sCtx)
		if err != nil {
			utils.LogError("summarizeInChunks: Chunk %d failed: %v", idx+1, err)
			continue
		}
		intermediateSummaries = append(intermediateSummaries, res.Message)
	}

	if len(intermediateSummaries) == 0 {
		return "", fmt.Errorf("all chunk summaries failed")
	}

	return strings.Join(intermediateSummaries, "\n\n---\n\n"), nil
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
		status.Summary = res.Summary
		status.PDFUrl = res.PDFUrl
		status.AgendaContext = res.AgendaContext

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
