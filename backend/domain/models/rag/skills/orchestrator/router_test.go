package orchestrator

import (
	"context"
	"sensio/domain/models/rag/skills"
	"testing"
)

type fakeSkill struct {
	name    string
	message string
}

func (f fakeSkill) Name() string { return f.name }
func (f fakeSkill) Description() string {
	return "fake skill"
}
func (f fakeSkill) Execute(_ *skills.SkillContext) (*skills.SkillResult, error) {
	return &skills.SkillResult{Message: f.message, HTTPStatusCode: 200}, nil
}

type panicLLM struct{}

func (p panicLLM) CallModel(_ context.Context, _ string, _ string) (string, error) {
	panic("LLM should not be called for deterministic device discovery routing")
}

func TestRouteAndExecute_DeviceDiscoveryRoutesToControl(t *testing.T) {
	registry := skills.NewSkillRegistry()
	registry.Register(fakeSkill{name: "Control", message: "control-response"})
	registry.Register(fakeSkill{name: "Identity", message: "identity-response"})

	router := NewRouter(registry, nil, nil)
	ctx := &skills.SkillContext{
		Ctx:      context.Background(),
		Prompt:   "device apa aja yang bisa saya control?",
		Language: "",
		LLM:      panicLLM{},
	}

	res, err := router.RouteAndExecute(ctx)
	if err != nil {
		t.Fatalf("RouteAndExecute returned error: %v", err)
	}
	if res == nil {
		t.Fatal("RouteAndExecute returned nil result")
	}
	if res.Message != "control-response" {
		t.Fatalf("unexpected message: got %q, want %q", res.Message, "control-response")
	}
}
