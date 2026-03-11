package usecases

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FinalizedUpload struct {
	MergedPath       string
	OriginalFileName string
	MimeType         string
	OwnerUID         string
}

type UploadSessionUseCase interface {
	CreateSession(req dtos.CreateUploadSessionRequest) (*dtos.UploadSessionResponseDTO, error)
	UploadChunk(sessionID string, chunkIndex int, ownerUID string, reader io.Reader) (*dtos.UploadChunkAckDTO, error)
	GetSessionStatus(sessionID string, ownerUID string) (*dtos.UploadSessionResponseDTO, error)
	FinalizeSession(sessionID string, ownerUID string) (*FinalizedUpload, error)
	CleanupExpiredSessions(now time.Time) (int, error) // Returns count of deleted sessions
}

type uploadSessionUseCase struct {
	cache     tasks.BadgerStore
	cfg       *utils.Config
	mu        sync.Mutex // For in-memory coordination if needed, but primary state is in Badger
	uploadDir string
}

type sessionMetadata struct {
	ID             string        `json:"id"`
	FileName       string        `json:"file_name"`
	TotalSizeBytes int64         `json:"total_size_bytes"`
	ChunkSizeBytes int           `json:"chunk_size_bytes"`
	TotalChunks    int           `json:"total_chunks"`
	ReceivedChunks map[int]int64 `json:"received_chunks"` // Map index to real received size
	State          string        `json:"state"`           // uploading, ready, consumed, aborted
	OwnerUID       string        `json:"owner_uid"`
	MimeType       string        `json:"mime_type"`
	CreatedAt      time.Time     `json:"created_at"`
	ExpiresAt      time.Time     `json:"expires_at"`
}

func NewUploadSessionUseCase(cache tasks.BadgerStore, cfg *utils.Config) UploadSessionUseCase {
	uploadDir := filepath.Join("tmp", "chunk_uploads")
	_ = os.MkdirAll(uploadDir, 0755)

	return &uploadSessionUseCase{
		cache:     cache,
		cfg:       cfg,
		uploadDir: uploadDir,
	}
}

func (u *uploadSessionUseCase) CreateSession(req dtos.CreateUploadSessionRequest) (*dtos.UploadSessionResponseDTO, error) {
	// Fase 3: Validation hardening
	if req.FileName == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	// Sanitize filename to prevent path traversal
	safeFileName := filepath.Base(req.FileName)

	if req.TotalSizeBytes <= 0 {
		return nil, fmt.Errorf("total_size_bytes must be greater than 0")
	}

	maxSizeGB := int64(u.cfg.ChunkUploadMaxFileSizeGB)
	if maxSizeGB == 0 {
		maxSizeGB = 20 // Default 20GB
	}
	if req.TotalSizeBytes > maxSizeGB*1024*1024*1024 {
		return nil, fmt.Errorf("file size exceeds maximum allowed (%d GB)", maxSizeGB)
	}

	chunkSize := req.ChunkSizeBytes
	if chunkSize <= 0 {
		chunkSize = u.cfg.ChunkUploadDefaultChunkMB * 1024 * 1024
	}

	// Clamp chunk size
	minChunk := u.cfg.ChunkUploadMinChunkMB * 1024 * 1024
	maxChunk := u.cfg.ChunkUploadMaxChunkMB * 1024 * 1024
	if chunkSize < minChunk {
		chunkSize = minChunk
	}
	if chunkSize > maxChunk {
		chunkSize = maxChunk
	}

	totalChunks := int((req.TotalSizeBytes + int64(chunkSize) - 1) / int64(chunkSize))
	sessionID := uuid.New().String()

	ttl, err := time.ParseDuration(u.cfg.ChunkUploadSessionTTL)
	if err != nil {
		ttl = 24 * time.Hour
	}
	expiresAt := time.Now().Add(ttl)

	meta := &sessionMetadata{
		ID:             sessionID,
		FileName:       safeFileName,
		TotalSizeBytes: req.TotalSizeBytes,
		ChunkSizeBytes: chunkSize,
		TotalChunks:    totalChunks,
		ReceivedChunks: make(map[int]int64),
		State:          "uploading",
		OwnerUID:       req.OwnerUID,
		MimeType:       req.MimeType,
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
	}

	if err := u.saveMetadata(meta); err != nil {
		return nil, err
	}

	// Create directory for chunks
	sessionDir := filepath.Join(u.uploadDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %v", err)
	}

	return &dtos.UploadSessionResponseDTO{
		SessionID:      sessionID,
		State:          meta.State,
		TotalChunks:    meta.TotalChunks,
		ChunkSizeBytes: meta.ChunkSizeBytes,
		TotalSizeBytes: meta.TotalSizeBytes,
		ExpiresAt:      meta.ExpiresAt,
	}, nil
}

