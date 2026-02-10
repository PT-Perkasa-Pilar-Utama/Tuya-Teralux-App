package usecases

import (
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/scene/repositories"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

// TuyaDeviceControlExecutor defines the interface for controlling Tuya devices
type TuyaDeviceControlExecutor interface {
	SendCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error)
}

type ControlSceneUseCase struct {
	repo    *repositories.SceneRepository
	tuyaCmd TuyaDeviceControlExecutor
	mqttSvc *infrastructure.MqttService
}

func NewControlSceneUseCase(
	repo *repositories.SceneRepository,
	tuyaCmd TuyaDeviceControlExecutor,
	mqttSvc *infrastructure.MqttService,
) *ControlSceneUseCase {
	return &ControlSceneUseCase{
		repo:    repo,
		tuyaCmd: tuyaCmd,
		mqttSvc: mqttSvc,
	}
}

func (u *ControlSceneUseCase) Execute(teraluxID, id, accessToken string) error {
	scene, err := u.repo.GetByID(teraluxID, id)
	if err != nil {
		return err
	}

	var errs []error
	for _, action := range scene.Actions {
		if action.Topic != "" {
			if u.mqttSvc != nil {
				if err := u.mqttSvc.Publish(action.Topic, 0, false, action.Value); err != nil {
					utils.LogError("Scene %s: Failed to publish to MQTT topic %s: %v", id, action.Topic, err)
					errs = append(errs, err)
				}
			}
		} else if action.DeviceID != "" {
			if action.RemoteID != "" {
				valInt, ok := utils.ToInt(action.Value)
				if !ok {
					err := fmt.Errorf("invalid value for IR command on device %s: %v", action.DeviceID, action.Value)
					utils.LogWarn("Scene %s: %v", id, err)
					errs = append(errs, err)
					continue
				}
				_, err := u.tuyaCmd.SendIRACCommand(accessToken, action.DeviceID, action.RemoteID, action.Code, valInt)
				if err != nil {
					utils.LogError("Scene %s: Failed to send IR command to %s: %v", id, action.DeviceID, err)
					errs = append(errs, err)
				}
			} else {
				cmd := tuya_dtos.TuyaCommandDTO{
					Code:  action.Code,
					Value: action.Value,
				}
				_, err := u.tuyaCmd.SendCommand(accessToken, action.DeviceID, []tuya_dtos.TuyaCommandDTO{cmd})
				if err != nil {
					utils.LogError("Scene %s: Failed to send command to %s: %v", id, action.DeviceID, err)
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("triggered %d errors during scene execution", len(errs))
	}

	return nil
}
