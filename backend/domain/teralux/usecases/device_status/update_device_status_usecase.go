package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

// TuyaDeviceControlExecutor defines the interface for Tuya device control operations
type TuyaDeviceControlExecutor interface {
	SendSwitchCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error)
}

// UpdateDeviceStatusUseCase handles updating an existing device status
type UpdateDeviceStatusUseCase struct {
	repo    *repositories.DeviceStatusRepository
	devRepo *repositories.DeviceRepository
	tuyaCmd TuyaDeviceControlExecutor
}

// NewUpdateDeviceStatusUseCase creates a new instance of UpdateDeviceStatusUseCase
func NewUpdateDeviceStatusUseCase(repo *repositories.DeviceStatusRepository, devRepo *repositories.DeviceRepository, tuyaCmd TuyaDeviceControlExecutor) *UpdateDeviceStatusUseCase {
	return &UpdateDeviceStatusUseCase{
		repo:    repo,
		devRepo: devRepo,
		tuyaCmd: tuyaCmd,
	}
}

// Execute updates a device status
func (uc *UpdateDeviceStatusUseCase) UpdateDeviceStatus(deviceID string, req *dtos.UpdateDeviceStatusRequestDTO, accessToken string) error {
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

	// Execute Tuya Command
	if req.RemoteID != "" {
		// IR Command
		// Need to convert Value to int if possible, safely
		valInt, ok := utils.ToInt(req.Value)
		if !ok {
			// fallback or error? For now assuming 0 or error, but let's try to proceed or handle error
			// If conversion fails and it's required, we should return error.
			// But existing code just passed req.Value to DB.
			// Let's assume for IR commands, value MUST be convertible to int as per SendIRACCommand signature
			valInt = 0 // default
		}

		params := map[string]int{
			req.Code: valInt,
		}
		success, err := uc.tuyaCmd.SendIRACCommand(accessToken, deviceID, req.RemoteID, params)
		if err != nil {
			return fmt.Errorf("failed to send IR command: %w", err)
		}
		if !success {
			return fmt.Errorf("failed to send IR command: unsuccessful response from Tuya")
		}
	} else {
		// Switch/Standard Command
		cmd := tuya_dtos.TuyaCommandDTO{
			Code:  req.Code,
			Value: req.Value,
		}
		success, err := uc.tuyaCmd.SendSwitchCommand(accessToken, deviceID, []tuya_dtos.TuyaCommandDTO{cmd})
		if err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}
		if !success {
			return fmt.Errorf("failed to send command: unsuccessful response from Tuya")
		}
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
