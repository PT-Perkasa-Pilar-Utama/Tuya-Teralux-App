package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sensio/domain/common/utils"
	"sensio/domain/rag/sensors"
	"sensio/domain/rag/services"
	tuyaDtos "sensio/domain/tuya/dtos"
	tuyaUsecases "sensio/domain/tuya/usecases"
	"strings"

	"gopkg.in/yaml.v3"
)

// MarkdownSkill is a generic skill that loads its definition from a Markdown file.
type MarkdownSkill struct {
	FilePath     string
	Metadata     SkillMetadata
	Prompt       string
	TuyaExecutor tuyaUsecases.TuyaDeviceControlExecutor
	TuyaAuth     tuyaUsecases.TuyaAuthUseCase
}

type SkillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func NewMarkdownSkill(path string, executor tuyaUsecases.TuyaDeviceControlExecutor, auth tuyaUsecases.TuyaAuthUseCase) (*MarkdownSkill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Simple frontmatter parsing
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown skill format in %s: missing frontmatter", path)
	}

	var meta SkillMetadata
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata in %s: %w", path, err)
	}

	return &MarkdownSkill{
		FilePath:     path,
		Metadata:     meta,
		Prompt:       strings.TrimSpace(parts[2]),
		TuyaExecutor: executor,
		TuyaAuth:     auth,
	}, nil
}

func (s *MarkdownSkill) Name() string {
	return s.Metadata.Name
}

func (s *MarkdownSkill) Description() string {
	return s.Metadata.Description
}

func (s *MarkdownSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	prompt := s.Prompt

	// 1. Fetch devices if needed by placeholders
	devices, deviceListStr := s.getDevices(ctx)

	// 2. Handle placeholders
	prompt = strings.ReplaceAll(prompt, "{{prompt}}", ctx.Prompt)
	prompt = strings.ReplaceAll(prompt, "{{history}}", strings.Join(ctx.History, "\n"))
	prompt = strings.ReplaceAll(prompt, "{{devices}}", deviceListStr)

	// Summary specific tokens
	if s.Name() == "Summary" {
		targetLangName := "Indonesian"
		if strings.EqualFold(ctx.Language, "en") {
			targetLangName = "English"
		}

		promptConfig := &services.PromptConfig{
			Assertiveness: 8,
			Audience:      "mixed",
			RiskScale:     "granular",
			Context:       ctx.Context,
			Style:         ctx.Style,
			Language:      targetLangName,
			Date:          ctx.Date,
			Location:      ctx.Location,
			Participants:  ctx.Participants,
		}

		prompt = strings.ReplaceAll(prompt, "{{audience_guidance}}", promptConfig.AudienceGuidance())
		prompt = strings.ReplaceAll(prompt, "{{risk_scoring_guidance}}", promptConfig.RiskScoringGuidance())
		prompt = strings.ReplaceAll(prompt, "{{assertiveness_phrasing}}", promptConfig.AssertivenessPhrasing())
		prompt = strings.ReplaceAll(prompt, "{{language}}", targetLangName)
		prompt = strings.ReplaceAll(prompt, "{{context}}", ctx.Context)
		prompt = strings.ReplaceAll(prompt, "{{style}}", ctx.Style)
		prompt = strings.ReplaceAll(prompt, "{{date}}", ctx.Date)
		prompt = strings.ReplaceAll(prompt, "{{location}}", ctx.Location)
		prompt = strings.ReplaceAll(prompt, "{{participants}}", ctx.Participants)
	}

	// Translation specific tokens
	if s.Name() == "Translation" {
		targetLang := "English"
		if ctx.Language != "" {
			switch {
			case strings.EqualFold(ctx.Language, "id") || strings.EqualFold(ctx.Language, "indonesian"):
				targetLang = "Indonesian"
			case strings.EqualFold(ctx.Language, "en") || strings.EqualFold(ctx.Language, "english"):
				targetLang = "English"
			default:
				targetLang = ctx.Language
			}
		} else if strings.Contains(strings.ToLower(ctx.Prompt), "indonesia") || strings.Contains(strings.ToLower(ctx.Prompt), "indo") {
			targetLang = "Indonesian"
		}
		prompt = strings.ReplaceAll(prompt, "{{target_lang}}", targetLang)
	}

	// 3. Call LLM
	model := "high"
	if s.Name() == "Refine" || s.Name() == "Translation" {
		model = "low"
	}

	res, err := ctx.LLM.CallModel(prompt, model)
	if err != nil {
		return nil, err
	}

	cleanRes := strings.TrimSpace(res)

	// 4. Handle Actions
	// ACTION:CONTROL[id]
	re := regexp.MustCompile(`ACTION:CONTROL\[([a-zA-Z0-9_-]+)\]`)
	matches := re.FindAllStringSubmatch(cleanRes, -1)

	if len(matches) > 0 && s.TuyaExecutor != nil {
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
				controlRes, err := s.executeControl(ctx, targetDevice)
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
			// Clean up the ACTION tags from the message if they were presented to the user
			finalMsg := re.ReplaceAllString(cleanRes, "")
			if strings.TrimSpace(finalMsg) != "" {
				combinedMsg = strings.TrimSpace(finalMsg) + "\n\n" + combinedMsg
			}

			return &SkillResult{
				Message:        combinedMsg,
				IsControl:      true,
				HTTPStatusCode: lastStatus,
			}, nil
		}
	}

	return &SkillResult{
		Message:        cleanRes,
		HTTPStatusCode: 200,
	}, nil
}

