package usecases

import (
	"regexp"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"

	"github.com/google/uuid"
)

// CreateTeraluxUseCase handles the business logic for creating a new teralux
type CreateTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewCreateTeraluxUseCase creates a new instance of CreateTeraluxUseCase
func NewCreateTeraluxUseCase(repository *repositories.TeraluxRepository) *CreateTeraluxUseCase {
	return &CreateTeraluxUseCase{
		repository: repository,
	}
}

// CreateTeralux creates a new teralux record
func (uc *CreateTeraluxUseCase) CreateTeralux(req *dtos.CreateTeraluxRequestDTO) (*dtos.CreateTeraluxResponseDTO, bool, error) {
	// Validation
	var details []utils.ValidationErrorDetail

	if strings.TrimSpace(req.Name) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name is required"})
	} else if len(req.Name) > 255 {
		details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name must be 255 characters or less"})
	}

	if strings.TrimSpace(req.MacAddress) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "mac_address is required"})
	} else {
		// Basic MAC address validation regex
		matched, _ := regexp.MatchString(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`, req.MacAddress)
		if !matched {
			details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "invalid mac address format"})
		}
	}

	if strings.TrimSpace(req.RoomID) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "room_id is required"})
	}

	if len(details) > 0 {
		return nil, false, utils.NewValidationError("Validation Error", details)
	}

	existing, err := uc.repository.GetByMacAddress(req.MacAddress)
	if err == nil && existing != nil {
		return &dtos.CreateTeraluxResponseDTO{
			TeraluxID: existing.ID,
		}, false, nil
	}

	// Generate UUID for the new teralux
	id := uuid.New().String()

	// Create entity
	teralux := &entities.Teralux{
		ID:         id,
		MacAddress: req.MacAddress,
		RoomID:     req.RoomID,
		Name:       req.Name,
	}

	// Save to database
	if err := uc.repository.Create(teralux); err != nil {
		return nil, false, err
	}

	// Return response DTO with only ID
	return &dtos.CreateTeraluxResponseDTO{
		TeraluxID: teralux.ID,
	}, true, nil
}
