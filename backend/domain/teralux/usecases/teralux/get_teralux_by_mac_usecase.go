package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetTeraluxByMACUseCase handles the business logic for retrieving a teralux by MAC address
type GetTeraluxByMACUseCase struct {
	repository TeraluxRepository
}

// NewGetTeraluxByMACUseCase creates a new instance of GetTeraluxByMACUseCase
func NewGetTeraluxByMACUseCase(repository TeraluxRepository) *GetTeraluxByMACUseCase {
	return &GetTeraluxByMACUseCase{
		repository: repository,
	}
}

// Execute retrieves a teralux record by MAC address
func (uc *GetTeraluxByMACUseCase) Execute(macAddress string) (*dtos.TeraluxResponseDTO, error) {
	teralux, err := uc.repository.GetByMacAddress(macAddress)
	if err != nil {
		return nil, err
	}

	return &dtos.TeraluxResponseDTO{
		ID:         teralux.ID,
		MacAddress: teralux.MacAddress,
		Name:       teralux.Name,
		CreatedAt:  teralux.CreatedAt,
		UpdatedAt:  teralux.UpdatedAt,
	}, nil
}
