package usecases

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/tuya/dtos"
	"sort"
	"time"
)

// TuyaDeviceControlExecutor (Duplicate from Terminal for internal use if needed, but better to use a local interface)
// This bridge satisfies domains that need a single executor for both command types.
type TuyaDeviceControlExecutor interface {
	SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error)
}

type tuyaDeviceControlBridge struct {
	sendCommandUC   TuyaCommandSwitchUseCase
	sendIRCommandUC TuyaSendIRCommandUseCase
	badger          *infrastructure.BadgerService
}

// NewTuyaDeviceControlBridge creates a bridge that implements both command types.
func NewTuyaDeviceControlBridge(sendCommandUC TuyaCommandSwitchUseCase, sendIRCommandUC TuyaSendIRCommandUseCase, badger *infrastructure.BadgerService) TuyaDeviceControlExecutor {
	return &tuyaDeviceControlBridge{
		sendCommandUC:   sendCommandUC,
		sendIRCommandUC: sendIRCommandUC,
		badger:          badger,
	}
}

func (b *tuyaDeviceControlBridge) SendSwitchCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error) {
	if b.badger != nil {
		// Canonicalize commands by sorting by Code for consistent hashing
		sortedCmds := make([]dtos.TuyaCommandDTO, len(commands))
		copy(sortedCmds, commands)
		sort.Slice(sortedCmds, func(i, j int) bool {
			return sortedCmds[i].Code < sortedCmds[j].Code
		})

		cmdsBytes, _ := json.Marshal(sortedCmds)
		hashInput := fmt.Sprintf("%s:switch:%s", deviceID, string(cmdsBytes))
		hash := sha256.Sum256([]byte(hashInput))
		cacheKey := fmt.Sprintf("action_guard:%x", hash)

		isNew, err := b.badger.SetIfAbsentWithTTL(cacheKey, []byte("1"), 3*time.Second)
		if err != nil {
			utils.LogError("ControlGuard: Duplicate check failed | error=%v", err)
		} else if !isNew {
			utils.LogInfo("ControlGuard: Duplicate control skipped | deviceID=%s | type=switch", deviceID)
			return true, nil
		}

		success, err := b.sendCommandUC.SendSwitchCommand(accessToken, deviceID, commands)
		if err != nil || !success {
			// Clear guard if execution failed so retry is possible
			_ = b.badger.Delete(cacheKey)
		}
		return success, err
	}
	return b.sendCommandUC.SendSwitchCommand(accessToken, deviceID, commands)
}

func (b *tuyaDeviceControlBridge) SendIRACCommand(accessToken, infraredID, remoteID string, params map[string]int) (bool, error) {
	if b.badger != nil {
		paramsBytes, _ := json.Marshal(params)
		hashInput := fmt.Sprintf("%s:ir:%s:%s", infraredID, remoteID, string(paramsBytes))
		hash := sha256.Sum256([]byte(hashInput))
		cacheKey := fmt.Sprintf("action_guard:%x", hash)

		isNew, err := b.badger.SetIfAbsentWithTTL(cacheKey, []byte("1"), 3*time.Second)
		if err != nil {
			utils.LogError("ControlGuard: Duplicate check failed | error=%v", err)
		} else if !isNew {
			utils.LogInfo("ControlGuard: Duplicate control skipped | remoteID=%s | type=ir", remoteID)
			return true, nil
		}

		success, err := b.sendIRCommandUC.SendIRACCommand(accessToken, infraredID, remoteID, params)
		if err != nil || !success {
			// Clear guard if execution failed so retry is possible
			_ = b.badger.Delete(cacheKey)
		}
		return success, err
	}
	return b.sendIRCommandUC.SendIRACCommand(accessToken, infraredID, remoteID, params)
}


