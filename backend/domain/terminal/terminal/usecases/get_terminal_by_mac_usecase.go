package usecases

import (
	"errors"
	"regexp"
	"strings"

	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
	"sensio/domain/terminal/terminal/services"

	"github.com/google/uuid"
)

// GetTerminalByMACUseCase handles the business logic for retrieving a terminal by MAC address
type GetTerminalByMACUseCase struct {
	repository repositories.ITerminalRepository
	mqttClient *services.MqttAuthClient
}

// NewGetTerminalByMACUseCase creates a new instance of GetTerminalByMACUseCase
func NewGetTerminalByMACUseCase(repository repositories.ITerminalRepository, mqttClient *services.MqttAuthClient) *GetTerminalByMACUseCase {
	return &GetTerminalByMACUseCase{
		repository: repository,
		mqttClient: mqttClient,
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

	// Fetch MQTT credentials from Rust Auth Service
	mqttUsername := terminal.MacAddress
	mqttPassword := ""

	creds, err := uc.mqttClient.GetMQTTCredentials(mqttUsername)
	if err != nil {
		utils.LogError("GetTerminalByMACUseCase: Failed to fetch MQTT credentials for %s: %v", mqttUsername, err)
		return nil, errors.New("failed to fetch mqtt credentials")
	}

	if creds == nil {
		// Scenario B: Terminal exists but MQTT user is missing — auth service creates it
		utils.LogDebug("GetTerminalByMACUseCase: MQTT user missing for %s, requesting creation (Scenario B)", mqttUsername)
		newPassword := strings.ReplaceAll(uuid.New().String(), "-", "")
		alreadyExists, createErr := uc.mqttClient.CreateMQTTUser(mqttUsername, newPassword)
		if createErr != nil {
			utils.LogError("GetTerminalByMACUseCase: Failed to create MQTT user for %s: %v", mqttUsername, createErr)
			return nil, errors.New("failed to create associated mqtt user")
		}
		if !alreadyExists {
			mqttPassword = newPassword
			utils.LogDebug("GetTerminalByMACUseCase: Created missing MQTT user (Scenario B) for %s", mqttUsername)
		}
	} else {
		mqttPassword = creds.Password
	}

	return &dtos.TerminalSingleResponseDTO{
		Terminal: dtos.TerminalResponseDTO{
			ID:           terminal.ID,
			MacAddress:   terminal.MacAddress,
			RoomID:       terminal.RoomID,
			Name:         terminal.Name,
			DeviceTypeID: terminal.DeviceTypeID,
			AiProvider:   terminal.AiProvider,
			MQTTUsername: mqttUsername,
			MQTTPassword: mqttPassword,
			CreatedAt:    terminal.CreatedAt,
			UpdatedAt:    terminal.UpdatedAt,
		},
	}, nil
}
