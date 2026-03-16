package usecases

import (
	"bytes"
	"os"
	"path/filepath"
	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	"testing"
	"time"
)

// MockBadgerService implements tasks.BadgerStore interface for testing
type MockBadgerService struct {
	data map[string][]byte
	ttls map[string]time.Duration
}

func NewMockBadgerService() *MockBadgerService {
	return &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}
}

func (m *MockBadgerService) Set(key string, value []byte) error {
	m.data[key] = value
	if _, exists := m.ttls[key]; !exists {
		m.ttls[key] = 3600 * time.Second // Default TTL
	}
	return nil
}

func (m *MockBadgerService) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	m.ttls[key] = ttl
	return nil
}

func (m *MockBadgerService) SetIfAbsentWithTTL(key string, value []byte, ttl time.Duration) (bool, error) {
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	m.data[key] = value
	m.ttls[key] = ttl
	return true, nil
}

func (m *MockBadgerService) SetPreserveTTL(key string, value []byte) error {
	m.data[key] = value
	if _, exists := m.ttls[key]; !exists {
		m.ttls[key] = 3600 * time.Second
	}
	return nil
}

func (m *MockBadgerService) GetWithTTL(key string) ([]byte, time.Duration, error) {
	data, ok := m.data[key]
	if !ok {
		return nil, 0, nil
	}
	ttl := m.ttls[key]
	return data, ttl, nil
}

