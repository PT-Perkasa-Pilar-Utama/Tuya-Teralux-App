package orchestrator

import (
	"context"
	"strings"
	"testing"

	"sensio/domain/common/infrastructure"
	"sensio/domain/models/rag/skills"
)

type captureLLM struct {
	lastPrompt string
}

func (c *captureLLM) CallModel(_ context.Context, prompt string, _ string) (string, error) {
	c.lastPrompt = prompt
	return "ok", nil
}

func TestBaseOrchestrator_ReplacesDevicesPlaceholder(t *testing.T) {
	vector := infrastructure.NewVectorService("")
	if err := vector.Upsert(
		"tuya:devices:uid:sg12345678",
		`{"devices":[{"id":"dev-1","name":"Lampu Kamar"}],"total_devices":1}`,
		nil,
	); err != nil {
		t.Fatalf("failed to seed vector: %v", err)
	}

	llm := &captureLLM{}
	orch := NewBaseOrchestrator()
	_, err := orch.Execute(&skills.SkillContext{
		Ctx:    context.Background(),
		UID:    "sg12345678",
		Vector: vector,
		LLM:    llm,
	}, "Devices:\n{{devices}}")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(llm.lastPrompt, "- Lampu Kamar (ID: dev-1)") {
		t.Fatalf("devices placeholder was not populated, prompt=%q", llm.lastPrompt)
	}
}

func TestBaseOrchestrator_UsesNoDevicesMessageWhenMissing(t *testing.T) {
	llm := &captureLLM{}
	orch := NewBaseOrchestrator()
	_, err := orch.Execute(&skills.SkillContext{
		Ctx:    context.Background(),
		UID:    "sg-not-found",
		Vector: infrastructure.NewVectorService(""),
		LLM:    llm,
	}, "Devices:\n{{devices}}")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(llm.lastPrompt, "No devices connected.") {
		t.Fatalf("expected fallback no-devices message, prompt=%q", llm.lastPrompt)
	}
}