func (u *uploadSessionUseCase) UploadChunk(sessionID string, chunkIndex int, ownerUID string, reader io.Reader) (*dtos.UploadChunkAckDTO, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	meta, err := u.getMetadata(sessionID)
	if err != nil {
		return nil, err
	}

	// Fase 3: Ownership check
	if meta.OwnerUID != "" && meta.OwnerUID != ownerUID {
		return nil, fmt.Errorf("unauthorized session access")
	}

	// Fase 3: Strict state machine
	if meta.State != "uploading" {
		return nil, fmt.Errorf("session not in uploading state")
	}

	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks {
		return nil, fmt.Errorf("invalid chunk index %d", chunkIndex)
	}

	// Save chunk to disk
	chunkPath := filepath.Join(u.uploadDir, sessionID, fmt.Sprintf("chunk_%d", chunkIndex))

	// Open file with O_CREATE|O_WRONLY|O_TRUNC
	out, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk file: %v", err)
	}
	defer out.Close()

	n, err := io.Copy(out, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write chunk: %v", err)
	}

	// Fase 3: Byte size validation
	expectedSize := int64(meta.ChunkSizeBytes)
	if chunkIndex == meta.TotalChunks-1 {
		expectedSize = meta.TotalSizeBytes - (int64(chunkIndex) * int64(meta.ChunkSizeBytes))
	}
	if n > expectedSize {
		return nil, fmt.Errorf("chunk size exceeds expected size (got %d, expected max %d)", n, expectedSize)
	}

	// Update metadata with actual received bytes
	isDuplicate := false
	if _, ok := meta.ReceivedChunks[chunkIndex]; ok {
		isDuplicate = true
	}

	meta.ReceivedChunks[chunkIndex] = n

	if len(meta.ReceivedChunks) == meta.TotalChunks {
		meta.State = "ready"
	}

	if err := u.saveMetadata(meta); err != nil {
		return nil, err
	}

	return &dtos.UploadChunkAckDTO{
		ReceivedChunks: len(meta.ReceivedChunks),
		ReceivedBytes:  u.calculateReceivedBytes(meta),
		IsDuplicate:    isDuplicate,
		State:          meta.State,
	}, nil
}

func (u *uploadSessionUseCase) GetSessionStatus(sessionID string, ownerUID string) (*dtos.UploadSessionResponseDTO, error) {
	meta, err := u.getMetadata(sessionID)
	if err != nil {
		return nil, err
	}

	// Fase 3: Ownership check
	if meta.OwnerUID != "" && meta.OwnerUID != ownerUID {
		return nil, fmt.Errorf("unauthorized session access")
	}

	missing := u.calculateMissingRanges(meta)

	return &dtos.UploadSessionResponseDTO{
		SessionID:      sessionID,
		State:          meta.State,
		TotalChunks:    meta.TotalChunks,
		ChunkSizeBytes: meta.ChunkSizeBytes,
		TotalSizeBytes: meta.TotalSizeBytes,
		ReceivedBytes:  u.calculateReceivedBytes(meta),
		MissingRanges:  missing,
		ExpiresAt:      meta.ExpiresAt,
	}, nil
}

