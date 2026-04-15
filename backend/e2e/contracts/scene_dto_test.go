package contracts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensio/domain/scene/dtos"
)

func TestCreateSceneRequestDTOContract(t *testing.T) {
	dto := dtos.CreateSceneRequestDTO{
		Name: "Evening Mode",
		Actions: []dtos.ActionDTO{
			{
				DeviceID: "dev-123",
				Topic:    "users/AA:BB:CC:DD:EE:FF/prod/custom",
				Value:    "on",
			},
		},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Evening Mode", decoded["name"])
	require.NotNil(t, decoded["actions"])

	actions := decoded["actions"].([]interface{})
	require.Len(t, actions, 1)

	action := actions[0].(map[string]interface{})
	assert.Equal(t, "dev-123", action["device_id"])
	assert.Equal(t, "on", action["value"])
}

func TestSceneResponseDTOContract(t *testing.T) {
	dto := dtos.SceneResponseDTO{
		ID:         "scene-123",
		TerminalID: "term-456",
		Name:       "Morning Routine",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "scene-123", decoded["id"])
	assert.Equal(t, "term-456", decoded["terminal_id"])
	assert.Equal(t, "Morning Routine", decoded["name"])
}

func TestUpdateSceneRequestDTOContract(t *testing.T) {
	dto := dtos.UpdateSceneRequestDTO{
		Name: "Updated Scene Name",
		Actions: []dtos.ActionDTO{
			{
				Topic: "users/AA:BB:CC:DD:EE:FF/prod/light",
				Value: "off",
			},
		},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Updated Scene Name", decoded["name"])
	require.NotNil(t, decoded["actions"])
}

func TestActionDTOContract(t *testing.T) {
	tests := []struct {
		name   string
		action dtos.ActionDTO
	}{
		{
			name: "device_control action",
			action: dtos.ActionDTO{
				DeviceID: "dev-123",
				Value:    "on",
			},
		},
		{
			name: "mqtt action",
			action: dtos.ActionDTO{
				Topic: "users/mac/prod/custom",
				Value: "on",
			},
		},
		{
			name: "ir action",
			action: dtos.ActionDTO{
				Code:  "power",
				Value: "on",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.action)
			require.NoError(t, err)

			var decoded dtos.ActionDTO
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.action.DeviceID, decoded.DeviceID)
			assert.Equal(t, tt.action.Code, decoded.Code)
			assert.Equal(t, tt.action.Topic, decoded.Topic)
		})
	}
}
