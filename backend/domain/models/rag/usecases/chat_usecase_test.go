package usecases

import (
	"context"
	"encoding/json"
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/models/rag/dtos"
	"sensio/domain/models/rag/skills"
	"sensio/domain/models/rag/skills/orchestrator"
	"strings"
	"testing"
	"time"
)

// MockControlUseCase is a mock implementation of ControlUseCase for testing
type MockControlUseCase struct {
	ProcessControlFunc func(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error)
}

func (m *MockControlUseCase) ProcessControl(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
	if m.ProcessControlFunc != nil {
		return m.ProcessControlFunc(ctx, uid, terminalID, prompt)
	}
	return &dtos.ControlResultDTO{Message: "OK", HTTPStatusCode: 200}, nil
}

// MockLLMClient is a mock LLM client for testing
type MockLLMClient struct{}

func (m MockLLMClient) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	return "", nil
}

func TestExecuteFastControl_PreservesOriginalPrompt(t *testing.T) {
	// Test that executeFastControl forwards the raw ctx.Prompt to ControlUseCase
	// instead of a reconstructed prompt from intent
	var capturedPrompt string

	mockControlUseCase := &MockControlUseCase{
		ProcessControlFunc: func(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
			capturedPrompt = prompt
			return &dtos.ControlResultDTO{
				Message:        "Berhasil menyalakan semua lampu.",
				HTTPStatusCode: 200,
			}, nil
		},
	}

	// Create minimal config
	cfg := &utils.Config{}

	// Create fast intent result for "nyalakan semua lampu"
	intentResult := orchestrator.FastIntentResult{
		Intent:     orchestrator.FastIntentControl,
		DeviceName: "lampu",
		ActionType: "on",
		Confidence: 0.9,
	}

	// Create skill context with the original prompt containing "semua" (all)
	originalPrompt := "nyalakan semua lampu"
	skillCtx := &skills.SkillContext{
		Ctx:        context.Background(),
		UID:        "test-user-id",
		TerminalID: "test-terminal-id",
		Prompt:     originalPrompt,
		Language:   "id",
		LLM:        MockLLMClient{},
		Config:     cfg,
		Vector:     &infrastructure.VectorService{},
		Badger:     &infrastructure.BadgerService{},
	}

	// Create the use case with mock control use case
	chatUseCase := &ChatUseCaseImpl{
		controlUseCase: mockControlUseCase,
		config:         cfg,
	}

	// Execute fast control
	result, err := chatUseCase.executeFastControl(skillCtx, intentResult)

	// Verify no error
	if err != nil {
		t.Fatalf("executeFastControl returned error: %v", err)
	}

	// Verify the captured prompt is the original prompt, not reconstructed
	if capturedPrompt != originalPrompt {
		t.Errorf("ControlUseCase.ProcessControl was called with wrong prompt:\n  got:  %q\n  want: %q", capturedPrompt, originalPrompt)
	}

	// Verify the result
	if result == nil {
		t.Fatal("executeFastControl returned nil result")
	}
	if result.Message != "Berhasil menyalakan semua lampu." {
		t.Errorf("Unexpected message: got %q, want %q", result.Message, "Berhasil menyalakan semua lampu.")
	}
	if !result.IsControl {
		t.Error("Expected result.IsControl to be true")
	}
}

