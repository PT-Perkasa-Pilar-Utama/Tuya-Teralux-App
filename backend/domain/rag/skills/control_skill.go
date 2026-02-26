package skills

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sensio/domain/rag/sensors"
	tuyaDtos "sensio/domain/tuya/dtos"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strings"
)

// ControlSkill handles device control requests using RAG and similarity search.
type ControlSkill struct {
	tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	tuyaAuth     tuyaUsecases.TuyaAuthUseCase
}

func NewControlSkill(tuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor, tuyaAuth tuyaUsecases.TuyaAuthUseCase) *ControlSkill {
	return &ControlSkill{
		tuyaExecutor: tuyaExecutor,
		tuyaAuth:     tuyaAuth,
	}
}

func (s *ControlSkill) Name() string {
	return "Control"
}

func (s *ControlSkill) Description() string {
	return "Handles all action requests to control devices, including lights, AC, media/music playback, or checking device status."
}

func (s *ControlSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	// 1. Get user's devices for filtering
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
	aggJSON, ok := ctx.Vector.Get(userDevicesID)
	if !ok {
		return &SkillResult{
			Message:        "Sorry, it seems there are no devices connected to your account. Please connect devices first through the Sensio app.",
			HTTPStatusCode: 404,
		}, nil
	}

	var aggResp tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &aggResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user devices: %w", err)
	}

	// 2. Hybrid Device Selection
	// Try fast-match first (exact/partial name matching in prompt)
	promptLower := strings.ToLower(ctx.Prompt)
	var fastMatches []tuyaDtos.TuyaDeviceDTO
	for _, d := range aggResp.Devices {
		nameLower := strings.ToLower(d.Name)
		// If prompt contains device name OR device name contains prompt words (excluding keywords)
		if strings.Contains(promptLower, nameLower) || (len(nameLower) > 3 && strings.Contains(nameLower, promptLower)) {
			fastMatches = append(fastMatches, d)
		}
	}

	// If we have a clear single fast-match, avoid LLM cost and potential hallucination
	if len(fastMatches) == 1 {
		fmt.Printf("DEBUG: ControlSkill: Fast-match hit for '%s'\n", fastMatches[0].Name)
		return s.executeOnTarget(ctx, &fastMatches[0])
	}

	// Otherwise, let the LLM decide from the full list (or the narrowed list if helpful)
	return s.selectDeviceWithLLM(ctx, aggResp.Devices)
}

