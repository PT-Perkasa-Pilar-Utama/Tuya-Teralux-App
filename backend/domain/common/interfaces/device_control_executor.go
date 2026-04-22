package interfaces

import (
	"sensio/domain/tuya/dtos"
)

// DeviceControlExecutor defines the interface for device control operations across domains
type DeviceControlExecutor interface {
	SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error)
}