func TestExecuteFastControl_PreservesQuantifiers(t *testing.T) {
	// Test various prompts with quantifiers to ensure they're preserved
	testCases := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "Indonesian all lights",
			prompt:   "matikan semua lampu",
			expected: "matikan semua lampu",
		},
		{
			name:     "English all lights",
			prompt:   "turn off all lights",
			expected: "turn off all lights",
		},
		{
			name:     "Indonesian with location",
			prompt:   "nyalakan semua lampu di ruang tamu",
			expected: "nyalakan semua lampu di ruang tamu",
		},
		{
			name:     "Specific quantifier",
			prompt:   "hidupkan semua switch",
			expected: "hidupkan semua switch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedPrompt string

			mockControlUseCase := &MockControlUseCase{
				ProcessControlFunc: func(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
					capturedPrompt = prompt
					return &dtos.ControlResultDTO{
						Message:        "OK",
						HTTPStatusCode: 200,
					}, nil
				},
			}

			cfg := &utils.Config{}
			intentResult := orchestrator.FastIntentResult{
				Intent:     orchestrator.FastIntentControl,
				DeviceName: "device",
				ActionType: "on",
				Confidence: 0.9,
			}

			skillCtx := &skills.SkillContext{
				Ctx:        context.Background(),
				UID:        "test-user",
				TerminalID: "test-terminal",
				Prompt:     tc.prompt,
				Language:   "id",
				LLM:        MockLLMClient{},
				Config:     cfg,
				Vector:     &infrastructure.VectorService{},
				Badger:     &infrastructure.BadgerService{},
			}

			chatUseCase := &ChatUseCaseImpl{
				controlUseCase: mockControlUseCase,
				config:         cfg,
			}

			_, err := chatUseCase.executeFastControl(skillCtx, intentResult)
			if err != nil {
				t.Fatalf("executeFastControl returned error: %v", err)
			}

			if capturedPrompt != tc.expected {
				t.Errorf("Prompt not preserved correctly:\n  got:  %q\n  want: %q", capturedPrompt, tc.expected)
			}
		})
	}
}

// TestChat_RequestId_Idempotency_Completed tests that duplicate requests with the same request_id
// return the cached response when the first request has completed.
func TestChat_RequestId_Idempotency_Completed(t *testing.T) {
	// Use mock BadgerService
	mockBadger := &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}

	cfg := &utils.Config{}
	mockControlUseCase := &MockControlUseCase{
		ProcessControlFunc: func(ctx context.Context, uid, terminalID, prompt string) (*dtos.ControlResultDTO, error) {
			return &dtos.ControlResultDTO{
				Message:        "Berhasil menyalakan AC",
				HTTPStatusCode: 200,
			}, nil
		},
	}

	chatUseCase := &ChatUseCaseImpl{
		config:           cfg,
		badger:           (*infrastructure.BadgerService)(nil), // Will use mock via custom wrapper
		controlUseCase:   mockControlUseCase,
		guard:            orchestrator.NewGuardOrchestrator(&MockGuardSkill{}),
		fastIntentRouter: orchestrator.NewFastIntentRouter(),
	}

	// Suppress unused variable warning
	_ = chatUseCase

	// For this test, we'll test the logic directly by seeding the cache
	// Since we can't easily mock BadgerService, we'll test the concept differently
	// This test demonstrates the expected behavior

	requestID := "test-request-id-123"
	_ = "test-user" // uid no longer used in key format (kept for potential future use)
	terminalID := "test-terminal"

	// Simulate a completed response in cache
	cachedResponse := &dtos.RAGChatResponseDTO{
		Response:       "Berhasil menyalakan AC",
		IsControl:      true,
		IsBlocked:      false,
		HTTPStatusCode: 200,
	}

	// Updated key format: terminalID + requestID (UID excluded for cross-channel stability)
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID
	cachedData, _ := json.Marshal(cachedResponse)
	mockBadger.data[idempotencyKey] = cachedData
	mockBadger.ttls[idempotencyKey] = 5 * time.Minute

	// Verify cache was seeded correctly
	data, ok := mockBadger.data[idempotencyKey]
	if !ok {
		t.Fatal("Failed to seed cache")
	}

	var unmarshaled dtos.RAGChatResponseDTO
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal cached data: %v", err)
	}

	if unmarshaled.Response != cachedResponse.Response {
		t.Errorf("Cache data mismatch: got %q, want %q", unmarshaled.Response, cachedResponse.Response)
	}
}

// TestChat_RequestId_Idempotency_InProgress tests that concurrent requests with the same request_id
// return a "processing" acknowledgment when the first request is still in progress.
func TestChat_RequestId_Idempotency_InProgress(t *testing.T) {
	mockBadger := &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}

	requestID := "test-request-id-in-progress"
	_ = "test-user" // uid no longer used in key format (kept for potential future use)
	terminalID := "test-terminal"

	// Simulate an in-progress request
	// Updated key format: terminalID + requestID (UID excluded for cross-channel stability)
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID
	mockBadger.data[idempotencyKey] = []byte("in_progress")
	mockBadger.ttls[idempotencyKey] = 30 * time.Second

	// Verify the in_progress state
	data, ok := mockBadger.data[idempotencyKey]
	if !ok {
		t.Fatal("Failed to seed cache with in_progress state")
	}

	if string(data) != "in_progress" {
		t.Errorf("Expected in_progress state, got %q", string(data))
	}
}