func (s *MarkdownSkill) getDevices(ctx *SkillContext) ([]tuyaDtos.TuyaDeviceDTO, string) {
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
	aggJSON, ok := ctx.Vector.Get(userDevicesID)
	if !ok {
		return nil, "No devices connected."
	}

	var aggResp tuyaDtos.TuyaDevicesResponseDTO
	if err := json.Unmarshal([]byte(aggJSON), &aggResp); err != nil {
		return nil, "Error loading devices."
	}

	var names []string
	for _, d := range aggResp.Devices {
		// For Identity, just names. For Control, ID + Status codes.
		if s.Name() == "Control" {
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
		} else {
			names = append(names, "- "+d.Name)
		}
	}

	return aggResp.Devices, strings.Join(names, "\n")
}

func (s *MarkdownSkill) executeControl(ctx *SkillContext, target *tuyaDtos.TuyaDeviceDTO) (*SkillResult, error) {
	token, err := s.TuyaAuth.GetTuyaAccessToken()
	if err != nil {
		return nil, err
	}

	category := strings.ToLower(target.Category)
	var deviceSensor sensors.DeviceSensor

	// Sensor selection logic (extracted from ControlSkill)
	switch {
	case target.RemoteID != "" || category == "rs" || category == "ac" || category == "cl":
		deviceSensor = sensors.NewIRACsensor()
	case category == "dj" || category == "xdd" || category == "fwd" || category == "ty":
		deviceSensor = sensors.NewLightSensor()
	case category == "kg" || category == "cz" || category == "pc" || category == "dlq":
		deviceSensor = sensors.NewSwitchSensor()
	case category == "ws" || category == "cs" || category == "mcs" || category == "wsdcg":
		deviceSensor = sensors.NewTemperatureSensor()
	default:
		deviceSensor = sensors.NewTerminalSensor()
	}

	res, err := deviceSensor.ExecuteControl(token, target, ctx.Prompt, ctx.History, s.TuyaExecutor)
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

// LoadSkillsFromDirectory scans the given directory for .md files and registers them.
func LoadSkillsFromDirectory(dir string, registry *SkillRegistry, executor tuyaUsecases.TuyaDeviceControlExecutor, auth tuyaUsecases.TuyaAuthUseCase) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}

	for _, f := range files {
		skill, err := NewMarkdownSkill(f, executor, auth)
		if err != nil {
			utils.LogError("Failed to load markdown skill from %s: %v", f, err)
			continue
		}
		registry.Register(skill)
		utils.LogInfo("Registered Markdown Skill: %s", skill.Name())
	}

	return nil
}
