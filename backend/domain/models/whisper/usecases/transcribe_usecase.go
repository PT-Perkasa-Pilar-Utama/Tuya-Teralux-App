package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sensio/domain/common/providers"
	"sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	ragUsecases "sensio/domain/models/rag/usecases"
	whisperdtos "sensio/domain/models/whisper/dtos"
	"strings"
	"sync"
	"time"
)

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// WhisperClient is the unified interface for all whisper transcription services
type WhisperClient interface {
	Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*whisperdtos.WhisperResult, error)
}

type TranscriptionMetadata struct {
	UID            string
	TerminalID     string
	MacAddress     string // Used for terminal-specific provider resolution
	RequestID      string
	Source         string // "mqtt", "rest", etc.
	Trigger        string // e.g., "/api/whisper/transcribe"
	DeleteAfter    bool   // Whether to delete the audio file after processing
	Diarize        bool   // Whether to perform speaker diarization
	IdempotencyKey string // Client-provided idempotency key
}

type TranscribeOptions struct {
	Language                string
	Diarize                 bool
	Refine                  bool
	IsPipeline              bool
	ProgressCallback        func(int)
	TerminalContext         []string
	DisableFallback         bool
	ForceSegmentSec         int
	ForceSegmentConcurrency int
}

type TranscribeUseCase interface {
	TranscribeAudio(ctx context.Context, inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error)
	TranscribeAudioSync(ctx context.Context, inputPath string, opts TranscribeOptions) (*whisperdtos.AsyncTranscriptionResultDTO, error)
	CheckIdempotency(idempotencyKey string, audioHash string, language string, terminalID string) (string, bool)
}

type transcribeUseCase struct {
	whisperClient       WhisperClient
	refineUC            ragUsecases.RefineUseCase
	store               *tasks.StatusStore[whisperdtos.AsyncTranscriptionStatusDTO]
	cache               *tasks.BadgerTaskCache
	config              *utils.Config
	mqttSvc             mqttPublisher
	providerResolver    providers.ProviderResolver
	audioAnalyzer       utils.AudioAnalyzer
	transcriptValidator utils.TranscriptValidator
}

func NewTranscribeUseCase(
	whisperClient WhisperClient,
	refineUC ragUsecases.RefineUseCase,
	store *tasks.StatusStore[whisperdtos.AsyncTranscriptionStatusDTO],
	cache *tasks.BadgerTaskCache,
	config *utils.Config,
	mqttSvc mqttPublisher,
	providerResolver providers.ProviderResolver,
	audioAnalyzer utils.AudioAnalyzer,
	transcriptValidator utils.TranscriptValidator,
) TranscribeUseCase {
	return &transcribeUseCase{
		whisperClient:       whisperClient,
		refineUC:            refineUC,
		store:               store,
		cache:               cache,
		config:              config,
		mqttSvc:             mqttSvc,
		providerResolver:    providerResolver,
		audioAnalyzer:       audioAnalyzer,
		transcriptValidator: transcriptValidator,
	}
}

// transcribeWithFallback attempts transcription respecting terminal AI preferences first, then gracefully failing over to remote provider candidates
// Returns both the result and the actual provider name used (may differ from resolvedProvider due to fallback)
func (uc *transcribeUseCase) transcribeWithFallback(ctx context.Context, processingPath string, language string, diarize bool, disableFallback bool, isPipeline bool, resolvedProvider string, macAddress string) (*whisperdtos.WhisperResult, string, error) {
	var finalResult *whisperdtos.WhisperResult
	var actualProvider string

	executable := func(resolvedSet *providers.ResolvedProviderSet) error {
		if resolvedSet.WhisperClient == nil {
			return fmt.Errorf("no Whisper client available for provider %s", resolvedSet.ProviderName)
		}
		res, err := resolvedSet.WhisperClient.Transcribe(ctx, processingPath, language, diarize)
		if err == nil {
			finalResult = res
			actualProvider = resolvedSet.ProviderName // Track actual provider used
		}
		return err
	}

	var err error
	if disableFallback && resolvedProvider != "" && resolvedProvider != "unknown" {
		// Direct execution without fallback - resolve the specific provider
		providerSet := uc.providerResolver.ResolveProvider(resolvedProvider)
		if providerSet != nil && providerSet.ProviderName == resolvedProvider && providerSet.WhisperClient != nil {
			finalResult, err = providerSet.WhisperClient.Transcribe(ctx, processingPath, language, diarize)
			if err == nil {
				actualProvider = resolvedProvider
			}
		} else {
			err = fmt.Errorf("provider %s not available", resolvedProvider)
		}
	} else if macAddress != "" {
		err = uc.providerResolver.ExecuteWithFallbackByMac(macAddress, executable)
	} else {
		err = uc.providerResolver.ExecuteWithFallback(executable)
	}

	return finalResult, actualProvider, err
}

