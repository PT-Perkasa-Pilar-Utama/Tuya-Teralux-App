package controllers

import (
	"net/http"
	"path/filepath"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineDtos "sensio/domain/models/pipeline/dtos"
	pipelineUsecases "sensio/domain/models/pipeline/usecases"
	speechdtos "sensio/domain/models/whisper/dtos"
	speechUsecases "sensio/domain/models/whisper/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PipelineController struct {
	pipelineUC      pipelineUsecases.PipelineUseCase
	statusUC        tasks.GenericStatusUseCase[pipelineDtos.PipelineStatusDTO]
	saveRecordingUC recordingUsecases.SaveRecordingUseCase
	uploadSessionUC speechUsecases.UploadSessionUseCase
	config          *utils.Config
}

func NewPipelineController(
	pipelineUC pipelineUsecases.PipelineUseCase,
	statusUC tasks.GenericStatusUseCase[pipelineDtos.PipelineStatusDTO],
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	uploadSessionUC speechUsecases.UploadSessionUseCase,
	cfg *utils.Config,
) *PipelineController {
	return &PipelineController{
		pipelineUC:      pipelineUC,
		statusUC:        statusUC,
		saveRecordingUC: saveRecordingUC,
		uploadSessionUC: uploadSessionUC,
		config:          cfg,
	}
}

// ExecuteJob handles POST /api/models/pipeline/job
// @Summary Run unified AI pipeline (Transcribe -> Refine -> Translate -> Summarize)
// @Description Submits a single job that orchestrates multiple AI stages.
// @Tags 04. Models
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file"
// @Param language formData string false "Source language (e.g. id, en)"
// @Param target_language formData string false "Target language for translation/summary"
// @Param summarize formData boolean false "Whether to generate a summary"
// @Param refine formData boolean false "Whether to refine text (grammar/spelling)"
// @Param diarize formData boolean false "Whether to diarize speakers"
// @Param context formData string false "Meeting context (for summary)"
// @Param style formData string false "Summary style"
// @Param date formData string false "Meeting date"
// @Param location formData string false "Meeting location"
// @Param participants formData string false "Comma-separated participants"
// @Param mac_address formData string false "Device MAC Address"
// @Param Idempotency-Key header string false "Idempotency key"
// @Success 202 {object} commonDtos.StandardResponse{data=pipelineDtos.PipelineResponseDTO}
// @Router /api/models/pipeline/job [post]
func (c *PipelineController) ExecuteJob(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "audio file is required",
		})
		return
	}

	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, commonDtos.StandardResponse{
			Status:  false,
			Message: "file too large",
		})
		return
	}

	macAddress := ctx.PostForm("mac_address")
	baseURL := utils.GetBaseURL(ctx)

	// Parse parameters
	summarize, _ := strconv.ParseBool(ctx.PostForm("summarize"))
	diarize, _ := strconv.ParseBool(ctx.PostForm("diarize"))

	var refine *bool
	refineStr := ctx.PostForm("refine")
	if refineStr != "" {
		b, _ := strconv.ParseBool(refineStr)
		refine = &b
	}

	participants := []string{}
	pStr := ctx.PostForm("participants")
	if pStr != "" {
		participants = strings.Split(pStr, ",")
		for i, p := range participants {
			participants[i] = strings.TrimSpace(p)
		}
	}

	req := pipelineDtos.PipelineRequestDTO{
		Language:       ctx.PostForm("language"),
		TargetLanguage: ctx.PostForm("target_language"),
		Context:        ctx.PostForm("context"),
		Style:          ctx.PostForm("style"),
		Date:           ctx.PostForm("date"),
		Location:       ctx.PostForm("location"),
		Participants:   participants,
		Diarize:        diarize,
		Refine:         refine,
		Summarize:      summarize,
		MacAddress:     macAddress,
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")

	// 1. Check Idempotency BEFORE saving recording
	if idempotencyKey != "" {
		f, err := file.Open()
		if err == nil {
			audioHash, _ := utils.HashReader(f)
			f.Close()
			if taskID, exists := c.pipelineUC.CheckIdempotency(idempotencyKey, audioHash, req); exists {
				utils.LogInfo("Pipeline.ExecuteJob: Duplicate request detected (pre-save) for key %s. Returning TaskID %s", idempotencyKey, taskID)
				ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
					Status:  true,
					Message: "Pipeline job already submitted",
					Data: pipelineDtos.PipelineResponseDTO{
						TaskID: taskID,
					},
				})
				return
			}
		}
	}

	// 2. Not a duplicate (or no key), proceed to save
	recording, err := c.saveRecordingUC.SaveRecording(file, macAddress, baseURL, recordingUsecases.SaveRecordingOption{
		NotifyBIG: summarize,
	})
	if err != nil {
		utils.LogError("Pipeline.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "failed to save recording",
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	taskID, err := c.pipelineUC.ExecutePipeline(ctx.Request.Context(), finalInputPath, req, idempotencyKey)
	if err != nil {
		utils.LogError("Pipeline.ExecutePipeline: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "pipeline execution failed",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
		Status:  true,
		Message: "Pipeline job submitted successfully",
		Data: pipelineDtos.PipelineResponseDTO{
			TaskID: taskID,
		},
	})
}

// GetStatus handles GET /api/models/pipeline/status/:task_id
// @Summary Get unified pipeline job status
// @Description Poll for status and results of a unified pipeline job.
// @Tags 04. Models
// @Security BearerAuth
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} commonDtos.StandardResponse{data=pipelineDtos.PipelineStatusDTO}
// @Failure      404  {object}  commonDtos.ErrorResponse
// @Router /api/models/pipeline/status/{task_id} [get]
func (c *PipelineController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	status, err := c.statusUC.GetTaskStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, commonDtos.StandardResponse{
			Status:  false,
			Message: "Task not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status: true,
		Data:   status,
	})
}

// ExecuteJobByUpload handles POST /api/models/pipeline/job/by-upload
func (c *PipelineController) ExecuteJobByUpload(ctx *gin.Context) {
	var req speechdtos.PipelineSubmitByUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error: " + err.Error(),
		})
		return
	}

	// 1. Finalize session
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
	recording, err := c.saveRecordingUC.SaveRecordingFromPath(finalized.MergedPath, finalized.OriginalFileName, req.MacAddress, baseURL, recordingUsecases.SaveRecordingOption{
		NotifyBIG: req.Summarize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to save recording: " + err.Error(),
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	// 3. Start pipeline
	pipelineReq := pipelineDtos.PipelineRequestDTO{
		Language:       req.Language,
		TargetLanguage: req.TargetLanguage,
		Context:        req.Context,
		Style:          req.Style,
		Date:           req.Date,
		Location:       req.Location,
		Participants:   req.Participants,
		Diarize:        req.Diarize,
		Refine:         req.Refine,
		Summarize:      req.Summarize,
		MacAddress:     req.MacAddress,
	}

	// For by-upload requests, pass session ID to prevent idempotency collisions across different upload sessions
	taskID, err := c.pipelineUC.ExecutePipelineWithSession(ctx.Request.Context(), finalInputPath, pipelineReq, req.IdempotencyKey, req.SessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Pipeline execution failed: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
		Status:  true,
		Message: "Pipeline job submitted successfully via upload session",
		Data: pipelineDtos.PipelineResponseDTO{
			TaskID: taskID,
		},
	})
}