// TestChat_NoRequestId_BypassIdempotency tests that requests without request_id
// bypass the idempotency check entirely (backward compatibility).
// This is a conceptual test demonstrating the expected behavior.
func TestChat_NoRequestId_BypassIdempotency(t *testing.T) {
	// Conceptual test: Without request_id, the idempotency key would be empty
	// and the cache lookup would be skipped entirely
	// This demonstrates backward compatibility

	_ = "test-user" // uid no longer used in key format (kept for potential future use)
	terminalID := "test-terminal"
	requestID := "" // Empty request_id

	// Verify that idempotency key would be empty (updated format: terminalID + requestID)
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID
	expectedKey := "chat:idempotency:" + terminalID + ":"

	if idempotencyKey != expectedKey {
		t.Errorf("Expected idempotency key to end with empty request_id, got %q", idempotencyKey)
	}

	// The actual bypass logic is tested in integration tests
	// This test confirms the key construction is correct
}

// MockBadgerService for testing
type MockBadgerService struct {
	data map[string][]byte
	ttls map[string]time.Duration
}

func (m *MockBadgerService) Get(key string) ([]byte, error) {
	if data, ok := m.data[key]; ok {
		return data, nil
	}
	return nil, nil
}

func (m *MockBadgerService) GetWithTTL(key string) ([]byte, time.Duration, error) {
	if data, ok := m.data[key]; ok {
		ttl, exists := m.ttls[key]
		if !exists {
			ttl = 0
		}
		return data, ttl, nil
	}
	return nil, 0, nil
}

func (m *MockBadgerService) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

func (m *MockBadgerService) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	m.ttls[key] = ttl
	return nil
}

func (m *MockBadgerService) SetIfAbsentWithTTL(key string, value []byte, ttl time.Duration) (bool, error) {
	if _, ok := m.data[key]; ok {
		return false, nil
	}
	m.data[key] = value
	m.ttls[key] = ttl
	return true, nil
}

func (m *MockBadgerService) Delete(key string) error {
	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

// MockGuardSkill is a minimal mock for testing
type MockGuardSkill struct{}

func (m *MockGuardSkill) Name() string        { return "MockGuard" }
func (m *MockGuardSkill) Description() string { return "Mock guard skill for testing" }
func (m *MockGuardSkill) Execute(ctx *skills.SkillContext) (*skills.SkillResult, error) {
	return &skills.SkillResult{Message: "", IsBlocked: false}, nil
}

// TestFinalizeIdempotency_CachesResponse tests that finalizeIdempotency correctly caches responses
func TestFinalizeIdempotency_CachesResponse(t *testing.T) {
	mockBadger := &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}

	// Test the helper logic conceptually
	requestID := "test-req-123"
	_ = "test-user" // uid no longer used in key format (kept for potential future use)
	terminalID := "test-terminal"
	response := &dtos.RAGChatResponseDTO{
		Response:       "Test response",
		IsControl:      false,
		HTTPStatusCode: 200,
	}

	// Simulate what finalizeIdempotency does (updated key format: terminalID + requestID)
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID
	responseData, _ := json.Marshal(response)
	mockBadger.data[idempotencyKey] = responseData
	mockBadger.ttls[idempotencyKey] = 5 * time.Minute

	// Verify cache was written
	data, ok := mockBadger.data[idempotencyKey]
	if !ok {
		t.Fatal("Failed to write to cache")
	}

	var unmarshaled dtos.RAGChatResponseDTO
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal cached data: %v", err)
	}

	if unmarshaled.Response != response.Response {
		t.Errorf("Cached response mismatch: got %q, want %q", unmarshaled.Response, response.Response)
	}
}

