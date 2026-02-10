package repositories

import (
	"encoding/json"
	"fmt"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/scene/entities"
)

const ScenePrefix = "scene:"

// SceneRepository handles persistent storage of scenes using BadgerService
type SceneRepository struct {
	cache *infrastructure.BadgerService
}

// NewSceneRepository creates a new instance of SceneRepository
func NewSceneRepository(cache *infrastructure.BadgerService) *SceneRepository {
	return &SceneRepository{cache: cache}
}

// Save persists a scene to the database
func (r *SceneRepository) Save(scene *entities.Scene) error {
	data, err := json.Marshal(scene)
	if err != nil {
		return fmt.Errorf("failed to marshal scene: %w", err)
	}
	return r.cache.SetPersistent(ScenePrefix+scene.ID, data)
}

// GetByID retrieves a scene by its unique identifier
func (r *SceneRepository) GetByID(id string) (*entities.Scene, error) {
	data, err := r.cache.Get(ScenePrefix + id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene from cache: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("scene not found")
	}
	var scene entities.Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scene: %w", err)
	}
	return &scene, nil
}

// GetAll retrieves all configured scenes
func (r *SceneRepository) GetAll() ([]entities.Scene, error) {
	keys, err := r.cache.GetAllKeysWithPrefix(ScenePrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene keys: %w", err)
	}
	var scenes []entities.Scene
	for _, key := range keys {
		data, err := r.cache.Get(key)
		if err != nil {
			continue
		}
		var scene entities.Scene
		if err := json.Unmarshal(data, &scene); err == nil {
			scenes = append(scenes, scene)
		}
	}
	return scenes, nil
}

// Delete removes a scene from the database
func (r *SceneRepository) Delete(id string) error {
	return r.cache.Delete(ScenePrefix + id)
}
