package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetAllDeviceStatusesUseCase handles retrieving all device statuses
type GetAllDeviceStatusesUseCase struct {
	repository DeviceStatusRepository
}

// NewGetAllDeviceStatusesUseCase creates a new instance of GetAllDeviceStatusesUseCase
func NewGetAllDeviceStatusesUseCase(repository DeviceStatusRepository) *GetAllDeviceStatusesUseCase {
	return &GetAllDeviceStatusesUseCase{
		repository: repository,
	}
}

// Execute retrieves all device statuses
func (uc *GetAllDeviceStatusesUseCase) Execute() (*dtos.DeviceStatusListResponseDTO, error) {
	statuses, err := uc.repository.GetAll()
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	var statusDTOs []dtos.DeviceStatusResponseDTO
	for _, status := range statuses {
		statusDTOs = append(statusDTOs, dtos.DeviceStatusResponseDTO{
			ID:        status.ID,
			DeviceID:  status.DeviceID,
			Name:      status.Name,
			Code:      status.Code,
			Value:     status.Value,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.UpdatedAt,
		})
	}

	// Ensure empty slice is not nil
	if statusDTOs == nil {
		statusDTOs = []dtos.DeviceStatusResponseDTO{}
	}

	return &dtos.DeviceStatusListResponseDTO{
		Statuses: statusDTOs,
		Total:    len(statusDTOs),
	}, nil
}
