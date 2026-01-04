package usecases

import (
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetTeraluxByMACUseCase handles the business logic for retrieving a teralux by MAC address
type GetTeraluxByMACUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewGetTeraluxByMACUseCase creates a new instance of GetTeraluxByMACUseCase
func NewGetTeraluxByMACUseCase(repository *repositories.TeraluxRepository) *GetTeraluxByMACUseCase {
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
