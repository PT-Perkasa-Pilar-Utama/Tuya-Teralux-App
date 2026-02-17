package skills

import "sync"

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
	r.skills[skill.Name()] = skill
}

// Get retrieves a skill by its unique name.
func (r *SkillRegistry) Get(name string) (Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
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
