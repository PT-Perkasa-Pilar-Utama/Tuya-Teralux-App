package utilities

import (
	"errors"
	"testing"
)

// MockLLMClient is a mock implementation of LLMClient for testing
type MockLLMClient struct {
	CallModelFunc  func(prompt string, model string) (string, error)
	HealthCheckVal bool
}

func (m *MockLLMClient) CallModel(prompt string, model string) (string, error) {
	if m.CallModelFunc != nil {
		return m.CallModelFunc(prompt, model)
	}
	return "mock response", nil
}

func (m *MockLLMClient) HealthCheck() bool {
	return m.HealthCheckVal
}

func TestNewLLMClientWithFallback(t *testing.T) {
	primary := &MockLLMClient{}
	secondary := &MockLLMClient{}
	tertiary := &MockLLMClient{}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	fallback, ok := client.(*LLMClientFallback)
	if !ok {
		t.Fatal("expected *LLMClientFallback type")
	}

	if fallback.primary != primary {
		t.Error("primary client not set correctly")
	}
	if fallback.secondary != secondary {
		t.Error("secondary client not set correctly")
	}
	if fallback.tertiary != tertiary {
		t.Error("tertiary client not set correctly")
	}
}

func TestLLMClientFallback_PrimarySucceeds(t *testing.T) {
	primary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "primary response", nil
		},
		HealthCheckVal: true,
	}
	secondary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			t.Error("secondary should not be called")
			return "", errors.New("should not reach here")
		},
	}
	tertiary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			t.Error("tertiary should not be called")
			return "", errors.New("should not reach here")
		},
	}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)
	result, err := client.CallModel("test prompt", "test-model")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "primary response" {
		t.Errorf("expected 'primary response', got '%s'", result)
	}
}

func TestLLMClientFallback_PrimaryFailsSecondarySucceeds(t *testing.T) {
	primary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("primary failed")
		},
		HealthCheckVal: true,
	}
	secondary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "secondary response", nil
		},
		HealthCheckVal: true,
	}
	tertiary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			t.Error("tertiary should not be called")
			return "", errors.New("should not reach here")
		},
	}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)
	result, err := client.CallModel("test prompt", "test-model")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "secondary response" {
		t.Errorf("expected 'secondary response', got '%s'", result)
	}
}

func TestLLMClientFallback_PrimaryUnhealthySecondarySucceeds(t *testing.T) {
	primary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			t.Error("primary should not be called when unhealthy")
			return "", errors.New("should not reach here")
		},
		HealthCheckVal: false,
	}
	secondary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "secondary response", nil
		},
		HealthCheckVal: true,
	}
	tertiary := &MockLLMClient{}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)
	result, err := client.CallModel("test prompt", "test-model")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "secondary response" {
		t.Errorf("expected 'secondary response', got '%s'", result)
	}
}

func TestLLMClientFallback_AllFailTertiarySucceeds(t *testing.T) {
	primary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("primary failed")
		},
		HealthCheckVal: true,
	}
	secondary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("secondary failed")
		},
		HealthCheckVal: true,
	}
	tertiary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "tertiary response", nil
		},
		HealthCheckVal: true,
	}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)
	result, err := client.CallModel("test prompt", "test-model")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "tertiary response" {
		t.Errorf("expected 'tertiary response', got '%s'", result)
	}
}

func TestLLMClientFallback_AllFail(t *testing.T) {
	primary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("primary failed")
		},
		HealthCheckVal: true,
	}
	secondary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("secondary failed")
		},
		HealthCheckVal: true,
	}
	tertiary := &MockLLMClient{
		CallModelFunc: func(prompt, model string) (string, error) {
			return "", errors.New("tertiary failed")
		},
		HealthCheckVal: true,
	}

	client := NewLLMClientWithFallback(primary, secondary, tertiary)
	_, err := client.CallModel("test prompt", "test-model")

	if err == nil {
		t.Fatal("expected error when all clients fail")
	}
	if err.Error() != "tertiary failed" {
		t.Errorf("expected last error 'tertiary failed', got '%v'", err)
	}
}

func TestLLMClientFallback_NoClients(t *testing.T) {
	client := NewLLMClientWithFallback(nil, nil, nil)
	_, err := client.CallModel("test prompt", "test-model")

	if err == nil {
		t.Fatal("expected error when no clients available")
	}

	expectedError := "all LLM clients failed or unavailable"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%v'", expectedError, err)
	}
}
