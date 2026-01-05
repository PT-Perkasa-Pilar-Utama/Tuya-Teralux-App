package usecases

import (
	"errors"
	"regexp"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// GetTeraluxByIDUseCase handles retrieving a single teralux
type GetTeraluxByIDUseCase struct {
	repository *repositories.TeraluxRepository
	devRepo    *repositories.DeviceRepository
}

// NewGetTeraluxByIDUseCase creates a new instance of GetTeraluxByIDUseCase
func NewGetTeraluxByIDUseCase(repository *repositories.TeraluxRepository, devRepo *repositories.DeviceRepository) *GetTeraluxByIDUseCase {
	return &GetTeraluxByIDUseCase{
		repository: repository,
		devRepo:    devRepo,
	}
}

// Execute retrieves a teralux by ID with its associated devices
func (uc *GetTeraluxByIDUseCase) Execute(id string) (*dtos.TeraluxSingleResponseDTO, error) {
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return nil, errors.New("Invalid ID format")
	}

	item, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &dtos.TeraluxSingleResponseDTO{
		Teralux: dtos.TeraluxResponseDTO{
			ID:         item.ID,
			MacAddress: item.MacAddress,
			RoomID:     item.RoomID,
			Name:       item.Name,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		},
	}, nil
}
