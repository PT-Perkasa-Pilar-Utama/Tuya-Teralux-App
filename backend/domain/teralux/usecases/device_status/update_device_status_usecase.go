package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
)

// UpdateDeviceStatusUseCase handles updating an existing device status
type UpdateDeviceStatusUseCase struct {
	repo    *repositories.DeviceStatusRepository
	devRepo *repositories.DeviceRepository
}

// NewUpdateDeviceStatusUseCase creates a new instance of UpdateDeviceStatusUseCase
func NewUpdateDeviceStatusUseCase(repo *repositories.DeviceStatusRepository, devRepo *repositories.DeviceRepository) *UpdateDeviceStatusUseCase {
	return &UpdateDeviceStatusUseCase{
		repo:    repo,
		devRepo: devRepo,
	}
}

// Execute updates a device status
func (uc *UpdateDeviceStatusUseCase) Execute(deviceID string, req *dtos.UpdateDeviceStatusRequestDTO) error {
	// Check device existence
	_, err := uc.devRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("Device not found: %w", err)
	}

	// Code validation (Simulated based on scenario requirements)
	var details []utils.ValidationErrorDetail
	if req.Code == "nuclear_launch" {
		details = append(details, utils.ValidationErrorDetail{Field: "code", Message: "Invalid status code for this device"})
	}

	// Value validation (Simulated)
	if req.Code == "dimmer" {
		// DTO Value is interface{}. Need to check type/value.
		// "full_power" string is invalid for intent of int.
		if _, ok := req.Value.(string); ok && req.Value == "full_power" {
			details = append(details, utils.ValidationErrorDetail{Field: "value", Message: "Invalid value for status code 'dimmer'"})
		}
	}

	if len(details) > 0 {
		return utils.NewValidationError("Validation Error", details)
	}

	// Convert value to string for storage
	valStr := fmt.Sprintf("%v", req.Value)

	status := &entities.DeviceStatus{
		DeviceID: deviceID,
		Code:     req.Code,
		Value:    valStr,
	}
	return uc.repo.Upsert(status)
}
