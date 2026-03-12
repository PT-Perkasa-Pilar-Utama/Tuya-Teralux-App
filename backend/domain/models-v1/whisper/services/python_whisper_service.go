package services

import (
	"context"
	"fmt"
	"os"
	"sensio/domain/common/utils"
	grpcpb "sensio/domain/models-v1/whisper/services/grpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcWhisperService handles communication with Python Whisper service via gRPC.
type GrpcWhisperService struct {
	client grpcpb.WhisperServiceClient
	conn   *grpc.ClientConn
}

// TranscribeResponse represents the response from Python Whisper service.
type TranscribeResponse struct {
	TaskID      string
	Status      string
	Transcript  string
	Error       string
	DurationMs  int64
}

// JobStatusResponse represents the job status response.
type JobStatusResponse struct {
	JobID     string
	Status    string
	Result    string
	Error     string
	FileName  string
	CreatedAt int64
	UpdatedAt int64
}

// UploadSessionResponse represents the upload session response.
type UploadSessionResponse struct {
	SessionID      string
	FileName       string
	TotalSize      int64
	ChunkCount     int32
	UploadedChunks int32
	Status         string
	CreatedAt      int64
	ExpiresAt      int64
}

// UploadChunkResponse represents the upload chunk response.
type UploadChunkResponse struct {
	SessionID      string
	ChunkIndex     int32
	Success        bool
	Error          string
	UploadedChunks int32
}

// FinalizeUploadSessionResponse represents the finalize session response.
type FinalizeUploadSessionResponse struct {
	SessionID      string
	MergedFilePath string
	FileName       string
	TotalSize      int64
	Success        bool
	Error          string
}

// NewGrpcWhisperService creates a new gRPC Whisper service.
func NewGrpcWhisperService(cfg *utils.Config) (*GrpcWhisperService, error) {
	conn, err := grpc.NewClient(
		cfg.PythonGrpcServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(50 * 1024 * 1024), // 50MB
			grpc.MaxCallSendMsgSize(50 * 1024 * 1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := grpcpb.NewWhisperServiceClient(conn)

	return &GrpcWhisperService{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection.
func (s *GrpcWhisperService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Transcribe sends audio file to Python service for transcription via gRPC.
func (s *GrpcWhisperService) Transcribe(audioPath, language string, diarize bool) (*TranscribeResponse, error) {
	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	// Create gRPC request
	req := &grpcpb.TranscribeRequest{
		AudioData:   audioData,
		FileName:    audioPath,
		Language:    language,
		Diarize:     diarize,
		CorrelationId: fmt.Sprintf("req_%d", time.Now().UnixNano()),
	}

	// Call gRPC service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s.client.Transcribe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC Transcribe failed: %w", err)
	}

	return &TranscribeResponse{
		TaskID:     resp.TaskId,
		Status:     resp.Status,
		Transcript: resp.Transcript,
		Error:      resp.Error,
		DurationMs: resp.DurationMs,
	}, nil
}

// GetStatus gets transcription status from Python service via gRPC.
func (s *GrpcWhisperService) GetStatus(taskID string) (*TranscribeResponse, error) {
	req := &grpcpb.GetJobStatusRequest{
		JobId:         taskID,
		CorrelationId: fmt.Sprintf("status_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.client.GetJobStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetJobStatus failed: %w", err)
	}

	return &TranscribeResponse{
		TaskID:     resp.JobId,
		Status:     resp.Status,
		Transcript: resp.Result,
		Error:      resp.Error,
	}, nil
}

// CreateUploadSession creates a new upload session via gRPC.
func (s *GrpcWhisperService) CreateUploadSession(fileName string, totalSize int64, chunkCount int32) (*UploadSessionResponse, error) {
	req := &grpcpb.CreateUploadSessionRequest{
		FileName:   fileName,
		TotalSize:  totalSize,
		ChunkCount: chunkCount,
		CorrelationId: fmt.Sprintf("session_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.client.CreateUploadSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC CreateUploadSession failed: %w", err)
	}

	return &UploadSessionResponse{
		SessionID:      resp.SessionId,
		FileName:       resp.FileName,
		TotalSize:      resp.TotalSize,
		ChunkCount:     resp.ChunkCount,
		UploadedChunks: resp.UploadedChunks,
		Status:         resp.Status,
		CreatedAt:      resp.CreatedAt,
		ExpiresAt:      resp.ExpiresAt,
	}, nil
}

// UploadChunk uploads a chunk of audio file via gRPC streaming.
func (s *GrpcWhisperService) UploadChunk(sessionID string, chunkIndex int32, chunkData []byte) (*UploadChunkResponse, error) {
	req := &grpcpb.UploadChunkRequest{
		SessionId:    sessionID,
		ChunkIndex:   chunkIndex,
		ChunkData:    chunkData,
		CorrelationId: fmt.Sprintf("chunk_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := s.client.UploadChunk(ctx)
	if err != nil {
		return nil, fmt.Errorf("gRPC UploadChunk stream failed: %w", err)
	}

	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("gRPC UploadChunk send failed: %w", err)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("gRPC UploadChunk recv failed: %w", err)
	}

	return &UploadChunkResponse{
		SessionID:      resp.SessionId,
		ChunkIndex:     resp.ChunkIndex,
		Success:        resp.Success,
		Error:          resp.Error,
		UploadedChunks: resp.UploadedChunks,
	}, nil
}

// GetSessionStatus gets upload session status via gRPC.
func (s *GrpcWhisperService) GetSessionStatus(sessionID string) (*UploadSessionResponse, error) {
	req := &grpcpb.GetUploadSessionStatusRequest{
		SessionId:    sessionID,
		CorrelationId: fmt.Sprintf("session_status_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.client.GetUploadSessionStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetUploadSessionStatus failed: %w", err)
	}

	return &UploadSessionResponse{
		SessionID:      resp.SessionId,
		FileName:       resp.FileName,
		TotalSize:      resp.TotalSize,
		ChunkCount:     resp.ChunkCount,
		UploadedChunks: resp.UploadedChunks,
		Status:         resp.Status,
		CreatedAt:      resp.CreatedAt,
		ExpiresAt:      resp.ExpiresAt,
	}, nil
}

// FinalizeSession finalizes upload session and merges chunks via gRPC.
func (s *GrpcWhisperService) FinalizeSession(sessionID string) (*FinalizeUploadSessionResponse, error) {
	req := &grpcpb.FinalizeUploadSessionRequest{
		SessionId:    sessionID,
		CorrelationId: fmt.Sprintf("finalize_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s.client.FinalizeUploadSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC FinalizeUploadSession failed: %w", err)
	}

	return &FinalizeUploadSessionResponse{
		SessionID:      resp.SessionId,
		MergedFilePath: resp.MergedFilePath,
		FileName:       resp.FileName,
		TotalSize:      resp.TotalSize,
		Success:        resp.Success,
		Error:          resp.Error,
	}, nil
}
