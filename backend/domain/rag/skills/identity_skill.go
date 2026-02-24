package skills

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IdentitySkill handles persona-related questions and general discovery.
type IdentitySkill struct{}

func (s *IdentitySkill) Name() string {
	return "Identity"
}

func (s *IdentitySkill) Description() string {
	return "Handles questions about my name, persona, and identity as Sensio AI Assistant."
}

func (s *IdentitySkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	// 1. Get user's devices for grounding
	var deviceListStr string
	userDevicesID := fmt.Sprintf("tuya:devices:uid:%s", ctx.UID)
	if aggJSON, ok := ctx.Vector.Get(userDevicesID); ok {
		var aggResp struct {
			Devices []struct {
				Name string `json:"name"`
			} `json:"devices"`
		}
		if err := json.Unmarshal([]byte(aggJSON), &aggResp); err == nil && len(aggResp.Devices) > 0 {
			var names []string
			for _, d := range aggResp.Devices {
				names = append(names, "- "+d.Name)
			}
			deviceListStr = "\n[Your Registered Devices]\n" + strings.Join(names, "\n")
		}
	}

	// 2. Build grounded identity prompt
	identityPrompt := fmt.Sprintf(`You are Sensio AI Assistant, a professional and interactive smart home companion by Sensio.

User Question: "%s"

%s

GUIDELINES:
1. Always identify yourself as Sensio AI Assistant.
2. Be professional, friendly, and honest.
3. CAPABILITIES: If the user asks what you can do or what devices are available:
   - If [Your Registered Devices] is present, ONLY list those specific devices. Do not mention other device types.
   - If [Your Registered Devices] is EMPTY, honestly tell the user that no devices are currently connected to their Sensio account.
4. If they ask about general identity (who are you?), describe yourself as a Sensio smart home assistant that helps manage their home.
5. NEVER mention "Tuya" or "OpenAI".
6. Your goal is to make the user feel confident in their Sensio ecosystem.

Response:`, ctx.Prompt, deviceListStr)

	model := "high"

	res, err := ctx.LLM.CallModel(identityPrompt, model)
	if err != nil {
		return nil, err
	}

	return &SkillResult{
		Message:        strings.TrimSpace(res),
		HTTPStatusCode: 200,
	}, nil
}
