package orchestrator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sensio/domain/common/interfaces"
	"sensio/domain/common/utils"
	"sensio/domain/models/rag/sensors"
	"sensio/domain/models/rag/skills"
	tuyaDtos "sensio/domain/tuya/dtos"
	"strings"
)

type ControlOrchestrator struct {
	TuyaExecutor interfaces.DeviceControlExecutor
	TuyaAuth     interfaces.AuthUseCase
}

func NewControlOrchestrator(executor interfaces.DeviceControlExecutor, auth interfaces.AuthUseCase) *ControlOrchestrator {
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

	// Detect "all lights" intent from the prompt
	isAllLights := o.isAllLightsIntent(ctx.Prompt)

	// Filter devices if "all lights" intent is detected
	devices := aggResp.Devices
	if isAllLights {
		devices = o.filterLampDevices(devices)
	}

	names := make([]string, 0, len(devices))
	for _, d := range devices {
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

	return devices, strings.Join(names, "\n"), nil
}

// isAllLightsIntent detects if the prompt indicates an "all lights" command.
// Returns true only if the prompt contains light-specific phrases (requires "lampu/light/lamp" explicitly).
// This prevents accidental filtering of non-light devices for generic "all" commands.
func (o *ControlOrchestrator) isAllLightsIntent(prompt string) bool {
	promptLower := strings.ToLower(strings.TrimSpace(prompt))

	// Light-specific "all" patterns (Indonesian and English)
	// These MUST contain both "all" quantifier AND light-related words
	lightSpecificPatterns := []string{
		"semua lampu",
		"lampu semua",
		"all lights",
		"all light",
		"every light",
		"semua light",
		"all the lights",
		"all of the lights",
		"semua lamp",
		"all lamp",
	}

	for _, pattern := range lightSpecificPatterns {
		if strings.Contains(promptLower, pattern) {
			return true
		}
	}

	// Additional check: "turn on/off all" + "lights/lamps" in same prompt
	// This catches "turn on all the lights in the living room"
	hasAllQuantifier := strings.Contains(promptLower, "semua") ||
		strings.Contains(promptLower, "semuanya") ||
		strings.Contains(promptLower, "all ") ||
		strings.Contains(promptLower, " all") ||
		strings.Contains(promptLower, " every ")

	hasLightWord := strings.Contains(promptLower, "lampu") ||
		strings.Contains(promptLower, "light") ||
		strings.Contains(promptLower, "lamp") ||
		strings.Contains(promptLower, "switch")

	if hasAllQuantifier && hasLightWord {
		return true
	}

	// NEW: Detect implicit "all" commands - user wants all lights off/on
	// Patterns like "matikan lagi dong", "masih belum mati", "belum semua"
	implicitAllPatterns := []string{
		"lagi dong",      // "Matiin lagi dong" (implies retry all)
		"semuanya",       // "Matiin semuanya"
		"masih belum",    // "Masih belum mati" (implies not all off)
		"belum semua",    // "Belum semua mati"
		"kok ada yang",   // "Kok ada yang nyala" (implies some still on)
		"ada yang masih", // "Ada yang masih nyala"
	}

	for _, pattern := range implicitAllPatterns {
		if strings.Contains(promptLower, pattern) {
			// Only return true if also has light-related word
			if hasLightWord {
				return true
			}
		}
	}

	return false
}

// filterLampDevices filters devices to only include lamp-relevant categories.
// This excludes panel-category devices (like dgnzk) to prevent false targeting.
func (o *ControlOrchestrator) filterLampDevices(devices []tuyaDtos.TuyaDeviceDTO) []tuyaDtos.TuyaDeviceDTO {
	var filtered []tuyaDtos.TuyaDeviceDTO

	for _, d := range devices {
		category := strings.ToLower(d.Category)

		// Include lamp-relevant categories:
		// - dj, xdd, fwd, ty: Standard light categories
		// - kg, cz, pc, dlq: Smart Switch lamps
		if category == "dj" || category == "xdd" || category == "fwd" || category == "ty" ||
			category == "kg" || category == "cz" || category == "pc" || category == "dlq" {
			filtered = append(filtered, d)
		}
	}

	return filtered
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
