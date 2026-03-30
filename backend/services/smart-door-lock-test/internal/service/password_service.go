package service

import (
	"sensio/backend/services/smart-door-lock-test/internal/domain"
)

// PasswordRepository defines the interface for password operations
type PasswordRepository interface {
	Generate(req *domain.PasswordRequest) (*domain.Password, error)
}

// PasswordService handles password-related business logic
type PasswordService struct {
	passwordRepo PasswordRepository
}

// NewPasswordService creates a new password service
func NewPasswordService(passwordRepo PasswordRepository) *PasswordService {
	return &PasswordService{passwordRepo: passwordRepo}
}

// GenerateDynamicPassword generates a one-time dynamic password (valid 5 minutes)
func (s *PasswordService) GenerateDynamicPassword(deviceID string) (*domain.Password, error) {
	req := &domain.PasswordRequest{
		Type:     domain.PasswordTypeDynamic,
		DeviceID: deviceID,
	}

	return s.passwordRepo.Generate(req)
}

// GenerateTemporaryPassword generates a temporary password with custom duration
func (s *PasswordService) GenerateTemporaryPassword(deviceID string, durationMinutes int, customPassword string) (*domain.Password, error) {
	req := &domain.PasswordRequest{
		Type:        domain.PasswordTypeTemporary,
		DeviceID:    deviceID,
		Duration:    durationMinutes,
		CustomValue: customPassword,
	}

	return s.passwordRepo.Generate(req)
}

// ValidatePassword checks if a password is still valid
func (s *PasswordService) ValidatePassword(password *domain.Password) error {
	if !password.IsValid() {
		return domain.ErrPasswordExpired
	}
	return nil
}
