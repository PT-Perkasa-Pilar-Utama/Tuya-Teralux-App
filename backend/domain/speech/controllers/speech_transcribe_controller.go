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

	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	recordingUsecases "sensio/domain/recordings/usecases"
	"sensio/domain/speech/dtos"
	"sensio/domain/speech/usecases"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SpeechTranscribeController handles POST /api/speech/transcribe.
type SpeechTranscribeController struct {
	transcribeUC    usecases.TranscribeUseCase
	saveRecordingUC recordingUsecases.SaveRecordingUseCase
	config          *utils.Config
	mqttSvc         *infrastructure.MqttService
}

func NewSpeechTranscribeController(
	transcribeUC usecases.TranscribeUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
	mqttSvc *infrastructure.MqttService,
) *SpeechTranscribeController {
	return &SpeechTranscribeController{
		transcribeUC:    transcribeUC,
		saveRecordingUC: saveRecordingUC,
		config:          cfg,
		mqttSvc:         mqttSvc,
	}
}

func (c *SpeechTranscribeController) StartMqttSubscription() error {
	if c.mqttSvc == nil {
		return nil
	}

	topic := fmt.Sprintf("users/+/%s/whisper", c.config.ApplicationEnvironment)
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		correlationID := uuid.New().String()

		utils.LogInfo("[%s] SpeechTranscribe MQTT: Received message on %s, payload size: %d", correlationID, msg.Topic(), len(payload))

		if len(payload) == 0 {
			return
		}

		var req dtos.WhisperMqttRequestDTO
		if err := json.Unmarshal(payload, &req); err != nil {
			utils.LogError("[%s] SpeechTranscribe MQTT: Failed to unmarshal JSON: %v", correlationID, err)
			return
		}

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
			utils.LogError("SpeechTranscribe MQTT: Missing audio or terminal_id")
			if mac != "" {
				c.publishMqttValidationError(mac, "audio/terminal_id", "audio and terminal_id are required")
			}
			return
		}

		// Immediately mark as active to prevent chat handler race condition.
		// It will be deleted either in the defer below (on failure) or by TranscribeAudio async processor.
		utils.ActiveTranscriptions.Store(req.TerminalID, true)

		// Create a local error flag to determine if we should clean up the transcription flag.
		// If we successfully start TranscribeAudio, it takes ownership of deleting the flag.
		var taskStarted bool
		defer func() {
			if !taskStarted {
				utils.ActiveTranscriptions.Delete(req.TerminalID)
			}
		}()

		// Decode Base64 audio
		audioBytes, err := base64.StdEncoding.DecodeString(req.Audio)
		if err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to decode base64: %v", err)
			if mac != "" {
				c.publishMqttValidationError(mac, "audio", "Failed to decode base64 audio")
			}
			return
		}

		language := req.Language
		if language == "" {
			language = "id"
		}

		// Generate a descriptive temporary filename
		uuidStr, _ := uuid.NewV7()
		tempFilename := fmt.Sprintf("mqtt_temp_%s_%s.wav", req.TerminalID, uuidStr.String())
		tempPath := filepath.Join("uploads", "audio", tempFilename)

		// Save audio bytes to disk manually (without DB entry)
		if err := os.WriteFile(tempPath, audioBytes, 0644); err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to save temporary audio: %v", err)
			if mac != "" {
				c.publishMqttError(mac, "Failed to process audio")
			}
			return
		}

		// Start transcription task using usecase
		taskID, err := c.transcribeUC.TranscribeAudio(context.Background(), tempPath, tempFilename, language, usecases.TranscriptionMetadata{
			UID:         req.UID,
			TerminalID:  mac,
			Source:      "mqtt",
			Trigger:     "mqtt:tera/transcribe",
			DeleteAfter: true, // Delete file after transcription
			Diarize:     req.Diarize,
		})
		if err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to start transcription: %v", err)
			if mac != "" {
				c.publishMqttError(mac, "Failed to start transcription task: "+err.Error())
			}
			_ = os.Remove(tempPath) // Clean up immediately on error
			return
		}

		taskStarted = true

		// Publish success status with empty RecordingID (since not saved in DB)
		if mac != "" {
			c.publishMqttResponse(mac, commonDtos.StandardResponse{
				Status:  true,
				Message: "Transcription task submitted successfully (Ephemeral)",
				Data: dtos.TranscriptionTaskResponseDTO{
					TaskID:      taskID,
					TaskStatus:  "pending",
					RecordingID: "", // No DB entry
				},
			})
		}

		utils.LogInfo("[%s] SpeechTranscribe MQTT: Started ephemeral task %s for file %s", correlationID, taskID, tempFilename)
	})

	// Subscribe to general task signaling as well
	taskTopic := fmt.Sprintf("users/+/%s/task", c.config.ApplicationEnvironment)
	_ = c.mqttSvc.Subscribe(taskTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		utils.LogInfo("Task Signaling MQTT: Received message on %s: %s", msg.Topic(), string(payload))
	})

	if err != nil {
		utils.LogError("SpeechTranscribe MQTT: Failed to subscribe to %s: %v", topic, err)
		return err
	}
	utils.LogInfo("SpeechTranscribe MQTT: Successfully subscribed to %s", topic)
	return nil
}

func (c *SpeechTranscribeController) publishMqttValidationError(mac, field, message string) {
	c.publishMqttResponse(mac, commonDtos.StandardResponse{
		Status:  false,
		Message: "Validation Error",
		Details: []utils.ValidationErrorDetail{
			{Field: field, Message: message},
		},
	})
}

func (c *SpeechTranscribeController) publishMqttError(mac, details string) {
	utils.LogError("SpeechTranscribe MQTT: %s", details)
	c.publishMqttResponse(mac, commonDtos.StandardResponse{
		Status:  false,
		Message: "Internal Server Error",
	})
}

func (c *SpeechTranscribeController) publishMqttResponse(mac string, resp commonDtos.StandardResponse) {
	if c.mqttSvc == nil {
		return
	}
	// Response topic matches device ACL: users/MAC/env/whisper/answer
	respTopic := fmt.Sprintf("users/%s/%s/whisper/answer", mac, c.config.ApplicationEnvironment)
	respData, _ := json.Marshal(resp)
	if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
		utils.LogError("SpeechTranscribe MQTT: Failed to publish response: %v", err)
	}
}

// Transcribe handles POST /api/speech/transcribe
// @Summary Transcribe audio file (Unified)
// @Description Start transcription of audio file using the configured LLM provider (LLM_PROVIDER). Asynchronous processing. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Speech
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
// @Router /api/speech/transcribe [post]
func (c *SpeechTranscribeController) Transcribe(ctx *gin.Context) {
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

	// Use the same TranscribeAudio with REST metadata
	taskID, err := c.transcribeUC.TranscribeAudio(ctx.Request.Context(), finalInputPath, file.Filename, language, usecases.TranscriptionMetadata{
		Source:         "rest",
		Trigger:        ctx.Request.URL.Path,
		TerminalID:     macAddress,
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
