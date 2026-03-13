package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	recordingUsecases "sensio/domain/recordings/usecases"
	"sensio/domain/models/whisper/dtos"
	"sensio/domain/models/whisper/usecases"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WhisperTranscribeController handles POST /api/whisper/transcribe.
type WhisperTranscribeController struct {
	transcribeUC    usecases.TranscribeUseCase
	saveRecordingUC recordingUsecases.SaveRecordingUseCase
	uploadSessionUC usecases.UploadSessionUseCase
	config          *utils.Config
	mqttSvc         *infrastructure.MqttService
}

func NewWhisperTranscribeController(
	transcribeUC usecases.TranscribeUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	uploadSessionUC usecases.UploadSessionUseCase,
	cfg *utils.Config,
	mqttSvc *infrastructure.MqttService,
) *WhisperTranscribeController {
	return &WhisperTranscribeController{
		transcribeUC:    transcribeUC,
		saveRecordingUC: saveRecordingUC,
		uploadSessionUC: uploadSessionUC,
		config:          cfg,
		mqttSvc:         mqttSvc,
	}
}

func (c *WhisperTranscribeController) StartMqttSubscription() error {
	if c.mqttSvc == nil {
		return nil
	}

	topic := fmt.Sprintf("users/+/%s/whisper", c.config.ApplicationEnvironment)
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		correlationID := uuid.New().String()
		handlerStart := time.Now()

		utils.LogInfo("[%s] WhisperTranscribe MQTT: Received message on %s, payload size: %d", correlationID, msg.Topic(), len(payload))

		if len(payload) == 0 {
			return
		}

		var req dtos.WhisperMqttRequestDTO
		parseStart := time.Now()
		if err := json.Unmarshal(payload, &req); err != nil {
			parseDuration := time.Since(parseStart)
			utils.LogError("[%s] WhisperTranscribe MQTT: Failed to unmarshal JSON: %v | parse_duration_ms=%d", correlationID, err, parseDuration.Milliseconds())
			return
		}
		parseDuration := time.Since(parseStart)
		utils.LogDebug("[%s] WhisperTranscribe MQTT: Payload parsed | parse_duration_ms=%d", correlationID, parseDuration.Milliseconds())

		// Extract MAC from topic: (optionally $share/group/)users/MAC/env/whisper
		topicParts := strings.Split(msg.Topic(), "/")
		mac := ""
		for i, part := range topicParts {
			if part == "users" && i+1 < len(topicParts) {
				mac = topicParts[i+1]
				break
			}
		}

		if req.Audio == "" || req.TerminalID == "" {
			utils.LogError("[%s] WhisperTranscribe MQTT: Missing audio or terminal_id | total_duration_ms=%d", correlationID, time.Since(handlerStart).Milliseconds())
			if mac != "" {
				c.publishMqttValidationError(mac, "audio/terminal_id", "audio and terminal_id are required")
			}
			return
		}

		// Immediately mark as active to prevent chat handler race condition.
		// It will be deleted either in the defer below (on failure) or by TranscribeAudio async processor.
		utils.ActiveTranscriptions.Store(mac, true)

		// Create a local error flag to determine if we should clean up the transcription flag.
		// If we successfully start TranscribeAudio, it takes ownership of deleting the flag.
		var taskStarted bool
		defer func() {
			if !taskStarted {
				utils.ActiveTranscriptions.Delete(mac)
			}
		}()

		// Decode Base64 audio
		decodeStart := time.Now()
		audioBytes, err := base64.StdEncoding.DecodeString(req.Audio)
		decodeDuration := time.Since(decodeStart)
		if err != nil {
			utils.LogError("[%s] WhisperTranscribe MQTT: Failed to decode base64: %v | decode_duration_ms=%d", correlationID, err, decodeDuration.Milliseconds())
			if mac != "" {
				c.publishMqttValidationError(mac, "audio", "Failed to decode base64 audio")
			}
			return
		}
		utils.LogDebug("[%s] WhisperTranscribe MQTT: Base64 decoded | decode_duration_ms=%d", correlationID, decodeDuration.Milliseconds())

		language := req.Language
		if language == "" {
			language = "id"
		}

		// Generate a descriptive temporary filename
		uuidStr, _ := uuid.NewV7()
		tempFilename := fmt.Sprintf("mqtt_temp_%s_%s.wav", req.TerminalID, uuidStr.String())
		tempPath := filepath.Join("uploads", "audio", tempFilename)

		// Save audio bytes to disk manually (without DB entry)
		fileWriteStart := time.Now()
		if err := os.WriteFile(tempPath, audioBytes, 0644); err != nil {
			fileWriteDuration := time.Since(fileWriteStart)
			utils.LogError("[%s] WhisperTranscribe MQTT: Failed to save temporary audio: %v | file_write_duration_ms=%d", correlationID, err, fileWriteDuration.Milliseconds())
			if mac != "" {
				c.publishMqttError(mac, "Failed to process audio")
			}
			return
		}
		fileWriteDuration := time.Since(fileWriteStart)
		utils.LogDebug("[%s] WhisperTranscribe MQTT: File saved | file_write_duration_ms=%d", correlationID, fileWriteDuration.Milliseconds())

		// Start transcription task using usecase
		transcribeSubmitStart := time.Now()
		taskID, err := c.transcribeUC.TranscribeAudio(context.Background(), tempPath, tempFilename, language, usecases.TranscriptionMetadata{
			UID:         req.UID,
			TerminalID:  req.TerminalID,
			MacAddress:  mac,
			RequestID:   req.RequestID,
			Source:      "mqtt",
			Trigger:     "mqtt:tera/transcribe",
			DeleteAfter: true, // Delete file after transcription
			Diarize:     req.Diarize,
		})
		transcribeSubmitDuration := time.Since(transcribeSubmitStart)
		if err != nil {
			utils.LogError("[%s] WhisperTranscribe MQTT: Failed to start transcription: %v | transcribe_submit_duration_ms=%d", correlationID, err, transcribeSubmitDuration.Milliseconds())
			if mac != "" {
				c.publishMqttError(mac, "Failed to start transcription task: "+err.Error())
			}
			_ = os.Remove(tempPath) // Clean up immediately on error
			return
		}
		utils.LogInfo("[%s] WhisperTranscribe MQTT: Transcription submitted | transcribe_submit_duration_ms=%d", correlationID, transcribeSubmitDuration.Milliseconds())

		taskStarted = true

		// Publish success status with empty RecordingID (since not saved in DB)
		ackPublishStart := time.Now()
		ackPublishSuccess := false
		if mac != "" {
			publishErr := c.publishMqttResponse(mac, commonDtos.StandardResponse{
				Status:  true,
				Message: "Transcription task submitted successfully (Ephemeral)",
				Data: dtos.TranscriptionTaskResponseDTO{
					TaskID:      taskID,
					TaskStatus:  "pending",
					RecordingID: "", // No DB entry
				},
			})
			ackPublishSuccess = (publishErr == nil)
		}
		ackPublishDuration := time.Since(ackPublishStart)

		totalDuration := time.Since(handlerStart)
		utils.LogInfo("[%s] WhisperTranscribe MQTT: Started ephemeral task %s for file %s | parse_duration_ms=%d | decode_duration_ms=%d | file_write_duration_ms=%d | transcribe_submit_duration_ms=%d | ack_publish_duration_ms=%d | ack_publish_success=%v | total_duration_ms=%d", correlationID, taskID, tempFilename, parseDuration.Milliseconds(), decodeDuration.Milliseconds(), fileWriteDuration.Milliseconds(), transcribeSubmitDuration.Milliseconds(), ackPublishDuration.Milliseconds(), ackPublishSuccess, totalDuration.Milliseconds())
	})

	// Subscribe to general task signaling as well
	taskTopic := fmt.Sprintf("users/+/%s/task", c.config.ApplicationEnvironment)
	_ = c.mqttSvc.Subscribe(taskTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		utils.LogInfo("Task Signaling MQTT: Received message on %s: %s", msg.Topic(), string(payload))
	})

	if err != nil {
		utils.LogError("WhisperTranscribe MQTT: Failed to subscribe to %s: %v", topic, err)
		return err
	}
	utils.LogInfo("WhisperTranscribe MQTT: Successfully subscribed to %s", topic)
	return nil
}

