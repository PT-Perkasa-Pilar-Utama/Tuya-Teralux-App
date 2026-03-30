package tuya

import (
	"encoding/json"
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"testing"
	"time"
)

// passwordRepositoryClient is an interface for testing
type passwordRepositoryClient interface {
	ExecuteRequest(method, urlPath string, body interface{}) ([]byte, error)
}

// MockClient for testing
type MockClient struct {
	ResponseToReturn []byte
	ErrorToReturn    error
	LastMethod       string
	LastURLPath      string
	LastBody         interface{}
}

func (m *MockClient) ExecuteRequest(method, urlPath string, body interface{}) ([]byte, error) {
	m.LastMethod = method
	m.LastURLPath = urlPath
	m.LastBody = body
	return m.ResponseToReturn, m.ErrorToReturn
}

func TestPasswordRepository_GenerateDynamic_Success(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": {
				"password": "123456",
				"expire_time": 1711881600000
			},
			"t": 1711881300000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	password, err := repo.GenerateDynamic("test-device-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "123456" {
		t.Errorf("expected password '123456', got '%s'", password.Value)
	}

	if password.Type != domain.PasswordTypeDynamic {
		t.Errorf("expected type 'dynamic', got '%s'", password.Type)
	}

	if password.ValidMinutes != 5 {
		t.Errorf("expected validMinutes 5, got %d", password.ValidMinutes)
	}

	expectedExpire := time.Unix(1711881600, 0)
	if !password.ExpireAt.Equal(expectedExpire) {
		t.Errorf("expected expireAt %v, got %v", expectedExpire, password.ExpireAt)
	}

	if mockClient.LastMethod != "GET" {
		t.Errorf("expected GET request, got %s", mockClient.LastMethod)
	}

	expectedURL := "/v1.0/devices/test-device-id/door-lock/dynamic-password"
	if mockClient.LastURLPath != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, mockClient.LastURLPath)
	}
}

func TestPasswordRepository_GenerateDynamic_ErrorResponse(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": false,
			"code": 28841101,
			"msg": "No permissions. This API is not subscribed.",
			"result": {},
			"t": 1711881300000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	_, err := repo.GenerateDynamic("test-device-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedError := "No permissions. This API is not subscribed. (code: 28841101)"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPasswordRepository_GenerateDynamic_NetworkError(t *testing.T) {
	mockClient := &MockClient{
		ErrorToReturn: &testError{message: "network timeout"},
	}
	repo := NewPasswordRepository(mockClient)

	_, err := repo.GenerateDynamic("test-device-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "network timeout" {
		t.Errorf("expected 'network timeout', got '%s'", err.Error())
	}
}

func TestPasswordRepository_GenerateDynamic_InvalidResultFormat(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": "invalid_string_result",
			"t": 1711881300000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	_, err := repo.GenerateDynamic("test-device-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "invalid result format" {
		t.Errorf("expected 'invalid result format', got '%s'", err.Error())
	}
}

func TestPasswordRepository_GenerateTemporary_Success(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": {
				"password": "654321",
				"expire_time": 1711885200000
			},
			"t": 1711881600000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	password, err := repo.GenerateTemporary("test-device-id", 60, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "654321" {
		t.Errorf("expected password '654321', got '%s'", password.Value)
	}

	if password.Type != domain.PasswordTypeTemporary {
		t.Errorf("expected type 'temporary', got '%s'", password.Type)
	}

	if password.ValidMinutes != 60 {
		t.Errorf("expected validMinutes 60, got %d", password.ValidMinutes)
	}

	bodyMap, ok := mockClient.LastBody.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}

	// Check password_type (stored as int in the map)
	if pt, ok := bodyMap["password_type"].(int); !ok || pt != 2 {
		t.Errorf("expected password_type 2, got %v (%T)", bodyMap["password_type"], bodyMap["password_type"])
	}

	// Check valid_time (stored as int in the map)
	if vt, ok := bodyMap["valid_time"].(int); !ok || vt != 60 {
		t.Errorf("expected valid_time 60, got %v (%T)", bodyMap["valid_time"], bodyMap["valid_time"])
	}

	if _, exists := bodyMap["password"]; exists {
		t.Error("expected no custom password in request")
	}
}

func TestPasswordRepository_GenerateTemporary_WithCustomPassword(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": {
				"password": "999999",
				"expire_time": 1711888800000
			},
			"t": 1711881600000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	password, err := repo.GenerateTemporary("test-device-id", 120, "999999")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.Value != "999999" {
		t.Errorf("expected password '999999', got '%s'", password.Value)
	}

	bodyMap, ok := mockClient.LastBody.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}

	if bodyMap["password"] != "999999" {
		t.Errorf("expected custom password '999999', got %v", bodyMap["password"])
	}
}

func TestPasswordRepository_GenerateTemporary_LongDuration(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": {
				"password": "111222",
				"expire_time": 1743417600000
			},
			"t": 1711881600000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	password, err := repo.GenerateTemporary("test-device-id", 525600, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password.ValidMinutes != 525600 {
		t.Errorf("expected validMinutes 525600 (1 year), got %d", password.ValidMinutes)
	}

	bodyMap, ok := mockClient.LastBody.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}

	// Check valid_time (stored as int in the map)
	if vt, ok := bodyMap["valid_time"].(int); !ok || vt != 525600 {
		t.Errorf("expected valid_time 525600, got %v (%T)", bodyMap["valid_time"], bodyMap["valid_time"])
	}
}

