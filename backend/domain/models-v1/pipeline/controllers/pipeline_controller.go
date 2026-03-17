package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	commonDtos "sensio/domain/common/dtos"
	pipelineDtos "sensio/domain/models-v1/pipeline/dtos"
	"sensio/domain/models-v1/pipeline/services"
	"sensio/domain/models-v1/pipeline/usecases"

	"github.com/gin-gonic/gin"
)

// Force usage for Swagger
var _ = pipelineDtos.V1PipelineStatusResponseDTO{}

// PipelineController handles pipeline requests.
type PipelineController struct {
	pipelineUC *usecases.PipelineUseCase
}

// NewPipelineController creates a new PipelineController.
func NewPipelineController(pipelineUC *usecases.PipelineUseCase) *PipelineController {
	return &PipelineController{
		pipelineUC: pipelineUC,
	}
}

// ExecuteJob handles POST /api/v1/models/pipeline/job
// @Summary Execute AI pipeline job
// @Description Execute a full AI pipeline (transcribe, translate, summarize)
// @Tags 05. Models-v1
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file"
// @Param language formData string false "Source language (default: id)"
// @Param target_language formData string false "Target language (default: en)"
// @Param summarize formData string false "Enable summarization (true/false)"
// @Param diarize formData string false "Enable speaker diarization (true/false)"
// @Param refine formData string false "Enable refinement (true/false)"
// @Param participants formData string false "Comma-separated participant names"
// @Success 202 {object} commonDtos.StandardResponse{data=commonDtos.TaskIDResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/models/pipeline/job [post]
func (c *PipelineController) ExecuteJob(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Audio file is required",
		})
		return
	}

	// Parse parameters
	language := ctx.PostForm("language")
	if language == "" {
		language = "id"
	}
	targetLanguage := ctx.PostForm("target_language")
	if targetLanguage == "" {
		targetLanguage = "en"
	}
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

	// Save file temporarily
	tempPath := filepath.Join("uploads", "audio", file.Filename)
	if err := ctx.SaveUploadedFile(file, tempPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to save audio file",
		})
		return
	}

	// Call usecase
	req := services.PipelineRequest{
		AudioPath:      tempPath,
		Language:       language,
		TargetLanguage: targetLanguage,
		Context:        ctx.PostForm("context"),
		Style:          ctx.PostForm("style"),
		Date:           ctx.PostForm("date"),
		Location:       ctx.PostForm("location"),
		Participants:   participants,
		Diarize:        diarize,
		Refine:         refine,
		Summarize:      summarize,
	}

	taskID, err := c.pipelineUC.ExecutePipeline(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to execute pipeline: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
		Status:  true,
		Message: "Pipeline job submitted",
		Data: map[string]string{
			"task_id": taskID,
		},
	})
}

// ExecuteJobByUpload handles POST /api/v1/models/pipeline/job/by-upload
func (c *PipelineController) ExecuteJobByUpload(ctx *gin.Context) {
	// TODO: Implement chunked upload support
	ctx.JSON(http.StatusNotImplemented, commonDtos.StandardResponse{
		Status:  false,
		Message: "Not implemented",
	})
}

// GetStatus handles GET /api/v1/models/pipeline/status/:task_id
// @Summary Get pipeline status
// @Description Get the status of a pipeline job by ID
// @Tags 05. Models-v1
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} commonDtos.StandardResponse{data=pipelineDtos.V1PipelineStatusResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      404  {object}  commonDtos.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/models/pipeline/status/{task_id} [get]
func (c *PipelineController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Task ID is required",
		})
		return
	}

	status, err := c.pipelineUC.GetPipelineStatus(ctx.Request.Context(), taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, commonDtos.StandardResponse{
			Status:  false,
			Message: "Task not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Pipeline status retrieved",
		Data:    status,
	})
}
