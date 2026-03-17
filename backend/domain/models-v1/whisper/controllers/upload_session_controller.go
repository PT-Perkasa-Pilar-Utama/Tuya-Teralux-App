package controllers

import (
	"io"
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	whisperDtos "sensio/domain/models-v1/whisper/dtos"
	"sensio/domain/models-v1/whisper/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UploadSessionController handles chunked upload sessions.
type UploadSessionController struct {
	grpcService *services.GrpcWhisperService
}

// NewUploadSessionController creates a new UploadSessionController.
func NewUploadSessionController(grpcService *services.GrpcWhisperService) *UploadSessionController {
	return &UploadSessionController{
		grpcService: grpcService,
	}
}

// CreateSession handles POST /api/v1/models/whisper/uploads/sessions
// @Summary      Create upload session (v1)
// @Description  Initialize a new chunked upload session for a large audio file
// @Tags         05. Models-v1
// @Accept       json
// @Produce      json
// @Param        request  body      whisperDtos.CreateUploadSessionRequest  true  "Upload session configuration"
// @Success      201  {object}  commonDtos.StandardResponse{data=whisperDtos.UploadSessionResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/v1/models/whisper/uploads/sessions [post]
// @Security     BearerAuth
func (c *UploadSessionController) CreateSession(ctx *gin.Context) {
	var req whisperDtos.CreateUploadSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	resp, err := c.grpcService.CreateUploadSession(req.FileName, req.TotalSizeBytes, int32(req.ChunkSizeBytes))
	if err != nil {
		ctx.JSON(mapUploadErrToStatus(err), commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to create upload session",
			Details: err.Error(),
		})
		return
	}

	dto := &whisperDtos.UploadSessionResponseDTO{
		SessionID:      resp.SessionID,
		State:          resp.Status,
		TotalChunks:    int(resp.ChunkCount),
		ChunkSizeBytes: 0,
		TotalSizeBytes: resp.TotalSize,
		ReceivedBytes:  0,
		ExpiresAt:      resp.CreatedAt,
	}

	ctx.JSON(http.StatusCreated, commonDtos.StandardResponse{
		Status:  true,
		Message: "Upload session created successfully",
		Data:    dto,
	})
}

// UploadChunk handles PUT /api/v1/models/whisper/uploads/sessions/:id/chunks/:index
// @Summary      Upload chunk (v1)
// @Description  Upload a single chunk of an audio file for a session
// @Tags         05. Models-v1
// @Accept       octet-stream
// @Produce      json
// @Param        id     path      string  true  "Session ID"
// @Param        index  path      int     true  "Chunk index"
// @Param        chunk  body      string  true  "Binary chunk data"
// @Success      200  {object}  commonDtos.StandardResponse{data=whisperDtos.UploadChunkAckDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/v1/models/whisper/uploads/sessions/{id}/chunks/{index} [put]
// @Security     BearerAuth
func (c *UploadSessionController) UploadChunk(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	indexStr := ctx.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid chunk index",
		})
		return
	}

	chunkData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to read chunk data",
		})
		return
	}

	resp, err := c.grpcService.UploadChunk(sessionID, int32(index), chunkData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to upload chunk",
			Details: err.Error(),
		})
		return
	}

	if !resp.Success {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Chunk upload failed",
			Details: resp.Error,
		})
		return
	}

	dto := &whisperDtos.UploadChunkAckDTO{
		ReceivedChunks: int(resp.UploadedChunks),
		ReceivedBytes:  0,
		IsDuplicate:    false,
		State:          "uploading",
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Chunk uploaded successfully",
		Data:    dto,
	})
}

// GetSessionStatus handles GET /api/v1/models/whisper/uploads/sessions/:id
// @Summary      Get upload session status (v1)
// @Description  Get the current status and progress of an upload session
// @Tags         05. Models-v1
// @Produce      json
// @Param        id   path      string  true  "Session ID"
// @Success      200  {object}  commonDtos.StandardResponse{data=whisperDtos.UploadSessionResponseDTO}
// @Failure      404  {object}  commonDtos.ErrorResponse
// @Router       /api/v1/models/whisper/uploads/sessions/{id} [get]
// @Security     BearerAuth
func (c *UploadSessionController) GetSessionStatus(ctx *gin.Context) {
	sessionID := ctx.Param("id")

	resp, err := c.grpcService.GetSessionStatus(sessionID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, commonDtos.StandardResponse{
			Status:  false,
			Message: "Session not found",
			Details: err.Error(),
		})
		return
	}

	dto := &whisperDtos.UploadSessionResponseDTO{
		SessionID:      resp.SessionID,
		State:          resp.Status,
		TotalChunks:    int(resp.ChunkCount),
		ChunkSizeBytes: 0,
		TotalSizeBytes: resp.TotalSize,
		ReceivedBytes:  0,
		ExpiresAt:      resp.CreatedAt,
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Session status retrieved successfully",
		Data:    dto,
	})
}

func mapUploadErrToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := err.Error()

	if msg == "session not found" {
		return http.StatusNotFound
	}

	if msg == "unauthorized session access" {
		return http.StatusForbidden
	}

	if msg == "session not in uploading state" ||
		msg == "session not ready" ||
		msg == "session conflict" ||
		msg == "session consumed" ||
		msg == "session already expired" {
		return http.StatusConflict
	}

	if msg == "file_name is required" ||
		msg == "total_size_bytes must be greater than 0" ||
		msg == "invalid chunk index" ||
		(len(msg) >= 19 && msg[:19] == "invalid chunk index") ||
		(len(msg) >= 28 && msg[:28] == "file size exceeds maximum allowed") ||
		(len(msg) >= 18 && msg[:18] == "chunk size exceeds") {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