func (m *MockBadgerService) Delete(key string) error {
	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

func (m *MockBadgerService) KeysWithPrefix(prefix string) ([]string, error) {
	var keys []string
	for k := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// Helper function to create a test usecase
func createTestUploadSessionUseCase(t *testing.T) (*uploadSessionUseCase, func()) {
	cfg := &utils.Config{
		ChunkUploadMaxFileSizeGB:        20,
		ChunkUploadDefaultChunkBytes:    1024,       // 1KB for testing
		ChunkUploadMinChunkBytes:        256,        // 256 bytes
		ChunkUploadMaxChunkBytes:        10 * 1024,  // 10KB
		ChunkUploadSessionTTL:           "24h",
	}

	// Use a temporary directory for test uploads
	tempDir := t.TempDir()
	
	// Create a mock badger store
	mockBadger := NewMockBadgerService()
	
	uc := &uploadSessionUseCase{
		cache:     mockBadger,
		cfg:       cfg,
		uploadDir: tempDir,
	}
	
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}
	
	return uc, cleanup
}

// Test UploadChunk with exact size validation
func TestUploadChunk_ExactSizeValidation(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	// Create a session
	req := dtos.CreateUploadSessionRequest{
		FileName:       "test.m4a",
		TotalSizeBytes: 3072, // 3 chunks of 1024 bytes
		ChunkSizeBytes: 1024,
		OwnerUID:       "test-user",
		MimeType:       "audio/mp4",
	}
	
	sessionResp, err := uc.CreateSession(req)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Run("non-last chunk with fewer bytes than expected is rejected", func(t *testing.T) {
		// Upload chunk 0 with only 512 bytes instead of 1024
		chunkData := make([]byte, 512)
		ack, err := uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
		
		if err == nil {
			t.Error("Expected error for undersized chunk, got nil")
		}
		if ack != nil {
			t.Error("Expected nil ack for failed upload")
		}
		
		// Verify chunk was not recorded
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		if _, exists := meta.ReceivedChunks[0]; exists {
			t.Error("Chunk 0 should not be recorded for undersized upload")
		}
		
		// Verify chunk file was deleted
		chunkPath := filepath.Join(uc.uploadDir, sessionResp.SessionID, "chunk_0")
		if _, err := os.Stat(chunkPath); !os.IsNotExist(err) {
			t.Error("Chunk file should be deleted for undersized upload")
		}
	})

	t.Run("non-last chunk with exact bytes is accepted", func(t *testing.T) {
		// Upload chunk 0 with exact 1024 bytes
		chunkData := make([]byte, 1024)
		ack, err := uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
		
		if err != nil {
			t.Errorf("Unexpected error for correctly-sized chunk: %v", err)
		}
		if ack == nil {
			t.Error("Expected ack for successful upload")
		}
		
		// Verify chunk was recorded
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		if size, exists := meta.ReceivedChunks[0]; !exists || size != 1024 {
			t.Errorf("Chunk 0 should be recorded with size 1024, got %d", size)
		}
	})

	t.Run("last chunk with wrong size is rejected", func(t *testing.T) {
		// Create a new session for this test
		req2 := dtos.CreateUploadSessionRequest{
			FileName:       "test2.m4a",
			TotalSizeBytes: 2048, // 2 chunks of 1024 bytes
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp2, _ := uc.CreateSession(req2)
		
		// Upload chunk 0 correctly
		chunkData0 := make([]byte, 1024)
		uc.UploadChunk(sessionResp2.SessionID, 0, "test-user", bytes.NewReader(chunkData0))
		
		// Upload chunk 1 (last) with wrong size (512 instead of 1024)
		chunkData1 := make([]byte, 512)
		ack, err := uc.UploadChunk(sessionResp2.SessionID, 1, "test-user", bytes.NewReader(chunkData1))
		
		if err == nil {
			t.Error("Expected error for undersized last chunk, got nil")
		}
		if ack != nil {
			t.Error("Expected nil ack for failed upload")
		}
	})

	t.Run("last chunk with exact size is accepted", func(t *testing.T) {
		// Create a new session for this test
		req3 := dtos.CreateUploadSessionRequest{
			FileName:       "test3.m4a",
			TotalSizeBytes: 2048, // 2 chunks of 1024 bytes
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp3, _ := uc.CreateSession(req3)
		
		// Upload chunk 0 correctly
		chunkData0 := make([]byte, 1024)
		uc.UploadChunk(sessionResp3.SessionID, 0, "test-user", bytes.NewReader(chunkData0))
		
		// Upload chunk 1 (last) with exact size
		chunkData1 := make([]byte, 1024)
		ack, err := uc.UploadChunk(sessionResp3.SessionID, 1, "test-user", bytes.NewReader(chunkData1))
		
		if err != nil {
			t.Errorf("Unexpected error for correctly-sized last chunk: %v", err)
		}
		if ack == nil {
			t.Error("Expected ack for successful upload")
		}
		
		// Verify session is now ready
		meta, _ := uc.getMetadata(sessionResp3.SessionID)
		if meta.State != "ready" {
			t.Errorf("Session state should be 'ready', got %s", meta.State)
		}
	})
}

// Test session integrity validation
func TestValidateSessionIntegrity(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	t.Run("fails when a stored chunk file is missing", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 2048,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload only chunk 0
		chunkData := make([]byte, 1024)
		uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
		
		// Manually corrupt metadata to claim chunk 1 exists
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		meta.ReceivedChunks[1] = 1024
		uc.saveMetadata(meta)
		
		// Validation should fail
		err := uc.validateSessionIntegrity(meta)
		if err == nil {
			t.Error("Expected error for missing chunk file, got nil")
		}
		if err.Error() != "chunk 1 file missing" {
			t.Errorf("Expected 'chunk 1 file missing' error, got: %v", err)
		}
	})

	t.Run("fails when stored chunk metadata differs from file size", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 2048,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload chunk 0 with 1024 bytes
		chunkData := make([]byte, 1024)
		uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
		
		// Manually corrupt metadata to claim different size
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		meta.ReceivedChunks[0] = 512 // Claim it's only 512 bytes
		uc.saveMetadata(meta)
		
		// Validation should fail
		err := uc.validateSessionIntegrity(meta)
		if err == nil {
			t.Error("Expected error for size mismatch, got nil")
		}
	})

	t.Run("fails when all chunk indexes exist but total received bytes do not match", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 3000, // Non-aligned size
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload all chunks correctly
		chunkData0 := make([]byte, 1024)
		chunkData1 := make([]byte, 1024)
		chunkData2 := make([]byte, 952) // Last chunk: 3000 - 2048 = 952
		
		uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData0))
		uc.UploadChunk(sessionResp.SessionID, 1, "test-user", bytes.NewReader(chunkData1))
		uc.UploadChunk(sessionResp.SessionID, 2, "test-user", bytes.NewReader(chunkData2))
		
		// Manually corrupt metadata to claim wrong total
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		meta.ReceivedChunks[2] = 1024 // Claim last chunk is full size
		uc.saveMetadata(meta)
		
		// Validation should fail
		err := uc.validateSessionIntegrity(meta)
		if err == nil {
			t.Error("Expected error for total bytes mismatch, got nil")
		}
	})
}

