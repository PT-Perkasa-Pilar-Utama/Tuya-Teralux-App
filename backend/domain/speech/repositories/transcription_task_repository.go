package repositories

import (
	"encoding/json"
	"fmt"
	"sync"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	speechdtos "teralux_app/domain/speech/dtos"
)

type TranscriptionTaskRepository interface {
	SaveShortTask(taskID string, status *speechdtos.AsyncTranscriptionStatusDTO) error
	GetShortTask(taskID string) (*speechdtos.AsyncTranscriptionStatusDTO, error)
	SaveLongTask(taskID string, status *speechdtos.AsyncTranscriptionLongStatusDTO) error
	GetLongTask(taskID string) (*speechdtos.AsyncTranscriptionLongStatusDTO, error)
}

type transcriptionTaskRepository struct {
	badger              *infrastructure.BadgerService
	tasksMutex          sync.RWMutex
	transcribeTasks     map[string]*speechdtos.AsyncTranscriptionStatusDTO
	transcribeLongTasks map[string]*speechdtos.AsyncTranscriptionLongStatusDTO
}

func NewTranscriptionTaskRepository(badger *infrastructure.BadgerService) TranscriptionTaskRepository {
	return &transcriptionTaskRepository{
		badger:              badger,
		transcribeTasks:     make(map[string]*speechdtos.AsyncTranscriptionStatusDTO),
		transcribeLongTasks: make(map[string]*speechdtos.AsyncTranscriptionLongStatusDTO),
	}
}

func (r *transcriptionTaskRepository) SaveShortTask(taskID string, status *speechdtos.AsyncTranscriptionStatusDTO) error {
	r.tasksMutex.Lock()
	r.transcribeTasks[taskID] = status
	r.tasksMutex.Unlock()

	if r.badger != nil {
		taskData, _ := json.Marshal(status)
		key := "transcribe:task:" + taskID
		// Use SetPreserveTTL if it's an update, or Set for new? 
		// For simplicity and consistency with previous logic, we'll try to preserve if exists or just set.
		// The previous logic used specific calls. Let's stick to SetPreserveTTL for updates if possible, 
		// but standard Set is safer for new items. 
		// Since this is a repository, we might want separate Create/Update, but simpler is:
		if err := r.badger.Set(key, taskData); err != nil {
			utils.LogWarn("TranscriptionTaskRepository: failed to save short task %s: %v", taskID, err)
			return err
		}
	}
	return nil
}

func (r *transcriptionTaskRepository) GetShortTask(taskID string) (*speechdtos.AsyncTranscriptionStatusDTO, error) {
	r.tasksMutex.RLock()
	status, exists := r.transcribeTasks[taskID]
	r.tasksMutex.RUnlock()

	if exists && status != nil {
		return status, nil
	}

	if r.badger != nil {
		key := "transcribe:task:" + taskID
		b, ttl, err := r.badger.GetWithTTL(key)
		if err == nil && len(b) > 0 {
			var st speechdtos.AsyncTranscriptionStatusDTO
			if err := json.Unmarshal(b, &st); err == nil {
				st.ExpiresInSecond = int64(ttl.Seconds())
				return &st, nil
			}
		}
	}

	return nil, fmt.Errorf("task not found")
}

func (r *transcriptionTaskRepository) SaveLongTask(taskID string, status *speechdtos.AsyncTranscriptionLongStatusDTO) error {
	r.tasksMutex.Lock()
	r.transcribeLongTasks[taskID] = status
	r.tasksMutex.Unlock()

	if r.badger != nil {
		taskData, _ := json.Marshal(status)
		key := "transcribe_long:task:" + taskID
		if err := r.badger.Set(key, taskData); err != nil {
			utils.LogWarn("TranscriptionTaskRepository: failed to save long task %s: %v", taskID, err)
			return err
		}
	}
	return nil
}

func (r *transcriptionTaskRepository) GetLongTask(taskID string) (*speechdtos.AsyncTranscriptionLongStatusDTO, error) {
	r.tasksMutex.RLock()
	status, exists := r.transcribeLongTasks[taskID]
	r.tasksMutex.RUnlock()

	if exists && status != nil {
		return status, nil
	}

	if r.badger != nil {
		key := "transcribe_long:task:" + taskID
		b, ttl, err := r.badger.GetWithTTL(key)
		if err == nil && len(b) > 0 {
			var st speechdtos.AsyncTranscriptionLongStatusDTO
			if err := json.Unmarshal(b, &st); err == nil {
				st.ExpiresInSecond = int64(ttl.Seconds())
				return &st, nil
			}
		}
	}

	return nil, fmt.Errorf("task not found")
}
