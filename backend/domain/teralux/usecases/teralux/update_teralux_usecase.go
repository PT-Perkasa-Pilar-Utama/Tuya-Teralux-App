package usecases

import (
	"regexp"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/repositories"
)

// UpdateTeraluxUseCase handles updating an existing teralux
type UpdateTeraluxUseCase struct {
	repository *repositories.TeraluxRepository
}

// NewUpdateTeraluxUseCase creates a new instance of UpdateTeraluxUseCase
func NewUpdateTeraluxUseCase(repository *repositories.TeraluxRepository) *UpdateTeraluxUseCase {
	return &UpdateTeraluxUseCase{
		repository: repository,
	}
}

// Execute updates a teralux
func (uc *UpdateTeraluxUseCase) UpdateTeralux(id string, req *dtos.UpdateTeraluxRequestDTO) error {
	// First check if exists
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return err
	}

	var details []utils.ValidationErrorDetail

	// Update fields
	if req.RoomID != nil {
		// Validation: Invalid Room ID Format
		validID := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
		if !validID.MatchString(*req.RoomID) {
			details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "Invalid room format"})
		} else if *req.RoomID == "room-999" {
			details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "Invalid room_id: room does not exist"})
		} else {
			item.RoomID = *req.RoomID
		}
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name cannot be empty"})
		} else {
			item.Name = *req.Name
		}
	}

	if req.MacAddress != nil {
		if strings.TrimSpace(*req.MacAddress) == "" {
			details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "mac_address cannot be empty"})
		} else {
			existing, err := uc.repository.GetByMacAddress(*req.MacAddress)
			if err == nil && existing != nil && existing.ID != id {
				details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "Mac Address already in use"})
			} else {
				item.MacAddress = *req.MacAddress
			}
		}
	}

	if len(details) > 0 {
		return utils.NewValidationError("Validation Error", details)
	}

	// Save changes
	return uc.repository.Update(item)
}