func (s *ControlSkill) selectDeviceWithLLM(ctx *SkillContext, devices []tuyaDtos.TuyaDeviceDTO) (*SkillResult, error) {
	var historyContext string
	if len(ctx.History) > 0 {
		historyContext = "Previous conversation:\n" + strings.Join(ctx.History, "\n") + "\n"
	}

	deviceList := make([]string, 0, len(devices))
	for _, d := range devices {
		// Include relevant status codes to help LLM distinguish multi-switch devices
		codes := []string{}
		for _, st := range d.Status {
			// Only include common/relevant codes to keep prompt small
			if strings.HasPrefix(st.Code, "switch") || strings.HasPrefix(st.Code, "bright") ||
				strings.HasPrefix(st.Code, "temp") || strings.HasPrefix(st.Code, "cur_") {
				codes = append(codes, st.Code)
			}
		}

		controlsStr := ""
		if len(codes) > 0 {
			controlsStr = fmt.Sprintf(" [Controls: %s]", strings.Join(codes, ", "))
		}
		deviceList = append(deviceList, fmt.Sprintf("- %s%s (ID: %s)", d.Name, controlsStr, d.ID))
	}

	reconcilePrompt := fmt.Sprintf(`You are Sensio AI Home Controller.
Your goal is to parse user requests and output the correct device execution commands.

User Prompt: "%s"

%s
[Available Devices]
%s

GUIDELINES:
1. MATCHING: Identify all devices the user wants to control.
2. MULTI-DEVICE: If the user says "all", "everything", or implies multiple devices (e.g., "turn on all lights"), identify ALL matching devices.
3. OUTPUT FORMAT: 
   - For EACH target device, output "EXECUTE:[Device ID]" on a NEW LINE.
   - You MAY include short natural language before or after the EXECUTE tags.
   - Example: "I will turn on the following devices:\nEXECUTE:id1\nEXECUTE:id2"
4. NO CONFIRMATION: If the command is clear (e.g., "turn on office light"), do NOT ask "Are you sure?" or "Do you want to continue?". Output the EXECUTE tag immediately.
5. AMBIGUITY: If the request is truly vague and could match different types of devices incorrectly, ask a short clarifying question.
6. NO HALLUCINATION: Only use Device IDs from the [Available Devices] list.

Response:`, ctx.Prompt, historyContext, strings.Join(deviceList, "\n"))

	model := "high"

	res, err := ctx.LLM.CallModel(reconcilePrompt, model)
	if err != nil {
		return nil, err
	}

	cleanRes := strings.TrimSpace(res)
	fmt.Printf("DEBUG: ControlSkill: LLM raw response: '%s'\n", cleanRes)

	// Regex to find all EXECUTE:[ID] patterns
	// Matches EXECUTE: followed by alphanumeric characters (IDs are typically alphanumeric)
	re := regexp.MustCompile(`EXECUTE:([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(cleanRes, -1)

	if len(matches) > 0 {
		var finalMessages []string
		var lastStatus int = 200
		executedCount := 0

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			deviceID := match[1]
			var targetDevice *tuyaDtos.TuyaDeviceDTO
			for _, d := range devices {
				if d.ID == deviceID {
					targetDevice = &d
					break
				}
			}

			if targetDevice != nil {
				res, err := s.executeOnTarget(ctx, targetDevice)
				if err == nil {
					finalMessages = append(finalMessages, res.Message)
					lastStatus = res.HTTPStatusCode
					executedCount++
				} else {
					finalMessages = append(finalMessages, fmt.Sprintf("Error controlling %s: %v", targetDevice.Name, err))
				}
			}
		}

		if executedCount > 0 {
			return &SkillResult{
				Message:        strings.Join(finalMessages, "\n"),
				IsControl:      true,
				HTTPStatusCode: lastStatus,
			}, nil
		}
	}

	// No execution tags found, return the LLM's full response (likely a question)
	return &SkillResult{
		Message: cleanRes,
	}, nil
}

func (s *ControlSkill) executeOnTarget(ctx *SkillContext, target *tuyaDtos.TuyaDeviceDTO) (*SkillResult, error) {
	token, err := s.tuyaAuth.GetTuyaAccessToken()
	if err != nil {
		return &SkillResult{
			Message:        fmt.Sprintf("Failed to get access token: %v", err),
			HTTPStatusCode: 401,
		}, nil
	}

	deviceSensor := s.selectDeviceSensor(target)
	res, err := deviceSensor.ExecuteControl(token, target, ctx.Prompt, ctx.History, s.tuyaExecutor)
	if err != nil {
		return nil, err
	}

	return &SkillResult{
		Message:        res.Message,
		Data:           map[string]interface{}{"device_id": target.ID},
		IsControl:      true,
		HTTPStatusCode: res.HTTPStatusCode,
	}, nil

}

func (s *ControlSkill) selectDeviceSensor(device *tuyaDtos.TuyaDeviceDTO) sensors.DeviceSensor {
	category := strings.ToLower(device.Category)

	// 1. IR Devices (AC, TV, Fan)
	if device.RemoteID != "" || category == "rs" || category == "ac" || category == "cl" {
		return sensors.NewIRACsensor()
	}

	// 2. Lights (dj = light, xdd = ceiling light, fwd = floor light, etc)
	if category == "dj" || category == "xdd" || category == "fwd" || category == "ty" {
		return sensors.NewLightSensor()
	}

	// 3. Switches & Sockets (kg = switch, cz = outlet/socket, pc = power strip, dlq = breaker)
	if category == "kg" || category == "cz" || category == "pc" || category == "dlq" {
		return sensors.NewSwitchSensor()
	}

	// 4. Sensors (ws = temp/humidity, cs = pir/motion, mcs = door/window, wsdcg = th sensor)
	if category == "ws" || category == "cs" || category == "mcs" || category == "wsdcg" {
		return sensors.NewTemperatureSensor()
	}

	// 5. Terminal Specific (dgnzk = voice/media control panel)
	if category == "dgnzk" {
		return sensors.NewTerminalSensor()
	}

	// Default fallback: TerminalSensor handles generic switches and controls better than SwitchSensor
	return sensors.NewTerminalSensor()
}
