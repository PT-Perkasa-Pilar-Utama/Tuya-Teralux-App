package usecases

import (
	"teralux_app/domain/tuya/dtos"
)

// TuyaDeviceControlExecutor (Duplicate from Teralux for internal use if needed, but better to use a local interface)
// This bridge satisfies domains that need a single executor for both command types.
type TuyaDeviceControlExecutor interface {
	SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error)
}

type tuyaDeviceControlBridge struct {
	sendCommandUC   TuyaCommandSwitchUseCase
	sendIRCommandUC TuyaSendIRCommandUseCase
}

// NewTuyaDeviceControlBridge creates a bridge that implements both command types.
func NewTuyaDeviceControlBridge(sendCommandUC TuyaCommandSwitchUseCase, sendIRCommandUC TuyaSendIRCommandUseCase) TuyaDeviceControlExecutor {
	return &tuyaDeviceControlBridge{
		sendCommandUC:   sendCommandUC,
		sendIRCommandUC: sendIRCommandUC,
	}
}

func (b *tuyaDeviceControlBridge) SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error) {
	return b.sendCommandUC.SendSwitchCommand(accessToken, deviceID, commands)
}

func (b *tuyaDeviceControlBridge) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	return b.sendIRCommandUC.SendIRACCommand(accessToken, infraredID, remoteID, params)
}