func (uc *transcribeUseCase) CheckIdempotency(idempotencyKey string, audioHash string, language string, terminalID string) (string, bool) {
	if idempotencyKey == "" {
		return "", false
	}
	hashInput := fmt.Sprintf("%s_%s_%s_%s", idempotencyKey, language, terminalID, audioHash)
	idempotencyHash := "idemp_transcribe_" + utils.HashString(hashInput)

	var existingTaskID string
	if _, exists, _ := uc.cache.GetWithTTL(idempotencyHash, &existingTaskID); exists && existingTaskID != "" {
		status, ok := uc.store.Get(existingTaskID)
		if !ok || status == nil {
			var cachedStatus whisperdtos.AsyncTranscriptionStatusDTO
			if _, cachedExists, _ := uc.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
				status = &cachedStatus
				ok = true
			}
		}
		if ok && status != nil && status.Status != "failed" {
			return existingTaskID, true
		}
	}
	return "", false
}

func (uc *transcribeUseCase) TranscribeAudio(ctx context.Context, inputPath string, fileName string, language string, metadata ...TranscriptionMetadata) (string, error) {
	if _, err := os.Stat(inputPath); err != nil {
		utils.LogError("Transcribe: Failed to stat audio file: %v", err)
		return "", fmt.Errorf("audio file not found")
	}

	var meta *TranscriptionMetadata
	if len(metadata) > 0 {
		meta = &metadata[0]
	}

	// 1. Idempotency Check
	var idempotencyHash string
	if meta != nil && meta.IdempotencyKey != "" {
		// Create a deterministic hash based on idempotency key, language, terminal ID, and audio content
		audioHash, _ := utils.HashFile(inputPath)
		hashInput := fmt.Sprintf("%s_%s_%s_%s", meta.IdempotencyKey, language, meta.TerminalID, audioHash)
		idempotencyHash = "idemp_transcribe_" + utils.HashString(hashInput)

		// Check if a task already exists for this idempotency key
		var existingTaskID string
		if _, exists, _ := uc.cache.GetWithTTL(idempotencyHash, &existingTaskID); exists && existingTaskID != "" {
			// Check task state - only return if NOT failed
			status, ok := uc.store.Get(existingTaskID)
			if !ok || status == nil {
				// Fallback to cache if memory store is empty
				var cachedStatus whisperdtos.AsyncTranscriptionStatusDTO
				if _, cachedExists, _ := uc.cache.GetWithTTL(existingTaskID, &cachedStatus); cachedExists {
					status = &cachedStatus
					uc.store.Set(existingTaskID, status)
					ok = true
				}
			}

			if ok && status != nil && status.Status != "failed" {
				utils.LogInfo("Transcribe Task: Duplicate request detected for IdempotencyKey %s. Returning existing TaskID %s (Status: %s)", meta.IdempotencyKey, existingTaskID, status.Status)
				return existingTaskID, nil
			}
			utils.LogInfo("Transcribe Task: Found existing task %s for key %s but it failed or could not be loaded. Proceeding with new task.", existingTaskID, meta.IdempotencyKey)
		}
	}

	taskID := utils.GenerateUUID()

	ttl, err := time.ParseDuration(uc.config.TaskStatusTTL)
	if err != nil {
		ttl = 24 * time.Hour
	}

	status := &whisperdtos.AsyncTranscriptionStatusDTO{
		Status:    "pending",
		Trigger:   "",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(ttl).Format(time.RFC3339),
	}

	if meta != nil {
		if meta.Trigger != "" {
			status.Trigger = meta.Trigger
		}
		if meta.MacAddress != "" {
			status.MacAddress = meta.MacAddress
		}
		if meta.TerminalID != "" {
			status.TerminalID = meta.TerminalID
		}
	}

	// 2. Mark as pending and save idempotency key map
	uc.store.Set(taskID, status)
	_ = uc.cache.Set(taskID, status)

	if idempotencyHash != "" {
		// Store the mapping from idempotency hash to task ID
		_ = uc.cache.Set(idempotencyHash, taskID)
	}

	// Mark terminal as actively transcribing if applicable
	if meta != nil && meta.TerminalID != "" && meta.Source == "mqtt" {
		utils.ActiveTranscriptions.Store(meta.TerminalID, true)
	}

	timeout, err := time.ParseDuration(uc.config.TranscribeAsyncTimeout)
	if err != nil {
		timeout = 8 * time.Hour
	}
	asyncCtx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		defer cancel()
		uc.processAsync(asyncCtx, taskID, inputPath, language, meta)
	}()

	return taskID, nil
}

