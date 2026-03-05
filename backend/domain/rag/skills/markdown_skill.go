package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sensio/domain/common/utils"
	"strings"

	"gopkg.in/yaml.v3"
)

// MarkdownOrchestrator defines the contract for specialized skill logic.
type MarkdownOrchestrator interface {
	Execute(ctx *SkillContext, prompt string) (*SkillResult, error)
}

// MarkdownSkill is a generic skill that loads its definition from a Markdown file.
type MarkdownSkill struct {
	FilePath     string
	Metadata     SkillMetadata
	Prompt       string
	Orchestrator MarkdownOrchestrator
}

type SkillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func NewMarkdownSkill(path string, orchestrator MarkdownOrchestrator) (*MarkdownSkill, error) {
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
		Orchestrator: orchestrator,
	}, nil
}

func (s *MarkdownSkill) Name() string {
	return s.Metadata.Name
}

func (s *MarkdownSkill) Description() string {
	return s.Metadata.Description
}

func (s *MarkdownSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
	if s.Orchestrator == nil {
		return nil, fmt.Errorf("no orchestrator configured for skill %s", s.Name())
	}
	return s.Orchestrator.Execute(ctx, s.Prompt)
}

// OrchestratorResolver is a function that returns an orchestrator for a given skill name.
type OrchestratorResolver func(skillName string) MarkdownOrchestrator

// LoadSkillsFromDirectory scans the given directory for .md files and registers them.
func LoadSkillsFromDirectory(dir string, registry *SkillRegistry, resolver OrchestratorResolver) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}

	for _, f := range files {
		// Temporary skill to get the name from frontmatter
		tempSkill, err := NewMarkdownSkill(f, nil)
		if err != nil {
			utils.LogError("Failed to peek markdown skill from %s: %v", f, err)
			continue
		}

		orchestrator := resolver(tempSkill.Name())
		skill, err := NewMarkdownSkill(f, orchestrator)
		if err != nil {
			utils.LogError("Failed to load markdown skill from %s: %v", f, err)
			continue
		}
		registry.Register(skill)
		utils.LogInfo("Registered Markdown Skill: %s", skill.Name())
	}

	return nil
}
