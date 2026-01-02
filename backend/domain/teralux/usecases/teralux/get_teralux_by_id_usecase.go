package usecases

import (
	"teralux_app/domain/teralux/dtos"
)

// GetTeraluxByIDUseCase handles retrieving a single teralux
type GetTeraluxByIDUseCase struct {
	repository TeraluxRepository
}

// NewGetTeraluxByIDUseCase creates a new instance of GetTeraluxByIDUseCase
func NewGetTeraluxByIDUseCase(repository TeraluxRepository) *GetTeraluxByIDUseCase {
	return &GetTeraluxByIDUseCase{
		repository: repository,
	}
}

// Execute retrieves a teralux by ID
func (uc *GetTeraluxByIDUseCase) Execute(id string) (*dtos.TeraluxResponseDTO, error) {
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &dtos.TeraluxResponseDTO{
		ID:         item.ID,
		MacAddress: item.MacAddress,
		Name:       item.Name,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}, nil
}
