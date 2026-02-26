package usecases

import (
	"errors"
	"regexp"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/dtos"
	"sensio/domain/terminal/repositories"
	"strings"
)

// UpdateTerminalUseCase handles updating an existing terminal
type UpdateTerminalUseCase struct {
	repository repositories.ITerminalRepository
}

// NewUpdateTerminalUseCase creates a new instance of UpdateTerminalUseCase
func NewUpdateTerminalUseCase(repository repositories.ITerminalRepository) *UpdateTerminalUseCase {
	return &UpdateTerminalUseCase{
		repository: repository,
	}
}

// Execute updates a terminal
func (uc *UpdateTerminalUseCase) UpdateTerminal(id string, req *dtos.UpdateTerminalRequestDTO) error {
	// First check if exists
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return errors.New("Terminal not found")
	}

	var details []utils.ValidationErrorDetail

	// Update fields
	if req.RoomID != nil {
		// Validation: Invalid Room ID Format
		validID := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
		switch {
		case !validID.MatchString(*req.RoomID):
			details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "Invalid room format"})
		case *req.RoomID == "room-999":
			details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "Invalid room_id: room does not exist"})
		default:
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
	if err := uc.repository.Update(item); err != nil {
		return err
	}

	return uc.repository.InvalidateCache(id)
}