// Test FinalizeSession with merged file validation
func TestFinalizeSession_MergedFileValidation(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	t.Run("finalize succeeds when merged output size equals total_size_bytes", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 3072,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload all chunks correctly
		for i := 0; i < 3; i++ {
			chunkData := make([]byte, 1024)
			_, err := uc.UploadChunk(sessionResp.SessionID, i, "test-user", bytes.NewReader(chunkData))
			if err != nil {
				t.Fatalf("Failed to upload chunk %d: %v", i, err)
			}
		}
		
		// Finalize should succeed
		finalized, err := uc.FinalizeSession(sessionResp.SessionID, "test-user")
		if err != nil {
			t.Fatalf("FinalizeSession failed: %v", err)
		}
		if finalized == nil {
			t.Error("Expected finalized upload result")
		}
		
		// Verify merged file exists and has correct size
		mergedInfo, err := os.Stat(finalized.MergedPath)
		if err != nil {
			t.Fatalf("Failed to stat merged file: %v", err)
		}
		if mergedInfo.Size() != 3072 {
			t.Errorf("Merged file size should be 3072, got %d", mergedInfo.Size())
		}
	})

	t.Run("finalize rejects corrupt sessions and does not create merged file", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 3072,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload chunks 0 and 1 correctly
		for i := 0; i < 2; i++ {
			chunkData := make([]byte, 1024)
			uc.UploadChunk(sessionResp.SessionID, i, "test-user", bytes.NewReader(chunkData))
		}
		
		// Manually corrupt: claim chunk 2 exists but don't upload it
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		meta.ReceivedChunks[2] = 1024
		meta.State = "ready"
		uc.saveMetadata(meta)
		
		// Finalize should fail
		finalized, err := uc.FinalizeSession(sessionResp.SessionID, "test-user")
		if err == nil {
			t.Error("Expected error for corrupt session, got nil")
		}
		if finalized != nil {
			t.Error("Expected nil finalized result for corrupt session")
		}
	})
}

// Test GetSessionStatus with integrity validation
func TestGetSessionStatus_IntegrityValidation(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	t.Run("status returns error for corrupt session", func(t *testing.T) {
		req := dtos.CreateUploadSessionRequest{
			FileName:       "test.m4a",
			TotalSizeBytes: 2048,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		
		sessionResp, _ := uc.CreateSession(req)
		
		// Upload chunk 0
		chunkData := make([]byte, 1024)
		uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
		
		// Manually corrupt metadata
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		meta.ReceivedChunks[0] = 512 // Wrong size
		uc.saveMetadata(meta)
		
		// Status should fail
		status, err := uc.GetSessionStatus(sessionResp.SessionID, "test-user")
		if err == nil {
			t.Error("Expected error for corrupt session, got nil")
		}
		if status != nil {
			t.Error("Expected nil status for corrupt session")
		}
	})
}

// Test invalidateCorruptSession
func TestInvalidateCorruptSession(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	req := dtos.CreateUploadSessionRequest{
		FileName:       "test.m4a",
		TotalSizeBytes: 2048,
		ChunkSizeBytes: 1024,
		OwnerUID:       "test-user",
		MimeType:       "audio/mp4",
	}
	
	sessionResp, _ := uc.CreateSession(req)
	
	// Upload chunk 0
	chunkData := make([]byte, 1024)
	uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(chunkData))
	
	// Get metadata before invalidation
	meta, _ := uc.getMetadata(sessionResp.SessionID)
	
	// Invalidate
	err := uc.invalidateCorruptSession(meta)
	if err != nil {
		t.Fatalf("invalidateCorruptSession failed: %v", err)
	}
	
	// Verify session directory is deleted
	sessionDir := filepath.Join(uc.uploadDir, sessionResp.SessionID)
	if _, err := os.Stat(sessionDir); !os.IsNotExist(err) {
		t.Error("Session directory should be deleted after invalidation")
	}
	
	// Verify metadata is deleted
	_, err = uc.getMetadata(sessionResp.SessionID)
	if err == nil {
		t.Error("Expected error getting invalidated session metadata")
	}
}