func (uc *transcribeUseCase) TranscribeAudioSync(ctx context.Context, inputPath string, opts TranscribeOptions) (*whisperdtos.AsyncTranscriptionResultDTO, error) {
	var rawTranscription string
	var detectedLang string
	var resultUtterances []whisperdtos.Utterance
	var resultSegments []whisperdtos.TranscriptSegment
	var resultTranscriptFormat whisperdtos.TranscriptFormat
	var resultConfidenceSummary *whisperdtos.ConfidenceSummary

	// Resolve provider from terminal context (MAC address) if available
	resolvedProvider := "unknown"

	// Use terminal context for provider resolution if provided
	if uc.providerResolver != nil && len(opts.TerminalContext) > 0 && opts.TerminalContext[0] != "" {
		macAddress := opts.TerminalContext[0]
		resolved, err := uc.providerResolver.ResolveByMacAddress(macAddress)
		if err != nil {
			utils.LogWarn("TranscribeUseCase: Provider resolution failed for MAC %s: %v, using default", macAddress, err)
		} else {
			resolvedProvider = resolved.ProviderName
			utils.LogInfo("TranscribeUseCase: Using terminal-specific provider '%s' for MAC %s", resolvedProvider, macAddress)
		}
	}

	// CRITICAL: If provider is still "unknown", resolve the default provider to get accurate size limits.
	// This ensures provider-aware segmentation uses the correct limit even without terminal context.
	if resolvedProvider == "unknown" && uc.providerResolver != nil {
		defaultResolved := uc.providerResolver.ResolveDefault()
		if defaultResolved != nil && defaultResolved.ProviderName != "" {
			resolvedProvider = defaultResolved.ProviderName
			utils.LogDebug("TranscribeUseCase: Using default provider '%s' for size limit calculation", resolvedProvider)
		}
	}

	// 1. Audio Normalization (WAV PCM 16k Mono)
	processingPath, audioCleanup, err := utils.NormalizeToWavPCM16k(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize audio: %w", err)
	}
	defer audioCleanup()

	fileInfo, err := os.Stat(processingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat normalized audio: %w", err)
	}

	normalizedSize := fileInfo.Size()

	// ASR Quality Gate: Pre-transcribe audio analysis
	// Analyze audio for silence/low-signal detection before calling provider
	audioClass := "active" // Default to active if analyzer is not available
	providerSkipped := false
	if uc.audioAnalyzer != nil {
		audioResult, analyzeErr := uc.audioAnalyzer.Analyze(processingPath)
		if analyzeErr != nil {
			utils.LogWarn("TranscribeSync: Audio analysis failed: %v, proceeding with transcription", analyzeErr)
		} else {
			audioClass = string(audioResult.Class)
			utils.LogInfo("TranscribeSync: Audio gate analysis | audio_class=%s | duration_sec=%.2f | mean_vol_db=%.1f | max_vol_db=%.1f | silence_pct=%.1f | longest_silence_sec=%.2f",
				audioClass,
				audioResult.Metrics.DurationSec,
				audioResult.Metrics.MeanVolumeDB,
				audioResult.Metrics.MaxVolumeDB,
				audioResult.Metrics.SilencePercentage,
				audioResult.Metrics.LongestSilenceSec)

			// Gate decision: skip provider for silent audio
			if audioResult.Class == utils.AudioClassSilent {
				utils.LogInfo("TranscribeSync: Audio gate REJECT - silent audio detected, skipping provider call")
				providerSkipped = true
			} else if audioResult.Class == utils.AudioClassNearSilent {
				utils.LogInfo("TranscribeSync: Audio gate WARNING - near-silent audio detected, will apply stricter post-gate validation")
			}
		}
	}

	// 2. Provider-aware segmentation decision
	// Segmentation is MANDATORY for oversized files regardless of AUDIO_SEGMENT_ENABLED flag.
	// The flag only controls segmentation for medium-sized files.
	useSegment := uc.shouldSegmentByProvider(normalizedSize, resolvedProvider)
	segmentationIsRequired := normalizedSize > uc.getProviderDirectLimit(resolvedProvider)

	if useSegment {
		// ASR Quality Gate: Skip segmentation entirely for silent audio
		if providerSkipped {
			utils.LogInfo("TranscribeSync: Audio gate REJECT - silent audio detected, skipping segmented transcription")
			rawTranscription = ""
			detectedLang = opts.Language
			resultTranscriptFormat = whisperdtos.TranscriptFormatPlainText
		} else {
			segmentSec := uc.config.AudioSegmentSec
			if opts.ForceSegmentSec > 0 {
				segmentSec = opts.ForceSegmentSec
			} else if opts.IsPipeline && resolvedProvider == "orion" {
				segmentSec = 180
			} else if segmentSec < 60 {
				segmentSec = 60
			}
			overlapSec := uc.config.AudioSegmentOverlapSec

			utils.LogInfo("TranscribeSync: Large audio file detected (%d bytes), starting segmented transcription (step=%ds, overlap=%ds)...", fileInfo.Size(), segmentSec, overlapSec)
			segments, splitErr := utils.SplitAudioSegments(processingPath, segmentSec, overlapSec)
			if splitErr != nil {
				// CRITICAL: If segmentation is mandatory (file exceeds provider limit), DO NOT fallback to full file.
				// This would send an oversized file to the provider and cause failure.
				if segmentationIsRequired {
					utils.LogError("TranscribeSync: Failed to split audio and segmentation is mandatory (size=%d bytes, provider=%s, limit=%d bytes). Aborting transcription.", normalizedSize, resolvedProvider, uc.getProviderDirectLimit(resolvedProvider))
					return nil, fmt.Errorf("failed to split oversized audio for segmented transcription: %w; cannot fallback to full-file as it exceeds provider limit", splitErr)
				}
				// Segmentation was optional (flag-based), safe to fallback
				utils.LogWarn("TranscribeSync: Failed to split audio, falling back to full file (segmentation was optional): %v", splitErr)
				useSegment = false
			} else {
				defer utils.CleanupSegments(segments)

				type segResult struct {
					index int
					text  string
					lang  string
					err   error
				}

				results := make([]segResult, len(segments))
				concurrency := utils.MaxInt(1, uc.config.AudioSegmentMaxConcurrency)
				if opts.ForceSegmentConcurrency > 0 {
					concurrency = opts.ForceSegmentConcurrency
				} else if opts.IsPipeline && resolvedProvider == "orion" {
					concurrency = 1
				}
				sem := make(chan struct{}, concurrency)
				var wg sync.WaitGroup

				var completed int
				var lastProgress int
				var mu sync.Mutex

				for i, seg := range segments {
					wg.Add(1)
					go func(idx int, path string) {
						defer wg.Done()

						// Check context before starting
						select {
						case <-ctx.Done():
							mu.Lock()
							results[idx] = segResult{index: idx, err: ctx.Err()}
							mu.Unlock()
							return
						case sem <- struct{}{}:
							defer func() { <-sem }()
						}

						var res *whisperdtos.WhisperResult
						var transErr error

						var macAddress string
						if len(opts.TerminalContext) > 0 && opts.TerminalContext[0] != "" {
							macAddress = opts.TerminalContext[0]
						}
						// Use health-aware fallback chain for each segment
						res, segmentProvider, transErr := uc.transcribeWithFallback(ctx, path, opts.Language, opts.Diarize, opts.DisableFallback, opts.IsPipeline, resolvedProvider, macAddress)
						if transErr == nil && actualProvider == "" {
							actualProvider = segmentProvider // Track first successful provider
						}

						mu.Lock()
						defer mu.Unlock()
						if transErr != nil {
							results[idx] = segResult{index: idx, err: transErr}
						} else {
							results[idx] = segResult{index: idx, text: res.Transcription, lang: res.DetectedLanguage}
						}

						completed++
						if opts.ProgressCallback != nil {
							progress := int((float64(completed) / float64(len(segments))) * 100)
							if progress > lastProgress {
								lastProgress = progress
								go opts.ProgressCallback(progress)
							}
						}
					}(i, seg.Path)
				}
				wg.Wait()

				// Merge and Dedup - now preserving segment structure
				var mergedText string
				var allSegments []whisperdtos.TranscriptSegment
				var allUtterances []whisperdtos.Utterance

				// Track cumulative offset for segment timing
				var cumulativeOffsetMs int64 = 0

				for i, r := range results {
					if r.err != nil {
						return nil, fmt.Errorf("segment %d failed: %w", r.index, r.err)
					}

					if i == 0 {
						mergedText = r.text
						detectedLang = r.lang
					} else {
						// Use utility merge function that handles overlap detection
						overlapChars := utils.FindOverlapLength(mergedText, r.text)
						if overlapChars > 0 {
							r.text = r.text[overlapChars:]
						}
						mergedText = mergedText + " " + strings.TrimSpace(r.text)
						if detectedLang == "" && r.lang != "" {
							detectedLang = r.lang
						}
					}

			// Merge and Dedup - now preserving segment structure
			var mergedText string
			var allSegments []whisperdtos.TranscriptSegment
			var allUtterances []whisperdtos.Utterance

			// Track cumulative offset for segment timing
			var cumulativeOffsetMs int64 = 0

			for i, r := range results {
				if r.err != nil {
					return nil, fmt.Errorf("segment %d failed: %w", r.index, r.err)
				}

				if i == 0 {
					mergedText = r.text
					detectedLang = r.lang
				} else {
					// Use utility merge function that handles overlap detection
					overlapChars := utils.FindOverlapLength(mergedText, r.text)
					if overlapChars > 0 {
						r.text = r.text[overlapChars:]
					}
					mergedText = mergedText + " " + strings.TrimSpace(r.text)
					if detectedLang == "" && r.lang != "" {
						detectedLang = r.lang
					}
					allSegments = append(allSegments, segment)

					// Parse utterances from this segment ONLY if diarization was requested
					// This prevents false-positive structured output when diarize=false
					if opts.Diarize {
						if segmentUtterances := utils.ParseUtterancesFromText(r.text); len(segmentUtterances) > 0 {
							// Adjust utterance timestamps to global timeline
							for j := range segmentUtterances {
								segmentUtterances[j].StartMs += cumulativeOffsetMs
								segmentUtterances[j].EndMs += cumulativeOffsetMs
							}
							segment.Utterances = segmentUtterances
							allUtterances = append(allUtterances, segmentUtterances...)
						}
					}

					cumulativeOffsetMs += int64(len(r.text) * 100)
				}

				rawTranscription = strings.TrimSpace(mergedText)

				// Store structured results from segmented transcription
				resultSegments = allSegments
				resultUtterances = allUtterances
				if len(allUtterances) > 0 {
					resultTranscriptFormat = whisperdtos.TranscriptFormatUtteranceList
					resultConfidenceSummary = utils.BuildConfidenceSummary(allUtterances, len(allSegments))
				} else {
					resultTranscriptFormat = whisperdtos.TranscriptFormatPlainText
				}

				// Create segment record with timing info (estimated)
				// WARNING: These are HEURISTIC ESTIMATES based on text length (~10 chars/sec).
				// They are NOT audio-aligned timestamps. Do NOT treat as precise evidence.
				segment := whisperdtos.TranscriptSegment{
					Index:   i,
					StartMs: cumulativeOffsetMs,
					EndMs:   cumulativeOffsetMs + int64(len(r.text)*100), // ~10 chars/sec estimate
					Text:    r.text,
				}
				allSegments = append(allSegments, segment)

				// Parse utterances from this segment ONLY if diarization was requested
				// This prevents false-positive structured output when diarize=false
				if opts.Diarize {
					if segmentUtterances := utils.ParseUtterancesFromText(r.text); len(segmentUtterances) > 0 {
						// Adjust utterance timestamps to global timeline
						for j := range segmentUtterances {
							segmentUtterances[j].StartMs += cumulativeOffsetMs
							segmentUtterances[j].EndMs += cumulativeOffsetMs
						}
						segment.Utterances = segmentUtterances
						allUtterances = append(allUtterances, segmentUtterances...)
					}
				}

				cumulativeOffsetMs += int64(len(r.text) * 100)
			}

			rawTranscription = strings.TrimSpace(mergedText)

			// Store structured results from segmented transcription
			resultSegments = allSegments
			resultUtterances = allUtterances
			if len(allUtterances) > 0 {
				resultTranscriptFormat = whisperdtos.TranscriptFormatUtteranceList
				resultConfidenceSummary = utils.BuildConfidenceSummary(allUtterances, len(allSegments))
			} else {
				resultTranscriptFormat = whisperdtos.TranscriptFormatPlainText
			}
		}
	}

	if !useSegment {
		var macAddress string
		if len(opts.TerminalContext) > 0 && opts.TerminalContext[0] != "" {
			macAddress = opts.TerminalContext[0]
		}
		// Use health-aware fallback chain for full-file transcription
		result, err := uc.transcribeWithFallback(ctx, processingPath, opts.Language, opts.Diarize, opts.DisableFallback, opts.IsPipeline, resolvedProvider, macAddress)
		if err != nil {
			return nil, err
		}
		rawTranscription = result.Transcription
		detectedLang = result.DetectedLanguage

		// Store structured artifacts from provider
		resultUtterances = result.Utterances
		resultSegments = result.Segments
		resultTranscriptFormat = result.TranscriptFormat
		resultConfidenceSummary = result.ConfidenceSummary

		if opts.ProgressCallback != nil {
			go opts.ProgressCallback(100)
		}
	}

	// Refine (Grammar/Spelling) - only if explicitly requested
	refined := rawTranscription
	normalizationApplied := false
	if opts.Refine {
		refineLang := detectedLang
		if opts.Language != "" {
			refineLang = opts.Language
		}
		// Pass terminalContext for terminal-specific provider resolution
		if len(opts.TerminalContext) > 0 && opts.TerminalContext[0] != "" {
			refined, _ = uc.refineUC.RefineText(ctx, rawTranscription, refineLang, opts.TerminalContext[0])
		} else {
			refined, _ = uc.refineUC.RefineText(ctx, rawTranscription, refineLang)
		}
		// Note: Current refine does full paraphrasing, so we don't mark as normalization
		// Normalization is reserved for safe punctuation/casing-only fixes
	}

	// Build structured result with backward-compatible fields
	return &whisperdtos.AsyncTranscriptionResultDTO{
		Transcription:        rawTranscription,
		RefinedText:          refined,
		DetectedLanguage:     detectedLang,
		Utterances:           resultUtterances,
		Segments:             resultSegments,
		TranscriptFormat:     resultTranscriptFormat,
		ConfidenceSummary:    resultConfidenceSummary,
		NormalizationApplied: normalizationApplied,
	}, nil
}

func (uc *transcribeUseCase) processAsync(ctx context.Context, taskID string, inputPath string, reqLanguage string, metadata *TranscriptionMetadata) {
	defer func() {
		// Defensive cleanup in case of panic or early exit
		if metadata != nil && metadata.TerminalID != "" && metadata.Source == "mqtt" {
			utils.ActiveTranscriptions.Delete(metadata.TerminalID)
		}
		if r := recover(); r != nil {
			utils.LogError("Transcribe Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", nil, fmt.Errorf("internal panic: %v", r))
		}
	}()

	diarize := false
	refine := false // Default to false to reduce latency
	if metadata != nil {
		diarize = metadata.Diarize
		// If we add Refine to metadata in the future, we should use it here.
	}

	// Pass terminal MAC for terminal-specific provider resolution in refine
	var refineContext string
	if metadata != nil && metadata.MacAddress != "" {
		refineContext = metadata.MacAddress
	}

	opts := TranscribeOptions{
		Language:        reqLanguage,
		Diarize:         diarize,
		Refine:          refine,
		IsPipeline:      false,
		TerminalContext: []string{refineContext},
	}

	finalResult, err := uc.TranscribeAudioSync(ctx, inputPath, opts)
	if err != nil {
		utils.LogError("Transcribe Task %s: Failed: %v", taskID, err)
		uc.updateStatus(taskID, "failed", nil, err)

		if metadata != nil && metadata.DeleteAfter {
			_ = os.Remove(inputPath)
		}
		return
	}

	utils.LogInfo("Transcribe Task %s: Finished successfully", taskID)

	if metadata != nil && metadata.DeleteAfter {
		_ = os.Remove(inputPath)
	}

	uc.updateStatus(taskID, "completed", finalResult, nil)

	// Explicitly clear the transcription flag BEFORE chaining to /chat
	// to prevent a race condition where the chat controller drops the request.
	if metadata != nil && metadata.TerminalID != "" && metadata.Source == "mqtt" {
		utils.ActiveTranscriptions.Delete(metadata.TerminalID)
	}

	// ASR Quality Gate: Stop empty/rejected transcripts from entering chat chain
	// This prevents hallucinated or silent audio from triggering chat responses
	if metadata != nil && metadata.Source == "mqtt" && metadata.MacAddress != "" && uc.mqttSvc != nil {
		// Get the final transcript (prefer refined, fallback to raw)
		finalTranscript := finalResult.RefinedText
		if finalTranscript == "" {
			finalTranscript = finalResult.Transcription
		}

		// Check if transcript was rejected or empty
		if finalTranscript == "" || !finalResult.TranscriptValid {
			utils.LogInfo("Transcribe Task %s: ASR gate BLOCKED chat chaining | transcript_valid=%v | rejection_reason=%s | audio_class=%s",
				taskID, finalResult.TranscriptValid, finalResult.TranscriptRejectionReason, finalResult.AudioClass)

			// NEW: Publish terminal completion via MQTT for rejected results
			// This ensures frontend receives explicit terminal state without waiting for HTTP fallback
			answerTopic := fmt.Sprintf("users/%s/%s/chat/answer", metadata.MacAddress, uc.config.ApplicationEnvironment)

			// Build rejection payload shaped like assistant final responses
			answerPayload := map[string]interface{}{
				"request_id": metadata.RequestID,
				"response":   nil,                // No assistant text for rejection
				"is_blocked": true,               // Signal terminal completion
				"source":     "WHISPER_REJECTED", // Canonical source for rejection
				"data": map[string]interface{}{
					"rejection_reason": finalResult.TranscriptRejectionReason,
					"audio_class":      finalResult.AudioClass,
					"provider_skipped": finalResult.ProviderSkipped,
				},
			}

			payload, marshalErr := json.Marshal(answerPayload)
			if marshalErr != nil {
				utils.LogError("Transcribe Task %s: Failed to marshal rejection payload: %v", taskID, marshalErr)
			} else if err := uc.mqttSvc.Publish(answerTopic, 0, false, payload); err != nil {
				utils.LogError("Transcribe Task %s: Failed to publish rejection to MQTT: %v", taskID, err)
			} else {
				utils.LogInfo("Transcribe Task %s: Published rejection completion to %s", taskID, answerTopic)
			}
			// Do NOT chain to /chat
		} else {
			// Chain to /chat with validated transcript
			chatTopic := fmt.Sprintf("users/%s/%s/chat", metadata.MacAddress, uc.config.ApplicationEnvironment)

			chatReq := map[string]string{
				"prompt":      finalTranscript,
				"terminal_id": metadata.TerminalID,
				"language":    finalResult.DetectedLanguage,
				"uid":         metadata.UID,
				"request_id":  metadata.RequestID,
			}
			payload, _ := json.Marshal(chatReq)
			if err := uc.mqttSvc.Publish(chatTopic, 0, false, payload); err != nil {
				utils.LogError("TranscribeUseCase: Failed to publish transcript to MQTT: %v", err)
			}
			utils.LogInfo("Transcribe Task %s: Chained result to %s", taskID, chatTopic)
		}
	}
}

func (uc *transcribeUseCase) updateStatus(taskID string, statusStr string, result *whisperdtos.AsyncTranscriptionResultDTO, err error) {
	// Try to get existing status to preserve StartedAt, TerminalID, and MacAddress
	var existing whisperdtos.AsyncTranscriptionStatusDTO
	_, _, _ = uc.cache.GetWithTTL(taskID, &existing)

	ttl, pErr := time.ParseDuration(uc.config.TaskStatusTTL)
	if pErr != nil {
		ttl = 24 * time.Hour
	}

	status := &whisperdtos.AsyncTranscriptionStatusDTO{
		Status:     statusStr,
		Result:     result,
		StartedAt:  existing.StartedAt,
		Trigger:    existing.Trigger,
		MacAddress: existing.MacAddress,
		TerminalID: existing.TerminalID,
		ExpiresAt:  time.Now().Add(ttl).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	} else if statusStr == "completed" {
		status.HTTPStatusCode = 200
	}

	// Calculate duration if finished
	if statusStr == "completed" || statusStr == "failed" {
		var duration float64
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			duration = time.Since(startTime).Seconds()
			status.DurationSeconds = duration
		}

		logMsg := fmt.Sprintf("Task %s: %s (Duration: %.2fs)", taskID, statusStr, duration)
		if status.TerminalID != "" {
			logMsg += fmt.Sprintf(", Terminal: %s", status.TerminalID)
		}
		if status.Trigger != "" {
			logMsg += fmt.Sprintf(", Trigger: %s", status.Trigger)
		}
		if err != nil {
			utils.LogError("%s, Error: %v", logMsg, err)
		} else {
			utils.LogInfo("%s", logMsg)
		}
		if status.MacAddress != "" && uc.mqttSvc != nil {
			taskTopic := fmt.Sprintf("users/%s/%s/task", status.MacAddress, uc.config.ApplicationEnvironment)
			msg := map[string]string{
				"event": "stop",
				"task":  "Transcribe",
			}
			payload, _ := json.Marshal(msg)
			if err := uc.mqttSvc.Publish(taskTopic, 0, false, payload); err != nil {
				utils.LogError("Transcribe Task %s: Failed to publish stop signal to MQTT: %v", taskID, err)
			} else {
				utils.LogInfo("Transcribe Task %s: Published stop signal to %s", taskID, taskTopic)
			}
		}
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.SetWithTTL(taskID, status, ttl)
}

// shouldSegmentByProvider determines if segmented transcription should be used based on:
// 1. Normalized file size
// 2. Resolved provider's direct upload limit
// 3. Provider-aware safety policy (mandatory for oversized files)
//
// Segmentation is MANDATORY when file size exceeds the provider's limit,
// regardless of the AUDIO_SEGMENT_ENABLED flag.
func (uc *transcribeUseCase) shouldSegmentByProvider(normalizedSize int64, provider string) bool {
	// Get provider-specific direct upload limit
	providerLimit := uc.getProviderDirectLimit(provider)

	// Log provider-aware size policy decision
	utils.LogInfo("TranscribeUseCase: Provider-aware size check | provider=%s | normalized_size=%d bytes | provider_limit=%d bytes", provider, normalizedSize, providerLimit)

	// Mandatory segmentation for oversized files
	if normalizedSize > providerLimit {
		utils.LogInfo("TranscribeUseCase: FORCING segmented transcription | provider=%s | size=%d bytes exceeds limit=%d bytes", provider, normalizedSize, providerLimit)
		return true
	}

	// For files within provider limit, respect the AUDIO_SEGMENT_ENABLED flag
	// for additional tuning (e.g., segment medium-sized files for better accuracy)
	const segmentThreshold = 20 * 1024 * 1024 // 20MB default threshold
	if normalizedSize > segmentThreshold && uc.config.AudioSegmentEnabled {
		utils.LogInfo("TranscribeUseCase: Using segmented transcription per AUDIO_SEGMENT_ENABLED flag | size=%d bytes", normalizedSize)
		return true
	}

	return false
}

// getProviderDirectLimit returns the direct upload limit in bytes for a given provider
func (uc *transcribeUseCase) getProviderDirectLimit(provider string) int64 {
	switch provider {
	case "gemini":
		return services.GeminiDirectUploadLimitBytes
	case "openai":
		return services.OpenAIDirectUploadLimitBytes
	case "groq":
		return services.GroqDirectUploadLimitBytes
	case "orion":
		return services.OrionDirectUploadLimitBytes
	default:
		// For local or unknown providers, use conservative 20MB default.
		// This is safer than allowing potentially oversized uploads.
		if provider != "" && provider != "local" {
			utils.LogWarn("TranscribeUseCase: Unknown provider '%s' for size limit check, using conservative 20MB default", provider)
		}
		return 20 * 1024 * 1024
	}
}

// mergeWithDedup handles the joining of two transcript segments by checking for shared word overlaps
// at the boundaries, which often occurs due to the segment overlap duration.
func (uc *transcribeUseCase) mergeWithDedup(prev, current string) string {
	prevWords := strings.Fields(prev)
	currentWords := strings.Fields(current)

	if len(prevWords) == 0 {
		return current
	}
	if len(currentWords) == 0 {
		return prev
	}

	// Try to find the longest overlap (max 10 words)
	maxOverlap := 10
	if len(prevWords) < maxOverlap {
		maxOverlap = len(prevWords)
	}
	if len(currentWords) < maxOverlap {
		maxOverlap = len(currentWords)
	}

	bestOverlap := 0
	for size := 1; size <= maxOverlap; size++ {
		match := true
		for i := 0; i < size; i++ {
			pWord := prevWords[len(prevWords)-size+i]
			cWord := currentWords[i]
			// Case-insensitive comparison for deduplication
			if !strings.EqualFold(pWord, cWord) {
				match = false
				break
			}
		}
		if match {
			bestOverlap = size
		}
	}

	if bestOverlap > 0 {
		remaining := currentWords[bestOverlap:]
		if len(remaining) == 0 {
			return prev
		}
		return prev + " " + strings.Join(remaining, " ")
	}

	return prev + " " + current
}
