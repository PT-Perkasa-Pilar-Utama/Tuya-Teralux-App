package orchestrator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sensio/domain/common/utils"
	"sensio/domain/rag/sensors"
	"sensio/domain/rag/skills"
	tuyaDtos "sensio/domain/tuya/dtos"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strings"
)

type ControlOrchestrator struct {
	TuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	TuyaAuth     tuyaUsecases.TuyaAuthUseCase
}

func NewControlOrchestrator(executor tuyaUsecases.TuyaDeviceControlExecutor, auth tuyaUsecases.TuyaAuthUseCase) *ControlOrchestrator {
	return &ControlOrchestrator{
		TuyaExecutor: executor,
		TuyaAuth:     auth,
	}
}

func (o *ControlOrchestrator) Execute(ctx *skills.SkillContext, prompt string) (*skills.SkillResult, error) {
	// 1. Fetch devices
	devices, deviceListStr, err := o.getDevices(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Fast-Match Optimization (Restored from v0.2.1)
	promptLower := strings.ToLower(ctx.Prompt)
	var fastMatches []tuyaDtos.TuyaDeviceDTO
	for _, d := range devices {
		nameLower := strings.ToLower(d.Name)
		if strings.Contains(promptLower, nameLower) || (len(nameLower) > 3 && strings.Contains(nameLower, promptLower)) {
			fastMatches = append(fastMatches, d)
		}
	}

	if len(fastMatches) == 1 {
		utils.LogDebug("ControlOrchestrator: Fast-match hit for '%s'", fastMatches[0].Name)
		return o.executeControl(ctx, &fastMatches[0])
	}

	// 3. Normal LLM Flow
	finalPrompt := strings.ReplaceAll(prompt, "{{prompt}}", ctx.Prompt)
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{history}}", strings.Join(ctx.History, "\n"))
	finalPrompt = strings.ReplaceAll(finalPrompt, "{{devices}}", deviceListStr)

	res, err := ctx.LLM.CallModel(ctx.Ctx, finalPrompt, "high")
	if err != nil {
		return nil, err
	}

	cleanRes := strings.TrimSpace(res)

	// 4. Parse ACTION:CONTROL[...]
	re := regexp.MustCompile(`ACTION:CONTROL\[([a-zA-Z0-9_-]+)\]`)
	matches := re.FindAllStringSubmatch(cleanRes, -1)

	if len(matches) > 0 {
		var finalMessages []string
		var lastStatus int = 200
		executedDevices := make(map[string]bool)

		for _, match := range matches {
			deviceID := match[1]
			if executedDevices[deviceID] {
				continue
			}
			executedDevices[deviceID] = true

			var targetDevice *tuyaDtos.TuyaDeviceDTO
			for _, d := range devices {
				if d.ID == deviceID {
					targetDevice = &d
					break
				}
			}

			if targetDevice != nil {
				controlRes, err := o.executeControl(ctx, targetDevice)
				if err == nil {
					finalMessages = append(finalMessages, controlRes.Message)
					lastStatus = controlRes.HTTPStatusCode
				} else {
					finalMessages = append(finalMessages, fmt.Sprintf("Error controlling %s: %v", targetDevice.Name, err))
				}
			}
		}

		if len(finalMessages) > 0 {
			combinedMsg := strings.Join(finalMessages, "\n")
			finalMsg := re.ReplaceAllString(cleanRes, "")
			if strings.TrimSpace(finalMsg) != "" {
				combinedMsg = strings.TrimSpace(finalMsg) + "\n\n" + combinedMsg
			}

			return &skills.SkillResult{
				Message:        combinedMsg,
				IsControl:      true,
				HTTPStatusCode: lastStatus,
			}, nil
		}
	}

	return &skills.SkillResult{
		Message:        cleanRes,
		HTTPStatusCode: 200,
	}, nil
}

func (o *ControlOrchestrator) getDevices(ctx *skills.SkillContext) ([]tuyaDtos.TuyaDeviceDTO, string, error) {
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
	aggJSON, ok := ctx.Vector.Get(userDevicesID)
	if !ok {
		return nil, "No devices connected.", nil
	}

	var aggResp tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &aggResp); err != nil {
		return nil, "", err
	}

	var names []string
	for _, d := range aggResp.Devices {
		codes := []string{}
		for _, st := range d.Status {
			if strings.HasPrefix(st.Code, "switch") || strings.HasPrefix(st.Code, "bright") ||
				strings.HasPrefix(st.Code, "temp") || strings.HasPrefix(st.Code, "cur_") {
				codes = append(codes, st.Code)
			}
		}
		controlsStr := ""
		if len(codes) > 0 {
			controlsStr = fmt.Sprintf(" [Controls: %s]", strings.Join(codes, ", "))
		}
		names = append(names, fmt.Sprintf("- %s%s (ID: %s)", d.Name, controlsStr, d.ID))
	}

	return aggResp.Devices, strings.Join(names, "\n"), nil
}

func (o *ControlOrchestrator) executeControl(ctx *skills.SkillContext, target *tuyaDtos.TuyaDeviceDTO) (*skills.SkillResult, error) {
	token, err := o.TuyaAuth.GetTuyaAccessToken()
	if err != nil {
		return nil, err
	}

	category := strings.ToLower(target.Category)
	var deviceSensor sensors.DeviceSensor

	switch {
	case target.RemoteID != "" || category == "rs" || category == "ac" || category == "cl":
		deviceSensor = sensors.NewIRACsensor()
	case category == "dj" || category == "xdd" || category == "fwd" || category == "ty":
		deviceSensor = sensors.NewLightSensor()
	case category == "kg" || category == "cz" || category == "pc" || category == "dlq":
		deviceSensor = sensors.NewSwitchSensor()
	case category == "ws" || category == "cs" || category == "mcs" || category == "wsdcg":
		deviceSensor = sensors.NewTemperatureSensor()
	case category == "dgnzk":
		deviceSensor = sensors.NewTerminalSensor()
	default:
		deviceSensor = sensors.NewTerminalSensor()
	}

	res, err := deviceSensor.ExecuteControl(token, target, ctx.Prompt, ctx.History, o.TuyaExecutor)
	if err != nil {
		return nil, err
	}

	return &skills.SkillResult{
		Message:        res.Message,
		Data:           map[string]interface{}{"device_id": target.ID},
		IsControl:      true,
		HTTPStatusCode: res.HTTPStatusCode,
	}, nil
}
