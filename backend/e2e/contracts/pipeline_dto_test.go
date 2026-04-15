package contracts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensio/domain/models/pipeline/dtos"
)

func TestPipelineRequestDTOContract(t *testing.T) {
	refine := true
	tests := []struct {
		name    string
		request dtos.PipelineRequestDTO
	}{
		{
			name: "valid request with all fields",
			request: dtos.PipelineRequestDTO{
				Language:       "id",
				TargetLanguage: "en",
				Context:        "meeting",
				Style:          "minutes",
				Date:           "2026-04-15",
				Location:       "Jakarta",
				Participants:   []string{"Alice", "Bob"},
				Diarize:        true,
				Refine:         &refine,
				Summarize:      true,
				MacAddress:     "AA:BB:CC:DD:EE:FF",
			},
		},
		{
			name: "minimal request",
			request: dtos.PipelineRequestDTO{
				Language: "id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			require.NoError(t, err)

			var decoded dtos.PipelineRequestDTO
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.request.Language, decoded.Language)
			assert.Equal(t, tt.request.TargetLanguage, decoded.TargetLanguage)
			assert.Equal(t, tt.request.Context, decoded.Context)
			assert.Equal(t, tt.request.Style, decoded.Style)
		})
	}
}

func TestPipelineStatusDTOContract(t *testing.T) {
	dto := dtos.PipelineStatusDTO{
		TaskID:        "task-123",
		OverallStatus: "completed",
		Stages: map[string]dtos.PipelineStageStatus{
			"transcribe": {
				Status:          "completed",
				StartedAt:       "2026-04-15T10:00:00Z",
				DurationSeconds: 5.2,
			},
			"summary": {
				Status:          "completed",
				StartedAt:       "2026-04-15T10:00:05Z",
				DurationSeconds: 3.1,
			},
		},
		StartedAt:       "2026-04-15T10:00:00Z",
		DurationSeconds: 8.3,
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "task-123", decoded["task_id"])
	assert.Equal(t, "completed", decoded["overall_status"])
	assert.NotNil(t, decoded["stages"])

	stages := decoded["stages"].(map[string]interface{})
	assert.Contains(t, stages, "transcribe")
	assert.Contains(t, stages, "summary")
}

func TestPipelineResponseDTOContract(t *testing.T) {
	dto := dtos.PipelineResponseDTO{
		TaskID: "task-123",
		TaskStatus: &dtos.PipelineStatusDTO{
			TaskID:        "task-123",
			OverallStatus: "completed",
			Stages:        map[string]dtos.PipelineStageStatus{},
		},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "task-123", decoded["task_id"])
	assert.NotNil(t, decoded["task_status"])
}
