package usecases

import (
	"errors"
	"regexp"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/repositories"
	"sensio/domain/terminal/terminal/services"
)

// DeleteTerminalUseCase handles deleting a terminal
type DeleteTerminalUseCase struct {
	repository repositories.ITerminalRepository
	mqttClient *services.MqttAuthClient
}

// NewDeleteTerminalUseCase creates a new instance of DeleteTerminalUseCase
func NewDeleteTerminalUseCase(
	repository repositories.ITerminalRepository,
	mqttClient *services.MqttAuthClient,
) *DeleteTerminalUseCase {
	return &DeleteTerminalUseCase{
		repository: repository,
		mqttClient: mqttClient,
	}
}

func (uc *DeleteTerminalUseCase) DeleteTerminal(id string) error {
	validID := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validID.MatchString(id) {
		return errors.New("Invalid ID format")
	}

	// Check existence
	terminal, err := uc.repository.GetByID(id)
	if err != nil || terminal == nil {
		return errors.New("Terminal not found")
	}

	// Delete MQTT User if MacAddress is available
	if terminal.MacAddress != "" && uc.mqttClient != nil {
		if err := uc.mqttClient.DeleteMQTTUser(terminal.MacAddress); err != nil {
			utils.LogError("DeleteTerminalUseCase: Failed to delete MQTT user for MAC %s: %v", terminal.MacAddress, err)
			// We continue even if MQTT deletion fails to avoid leaving ghost terminals in DB
		} else {
			utils.LogInfo("DeleteTerminalUseCase: Successfully deleted MQTT user for MAC %s", terminal.MacAddress)
		}
	}

	if err := uc.repository.Delete(id); err != nil {
		return err
	}

	return uc.repository.InvalidateCache(id)
}
