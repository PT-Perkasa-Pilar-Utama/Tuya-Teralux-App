package repositories

import (
	"sensio/domain/scene/entities"

	"gorm.io/gorm"
)

// ISceneRepository defines the interface for scene storage operations
type ISceneRepository interface {
	Save(scene *entities.Scene) error
	GetByID(terminalID, id string) (*entities.Scene, error)
	GetAll(terminalID string) ([]entities.Scene, error)
	Delete(terminalID, id string) error
	GetAllGrouped() (map[string][]entities.Scene, error)
}

// SceneRepository handles persistent storage of scenes using GORM/MySQL
type SceneRepository struct {
	db *gorm.DB
}

// NewSceneRepository creates a new instance of SceneRepository
func NewSceneRepository(db *gorm.DB) *SceneRepository {
	return &SceneRepository{db: db}
}

// Save persists a scene to the database (Upsert)
func (r *SceneRepository) Save(scene *entities.Scene) error {
	return r.db.Save(scene).Error
}

// GetByID retrieves a scene by its unique identifier and TerminalID
func (r *SceneRepository) GetByID(terminalID, id string) (*entities.Scene, error) {
	var scene entities.Scene
	if err := r.db.Where("id = ? AND terminal_id = ?", id, terminalID).First(&scene).Error; err != nil {
		return nil, err
	}
	return &scene, nil
}

// GetAll retrieves all configured scenes for a specific Terminal device
func (r *SceneRepository) GetAll(terminalID string) ([]entities.Scene, error) {
	var scenes []entities.Scene
	if err := r.db.Where("terminal_id = ?", terminalID).Find(&scenes).Error; err != nil {
		return nil, err
	}
	return scenes, nil
}

// Delete removes a scene from the database if it belongs to the specified Terminal device
func (r *SceneRepository) Delete(terminalID, id string) error {
	return r.db.Where("id = ? AND terminal_id = ?", id, terminalID).Delete(&entities.Scene{}).Error
}

// GetAllGrouped retrieves all scenes across all Terminal devices, grouped by TerminalID
func (r *SceneRepository) GetAllGrouped() (map[string][]entities.Scene, error) {
	var scenes []entities.Scene
	if err := r.db.Order("terminal_id").Find(&scenes).Error; err != nil {
		return nil, err
	}
	grouped := make(map[string][]entities.Scene)
	for _, s := range scenes {
		grouped[s.TerminalID] = append(grouped[s.TerminalID], s)
	}
	return grouped, nil
}