func (u *uploadSessionUseCase) FinalizeSession(sessionID string, ownerUID string) (*FinalizedUpload, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	meta, err := u.getMetadata(sessionID)
	if err != nil {
		return nil, err
	}

	// Phase 2: Ownership check
	if meta.OwnerUID != "" && meta.OwnerUID != ownerUID {
		return nil, fmt.Errorf("unauthorized session access")
	}

	// Fase 3: Transition to ready if all chunks ok
	if meta.State != "ready" && len(meta.ReceivedChunks) < meta.TotalChunks {
		return nil, fmt.Errorf("session not ready: %d/%d chunks received", len(meta.ReceivedChunks), meta.TotalChunks)
	}

	// Merge chunks
	// Secure filename for merged file
	safeFileName := filepath.Base(meta.FileName)
	if safeFileName == "." || safeFileName == "/" {
		safeFileName = fmt.Sprintf("upload_%s.wav", sessionID)
	}

	mergedPath := filepath.Join(u.uploadDir, sessionID, safeFileName)
	out, err := os.Create(mergedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create merged file: %v", err)
	}
	defer out.Close()

	for i := 0; i < meta.TotalChunks; i++ {
		chunkPath := filepath.Join(u.uploadDir, sessionID, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open chunk %d: %v", i, err)
		}
		_, err = io.Copy(out, chunkFile)
		chunkFile.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to merge chunk %d: %v", i, err)
		}
	}

	meta.State = "consumed"
	if err := u.saveMetadata(meta); err != nil {
		return nil, err
	}

	return &FinalizedUpload{
		MergedPath:       mergedPath,
		OriginalFileName: safeFileName,
		MimeType:         meta.MimeType,
		OwnerUID:         meta.OwnerUID,
	}, nil
}

func (u *uploadSessionUseCase) calculateReceivedBytes(meta *sessionMetadata) int64 {
	var total int64
	for _, size := range meta.ReceivedChunks {
		total += size
	}
	return total
}

func (u *uploadSessionUseCase) CleanupExpiredSessions(now time.Time) (int, error) {
	prefix := "cache:upload:session:"
	keys, err := u.cache.KeysWithPrefix(prefix)
	if err != nil {
		return 0, fmt.Errorf("failed to list session keys: %v", err)
	}

	deletedCount := 0
	for _, key := range keys {
		data, _, err := u.cache.GetWithTTL(key)
		if err != nil || data == nil {
			continue
		}

		var meta sessionMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// State terminal: consumed or aborted
		isTerminal := meta.State == "consumed" || meta.State == "aborted"
		isExpired := meta.ExpiresAt.Before(now)

		if isExpired || isTerminal {
			// Remove folder
			sessionDir := filepath.Join(u.uploadDir, meta.ID)
			_ = os.RemoveAll(sessionDir)

			// Remove metadata
			_ = u.cache.Delete(key)
			deletedCount++
		}
	}

	return deletedCount, nil
}

// Helpers

func (u *uploadSessionUseCase) saveMetadata(meta *sessionMetadata) error {
	key := fmt.Sprintf("cache:upload:session:%s", meta.ID)
	data, _ := json.Marshal(meta)

	ttl := time.Until(meta.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	return u.cache.SetWithTTL(key, data, ttl)
}

func (u *uploadSessionUseCase) getMetadata(sessionID string) (*sessionMetadata, error) {
	key := fmt.Sprintf("cache:upload:session:%s", sessionID)
	data, _, err := u.cache.GetWithTTL(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("session not found")
	}

	var meta sessionMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (u *uploadSessionUseCase) calculateMissingRanges(meta *sessionMetadata) []string {
	if len(meta.ReceivedChunks) == meta.TotalChunks {
		return nil
	}

	var missing []int
	for i := 0; i < meta.TotalChunks; i++ {
		if _, ok := meta.ReceivedChunks[i]; !ok {
			missing = append(missing, i)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	sort.Ints(missing)

	var ranges []string
	if len(missing) == 0 {
		return ranges
	}

	start := missing[0]
	end := missing[0]

	for i := 1; i < len(missing); i++ {
		if missing[i] == end+1 {
			end = missing[i]
		} else {
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = missing[i]
			end = missing[i]
		}
	}

	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	return ranges
}
