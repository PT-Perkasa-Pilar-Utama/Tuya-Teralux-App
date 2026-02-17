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
	return "Handles requests to control smart home devices, such as turning on lights, setting AC temperature, or checking device status."
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

	allowedIDs := make(map[string]tuyaDtos.TuyaDeviceDTO)
	for _, d := range aggResp.Devices {
		allowedIDs["tuya:device:"+d.ID] = d
	}

	// 2. Initial Similarity Search
	matches, _ := ctx.Vector.Search(ctx.Prompt)
	var validMatches []tuyaDtos.TuyaDeviceDTO
	for _, m := range matches {
		if dev, exists := allowedIDs[m]; exists {
			validMatches = append(validMatches, dev)
		}
	}

	// 3. Handle Ambiguity logic
	if len(validMatches) == 0 {
		return s.handleNoInitialMatches(ctx, aggResp.Devices)
	}

	if len(validMatches) > 1 {
		var names []string
		for _, v := range validMatches {
			names = append(names, fmt.Sprintf("- **%s**", v.Name))
		}
		return &SkillResult{
			Message:        fmt.Sprintf("I found %d matching devices:\n%s\nWhich one would you like to control?", len(validMatches), strings.Join(names, "\n")),
			HTTPStatusCode: 400,
		}, nil
	}

	// 4. Single match found - Execute command
	target := validMatches[0]
	return s.executeOnTarget(ctx, &target)
}

func (s *ControlSkill) handleNoInitialMatches(ctx *SkillContext, devices []tuyaDtos.TuyaDeviceDTO) (*SkillResult, error) {
	var historyContext string
	if len(ctx.History) > 0 {
		historyContext = "Previous conversation:\n" + strings.Join(ctx.History, "\n") + "\n"
	}

	var deviceList []string
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
1. KAPABILITAS: Jika ditanya apa yang bisa dikontrol atau perangkat apa yang tersedia, sebutkan perangkat dari daftar [Daftar Perangkat Tersedia] di atas.
2. KONTROL: 
   - Jika pengguna jelas merujuk pada SATU perangkat spesifik dari daftar, kembalikan: "EXECUTE:[Device ID]".
   - Jika pengguna menjawab pertanyaan tindak lanjut untuk memperjelas perangkat, identifikasi dan kembalikan: "EXECUTE:[Device ID]".
3. AMBIGUITAS: Jika permintaan samar tetapi berkaitan dengan kontrol rumah pintar, ajukan pertanyaan klarifikasi yang sopan dan profesional.
4. KEJUJURAN: Hanya bicara tentang perangkat yang ada di daftar [Daftar Perangkat Tersedia]. Jika perangkat tidak ada, katakan dengan jujur bahwa perangkat tidak ditemukan dalam sistem Sensio mereka.
5. NO HALLUCINATION: Jika prompt tentang identitas Anda, abaikan dan fokus ke perangkat.

Response (Bahasa Indonesia):`, ctx.Prompt, historyContext, strings.Join(deviceList, "\n"))
	} else {
		reconcilePrompt = fmt.Sprintf(`You are Sensio AI Assistant, a professional and interactive smart home companion by Sensio.
Your goal is to help the user manage their smart home devices efficiently.

User Prompt: "%s"

%s
[Available Devices]
%s

GUIDELINES:
1. CAPABILITIES: If asked what you can control or what devices are available, list the devices from the [Available Devices] section above.
2. CONTROL: 
   - If the user is clearly referring to ONE specific device from the list, return: "EXECUTE:[Device ID]".
   - If they are answering a follow-up question to clarify a device, identify it and return: "EXECUTE:[Device ID]".
3. AMBIGUITY: If the request is vague but relates to smart home control, ask a professional follow-up question.
4. HONESTY: Only talk about devices present in the [Available Devices] list. If a device isn't there, be direct and honest: tell the user that the device is not found in their Sensio system.
5. NO HALLUCINATION: If the prompt is about your identity or what you are, the Orchestrator should have routed you to the Identity skill. If you are here, focus on device-related matters.

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
	// If device has a RemoteID, it's an IR device (AC, TV, Fan, etc.)
	if device.RemoteID != "" || category == "rs" || category == "ac" || category == "cl" {
		return sensors.NewIRACsensor()
	}
	// Default to Teralux sensor for other devices
	return sensors.NewTeraluxSensor()
}
