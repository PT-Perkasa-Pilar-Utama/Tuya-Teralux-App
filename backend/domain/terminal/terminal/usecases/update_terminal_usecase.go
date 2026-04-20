package usecases

import (
	"errors"
	"regexp"
	"sensio/domain/common/providers"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
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

// Execute updates a terminal and returns the updated DTO
func (uc *UpdateTerminalUseCase) UpdateTerminal(id string, req *dtos.UpdateTerminalRequestDTO) (*dtos.TerminalResponseDTO, error) {
	// First check if exists
	item, err := uc.repository.GetByID(id)
	if err != nil {
		return nil, errors.New("Terminal not found")
	}
	if item == nil {
		return nil, errors.New("Terminal not found")
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
	if req.DeviceTypeID != nil {
		item.DeviceTypeID = *req.DeviceTypeID
	}

	if req.AiProvider != nil {
		// Validate ai_provider if provided
		if *req.AiProvider == "" {
			// Empty string means clear the preference, set to nil
			item.AiProvider = nil
		} else {
			normalizedProvider := providers.NormalizeProvider(*req.AiProvider)
			// Validate provider - only remote providers are supported
			if !providers.IsValidProvider(normalizedProvider) {
				details = append(details, utils.ValidationErrorDetail{
					Field:   "ai_provider",
					Message: "Invalid ai_provider. Supported values: gemini, openai, groq, orion",
				})
			} else {
				item.AiProvider = &normalizedProvider
			}
		}
	}

	if len(details) > 0 {
		return nil, utils.NewValidationError("Validation Error", details)
	}

	// Save changes
	if err := uc.repository.Update(item); err != nil {
		return nil, err
	}

	if err := uc.repository.InvalidateCache(id); err != nil {
		return nil, err
	}

	// Convert entity to DTO
	dto := &dtos.TerminalResponseDTO{
		ID:           item.ID,
		MacAddress:   item.MacAddress,
		RoomID:       item.RoomID,
		Name:         item.Name,
		DeviceTypeID: item.DeviceTypeID,
		AiProvider:   item.AiProvider,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}

	return dto, nil
}
