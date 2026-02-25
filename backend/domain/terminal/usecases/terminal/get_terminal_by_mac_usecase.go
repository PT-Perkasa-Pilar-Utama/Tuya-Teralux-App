package usecases

import (
	"errors"
	"regexp"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/repositories"
)

// GetTerminalByMACUseCase handles the business logic for retrieving a terminal by MAC address
type GetTerminalByMACUseCase struct {
	repository repositories.ITerminalRepository
}

// NewGetTerminalByMACUseCase creates a new instance of GetTerminalByMACUseCase
func NewGetTerminalByMACUseCase(repository repositories.ITerminalRepository) *GetTerminalByMACUseCase {
	return &GetTerminalByMACUseCase{
		repository: repository,
	}
}

func (uc *GetTerminalByMACUseCase) GetTerminalByMAC(macAddress string) (*dtos.TerminalSingleResponseDTO, error) {
	// Allow standard MAC (AA:BB:CC:DD:EE:FF) OR Android ID (16 hex chars)
	// Regex: ^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$  <-- MAC
	//        ^[0-9a-fA-F]{16}$                     <-- Android ID
	validMAC := regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$|^[0-9a-fA-F]{16}$`)

	if !validMAC.MatchString(macAddress) {
		return nil, errors.New("invalid mac address or device id format")
	}

	terminal, err := uc.repository.GetByMacAddress(macAddress)
	if err != nil {
		return nil, errors.New("Terminal not found")
	}

	return &dtos.TerminalSingleResponseDTO{
		Terminal: dtos.TerminalResponseDTO{
			ID:         terminal.ID,
			MacAddress: terminal.MacAddress,
			RoomID:     terminal.RoomID,
			Name:       terminal.Name,
			CreatedAt:  terminal.CreatedAt,
			UpdatedAt:  terminal.UpdatedAt,
		},
	}, nil
}
