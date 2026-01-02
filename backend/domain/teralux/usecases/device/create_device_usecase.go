package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"

	"github.com/google/uuid"
)

// CreateDeviceUseCase handles the business logic for creating a new device
type CreateDeviceUseCase struct {
	repository DeviceRepository
}

// NewCreateDeviceUseCase creates a new instance of CreateDeviceUseCase
func NewCreateDeviceUseCase(repository DeviceRepository) *CreateDeviceUseCase {
	return &CreateDeviceUseCase{
		repository: repository,
	}
}

// Execute creates a new device record
func (uc *CreateDeviceUseCase) Execute(req *dtos.CreateDeviceRequestDTO) (*dtos.CreateDeviceResponseDTO, error) {
	// Check if device already exists by TeraluxID
	existingDevices, err := uc.repository.GetByTeraluxID(req.TeraluxID)
	if err == nil && len(existingDevices) > 0 {
		return &dtos.CreateDeviceResponseDTO{
			ID: existingDevices[0].ID,
		}, nil
	}

	// Generate UUID for the new device
	id := uuid.New().String()

	// Create entity
	device := &entities.Device{
		ID:        id,
		TeraluxID: req.TeraluxID,
		Name:      req.Name,
	}

	// Save to database
	if err := uc.repository.Create(device); err != nil {
		return nil, err
	}

	// Return response DTO with only ID
	return &dtos.CreateDeviceResponseDTO{
		ID: device.ID,
	}, nil
}
