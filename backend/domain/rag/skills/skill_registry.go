package skills

import (
	"sensio/domain/common/utils"
	"strings"
	"sync"
)

// SkillRegistry manages a collection of available skills.
type SkillRegistry struct {
	skills map[string]Skill
	mu     sync.RWMutex
}

// NewSkillRegistry creates a new, empty skill registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]Skill),
	}
}

// Register adds a new skill to the registry.
func (r *SkillRegistry) Register(skill Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Store with normalized name (lowercase) for case-insensitive lookup
	name := strings.ToLower(skill.Name())
	r.skills[name] = skill
	utils.LogInfo("SkillRegistry: Registered skill '%s' (normalized: '%s')", skill.Name(), name)
}

// Get retrieves a skill by its unique name (case-insensitive).
func (r *SkillRegistry) Get(name string) (Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[strings.ToLower(name)]
	return s, ok
}

// GetAll returns all registered skills.
func (r *SkillRegistry) GetAll() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		all = append(all, s)
	}
	return all
}
