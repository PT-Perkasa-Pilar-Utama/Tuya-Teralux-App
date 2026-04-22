package usecases

import (
	"fmt"
	"sensio/domain/common/interfaces"
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/scene/repositories"
	tuyaDtos "sensio/domain/tuya/dtos"
)

type ControlSceneUseCase struct {
	repo    *repositories.SceneRepository
	tuyaCmd interfaces.DeviceControlExecutor
	mqttSvc *infrastructure.MqttService
}

func NewControlSceneUseCase(
	repo *repositories.SceneRepository,
	tuyaCmd interfaces.DeviceControlExecutor,
	mqttSvc *infrastructure.MqttService,
) *ControlSceneUseCase {
	return &ControlSceneUseCase{
		repo:    repo,
		tuyaCmd: tuyaCmd,
		mqttSvc: mqttSvc,
	}
}

func (u *ControlSceneUseCase) ControlScene(terminalID, id, accessToken string) error {
	scene, err := u.repo.GetByID(terminalID, id)
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
				params := map[string]int{
					action.Code: valInt,
				}
				_, err := u.tuyaCmd.SendIRACCommand(accessToken, action.DeviceID, action.RemoteID, params)
				if err != nil {
					utils.LogError("Scene %s: Failed to send IR command to %s: %v", id, action.DeviceID, err)
					errs = append(errs, err)
				}
			} else {
				cmd := tuyaDtos.TuyaCommandDTO{
					Code:  action.Code,
					Value: action.Value,
				}
				_, err := u.tuyaCmd.SendSwitchCommand(accessToken, action.DeviceID, []tuyaDtos.TuyaCommandDTO{cmd})
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