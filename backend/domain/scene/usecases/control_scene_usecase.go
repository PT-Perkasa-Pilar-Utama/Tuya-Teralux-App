package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/scene/repositories"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

// TuyaDeviceControlExecutor matches the interface in Teralux domain for consistency
type TuyaDeviceControlExecutor interface {
	SendCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error)
}

// MqttPublisher defines the interface for MQTT publishing
type MqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

type ControlSceneUseCase struct {
	repo    *repositories.SceneRepository
	tuyaCmd TuyaDeviceControlExecutor
	mqttSvc MqttPublisher
}

func NewControlSceneUseCase(repo *repositories.SceneRepository, tuyaCmd TuyaDeviceControlExecutor, mqttSvc MqttPublisher) *ControlSceneUseCase {
	return &ControlSceneUseCase{
		repo:    repo,
		tuyaCmd: tuyaCmd,
		mqttSvc: mqttSvc,
	}
}

func (uc *ControlSceneUseCase) Execute(id string, accessToken string) error {
	scene, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("scene not found: %w", err)
	}

	for _, action := range scene.Actions {
		if action.Topic != "" {
			// MQTT Action
			if uc.mqttSvc != nil {
				if err := uc.mqttSvc.Publish(action.Topic, 0, false, action.Value); err != nil {
					utils.LogError("Scene %s: Failed to publish to MQTT topic %s: %v", id, action.Topic, err)
				}
			}
		} else if action.DeviceID != "" {
			// Tuya Action
			if action.RemoteID != "" {
				// IR Command
				valInt, ok := utils.ToInt(action.Value)
				if !ok {
					utils.LogWarn("Scene %s: Invalid value for IR command on device %s: %v", id, action.DeviceID, action.Value)
					continue
				}
				_, err := uc.tuyaCmd.SendIRACCommand(accessToken, action.DeviceID, action.RemoteID, action.Code, valInt)
				if err != nil {
					utils.LogError("Scene %s: Failed to send IR command to %s: %v", id, action.DeviceID, err)
				}
			} else {
				// Standard Command
				cmd := tuya_dtos.TuyaCommandDTO{
					Code:  action.Code,
					Value: action.Value,
				}
				_, err := uc.tuyaCmd.SendCommand(accessToken, action.DeviceID, []tuya_dtos.TuyaCommandDTO{cmd})
				if err != nil {
					utils.LogError("Scene %s: Failed to send command to %s: %v", id, action.DeviceID, err)
				}
			}
		}
	}

	return nil
}
