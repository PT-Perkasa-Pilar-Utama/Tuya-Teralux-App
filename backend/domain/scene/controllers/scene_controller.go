package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	scene_dtos "teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/usecases"
	tuya_dtos "teralux_app/domain/tuya/dtos"

	"github.com/gin-gonic/gin"
)

// TuyaDeviceControlExecutor defines the interface for controlling Tuya devices
type TuyaDeviceControlExecutor interface {
	SendCommand(accessToken, deviceID string, commands []tuya_dtos.TuyaCommandDTO) (bool, error)
	SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error)
}

type SceneController struct {
	addUseCase     *usecases.AddSceneUseCase
	updateUseCase  *usecases.UpdateSceneUseCase
	deleteUseCase  *usecases.DeleteSceneUseCase
	getAllUseCase  *usecases.GetAllScenesUseCase
	controlUseCase *usecases.ControlSceneUseCase
}

func NewSceneController(
	addUC *usecases.AddSceneUseCase,
	updateUC *usecases.UpdateSceneUseCase,
	deleteUC *usecases.DeleteSceneUseCase,
	getAllUC *usecases.GetAllScenesUseCase,
	controlUC *usecases.ControlSceneUseCase,
) *SceneController {
	return &SceneController{
		addUseCase:     addUC,
		updateUseCase:  updateUC,
		deleteUseCase:  deleteUC,
		getAllUseCase:  getAllUC,
		controlUseCase: controlUC,
	}
}

// AddScene handles POST /api/teralux/:teralux_id/scenes
// @Summary Create a new scene
// @Description Create a scene with a name and a list of actions (Tuya/MQTT)
// @Tags 03. Scenes
// @Accept json
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene body scene_dtos.CreateSceneRequestDTO true "Scene configuration"
// @Success 201 {object} dtos.StandardResponse "Scene created"
// @Failure 400 {object} dtos.StandardResponse "Invalid request"
// @Failure 500 {object} dtos.StandardResponse "Internal error"
// @Router /api/teralux/{id}/scenes [post]
func (c *SceneController) AddScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	var req scene_dtos.CreateSceneRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	actions := make(entities.Actions, len(req.Actions))
	for i, a := range req.Actions {
		actions[i] = entities.Action(a)
	}

	sceneID, err := c.addUseCase.Execute(teraluxID, req.Name, actions)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Scene created successfully",
		Data:    gin.H{"scene_id": sceneID},
	})
}

// UpdateScene handles PUT /api/teralux/:id/scenes/:scene_id
// @Summary Update an existing scene
// @Description Update the configuration of a specific scene
// @Tags 03. Scenes
// @Accept json
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene_id path string true "Scene UUID"
// @Param scene body scene_dtos.UpdateSceneRequestDTO true "Updated scene configuration"
// @Success 200 {object} dtos.StandardResponse "Scene updated"
// @Failure 404 {object} dtos.StandardResponse "Scene not found"
// @Router /api/teralux/{id}/scenes/{scene_id} [put]
func (c *SceneController) UpdateScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	id := ctx.Param("scene_id")
	var req scene_dtos.UpdateSceneRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	var actions entities.Actions
	if len(req.Actions) > 0 {
		actions = make(entities.Actions, len(req.Actions))
		for i, a := range req.Actions {
			actions[i] = entities.Action(a)
		}
	}

	if err := c.updateUseCase.Execute(teraluxID, id, req.Name, actions); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" { // Assuming usecase returns this specific error string for not found
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene updated successfully",
		Data:    gin.H{"scene_id": id},
	})
}

// DeleteScene handles DELETE /api/teralux/:id/scenes/:scene_id
// @Summary Delete a scene
// @Description Remove a specific scene configuration
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene_id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene deleted"
// @Router /api/teralux/{id}/scenes/{scene_id} [delete]
func (c *SceneController) DeleteScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	id := ctx.Param("scene_id")
	if err := c.deleteUseCase.Execute(teraluxID, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene deleted successfully",
	})
}

// GetAllScenes handles GET /api/teralux/:id/scenes
// @Summary List all scenes for a Teralux device
// @Description Retrieve a list of all configured scenes for a specific Teralux device
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Teralux UUID"
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.SceneListResponseDTO} "List of scenes"
// @Router /api/teralux/{id}/scenes [get]
func (c *SceneController) GetAllScenes(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	scenes, err := c.getAllUseCase.Execute(teraluxID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	response := make([]scene_dtos.SceneListResponseDTO, len(scenes))
	for i, s := range scenes {
		response[i] = scene_dtos.SceneListResponseDTO{
			ID:        s.ID,
			TeraluxID: s.TeraluxID,
			Name:      s.Name,
		}
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scenes retrieved successfully",
		Data:    response,
	})
}

// ControlScene handles GET /api/teralux/:id/scenes/:scene_id/control
// @Summary Apply/Trigger a scene
// @Description Trigger all actions defined in a specific scene
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene_id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene applied"
// @Router /api/teralux/{id}/scenes/{scene_id}/control [get]
func (c *SceneController) ControlScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	id := ctx.Param("scene_id")
	accessToken := ctx.GetString("access_token")

	if err := c.controlUseCase.Execute(teraluxID, id, accessToken); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" { // Assuming usecase returns this specific error string for not found
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene applied successfully",
	})
}
