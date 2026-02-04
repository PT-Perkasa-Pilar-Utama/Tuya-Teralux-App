package repositories

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"teralux_app/domain/common/utils"
	"testing"
)

func TestAntigravityRepository_CallModel(t *testing.T) {
	// 1. Setup a Mock Server that mimics the Local API Proxy (OpenAI format)
	expectedResponse := "Hello from Antigravity"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Decode Body
		var req OpenAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if req.Model != "gemini-2.5-flash" {
			t.Errorf("Expected model gemini-2.5-flash, got %s", req.Model)
		}
		if len(req.Messages) == 0 || req.Messages[0].Content != "Testing" {
			t.Errorf("Expected message 'Testing', got %v", req.Messages)
		}

		// Send Response
		resp := OpenAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: expectedResponse,
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 2. Setup Config to point to Mock Server
	// We need to overwrite the global AppConfig or mock GetConfig (which reads global).
	// Since GetConfig() uses global AppConfig, we can just modify it if checks allow.
	// But it's safer to rely on env vars + LoadConfig if possible, OR just manually set it if exposed.
	// However, usually Config is loaded from Env.
	os.Setenv("LLM_PROVIDER", "antigravity")
	os.Setenv("LLM_BASE_URL", mockServer.URL+"/v1") // mockServer.URL is http://127.0.0.1:xxxxx
	os.Setenv("LLM_API_KEY", "test-key")
	os.Setenv("LLM_MODEL", "gemini-2.5-flash")

	// Helper to reset config
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("LLM_BASE_URL")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("LLM_MODEL")
		utils.AppConfig = nil // reset singleton
	}()

	// Force reload config
	utils.AppConfig = nil
	utils.LoadConfig()

	// 3. Initialize Repository
	repo := NewAntigravityRepository()

	// 4. Test CallModel
	response, err := repo.CallModel("Testing", "gemini-2.5-flash")
	if err != nil {
		t.Fatalf("CallModel failed: %v", err)
	}

	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}
}