func (c *WhisperTranscribeController) publishMqttValidationError(mac, field, message string) {
	c.publishMqttResponse(mac, commonDtos.StandardResponse{
		Status:  false,
		Message: "Validation Error",
		Details: []utils.ValidationErrorDetail{
			{Field: field, Message: message},
		},
	})
}

func (c *WhisperTranscribeController) publishMqttError(mac, details string) {
	utils.LogError("WhisperTranscribe MQTT: %s", details)
	c.publishMqttResponse(mac, commonDtos.StandardResponse{
		Status:  false,
		Message: "Internal Server Error",
	})
}

func (c *WhisperTranscribeController) publishMqttResponse(mac string, resp commonDtos.StandardResponse) error {
	if c.mqttSvc == nil {
		return nil
	}
	// Response topic matches device ACL: users/MAC/env/whisper/answer
	respTopic := fmt.Sprintf("users/%s/%s/whisper/answer", mac, c.config.ApplicationEnvironment)
	respData, _ := json.Marshal(resp)
	if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
		utils.LogError("WhisperTranscribe MQTT: Failed to publish response: %v", err)
		return err
	}
	return nil
}

// Transcribe handles POST /api/models/whisper/transcribe
// @Summary Transcribe audio file (Unified)
// @Description Start transcription of audio file using the configured LLM provider (LLM_PROVIDER). Asynchronous processing. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Models
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Param language formData string false "Language code (e.g. id, en)"
// @Param mac_address formData string false "Device MAC Address"
// @Param diarize formData boolean false "Identify speakers in transcription"
// @Param Idempotency-Key header string false "Idempotency key to deduplicate requests"
// @Success 202 {object} commonDtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 413 {object} commonDtos.StandardResponse
// @Failure 415 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/models/whisper/transcribe [post]
func (c *WhisperTranscribeController) Transcribe(ctx *gin.Context) {
	macAddress := ctx.PostForm("mac_address")
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "audio", Message: "Audio file is required: " + err.Error()},
			},
		})
		return
	}

	if file.Size > c.config.MaxFileSize {
		c.publishMqttError(macAddress, fmt.Sprintf("File too large: %d bytes", file.Size))
		ctx.JSON(http.StatusRequestEntityTooLarge, commonDtos.StandardResponse{
			Status:  false,
			Message: "File too large",
		})
		return
	}

	ext := filepath.Ext(file.Filename)
	supportedExts := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".ogg":  true,
		".flac": true,
	}
	if !supportedExts[ext] {
		ctx.JSON(http.StatusUnsupportedMediaType, commonDtos.StandardResponse{
			Status:  false,
			Message: "Unsupported Media Type",
		})
		return
	}

	language := ctx.PostForm("language")
	if language == "" {
		language = "id"
	}

	diarizeStr := ctx.PostForm("diarize")
	diarize := diarizeStr == "true" || diarizeStr == "1"

	idempotencyKey := ctx.GetHeader("Idempotency-Key")

	// 1. Check Idempotency BEFORE saving recording
	if idempotencyKey != "" {
		f, err := file.Open()
		if err == nil {
			audioHash, _ := utils.HashReader(f)
			f.Close()
			if taskID, exists := c.transcribeUC.CheckIdempotency(idempotencyKey, audioHash, language, macAddress); exists {
				utils.LogInfo("Transcribe.Transcribe: Duplicate request detected (pre-save) for key %s. Returning TaskID %s", idempotencyKey, taskID)
				// We don't have the RecordingID here because it wasn't saved this time,
				// but the client usually only cares about TaskID for polling.
				ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
					Status:  true,
					Message: "Transcription task already submitted",
					Data: dtos.TranscriptionTaskResponseDTO{
						TaskID:     taskID,
						TaskStatus: "pending",
					},
				})
				return
			}
		}
	}

	// 2. Not a duplicate, proceed to save
	baseURL := utils.GetBaseURL(ctx)
	recording, err := c.saveRecordingUC.SaveRecording(file, macAddress, baseURL)
	if err != nil {
		utils.LogError("Transcribe.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	// 3. Content Validation via ffprobe
	probe, err := utils.ProbeAudio(finalInputPath)
	if err != nil {
		utils.LogError("Transcribe.ProbeAudio: %v", err)
		_ = os.Remove(finalInputPath) // Cleanup invalid file
		ctx.JSON(http.StatusUnsupportedMediaType, commonDtos.StandardResponse{
			Status:  false,
			Message: "Unsupported Media Type: actual content check failed",
		})
		return
	}

	if !probe.HasAudio {
		_ = os.Remove(finalInputPath)
		ctx.JSON(http.StatusUnsupportedMediaType, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid Audio: no audio stream found",
		})
		return
	}

	// Extension mismatch check
	realExt := filepath.Ext(recording.Filename)
	// Compare actual format to extension (simple check)
	if !strings.Contains(strings.ToLower(probe.FormatName), strings.TrimPrefix(strings.ToLower(realExt), ".")) {
		utils.LogWarn("Transcribe.Validation: Extension mismatch detected for %s. Extension: %s, Actual Format: %s. Proceeding anyway.", recording.Filename, realExt, probe.FormatName)
	}

	// 4. Start transcription
	taskID, err := c.transcribeUC.TranscribeAudio(ctx.Request.Context(), finalInputPath, file.Filename, language, usecases.TranscriptionMetadata{
		Source:         "rest",
		Trigger:        ctx.Request.URL.Path,
		TerminalID:     macAddress,
		MacAddress:     macAddress,
		Diarize:        diarize,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		utils.LogError("Transcribe.TranscribeAudio: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	resp := commonDtos.StandardResponse{
		Status:  true,
		Message: "Transcription task submitted successfully",
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	}

	// Also publish to MQTT if service is available
	c.publishMqttResponse(macAddress, resp)

	ctx.JSON(http.StatusAccepted, resp)
}

// TranscribeByUpload handles POST /api/models/whisper/transcribe/by-upload
func (c *WhisperTranscribeController) TranscribeByUpload(ctx *gin.Context) {
	var req dtos.SubmitByUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error: " + err.Error(),
		})
		return
	}

	// 1. Finalize session (merge chunks)
	uid, _ := ctx.Get("uid")
	ownerUID := ""
	if uid != nil {
		ownerUID = uid.(string)
	}

	finalized, err := c.uploadSessionUC.FinalizeSession(req.SessionID, ownerUID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "unauthorized session access" {
			statusCode = http.StatusForbidden
		}
		ctx.JSON(statusCode, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to finalize upload: " + err.Error(),
		})
		return
	}

	// 2. Save as Recording (moves file)
	baseURL := utils.GetBaseURL(ctx)
	recording, err := c.saveRecordingUC.SaveRecordingFromPath(finalized.MergedPath, finalized.OriginalFileName, req.MacAddress, baseURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to save recording: " + err.Error(),
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	// 3. Start transcription
	taskID, err := c.transcribeUC.TranscribeAudio(ctx.Request.Context(), finalInputPath, recording.OriginalName, req.Language, usecases.TranscriptionMetadata{
		Source:         "by-upload",
		Trigger:        ctx.Request.URL.Path,
		TerminalID:     req.MacAddress,
		MacAddress:     req.MacAddress,
		Diarize:        req.Diarize,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to start transcription: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
		Status:  true,
		Message: "Transcription task submitted via upload session",
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	})
}
