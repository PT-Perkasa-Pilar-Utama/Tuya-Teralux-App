package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func (c *SpeechTranscribeController) StartMqttSubscription() {
	if c.mqttSvc == nil {
		return
	}

	topic := "users/+/whisper"
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		if len(payload) == 0 {
			return
		}

		var req dtos.WhisperMqttRequestDTO
		if err := json.Unmarshal(payload, &req); err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to unmarshal JSON: %v", err)
			return
		}

		// Extract MAC from topic: users/MAC/whisper
		topicParts := strings.Split(msg.Topic(), "/")
		mac := ""
		if len(topicParts) >= 2 && topicParts[0] == "users" {
			mac = topicParts[1]
		}

		if req.Audio == "" || req.TerminalID == "" {
			utils.LogError("SpeechTranscribe MQTT: Missing audio or terminal_id")
			if mac != "" {
				c.publishMqttValidationError(mac, "audio/terminal_id", "audio and terminal_id are required")
			}
			return
		}

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
		taskID, err := c.transcribeUC.TranscribeAudio(tempPath, tempFilename, language, usecases.TranscriptionMetadata{
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

		// Publish success status with empty RecordingID (since not saved in DB)
		if mac != "" {
			c.publishMqttResponse(mac, dtos.StandardResponse{
				Status:  true,
				Message: "Transcription task submitted successfully (Ephemeral)",
				Data: dtos.TranscriptionTaskResponseDTO{
					TaskID:      taskID,
					TaskStatus:  "pending",
					RecordingID: "", // No DB entry
				},
			})
		}

		utils.LogInfo("SpeechTranscribe MQTT: Started ephemeral task %s for file %s", taskID, tempFilename)
	})

	if err != nil {
		utils.LogError("SpeechTranscribe MQTT: Failed to subscribe to %s: %v", topic, err)
	}
}

func (c *SpeechTranscribeController) publishMqttValidationError(mac, field, message string) {
	c.publishMqttResponse(mac, dtos.StandardResponse{
		Status:  false,
		Message: "Validation Error",
		Details: []utils.ValidationErrorDetail{
			{Field: field, Message: message},
		},
	})
}

func (c *SpeechTranscribeController) publishMqttError(mac, details string) {
	utils.LogError("SpeechTranscribe MQTT: %s", details)
	c.publishMqttResponse(mac, dtos.StandardResponse{
		Status:  false,
		Message: "Internal Server Error",
	})
}

func (c *SpeechTranscribeController) publishMqttResponse(mac string, resp dtos.StandardResponse) {
	if c.mqttSvc == nil {
		return
	}
	// Response topic matches device ACL: users/MAC/whisper/answer
	respTopic := fmt.Sprintf("users/%s/whisper/answer", mac)
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
// @Success 202 {object} dtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe [post]
func (c *SpeechTranscribeController) Transcribe(ctx *gin.Context) {
	macAddress := ctx.PostForm("mac_address")
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
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
		ctx.JSON(http.StatusRequestEntityTooLarge, dtos.StandardResponse{
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
		ctx.JSON(http.StatusUnsupportedMediaType, dtos.StandardResponse{
			Status:  false,
			Message: "Unsupported Media Type",
		})
		return
	}

	macAddress = ctx.PostForm("mac_address")
	baseURL := utils.GetBaseURL(ctx)
	recording, err := c.saveRecordingUC.SaveRecording(file, macAddress, baseURL)
	if err != nil {
		utils.LogError("Transcribe.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	language := ctx.PostForm("language")
	if language == "" {
		language = "id"
	}

	diarizeStr := ctx.PostForm("diarize")
	diarize := diarizeStr == "true" || diarizeStr == "1"

	// Use the same TranscribeAudio with REST metadata
	taskID, err := c.transcribeUC.TranscribeAudio(finalInputPath, file.Filename, language, usecases.TranscriptionMetadata{
		Source:  "rest",
		Trigger: ctx.Request.URL.Path,
		Diarize: diarize,
	})
	if err != nil {
		utils.LogError("Transcribe.TranscribeAudio: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	resp := dtos.StandardResponse{
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