func TestPasswordRepository_GenerateTemporary_ErrorResponse(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": false,
			"code": 2001,
			"msg": "device is offline",
			"result": {},
			"t": 1711881300000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	_, err := repo.GenerateTemporary("test-device-id", 60, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedError := "device is offline (code: 2001)"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPasswordRepository_Generate_DelegatesToCorrectMethod(t *testing.T) {
	mockClient := &MockClient{
		ResponseToReturn: []byte(`{
			"success": true,
			"code": 0,
			"msg": "",
			"result": {
				"password": "123456",
				"expire_time": 1711881600000
			},
			"t": 1711881300000
		}`),
	}
	repo := NewPasswordRepository(mockClient)

	tests := []struct {
		name           string
		requestType    domain.PasswordType
		expectedMethod string
		expectedURL    string
	}{
		{
			name:           "Dynamic password uses GET",
			requestType:    domain.PasswordTypeDynamic,
			expectedMethod: "GET",
			expectedURL:    "/v1.0/devices/test-device/door-lock/dynamic-password",
		},
		{
			name:           "Temporary password uses POST",
			requestType:    domain.PasswordTypeTemporary,
			expectedMethod: "POST",
			expectedURL:    "/v1.0/devices/test-device/door-lock/temp-password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.LastMethod = ""
			mockClient.LastURLPath = ""

			_, err := repo.Generate(&domain.PasswordRequest{
				Type:     tt.requestType,
				DeviceID: "test-device",
				Duration: 60,
			})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if mockClient.LastMethod != tt.expectedMethod {
				t.Errorf("expected method '%s', got '%s'", tt.expectedMethod, mockClient.LastMethod)
			}

			if mockClient.LastURLPath != tt.expectedURL {
				t.Errorf("expected URL '%s', got '%s'", tt.expectedURL, mockClient.LastURLPath)
			}
		})
	}
}

func TestPasswordRepository_Generate_InvalidType(t *testing.T) {
	mockClient := &MockClient{}
	repo := NewPasswordRepository(mockClient)

	_, err := repo.Generate(&domain.PasswordRequest{
		Type:     "invalid_type",
		DeviceID: "test-device",
		Duration: 60,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "invalid password type" {
		t.Errorf("expected 'invalid password type', got '%s'", err.Error())
	}
}

func TestParseResponse_Success(t *testing.T) {
	jsonData := `{
		"success": true,
		"code": 0,
		"msg": "",
		"result": {"key": "value"},
		"t": 1711881300000
	}`

	resp, err := ParseResponse([]byte(jsonData))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Success != true {
		t.Errorf("expected Success=true, got %v", resp.Success)
	}

	if resp.Code != 0 {
		t.Errorf("expected Code=0, got %d", resp.Code)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be map[string]interface{}")
	}

	if result["key"] != "value" {
		t.Errorf("expected result.key='value', got %v", result["key"])
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	_, err := ParseResponse([]byte(invalidJSON))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAPIResponse_CheckError(t *testing.T) {
	tests := []struct {
		name        string
		response    *APIResponse
		expectError bool
		errorMsg    string
	}{
		{
			name: "Success response returns no error",
			response: &APIResponse{
				Success: true,
				Code:    0,
				Msg:     "",
			},
			expectError: false,
		},
		{
			name: "Error response returns error with message and code",
			response: &APIResponse{
				Success: false,
				Code:    28841101,
				Msg:     "No permissions",
			},
			expectError: true,
			errorMsg:    "No permissions (code: 28841101)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.CheckError()

			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.expectError && err.Error() != tt.errorMsg {
				t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestPasswordRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *domain.PasswordRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid dynamic password request",
			request: &domain.PasswordRequest{
				Type:     domain.PasswordTypeDynamic,
				DeviceID: "test-device",
			},
			wantErr: false,
		},
		{
			name: "Valid temporary password request",
			request: &domain.PasswordRequest{
				Type:     domain.PasswordTypeTemporary,
				DeviceID: "test-device",
				Duration: 60,
			},
			wantErr: false,
		},
		{
			name: "Missing device ID",
			request: &domain.PasswordRequest{
				Type: domain.PasswordTypeDynamic,
			},
			wantErr: true,
			errMsg:  "device_id is required",
		},
		{
			name: "Temporary password with zero duration",
			request: &domain.PasswordRequest{
				Type:     domain.PasswordTypeTemporary,
				DeviceID: "test-device",
				Duration: 0,
			},
			wantErr: true,
			errMsg:  "duration must be positive",
		},
		{
			name: "Temporary password with negative duration",
			request: &domain.PasswordRequest{
				Type:     domain.PasswordTypeTemporary,
				DeviceID: "test-device",
				Duration: -10,
			},
			wantErr: true,
			errMsg:  "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("expected error '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestPasswordRequest_JSONSerialization(t *testing.T) {
	req := &domain.PasswordRequest{
		Type:        domain.PasswordTypeTemporary,
		DeviceID:    "test-device-123",
		Duration:    120,
		CustomValue: "123456",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled domain.PasswordRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Type != req.Type {
		t.Errorf("expected Type '%s', got '%s'", req.Type, unmarshaled.Type)
	}

	if unmarshaled.DeviceID != req.DeviceID {
		t.Errorf("expected DeviceID '%s', got '%s'", req.DeviceID, unmarshaled.DeviceID)
	}

	if unmarshaled.Duration != req.Duration {
		t.Errorf("expected Duration %d, got %d", req.Duration, unmarshaled.Duration)
	}

	if unmarshaled.CustomValue != req.CustomValue {
		t.Errorf("expected CustomValue '%s', got '%s'", req.CustomValue, unmarshaled.CustomValue)
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
