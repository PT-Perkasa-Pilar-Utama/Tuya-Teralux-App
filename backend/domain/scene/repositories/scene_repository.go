package repositories

import (
	"teralux_app/domain/scene/entities"

	"gorm.io/gorm"
)

// SceneRepository handles persistent storage of scenes using GORM/SQLite
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

// GetByID retrieves a scene by its unique identifier and TeraluxID
func (r *SceneRepository) GetByID(teraluxID, id string) (*entities.Scene, error) {
	var scene entities.Scene
	if err := r.db.Where("id = ? AND teralux_id = ?", id, teraluxID).First(&scene).Error; err != nil {
		return nil, err
	}
	return &scene, nil
}

// GetAll retrieves all configured scenes for a specific Teralux device
func (r *SceneRepository) GetAll(teraluxID string) ([]entities.Scene, error) {
	var scenes []entities.Scene
	if err := r.db.Where("teralux_id = ?", teraluxID).Find(&scenes).Error; err != nil {
		return nil, err
	}
	return scenes, nil
}

// Delete removes a scene from the database if it belongs to the specified Teralux device
func (r *SceneRepository) Delete(teraluxID, id string) error {
	return r.db.Where("id = ? AND teralux_id = ?", id, teraluxID).Delete(&entities.Scene{}).Error
}

// GetAllGrouped retrieves all scenes across all Teralux devices, grouped by TeraluxID
func (r *SceneRepository) GetAllGrouped() (map[string][]entities.Scene, error) {
	var scenes []entities.Scene
	if err := r.db.Order("teralux_id").Find(&scenes).Error; err != nil {
		return nil, err
	}
	grouped := make(map[string][]entities.Scene)
	for _, s := range scenes {
		grouped[s.TeraluxID] = append(grouped[s.TeraluxID], s)
	}
	return grouped, nil
}
