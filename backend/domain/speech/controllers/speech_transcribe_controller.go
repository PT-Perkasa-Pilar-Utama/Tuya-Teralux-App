package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
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

	topic := "users/teralux/whisper"
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		if len(payload) == 0 {
			return
		}

		var req dtos.WhisperMqttRequestDTO
		if err := json.Unmarshal(payload, &req); err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to unmarshal JSON: %v", err)
			c.publishMqttValidationError("payload", "Invalid JSON payload: "+err.Error())
			return
		}

		if req.Audio == "" || req.TeraluxID == "" {
			utils.LogError("SpeechTranscribe MQTT: Missing audio or teralux_id")
			c.publishMqttValidationError("audio/teralux_id", "audio and teralux_id are required")
			return
		}

		// Decode Base64 audio
		audioBytes, err := base64.StdEncoding.DecodeString(req.Audio)
		if err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to decode base64: %v", err)
			c.publishMqttValidationError("audio", "Failed to decode base64 audio")
			return
		}

		language := req.Language
		if language == "" {
			language = "id"
		}

		// Generate a descriptive original name for logging
		filename := fmt.Sprintf("mqtt_%s_%d.wav", req.TeraluxID, time.Now().UnixNano())

		// Save recording (File + DB) - Same as REST
		recording, err := c.saveRecordingUC.SaveRecordingFromBytes(audioBytes, filename)
		if err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to save recording: %v", err)
			c.publishMqttError("Failed to save recording: " + err.Error())
			return
		}

		finalPath := filepath.Join("uploads", "audio", recording.Filename)

		// Start transcription task using usecase - Same metadata logic
		taskID, err := c.transcribeUC.TranscribeAudio(finalPath, filename, language, usecases.TranscriptionMetadata{
			UID:       req.UID,
			TeraluxID: req.TeraluxID,
			Source:    "mqtt",
			Trigger:   "mqtt:tera/transcribe", // Assuming this is the topic or some descriptive name
		})
		if err != nil {
			utils.LogError("SpeechTranscribe MQTT: Failed to start transcription: %v", err)
			c.publishMqttError("Failed to start transcription task: " + err.Error())
			return
		}

		// Publish success status with RecordingID - Same as REST
		c.publishMqttResponse(dtos.StandardResponse{
			Status:  true,
			Message: "Transcription task submitted successfully",
			Data: dtos.TranscriptionTaskResponseDTO{
				TaskID:      taskID,
				TaskStatus:  "pending",
				RecordingID: recording.ID,
			},
		})

		utils.LogInfo("SpeechTranscribe MQTT: Started task %s for file %s", taskID, filename)
	})

	if err != nil {
		utils.LogError("SpeechTranscribe MQTT: Failed to subscribe to %s: %v", topic, err)
	}
}

func (c *SpeechTranscribeController) publishMqttValidationError(field, message string) {
	c.publishMqttResponse(dtos.StandardResponse{
		Status:  false,
		Message: "Validation Error",
		Details: []utils.ValidationErrorDetail{
			{Field: field, Message: message},
		},
	})
}

func (c *SpeechTranscribeController) publishMqttError(details string) {
	utils.LogError("SpeechTranscribe MQTT: %s", details)
	c.publishMqttResponse(dtos.StandardResponse{
		Status:  false,
		Message: "Internal Server Error",
	})
}

func (c *SpeechTranscribeController) publishMqttResponse(resp dtos.StandardResponse) {
	if c.mqttSvc == nil {
		return
	}
	respTopic := "users/teralux/whisper/answer"
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
// @Success 202 {object} dtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe [post]
func (c *SpeechTranscribeController) Transcribe(ctx *gin.Context) {
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
		c.publishMqttError(fmt.Sprintf("File too large: %d bytes", file.Size))
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

	recording, err := c.saveRecordingUC.SaveRecording(file)
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

	// Use the same TranscribeAudio with REST metadata
	taskID, err := c.transcribeUC.TranscribeAudio(finalInputPath, file.Filename, language, usecases.TranscriptionMetadata{
		Source:  "rest",
		Trigger: ctx.Request.URL.Path,
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
	c.publishMqttResponse(resp)

	ctx.JSON(http.StatusAccepted, resp)
}