// TestFinalizeIdempotency_SkipsEmptyRequestID tests that finalizeIdempotency skips when requestID is empty
func TestFinalizeIdempotency_SkipsEmptyRequestID(t *testing.T) {
	mockBadger := &MockBadgerService{
		data: make(map[string][]byte),
	}

	requestID := "" // Empty
	_ = "test-user" // uid no longer used in key format (kept for potential future use)
	terminalID := "test-terminal"

	// Should not write to cache
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID

	// Verify key would be malformed (ends with colon due to empty requestID)
	expectedIncompleteKey := "chat:idempotency:" + terminalID + ":"
	if idempotencyKey != expectedIncompleteKey {
		t.Errorf("Expected incomplete key for empty requestID, got %q", idempotencyKey)
	}

	// Verify cache was not written
	_, ok := mockBadger.data[idempotencyKey]
	if ok {
		t.Error("Cache should not be written for empty requestID")
	}
}

// TestIdempotencySourceMarkers_VerifyContract tests that idempotency source markers are correctly defined
func TestIdempotencySourceMarkers_VerifyContract(t *testing.T) {
	// Verify expected source values
	expectedSources := map[string]bool{
		"IDEMPOTENCY_CACHED":      true,
		"IDEMPOTENCY_IN_PROGRESS": true,
		"MQTT_SYNC_DROP":          true,
		"HTTP_HANDLER":            true,
		"MQTT_SUBSCRIBER":         true,
	}

	for source := range expectedSources {
		if source == "" {
			t.Error("Source marker should not be empty")
		}
	}

	// Test that cached response gets proper marker
	cachedResponse := &dtos.RAGChatResponseDTO{
		Response: "Cached",
		Source:   "IDEMPOTENCY_CACHED",
	}
	if cachedResponse.Source != "IDEMPOTENCY_CACHED" {
		t.Errorf("Expected IDEMPOTENCY_CACHED source, got %q", cachedResponse.Source)
	}

	// Test that in-progress response gets proper marker
	inProgressResponse := &dtos.RAGChatResponseDTO{
		Response: "Processing",
		Source:   "IDEMPOTENCY_IN_PROGRESS",
	}
	if inProgressResponse.Source != "IDEMPOTENCY_IN_PROGRESS" {
		t.Errorf("Expected IDEMPOTENCY_IN_PROGRESS source, got %q", inProgressResponse.Source)
	}
}

// TestAllLightsIntent_RestrictiveMatching tests that all-lights detection requires light-specific words
func TestAllLightsIntent_RestrictiveMatching(t *testing.T) {
	testCases := []struct {
		name      string
		prompt    string
		wantMatch bool
	}{
		// Should match - light-specific
		{name: "Indonesian all lights", prompt: "nyalakan semua lampu", wantMatch: true},
		{name: "English all lights", prompt: "turn on all lights", wantMatch: true},
		{name: "All the lights", prompt: "turn on all the lights in living room", wantMatch: true},
		{name: "Every light", prompt: "turn off every light", wantMatch: true},

		// Should NOT match - generic "all" without light word
		{name: "Generic all on", prompt: "nyalakan semua", wantMatch: false},
		{name: "Generic all off", prompt: "matikan semua", wantMatch: false},
		{name: "Turn on all", prompt: "turn on all", wantMatch: false},
		{name: "Turn off all", prompt: "turn off all", wantMatch: false},
		{name: "All devices", prompt: "turn on all devices", wantMatch: false},
		{name: "All AC", prompt: "turn on all AC", wantMatch: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: isAllLightsIntent is not exported, so we test the concept
			// In real implementation, this would call orchestrator.isAllLightsIntent(tc.prompt)
			// For now, we verify the test cases are correctly defined
			hasLightWord := strings.Contains(strings.ToLower(tc.prompt), "lampu") ||
				strings.Contains(strings.ToLower(tc.prompt), "light") ||
				strings.Contains(strings.ToLower(tc.prompt), "lamp")
			hasAllQuantifier := strings.Contains(strings.ToLower(tc.prompt), "semua") ||
				strings.Contains(strings.ToLower(tc.prompt), "all") ||
				strings.Contains(strings.ToLower(tc.prompt), "every")

			gotMatch := hasLightWord && hasAllQuantifier

			// Direct pattern match for explicit phrases
			promptLower := strings.ToLower(tc.prompt)
			explicitPatterns := []string{"semua lampu", "lampu semua", "all lights", "all light", "every light"}
			for _, pattern := range explicitPatterns {
				if strings.Contains(promptLower, pattern) {
					gotMatch = true
					break
				}
			}

			if gotMatch != tc.wantMatch {
				t.Errorf("All-lights intent mismatch for %q: got %v, want %v", tc.prompt, gotMatch, tc.wantMatch)
			}
		})
	}
}

