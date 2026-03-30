package tuya

import (
	"fmt"
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"time"
)

// PasswordRepositoryClient defines the interface for Tuya API client
type PasswordRepositoryClient interface {
	ExecuteRequest(method, urlPath string, body interface{}) ([]byte, error)
}

// PasswordRepository handles password-related API operations
type PasswordRepository struct {
	client PasswordRepositoryClient
}

// NewPasswordRepository creates a new password repository
func NewPasswordRepository(client PasswordRepositoryClient) *PasswordRepository {
	return &PasswordRepository{client: client}
}

// GenerateDynamic generates a one-time dynamic password (valid 5 minutes)
func (r *PasswordRepository) GenerateDynamic(deviceID string) (*domain.Password, error) {
	urlPath := fmt.Sprintf("/v1.0/devices/%s/door-lock/dynamic-password", deviceID)

	respBody, err := r.client.ExecuteRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	apiResp, err := ParseResponse(respBody)
	if err != nil {
		return nil, err
	}

	if err := apiResp.CheckError(); err != nil {
		return nil, err
	}

	result, ok := apiResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	password := getString(result, "password")
	expireTime := getInt(result, "expire_time")

	return &domain.Password{
		Value:        password,
		Type:         domain.PasswordTypeDynamic,
		ExpireAt:     time.Unix(int64(expireTime)/1000, 0),
		ValidMinutes: 5,
	}, nil
}

// GenerateTemporary generates a temporary password with custom duration
func (r *PasswordRepository) GenerateTemporary(deviceID string, durationMinutes int, customPassword string) (*domain.Password, error) {
	urlPath := fmt.Sprintf("/v1.0/devices/%s/door-lock/temp-password", deviceID)

	body := map[string]interface{}{
		"password_type": 2, // 2 = temporary
		"valid_time":    durationMinutes,
	}

	if customPassword != "" {
		body["password"] = customPassword
	}

	respBody, err := r.client.ExecuteRequest("POST", urlPath, body)
	if err != nil {
		return nil, err
	}

	apiResp, err := ParseResponse(respBody)
	if err != nil {
		return nil, err
	}

	if err := apiResp.CheckError(); err != nil {
		return nil, err
	}

	result, ok := apiResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	password := getString(result, "password")
	expireTime := getInt(result, "expire_time")

	return &domain.Password{
		Value:        password,
		Type:         domain.PasswordTypeTemporary,
		ExpireAt:     time.Unix(int64(expireTime)/1000, 0),
		ValidMinutes: durationMinutes,
	}, nil
}

// Generate generates a password based on request type
func (r *PasswordRepository) Generate(req *domain.PasswordRequest) (*domain.Password, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	switch req.Type {
	case domain.PasswordTypeDynamic:
		return r.GenerateDynamic(req.DeviceID)
	case domain.PasswordTypeTemporary:
		return r.GenerateTemporary(req.DeviceID, req.Duration, req.CustomValue)
	default:
		return nil, &domain.ValidationError{"invalid password type"}
	}
}
