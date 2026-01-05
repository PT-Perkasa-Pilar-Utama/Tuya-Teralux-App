package usecases

import (
	"errors"
	"regexp"
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

func (uc *GetTeraluxByMACUseCase) Execute(macAddress string) (*dtos.TeraluxSingleResponseDTO, error) {
	validMAC := regexp.MustCompile(`^([0-9a-zA-Z]{2}:){5}[0-9a-zA-Z]{2}$`)
	if !validMAC.MatchString(macAddress) {
		return nil, errors.New("Invalid MAC address format")
	}

	teralux, err := uc.repository.GetByMacAddress(macAddress)
	if err != nil {
		return nil, err
	}

	return &dtos.TeraluxSingleResponseDTO{
		Teralux: dtos.TeraluxResponseDTO{
			ID:         teralux.ID,
			MacAddress: teralux.MacAddress,
			RoomID:     teralux.RoomID,
			Name:       teralux.Name,
			CreatedAt:  teralux.CreatedAt,
			UpdatedAt:  teralux.UpdatedAt,
		},
	}, nil
}
