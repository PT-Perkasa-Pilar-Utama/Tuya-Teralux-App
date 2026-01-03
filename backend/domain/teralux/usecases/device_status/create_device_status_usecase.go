package usecases

import (
	"errors"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"

	"gorm.io/gorm"
)

// CreateDeviceStatusUseCase handles the business logic for creating a new device status
type CreateDeviceStatusUseCase struct {
	repository DeviceStatusRepository
}

// NewCreateDeviceStatusUseCase creates a new instance of CreateDeviceStatusUseCase
func NewCreateDeviceStatusUseCase(repository DeviceStatusRepository) *CreateDeviceStatusUseCase {
	return &CreateDeviceStatusUseCase{
		repository: repository,
	}
}

// Execute creates a new device status record
func (uc *CreateDeviceStatusUseCase) Execute(req *dtos.CreateDeviceStatusRequestDTO) (*dtos.CreateDeviceStatusResponseDTO, error) {
	// Check if status with same code exists for device
	existing, err := uc.repository.GetByDeviceIDAndCode(req.DeviceID, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New("device status with this code already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create entity
	status := &entities.DeviceStatus{
		DeviceID: req.DeviceID,
		Code:     req.Code,
		Value:    req.Value,
	}

	// Save to database
	if err := uc.repository.Create(status); err != nil {
		return nil, err
	}

	return &dtos.CreateDeviceStatusResponseDTO{
		DeviceID: status.DeviceID,
		Code:     status.Code,
	}, nil
}