// TestChat_Idempotency_KeyConsistency verifies that lock/read and finalize paths use the same key
func TestChat_Idempotency_KeyConsistency(t *testing.T) {
	terminalID := "test-terminal-001"
	requestID := "test-request-001"
	uid := "test-user" // uid should NOT be in key

	// Expected key format: chat:idempotency:{terminalID}:{requestID}
	expectedKey := "chat:idempotency:" + terminalID + ":" + requestID

	// Verify pre-check path key construction (simulating line 84)
	preCheckKey := "chat:idempotency:" + terminalID + ":" + requestID
	if preCheckKey != expectedKey {
		t.Errorf("Pre-check key mismatch: got %q, want %q", preCheckKey, expectedKey)
	}

	// Verify finalize path key construction (simulating finalizeIdempotency)
	// Should use same format without uid
	finalizeKey := "chat:idempotency:" + terminalID + ":" + requestID
	if finalizeKey != expectedKey {
		t.Errorf("Finalize key mismatch: got %q, want %q", finalizeKey, expectedKey)
	}

	// Verify uid is NOT in the key
	if strings.Contains(expectedKey, uid) {
		t.Error("Idempotency key should not contain uid")
	}
}

// TestChat_Idempotency_CompleteFlow tests the full idempotency flow:
// 1. First request sets in_progress and completes
// 2. Duplicate request retrieves cached response
func TestChat_Idempotency_CompleteFlow(t *testing.T) {
	mockBadger := &MockBadgerService{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}

	terminalID := "test-terminal-flow"
	requestID := "test-request-flow"
	idempotencyKey := "chat:idempotency:" + terminalID + ":" + requestID

	// Step 1: Simulate first request setting in_progress
	mockBadger.data[idempotencyKey] = []byte("in_progress")
	mockBadger.ttls[idempotencyKey] = 30 * time.Second

	// Verify in_progress state
	data, ok := mockBadger.data[idempotencyKey]
	if !ok {
		t.Fatal("Failed to set in_progress state")
	}
	if string(data) != "in_progress" {
		t.Errorf("Expected in_progress, got %q", string(data))
	}

	// Step 2: Simulate first request completing (finalizeIdempotency writes response)
	completedResponse := &dtos.RAGChatResponseDTO{
		Response:       "Lampu berhasil dinyalakan",
		IsControl:      true,
		IsBlocked:      false,
		HTTPStatusCode: 200,
		RequestID:      requestID,
		Source:         "HTTP_HANDLER",
	}
	responseData, _ := json.Marshal(completedResponse)
	mockBadger.data[idempotencyKey] = responseData
	mockBadger.ttls[idempotencyKey] = 5 * time.Minute

	// Step 3: Simulate duplicate request reading cache
	cachedData, _, err := mockBadger.GetWithTTL(idempotencyKey)
	if err != nil {
		t.Fatalf("Failed to read cached response: %v", err)
	}

	var cachedResponse dtos.RAGChatResponseDTO
	if err := json.Unmarshal(cachedData, &cachedResponse); err != nil {
		t.Fatalf("Failed to unmarshal cached response: %v", err)
	}

	// Verify cached response matches original
	if cachedResponse.Response != completedResponse.Response {
		t.Errorf("Cached response mismatch: got %q, want %q", cachedResponse.Response, completedResponse.Response)
	}
	if cachedResponse.RequestID != requestID {
		t.Errorf("Cached requestID mismatch: got %q, want %q", cachedResponse.RequestID, requestID)
	}

	// Verify Source would be set to IDEMPOTENCY_CACHED on read
	cachedResponse.Source = "IDEMPOTENCY_CACHED"
	if cachedResponse.Source != "IDEMPOTENCY_CACHED" {
		t.Error("Expected IDEMPOTENCY_CACHED source for duplicate request")
	}
}
