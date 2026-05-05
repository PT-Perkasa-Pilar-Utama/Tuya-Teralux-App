package usecases

import (
	"errors"
	"regexp"

	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
	"sensio/domain/terminal/terminal/services"

	"github.com/google/uuid"
)

type GetOrCreateMQTTCredentialsUseCase struct {
	repository repositories.ITerminalRepository
	mqttClient services.IMqttAuthClient
}

func NewGetOrCreateMQTTCredentialsUseCase(
	repository repositories.ITerminalRepository,
	mqttClient services.IMqttAuthClient,
) *GetOrCreateMQTTCredentialsUseCase {
	return &GetOrCreateMQTTCredentialsUseCase{
		repository: repository,
		mqttClient: mqttClient,
	}
}

func (uc *GetOrCreateMQTTCredentialsUseCase) GetOrCreateMQTTCredentials(macAddress string) (*dtos.MQTTCredentialsResponseDTO, error) {
	validMAC := regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$|^[0-9a-fA-F]{16}$`)

	if !validMAC.MatchString(macAddress) {
		return nil, errors.New("invalid mac address or device id format")
	}

	terminal, err := uc.repository.GetByMacAddress(macAddress)
	if err != nil {
		utils.LogError("GetOrCreateMQTTCredentialsUseCase: Terminal not found for MAC %s: %v", macAddress, err)
		return nil, errors.New("Terminal not found")
	}

	creds, err := uc.mqttClient.GetMQTTCredentials(terminal.MacAddress)
	if err != nil {
		utils.LogError("GetOrCreateMQTTCredentialsUseCase: Failed to fetch MQTT credentials for %s: %v", terminal.MacAddress, err)
		return nil, errors.New("failed to fetch mqtt credentials")
	}

	if creds == nil {
		newPassword := uuid.New().String()
		newPassword = newPassword[:32]

		alreadyExists, createErr := uc.mqttClient.CreateMQTTUser(terminal.MacAddress, newPassword)
		if createErr != nil {
			utils.LogError("GetOrCreateMQTTCredentialsUseCase: Failed to create MQTT user for %s: %v", terminal.MacAddress, createErr)
			return nil, errors.New("failed to create mqtt user")
		}

		if alreadyExists {
			utils.LogDebug("GetOrCreateMQTTCredentialsUseCase: Race condition - MQTT user already exists for %s, fetching credentials", terminal.MacAddress)
			creds, err = uc.mqttClient.GetMQTTCredentials(terminal.MacAddress)
			if err != nil {
				utils.LogError("GetOrCreateMQTTCredentialsUseCase: Failed to fetch MQTT credentials after 409 for %s: %v", terminal.MacAddress, err)
				return nil, errors.New("failed to fetch mqtt credentials after conflict")
			}
			if creds == nil {
				return nil, errors.New("mqtt credentials not found after conflict resolution")
			}
		} else {
			creds = &services.MQTTCredentials{
				Username: terminal.MacAddress,
				Password: newPassword,
			}
		}
	}

	return &dtos.MQTTCredentialsResponseDTO{
		Username: creds.Username,
		Password: creds.Password,
	}, nil
}