// Test CleanupExpiredSessions handles aborted sessions
func TestCleanupExpiredSessions_AbortedState(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	req := dtos.CreateUploadSessionRequest{
		FileName:       "test.m4a",
		TotalSizeBytes: 2048,
		ChunkSizeBytes: 1024,
		OwnerUID:       "test-user",
		MimeType:       "audio/mp4",
	}

	sessionResp, _ := uc.CreateSession(req)

	// Manually set state to aborted
	meta, _ := uc.getMetadata(sessionResp.SessionID)
	meta.State = "aborted"
	uc.saveMetadata(meta)

	// Cleanup should remove the session
	count, err := uc.CleanupExpiredSessions(time.Now())
	if err != nil {
		t.Fatalf("CleanupExpiredSessions failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session cleaned up, got %d", count)
	}
}

// TestUploadChunk_RetryableInterruption validates that partial chunk uploads are properly handled
func TestUploadChunk_RetryableInterruption(t *testing.T) {
	uc, cleanup := createTestUploadSessionUseCase(t)
	defer cleanup()

	// Create a session
	req := dtos.CreateUploadSessionRequest{
		FileName:       "test.m4a",
		TotalSizeBytes: 2048,
		ChunkSizeBytes: 1024,
		OwnerUID:       "test-user",
		MimeType:       "audio/mp4",
	}

	sessionResp, err := uc.CreateSession(req)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Run("partial chunk upload returns error and removes file", func(t *testing.T) {
		// Simulate partial upload by sending only 512 bytes of expected 1024
		// This simulates a connection interruption mid-upload
		partialData := make([]byte, 512)
		
		ack, err := uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(partialData))
		
		// Should return error for incomplete chunk
		if err == nil {
			t.Error("Expected error for partial chunk upload, got nil")
		}
		if ack != nil {
			t.Error("Expected nil ack for failed upload")
		}
		
		// Verify chunk was NOT recorded in metadata
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		if _, exists := meta.ReceivedChunks[0]; exists {
			t.Error("Partial chunk should not be recorded in metadata")
		}
		
		// Verify chunk file was deleted
		chunkPath := filepath.Join(uc.uploadDir, sessionResp.SessionID, "chunk_0")
		if _, err := os.Stat(chunkPath); !os.IsNotExist(err) {
			t.Error("Partial chunk file should be deleted")
		}
	})

	t.Run("full retry after partial failure succeeds", func(t *testing.T) {
		// After partial failure, full retry should succeed
		fullData := make([]byte, 1024)
		
		ack, err := uc.UploadChunk(sessionResp.SessionID, 0, "test-user", bytes.NewReader(fullData))
		
		if err != nil {
			t.Errorf("Unexpected error for full retry: %v", err)
		}
		if ack == nil {
			t.Error("Expected ack for successful upload")
		}
		
		// Verify chunk was recorded
		meta, _ := uc.getMetadata(sessionResp.SessionID)
		if size, exists := meta.ReceivedChunks[0]; !exists || size != 1024 {
			t.Errorf("Chunk 0 should be recorded with size 1024, got %d", size)
		}
		
		// Verify chunk file exists
		chunkPath := filepath.Join(uc.uploadDir, sessionResp.SessionID, "chunk_0")
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			t.Error("Full chunk file should exist")
		}
	})

	t.Run("session remains valid after partial upload", func(t *testing.T) {
		// Create new session for this test
		req2 := dtos.CreateUploadSessionRequest{
			FileName:       "test2.m4a",
			TotalSizeBytes: 2048,
			ChunkSizeBytes: 1024,
			OwnerUID:       "test-user",
			MimeType:       "audio/mp4",
		}
		sessionResp2, _ := uc.CreateSession(req2)
		
		// Partial upload
		partialData := make([]byte, 256)
		uc.UploadChunk(sessionResp2.SessionID, 0, "test-user", bytes.NewReader(partialData))
		
		// Session should still be in "uploading" state
		meta, _ := uc.getMetadata(sessionResp2.SessionID)
		if meta.State != "uploading" {
			t.Errorf("Session should remain in uploading state after partial upload, got %s", meta.State)
		}
		
		// Full retry should work
		fullData := make([]byte, 1024)
		ack, err := uc.UploadChunk(sessionResp2.SessionID, 0, "test-user", bytes.NewReader(fullData))
		
		if err != nil {
			t.Errorf("Full retry failed: %v", err)
		}
		if ack == nil {
			t.Error("Expected ack for successful retry")
		}
	})
}

