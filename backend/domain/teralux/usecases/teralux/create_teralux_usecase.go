package usecases

import (
	"fmt"
	"regexp"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/teralux/dtos"
	"teralux_app/domain/teralux/entities"
	"teralux_app/domain/teralux/repositories"
	"teralux_app/domain/teralux/services"

	"strconv"

	"github.com/google/uuid"
)

// CreateTeraluxUseCase handles the business logic for creating a new teralux
type CreateTeraluxUseCase struct {
	repository      *repositories.TeraluxRepository
	externalService *services.TeraluxExternalService
}

// NewCreateTeraluxUseCase creates a new instance of CreateTeraluxUseCase
func NewCreateTeraluxUseCase(
	repository *repositories.TeraluxRepository,
	externalService *services.TeraluxExternalService,
) *CreateTeraluxUseCase {
	return &CreateTeraluxUseCase{
		repository:      repository,
		externalService: externalService,
	}
}

// CreateTeralux creates a new teralux record
func (uc *CreateTeraluxUseCase) CreateTeralux(req *dtos.CreateTeraluxRequestDTO) (*dtos.CreateTeraluxResponseDTO, bool, error) {
	// Normalization
	req.MacAddress = strings.ToUpper(strings.TrimSpace(req.MacAddress))

	// Validation
	var details []utils.ValidationErrorDetail

	if strings.TrimSpace(req.Name) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name is required"})
	} else if len(req.Name) > 255 {
		details = append(details, utils.ValidationErrorDetail{Field: "name", Message: "name must be 255 characters or less"})
	}

	if req.MacAddress == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "mac_address is required"})
	} else {
		// MAC address validation regex:
		// 1. AA:BB:CC:DD:EE:FF or AA-BB-CC-DD-EE-FF
		// 2. 12 or 16 hex digits (raw MAC or Android ID)
		matched, _ := regexp.MatchString(`^([0-9A-F]{2}[:-]){5}([0-9A-F]{2})$|^[0-9A-F]{12}$|^[0-9A-F]{16}$`, req.MacAddress)
		if !matched {
			details = append(details, utils.ValidationErrorDetail{Field: "mac_address", Message: "invalid mac address format"})
		}
	}

	if strings.TrimSpace(req.RoomID) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "room_id", Message: "room_id is required"})
	}

	if strings.TrimSpace(req.DeviceTypeID) == "" {
		details = append(details, utils.ValidationErrorDetail{Field: "device_type_id", Message: "device_type_id is required"})
	}

	if len(details) > 0 {
		return nil, false, utils.NewValidationError("Validation Error", details)
	}

	// Convert string IDs to int for external service
	roomIDInt, err := strconv.Atoi(req.RoomID)
	if err != nil {
		return nil, false, utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
			{Field: "room_id", Message: "room_id must be a numeric string"},
		})
	}

	deviceTypeIDInt, err := strconv.Atoi(req.DeviceTypeID)
	if err != nil {
		return nil, false, utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
			{Field: "device_type_id", Message: "device_type_id must be a numeric string"},
		})
	}

	// Check existing before everything (normalization ensures case-insensitive check)
	existing, err := uc.repository.GetByMacAddress(req.MacAddress)
	if err == nil && existing != nil {
		utils.LogDebug("CreateTeraluxUseCase: MAC conflict detected for %s", req.MacAddress)
		return nil, false, utils.NewAPIError(409, "Mac Address already in use")
	}

	// Call external service before saving to DB
	utils.LogDebug("CreateTeraluxUseCase: Attempting external registration for MAC: %s", req.MacAddress)
	if err := uc.externalService.ProcInsertMacAddress(roomIDInt, req.MacAddress, deviceTypeIDInt); err != nil {
		utils.LogError("CreateTeraluxUseCase: External registration failed for MAC %s: %v", req.MacAddress, err)
		return nil, false, fmt.Errorf("failed to register device externally: %w", err)
	}
	utils.LogDebug("CreateTeraluxUseCase: External registration success for MAC: %s", req.MacAddress)

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
		utils.LogError("CreateTeraluxUseCase: DB save failed for MAC %s: %v", req.MacAddress, err)
		return nil, false, err
	}

	utils.LogDebug("CreateTeraluxUseCase: Successfully created Teralux ID %s for MAC %s", id, req.MacAddress)

	// Return response DTO with only ID
	return &dtos.CreateTeraluxResponseDTO{
		TeraluxID: teralux.ID,
	}, true, nil
}
