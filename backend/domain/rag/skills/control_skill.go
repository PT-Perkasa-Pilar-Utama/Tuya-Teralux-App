package skills

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/rag/sensors"
	tuyaDtos "teralux_app/domain/tuya/dtos"
	tuyaUsecases "teralux_app/domain/tuya/usecases"
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
		deviceList = append(deviceList, fmt.Sprintf("- %s (ID: %s)", d.Name, d.ID))
	}

	var reconcilePrompt string
	if ctx.Language == "id" {
		reconcilePrompt = fmt.Sprintf(`Anda adalah Sensio AI Assistant, asisten rumah pintar yang profesional.
Tujuan Anda adalah membantu pengguna mengelola perangkat rumah pintar mereka dengan efisien.

User Prompt: "%s"

%s
[Daftar Perangkat Tersedia]
%s

PANDUAN:
1. KOCARKIR (MATCHING): Cocokkan nama perangkat yang diminta dengan nama di [Daftar Perangkat Tersedia]. Nama mungkin tidak sama persis (misal: "Teralux" cocok dengan "Teralux (Receptionist)").
2. KONTROL: 
   - Jika pengguna merujuk pada perangkat dari daftar, kembalikan: "EXECUTE:[Device ID]".
   - Jika pengguna menjawab pertanyaan tindak lanjut, pilih ID-nya dan kembalikan: "EXECUTE:[Device ID]".
3. AMBIGUITAS: Jika permintaan samar atau ada banyak kemiripan, ajukan pertanyaan klarifikasi yang sopan.
4. KEJUJURAN: Hanya bicara tentang perangkat yang ada di daftar. Jika BENAR-BENAR tidak ada yang mirip, katakan perangkat tidak ditemukan.
5. NO HALLUCINATION: Jangan mengarang Device ID. Gunakan hanya ID dari daftar di atas.

Response (Bahasa Indonesia):`, ctx.Prompt, historyContext, strings.Join(deviceList, "\n"))
	} else {
		reconcilePrompt = fmt.Sprintf(`You are Sensio AI Assistant, a professional and interactive smart home companion by Sensio.
Your goal is to help the user manage their smart home devices efficiently.

User Prompt: "%s"

%s
[Available Devices]
%s

GUIDELINES:
1. MATCHING: Match the requested device name with the names in [Available Devices]. Names might not match exactly (e.g., "Teralux" matches "Teralux (Receptionist)").
2. CONTROL: 
   - If the user refers to a device from the list, return: "EXECUTE:[Device ID]".
   - If they are answering a follow-up question, identify it and return: "EXECUTE:[Device ID]".
3. AMBIGUITY: If the request is vague or matches multiple devices, ask a professional follow-up question.
4. HONESTY: Only talk about devices present in the list. If it's CLEARLY not there, say it's not found.
5. NO HALLUCINATION: Only use Device IDs from the [Available Devices] list.

Response (English):`, ctx.Prompt, historyContext, strings.Join(deviceList, "\n"))
	}

	model := ctx.Config.LLMModel
	if model == "" {
		model = "default"
	}

	res, err := ctx.LLM.CallModel(reconcilePrompt, model)
	if err != nil {
		return nil, err
	}

	cleanRes := strings.TrimSpace(res)
	fmt.Printf("DEBUG: ControlSkill: LLM raw response: '%s'\n", cleanRes)

	if strings.HasPrefix(cleanRes, "EXECUTE:") {
		deviceID := strings.TrimPrefix(cleanRes, "EXECUTE:")
		var targetDevice *tuyaDtos.TuyaDeviceDTO
		for _, d := range devices {
			if d.ID == deviceID {
				targetDevice = &d
				break
			}
		}

		if targetDevice == nil {
			return &SkillResult{
				Message:        "Device not found.",
				HTTPStatusCode: 404,
			}, nil
		}

		return s.executeOnTarget(ctx, targetDevice)
	}

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

	// 3. Switches & Sockets (kg = switch, cz = outlet/socket, pc = power strip)
	if category == "kg" || category == "cz" || category == "pc" {
		return sensors.NewSwitchSensor()
	}

	// 4. Sensors (ws = temp/humidity, cs = pir/motion, mcs = door/window)
	if category == "ws" || category == "cs" || category == "mcs" {
		return sensors.NewTemperatureSensor()
	}

	// 5. Teralux Specific (dgnzk = voice/media control panel)
	if category == "dgnzk" {
		return sensors.NewTeraluxSensor()
	}

	// Default fallback: Try to handle as a switch if it has switch capabilities,
	// otherwise defaults to basic TeraluxSensor which handles generic switches too.
	return sensors.NewSwitchSensor()
}
